"""Matching analysis endpoint with SHA-256 caching."""
import hashlib
import logging
from asyncio import Lock
from fastapi import APIRouter, HTTPException
from models import MatchRequest
from services.resume_parser import ResumeParser
from services.jd_parser import JDParser
from services.matcher import Matcher
from services.optimizer import Optimizer
from services.gap_analyzer import GapAnalyzer

router = APIRouter(prefix="/api", tags=["match"])
logger = logging.getLogger(__name__)


_resume_parser: ResumeParser | None = None
_jd_parser: JDParser | None = None
_matcher: Matcher | None = None
_optimizer: Optimizer | None = None
_gap_analyzer: GapAnalyzer | None = None

_cache: dict[str, dict] = {}
_cache_lock = Lock()


def _get_services():
    global _resume_parser, _jd_parser, _matcher, _optimizer, _gap_analyzer
    if _resume_parser is None:
        _resume_parser = ResumeParser()
        _jd_parser = JDParser()
        _matcher = Matcher()
        _optimizer = Optimizer()
        _gap_analyzer = GapAnalyzer()
    return _resume_parser, _jd_parser, _matcher, _optimizer, _gap_analyzer


def _cache_key(resume_text: str, job_text: str) -> str:
    return hashlib.sha256(f"{resume_text}||{job_text}".encode()).hexdigest()


@router.post("/match")
async def analyze_match(req: MatchRequest):
    logger.info("Match request: resume_text len=%d, job_text len=%d", len(req.resume_text), len(req.job_text))

    key = _cache_key(req.resume_text, req.job_text)
    async with _cache_lock:
        if key in _cache:
            return {"report": _cache[key], "cached": True}

    resume_parser, jd_parser, matcher, optimizer, gap_analyzer = _get_services()

    try:
        resume = await resume_parser.parse(req.resume_text)
        job = await jd_parser.parse(req.job_text)

        report = await matcher.analyze(resume, job)

        tips = await optimizer.generate_tips(resume, job, report)
        report.optimization_tips = tips

        gaps = await gap_analyzer.analyze(resume, job)
        report.gap_analysis = gaps

        report_dict = report.model_dump()

        async with _cache_lock:
            _cache[key] = report_dict

        return {"report": report_dict}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
