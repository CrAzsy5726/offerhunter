"""Resume optimization service using LangChain + DeepSeek."""
import json
from langchain_core.messages import SystemMessage, HumanMessage
from models import Resume, Job, MatchReport, OptimizationTip
from prompts.optimization import OPTIMIZATION_PROMPT, OPTIMIZATION_INSTRUCTION
from services.extract_json import extract_json
from services.llm_factory import create_deepseek_llm


class Optimizer:
    def __init__(self):
        self._llm = create_deepseek_llm(temperature=0.3)

    async def generate_tips(self, resume: Resume, job: Job, report: MatchReport) -> list[OptimizationTip]:
        prompt = (
            OPTIMIZATION_PROMPT
            .replace("{resume_json}", resume.model_dump_json(exclude_none=True))
            .replace("{job_json}", job.model_dump_json(exclude_none=True))
            .replace("{match_json}", report.model_dump_json(exclude_none=True))
        )

        messages = [
            SystemMessage(content=prompt),
            HumanMessage(content=OPTIMIZATION_INSTRUCTION),
        ]
        response = await self._llm.ainvoke(messages)
        content = response.content if hasattr(response, 'content') else str(response)

        try:
            data = json.loads(content)
        except json.JSONDecodeError:
            json_str = extract_json(content)
            if not json_str:
                raise ValueError(f"failed to parse LLM response as JSON")
            data = json.loads(json_str)

        return [OptimizationTip(**tip) for tip in data.get("optimization_tips", [])]
