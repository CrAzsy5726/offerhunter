"""LangChain @tool wrappers around Offer-Hunter services.

Each tool returns a natural-language summary string for the agent to read.
Parsed objects are stored in _thread_cache keyed by thread_id so the
composite analyze_match tool can retrieve them without the LLM re-passing
large text blobs between tool calls.
"""
import logging
from typing import Optional
from langchain.tools import tool
from langchain_core.runnables import RunnableConfig

from models import Resume, Job, MatchReport
from services.resume_parser import ResumeParser
from services.jd_parser import JDParser
from services.matcher import Matcher
from services.optimizer import Optimizer
from services.gap_analyzer import GapAnalyzer
from services.market_research import MarketResearchService

logger = logging.getLogger(__name__)

# Per-conversation cache: thread_id -> { "resume_text", "resume", "job_text", "job", "report" }
_thread_cache: dict[str, dict] = {}

# Lazy-initialized service singletons
_resume_parser: Optional[ResumeParser] = None
_jd_parser: Optional[JDParser] = None
_matcher: Optional[Matcher] = None
_optimizer: Optional[Optimizer] = None
_gap_analyzer: Optional[GapAnalyzer] = None
_market_service: Optional[MarketResearchService] = None


def _get_services():
    global _resume_parser, _jd_parser, _matcher, _optimizer, _gap_analyzer, _market_service
    if _resume_parser is None:
        _resume_parser = ResumeParser()
        _jd_parser = JDParser()
        _matcher = Matcher()
        _optimizer = Optimizer()
        _gap_analyzer = GapAnalyzer()
        _market_service = MarketResearchService()
    return _resume_parser, _jd_parser, _matcher, _optimizer, _gap_analyzer, _market_service


def _get_thread_id(config: Optional[RunnableConfig]) -> str:
    """Extract thread_id from RunnableConfig, or return 'default'."""
    if config and "configurable" in config:
        tid = config["configurable"].get("thread_id", "")
        if tid:
            return tid
    return "default"


@tool
async def parse_resume(
    text: str,
    config: RunnableConfig = None,
) -> str:
    """解析用户的简历文本，提取结构化信息（教育背景、技能、项目经历、实习经历等）。
    调用时机：用户提供了一份简历（文本内容较长，包含教育/技能/项目/实习等信息）。
    text: 完整的简历原文，必须传入原始文本，不要做任何删减或摘要。"""

    rp, _, _, _, _, _ = _get_services()
    resume = await rp.parse(text)
    tid = _get_thread_id(config)
    _thread_cache.setdefault(tid, {})["resume_text"] = text
    _thread_cache[tid]["resume"] = resume

    skills = resume.metadata.top_skills or resume.skills[:5]
    projects = resume.metadata.total_projects
    exps = resume.metadata.total_experience
    level = resume.metadata.experience_level or "未知"
    level_cn = {"entry": "应届", "junior": "1-3年", "mid": "3-5年", "senior": "5年+"}.get(level, level)

    return (
        f"简历解析成功。经验水平: {level_cn}，"
        f"核心技能: {', '.join(skills[:6])}，"
        f"项目经历: {projects}个，实习/工作经历: {exps}个。"
    )


@tool
async def parse_jd(
    text: str,
    config: RunnableConfig = None,
) -> str:
    """解析岗位描述(JD)文本，提取岗位名称、公司、硬性要求、软性要求、隐性偏好等信息。
    调用时机：用户提供了一份招聘JD（包含岗位职责、任职要求、加分项等）。
    text: 完整的JD原文，必须传入原始文本，不要做任何删减或摘要。"""

    _, jp, _, _, _, _ = _get_services()
    job = await jp.parse(text)
    tid = _get_thread_id(config)
    _thread_cache.setdefault(tid, {})["job_text"] = text
    _thread_cache[tid]["job"] = job

    hard_reqs = [r for r in job.requirements if r.is_hard]
    title = job.title or "未知岗位"
    company = f"@{job.company}" if job.company else ""
    return (
        f"JD解析成功。岗位: {title} {company}，"
        f"硬性要求: {len(hard_reqs)}项（{', '.join(r.value[:20] for r in hard_reqs[:4])}），"
        f"软性要求: {len(job.requirements) - len(hard_reqs)}项，"
        f"加分项: {len(job.bonus_items)}项。"
    )


@tool
async def analyze_match(
    resume_text: str = "",
    job_text: str = "",
    config: RunnableConfig = None,
) -> str:
    """执行完整的简历-岗位匹配分析。包括三层分析：硬性门槛判定、竞争力评分（项目深度/实习含金量/技术栈匹配）、隐性契合分析，
    以及简历优化建议和能力差距分析。
    调用时机：当简历和JD都已就绪，用户表达了匹配意愿，或双方信息已经完整时调用。
    resume_text: 简历原文（如果之前已通过parse_resume解析过，可留空让工具自动从上下文获取）。
    job_text: JD原文（如果之前已通过parse_jd解析过，可留空让工具自动从上下文获取）。"""

    tid = _get_thread_id(config)
    cache = _thread_cache.get(tid, {})

    # Resolve resume text
    if not resume_text:
        resume_text = cache.get("resume_text", "")
    if not resume_text:
        return "错误：没有找到简历信息。请先要求用户提供简历。"

    # Resolve job text
    if not job_text:
        job_text = cache.get("job_text", "")
    if not job_text:
        return "错误：没有找到岗位信息。请先要求用户提供JD。"

    rp, jp, mat, opt, gap, _ = _get_services()

    # Full pipeline
    resume = await rp.parse(resume_text)
    job = await jp.parse(job_text)
    report = await mat.analyze(resume, job)
    tips = await opt.generate_tips(resume, job, report)
    gaps = await gap.analyze(resume, job)
    report.optimization_tips = tips
    report.gap_analysis = gaps

    # Cache for later retrieval
    cache["report"] = report
    cache["resume_text"] = resume_text
    cache["job_text"] = job_text

    # Score labels
    def grade(s: int) -> str:
        if s >= 80:
            return "优秀"
        if s >= 60:
            return "良好"
        if s >= 40:
            return "一般"
        return "需改进"

    c = report.competitiveness
    ev_tips = [t for t in tips if t.has_evidence]
    learn_tips = [t for t in tips if not t.has_evidence]

    return (
        f"匹配分析完成！\n"
        f"整体匹配度: {report.overall_score}% ({grade(report.overall_score)})。\n"
        f"硬性门槛: {sum(1 for it in report.threshold_check.items if it.status == 'pass')}/{len(report.threshold_check.items)} 通过。\n"
        f"竞争力: 项目深度 {c.project_depth.score} / 实习含金量 {c.intern_quality.score} / 技术栈匹配 {c.tech_stack_fit.score} / 整体 {c.overall_power}。\n"
        f"隐性契合度: {report.hidden_fit.score}%。\n"
        f"优化建议: {len(ev_tips)}条有据可依 + {len(learn_tips)}条技能提升。\n"
        f"能力差距: {len(gaps)}项需弥补。\n"
        f"参考线: {report.benchmark.pass_line}分，排名: {report.benchmark.percentile or '未评估'}。"
    )


@tool
async def research_market(
    job_info: str,
    job_preference: str = "",
) -> str:
    """查询某个岗位的市场信息，包括薪资范围、热门技能、招聘趋势、竞争度。
    调用时机：用户询问某个岗位/行业的薪资水平、市场需求、技能要求、就业前景时。
    job_info: 岗位信息（如"Java后端开发"或"北京前端工程师"），尽量包含城市和岗位名称。
    job_preference: 用户的求职偏好（如"应届生"、"3年经验"），可选。"""

    _, _, _, _, _, ms = _get_services()
    info = await ms.research(job_info, job_preference)

    skills = ", ".join(info.hot_skills[:5]) if info.hot_skills else "未获取"
    return (
        f"市场分析结果：\n"
        f"薪资范围: {info.salary_range or '未获取'}。\n"
        f"热门技能: {skills}。\n"
        f"招聘趋势: {info.trend or '未获取'}。\n"
        f"竞争度: {info.competition or '未获取'}。\n"
        f"综合: {info.summary or '未获取'}"
    )


# Export list used by the agent factory
ALL_TOOLS = [parse_resume, parse_jd, analyze_match, research_market]
