"""Three-layer matching engine using LangChain + DeepSeek."""
import json
from langchain_core.messages import SystemMessage, HumanMessage
from models import Resume, Job, MatchReport
from prompts.matching import MATCHING_PROMPT, MATCHING_INSTRUCTION
from services.extract_json import extract_json
from services.llm_factory import create_deepseek_llm


class Matcher:
    def __init__(self):
        self._llm = create_deepseek_llm(temperature=0)

    async def analyze(self, resume: Resume, job: Job) -> MatchReport:
        resume_json = resume.model_dump_json(exclude_none=True)
        job_json = job.model_dump_json(exclude_none=True)

        prompt = MATCHING_PROMPT.replace("{resume_json}", resume_json).replace("{job_json}", job_json)

        messages = [
            SystemMessage(content=prompt),
            HumanMessage(content=MATCHING_INSTRUCTION),
        ]
        response = await self._llm.ainvoke(messages)
        content = response.content if hasattr(response, 'content') else str(response)

        try:
            report = MatchReport.model_validate_json(content)
        except Exception:
            json_str = extract_json(content)
            if not json_str:
                raise ValueError(f"failed to parse LLM response as JSON")
            report = MatchReport.model_validate_json(json_str)

        # Deterministic weighted scoring (same formula as Go version)
        c = report.competitiveness
        c.overall_power = (c.project_depth.score + c.intern_quality.score + c.tech_stack_fit.score) // 3
        w = (
            c.project_depth.score * 25
            + c.intern_quality.score * 20
            + c.tech_stack_fit.score * 35
            + report.hidden_fit.score * 20
        )
        report.overall_score = w // 100

        return report
