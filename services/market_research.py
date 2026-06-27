"""Market research service using LangChain + DuckDuckGo search."""
import re
import json
import httpx
from langchain_core.messages import SystemMessage, HumanMessage
from models import MarketInfo
from prompts.market_research import MARKET_RESEARCH_PROMPT, MARKET_RESEARCH_INSTRUCTION
from services.extract_json import extract_json
from services.llm_factory import create_deepseek_llm


class MarketResearchService:
    def __init__(self):
        self._llm = create_deepseek_llm(temperature=0.3)

    async def research(self, job_info: str, job_preference: str = "") -> MarketInfo:
        # Step 1: Search DuckDuckGo
        search_query = f"{job_info} {job_preference} 薪资 技能要求 招聘 应届生"
        search_results = await self._search_ddg(search_query)

        # Step 2: Build prompt with search results
        prompt = (
            MARKET_RESEARCH_PROMPT
            .replace("{job_info}", job_info)
            .replace("{job_preference}", job_preference or "不限")
            .replace("{search_results}", search_results or "（无实时搜索结果）")
        )

        messages = [
            SystemMessage(content=prompt),
            HumanMessage(content=MARKET_RESEARCH_INSTRUCTION),
        ]
        response = await self._llm.ainvoke(messages)
        content = response.content if hasattr(response, 'content') else str(response)

        try:
            return MarketInfo.model_validate_json(content)
        except Exception:
            json_str = extract_json(content)
            if not json_str:
                raise ValueError(f"failed to parse market research response")
            return MarketInfo.model_validate_json(json_str)

    async def _search_ddg(self, query: str) -> str:
        url = "https://html.duckduckgo.com/html/"
        async with httpx.AsyncClient(timeout=10.0) as client:
            try:
                resp = await client.get(
                    url,
                    params={"q": query},
                    headers={"User-Agent": "OfferHunter/1.0 (job-search-tool)"},
                )
                return self._extract_snippets(resp.text)
            except Exception:
                return ""

    def _extract_snippets(self, html: str) -> str:
        snippets = re.findall(r'class="result__snippet"[^>]*>(.*?)</a>', html)
        if not snippets:
            snippets = re.findall(r'class="result__snippet">(.+?)</a>', html)

        result_lines = []
        count = 0
        for s in snippets:
            text = re.sub(r"<[^>]*>", "", s).strip()
            if len(text) > 20:
                count += 1
                result_lines.append(f"{count}. {text}")
                if count >= 5:
                    break
        return "\n".join(result_lines)
