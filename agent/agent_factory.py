"""Agent factory -- one agent, one InMemorySaver, tool filtering via middleware.

chat_handler classifies intent (Round 1), stores it via set_intent(),
then the middleware reads it to narrow tools per LLM call (Round 2).
Single agent + single InMemorySaver = full cross-intent conversation memory.
"""
import contextvars
import logging
from langchain.agents import create_agent
from langchain.agents.middleware import wrap_model_call
from langchain_openai import ChatOpenAI
from langchain_core.messages import HumanMessage
from langgraph.checkpoint.memory import InMemorySaver

import config
from services.agent_tools import parse_resume, parse_jd, analyze_match, research_market
from services.agent_tools_advanced import (
    generate_interview_questions,
    search_job_listings,
    generate_job_search_plan,
    analyze_skill_transferability,
    recommend_transition_roles,
    get_salary_benchmark,
    generate_negotiation_strategy,
)
from agent.middleware import LoggingTimingMiddleware

logger = logging.getLogger(__name__)

ALL_TOOLS_FLAT = [
    parse_resume, parse_jd, analyze_match, research_market,
    generate_interview_questions,
    search_job_listings, generate_job_search_plan,
    analyze_skill_transferability, recommend_transition_roles,
    get_salary_benchmark, generate_negotiation_strategy,
]

# ---- Intent classification ----

INTENT_CLASSIFY_PROMPT = """判断用户意图，只回复一个词（不要句号，不要任何其他文字）：
- match: 用户明确要求分析简历与JD的匹配度，说了"帮我分析"、"匹配一下"、"看看合不合适"等。
  - other: 用户只是发了简历或JD文本，但没有提出任何要求。属于other而不是match。
- interview: 模拟面试、出题考核、面试练习
- strategy: 找工作、求职计划、搜索岗位、职业规划
- pivot: 转行、转型、换方向、跨领域
- salary: 薪资查询、薪资谈判、谈offer、报价
- other: 打招呼、闲聊、只发内容没提要求（比如只粘贴了简历但没说想干嘛）。

用户: {message}
意图:"""

INTENT_LABELS = {
    "match": "简历分析", "interview": "模拟面试",
    "strategy": "求职策略", "pivot": "职业转型",
    "salary": "薪资谈判", "other": "通用助手",
}

# Per-intent tool subsets (max 4-5 for DeepSeek stability)
INTENT_TOOL_NAMES = {
    "match":     {"parse_resume", "parse_jd", "analyze_match", "research_market"},
    "interview": {"generate_interview_questions", "research_market"},
    "strategy":  {"search_job_listings", "generate_job_search_plan", "research_market"},
    "pivot":     {"analyze_skill_transferability", "recommend_transition_roles", "research_market"},
    "salary":    {"get_salary_benchmark", "generate_negotiation_strategy", "research_market"},
    "other":     {"research_market"},
}

# ---- Per-request intent (contextvars: async-safe) ----

_current_intent: contextvars.ContextVar[str] = contextvars.ContextVar("intent", default="other")

def set_intent(intent: str):
    _current_intent.set(intent)

# ---- Tool-filter middleware ----

@wrap_model_call
async def intent_tool_filter(request, handler):
    intent = _current_intent.get()
    allowed = INTENT_TOOL_NAMES.get(intent, INTENT_TOOL_NAMES["other"])
    filtered = [t for t in request.tools if t.name in allowed]
    logger.debug("Tool filter [%s]: %d -> %d tools", intent, len(request.tools), len(filtered))
    return await handler(request.override(tools=filtered))

# ---- Classifier LLM ----

_classifier_llm: ChatOpenAI | None = None

def _get_classifier():
    global _classifier_llm
    if _classifier_llm is None:
        _classifier_llm = ChatOpenAI(
            model=config.DEEPSEEK_MODEL,
            api_key=config.DEEPSEEK_API_KEY,
            base_url=config.DEEPSEEK_BASE_URL,
            temperature=0, max_tokens=10,
        )
    return _classifier_llm

async def classify_intent(message: str) -> tuple[str, str]:
    llm = _get_classifier()
    # Take first 200 and last 300 chars to capture both the content and the user's request
    short = (message[:200] + " ... " + message[-300:]) if len(message) > 500 else message
    prompt = INTENT_CLASSIFY_PROMPT.format(message=short)
    try:
        resp = await llm.ainvoke([HumanMessage(content=prompt)])
        raw = resp.content.strip().lower() if hasattr(resp, "content") else str(resp).strip().lower()
    except Exception:
        logger.exception("Intent classification failed")
        return "other", INTENT_LABELS["other"]
    for key in INTENT_LABELS:
        if key in raw:
            logger.info("Intent classified: %s (%s) for msg[:200]=%r", key, INTENT_LABELS[key], short[:80])
            return key, INTENT_LABELS[key]
    logger.warning("Intent fallback to 'other', raw response: %r", raw)
    return "other", INTENT_LABELS["other"]

# ---- Single agent ----

SYSTEM_PROMPT = """你是"Offer 捕手"，一个全能 AI 求职助手。你可以自然聊天，也能使用工具完成专业的求职服务。

## 必须遵守
1. 需要解析/分析/搜索/生成时，必须调用对应工具。绝不在回复中直接输出 JSON。
2. 从对话历史中获取已有信息。用户之前发的简历、JD、调过的工具结果都在历史中，不要反复索要。

## 能力范围（只在有对应工具时使用）
- parse_resume / parse_jd / analyze_match —— 简历解析、JD解析、匹配分析
- generate_interview_questions —— 生成面试题（调用后逐题提问评分）
- search_job_listings / generate_job_search_plan —— 搜索岗位、制定求职计划
- analyze_skill_transferability / recommend_transition_roles —— 技能迁移、过渡岗位
- get_salary_benchmark / generate_negotiation_strategy —— 薪资基准、谈判策略
- research_market —— 市场调研

## 工具边界
当前能用的工具是动态过滤的。如果你发现你想做的操作没有对应工具可用，不要硬着头皮用文字模拟（比如口述一份"分析报告"）。正确做法是：简要列出你还能帮用户做什么，然后问"你想先试试哪个？"。

## 对话流程
- 用户发了简历/文件但没说"分析"/"面试"等 → 先解析，然后问"简历已收到。你想让我帮你做什么？比如匹配分析、模拟面试、求职策略……"不要擅自决定。
- 用户发了简历+JD但没说要求 → 同上，解析后询问，不要自动分析。
- 简历+JD都已就绪 + 用户说"分析/匹配" → 立即 analyze_match
- 说"模拟面试" → generate_interview_questions → 逐题提问评分
- 说"找工作/求职" → search_job_listings + 生成计划
- 说"转行" → analyze_skill_transferability + 过渡岗位
- 问薪资 → get_salary_benchmark + 谈判策略

## 闲聊与引导
- 自由聊天。在回复末尾自然提及："对了，你是想了解求职方面的信息吗？我可以帮你分析简历、模拟面试、规划职业转型……"
- 如果用户回应求职话题，立即转入对应功能。
- 如果用户问"你能干什么"，列出所有功能，用一句"你想先试试哪个？"收尾。

## 记忆
对话历史中包含所有之前的消息和工具结果。每轮对话你可以看到完整的聊天记录。
"""


_agent = None

def get_agent():
    global _agent
    if _agent is None:
        llm = ChatOpenAI(
            model=config.DEEPSEEK_MODEL,
            api_key=config.DEEPSEEK_API_KEY,
            base_url=config.DEEPSEEK_BASE_URL,
            temperature=0, max_tokens=4096,
            model_kwargs={"tool_choice": "auto"},
        )
        _agent = create_agent(
            model=llm,
            tools=ALL_TOOLS_FLAT,
            system_prompt=SYSTEM_PROMPT,
            middleware=[intent_tool_filter, LoggingTimingMiddleware()],
            checkpointer=InMemorySaver(),
        )
        logger.info("Unified Agent created -- %d tools (filtered per intent)", len(ALL_TOOLS_FLAT))
    return _agent
