"""JD parsing service using LangChain + DeepSeek."""
import json
from langchain_core.messages import SystemMessage, HumanMessage
from models import Job
from prompts.jd_extraction import JD_EXTRACTION_PROMPT, JD_EXTRACTION_INSTRUCTION
from services.extract_json import extract_json
from services.llm_factory import create_deepseek_llm


class JDParser:
    def __init__(self):
        self._llm = create_deepseek_llm(temperature=0)

    async def parse(self, text: str) -> Job:
        text = text.strip()
        if not text:
            raise ValueError("empty JD text")

        messages = [
            SystemMessage(content=JD_EXTRACTION_PROMPT),
            HumanMessage(content=text + "\n\n" + JD_EXTRACTION_INSTRUCTION),
        ]
        response = await self._llm.ainvoke(messages)
        content = response.content if hasattr(response, 'content') else str(response)

        try:
            job = Job.model_validate_json(content)
        except Exception:
            json_str = extract_json(content)
            if not json_str:
                raise ValueError(f"failed to parse LLM response as JSON")
            job = Job.model_validate_json(json_str)

        job.description = text
        return job
