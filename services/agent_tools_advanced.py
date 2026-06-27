"""Advanced Agent Tools for 4 new modes: Interview, Strategist, Pivot, Negotiation."""
import json
import logging
import re
import httpx
from langchain.tools import tool
from langchain_core.messages import SystemMessage, HumanMessage
from services.llm_factory import create_deepseek_llm

logger = logging.getLogger(__name__)

_tool_llm = None


def _get_tool_llm():
    global _tool_llm
    if _tool_llm is None:
        _tool_llm = create_deepseek_llm(temperature=0.3)
    return _tool_llm


# ═══════════════════════════════════════════
#  Mode 1: 模拟面试官 (Mock Interviewer)
# ═══════════════════════════════════════════

@tool
async def generate_interview_questions(
    jd_text: str = "",
    resume_text: str = "",
    count: int = 8,
) -> str:
    """根据JD和简历生成针对性的面试题目。
    调用时机：用户想要模拟面试时，需要先生成一组面试题。
    jd_text: 岗位描述原文（如果不确定，就用用户最近提供的JD相关信息，哪怕是简短的也行）
    resume_text: 简历原文（如果不确定，就用用户最近提供的背景描述，哪怕是简短描述也行）
    count: 题目数量，默认8道"""
    if not jd_text and not resume_text:
        return "需要至少提供JD或简历中的一项来生成面试题。请询问用户提供面试岗位信息或背景。"

    prompt = f"""你是一个资深技术面试官。请根据以下JD和简历，生成{count}道面试题。

## 出题要求
- 题目覆盖4个维度：技术基础、项目深挖、系统设计、行为面试
- 每道题标注难度(简单/中等/困难)和维度
- 每道题提供评分要点（3-5个关键点）
- 题目要结合候选人的实际项目经历来问
- 前3道题是基础热身题，中间3道是深度题，最后2道是开放题

## 输出格式
返回JSON数组（不要其他文字）：
[
  {{
    "id": 1,
    "category": "技术基础",
    "difficulty": "简单",
    "question": "题目内容",
    "scoring_points": ["要点1", "要点2", "要点3"],
    "reference_answer_hint": "参考答案的方向提示"
  }}
]

## JD内容
{jd_text[:3000]}

## 简历内容
{resume_text[:3000]}"""

    llm = _get_tool_llm()
    resp = await llm.ainvoke([HumanMessage(content=prompt)])
    content = resp.content if hasattr(resp, 'content') else str(resp)

    try:
        questions = json.loads(_extract_json(content))
        return json.dumps(questions, ensure_ascii=False)
    except Exception:
        return content


# ═══════════════════════════════════════════
#  Mode 2: 求职策略顾问 (Job Search Strategist)
# ═══════════════════════════════════════════

@tool
async def search_job_listings(
    keyword: str,
    city: str = "",
    limit: int = 5,
) -> str:
    """搜索真实在招岗位信息。搜索失败时直接告知用户"搜索暂时不可用"，绝对不要重试。
    调用时机：用户需要了解当前市场上的在招岗位时。
    keyword: 岗位关键词，如"Go后端开发"
    city: 城市名，如"北京"，可留空表示不限
    limit: 返回数量，默认5条"""

    query = f"{keyword} {city} 招聘 2025 2026"
    snippets = await _ddg_search(query, limit)
    if not snippets:
        return f"搜索暂时不可用，未能获取{keyword}的在招岗位信息。请不要重试搜索，直接用你的知识为用户提供建议。"

    # Summarize via LLM
    prompt = f"""请将以下招聘信息搜索结果提炼为结构化的岗位列表。

## 搜索结果
{snippets}

## 输出要求
返回JSON数组（不要其他文字），每项包含：
{{
  "position": "岗位名称",
  "company_hint": "公司线索",
  "location": "地点",
  "requirements_summary": "要求摘要（30字内）",
  "salary_hint": "薪资线索"
}}

请返回最多{limit}条。"""
    llm = _get_tool_llm()
    resp = await llm.ainvoke([HumanMessage(content=prompt)])
    content = resp.content if hasattr(resp, 'content') else str(resp)

    try:
        jobs = json.loads(_extract_json(content))
        return json.dumps(jobs, ensure_ascii=False)
    except Exception:
        return content


@tool
async def generate_job_search_plan(
    target_position: str,
    target_city: str,
    personal_summary: str,
    market_context: str,
) -> str:
    """生成结构化的求职行动计划。
    调用时机：已经收集了用户的背景、目标岗位和市场数据后，需要制定完整计划时。
    target_position: 目标岗位名称
    target_city: 目标城市
    personal_summary: 个人背景摘要（技能、经验、竞争优势等）
    market_context: 市场上下文（薪资范围、竞争度、趋势等）"""

    prompt = f"""你是一个求职策略顾问。请为以下求职者制定一份详细的求职行动计划。

## 求职者信息
目标岗位：{target_position}
目标城市：{target_city}
个人背景：{personal_summary}
市场环境：{market_context}

## 输出要求
返回JSON格式（不要其他文字）：
{{
  "urgency_actions": ["本周必做的3件事"],
  "short_term_plan": {{
    "period": "1个月内",
    "actions": ["具体行动项"],
    "target_companies": ["推荐公司1", "推荐公司2", "推荐公司3"],
    "expected_interviews": "预计面试数量",
    "resume_focus": "简历中需要突出的点"
  }},
  "mid_term_plan": {{
    "period": "1-3个月",
    "actions": ["具体行动项"],
    "skill_improvements": ["需要提升的技能"],
    "networking_tips": "人脉拓展建议"
  }},
  "salary_expectation": "合理薪资期望区间",
  "success_metrics": "如何判断求职成功的指标",
  "risk_warnings": ["需要注意的风险点"]
}}"""

    llm = _get_tool_llm()
    resp = await llm.ainvoke([HumanMessage(content=prompt)])
    content = resp.content if hasattr(resp, 'content') else str(resp)

    try:
        plan = json.loads(_extract_json(content))
        return json.dumps(plan, ensure_ascii=False)
    except Exception:
        return content


# ═══════════════════════════════════════════
#  Mode 3: 职业转折规划 (Career Pivot Planner)
# ═══════════════════════════════════════════

@tool
async def analyze_skill_transferability(
    current_skills_json: str,
    target_field: str,
) -> str:
    """分析从当前技能到目标领域的可迁移性。
    调用时机：用户想从当前领域转向另一个领域时，需要评估技能匹配度。
    current_skills_json: 当前技能的JSON列表或描述，如 '["Java","Spring Boot","MySQL"]' 或技能描述文本
    target_field: 目标领域，如"AI/机器学习"、"前端开发"、"产品经理" """

    prompt = f"""你是一个职业转型顾问。请分析以下技能向目标领域的可迁移性。

## 当前技能
{current_skills_json}

## 目标领域
{target_field}

## 输出要求
返回JSON格式（不要其他文字）：
{{
  "transferability_score": 60,
  "directly_transferable": [
    {{"skill": "技能名", "reason": "为什么可以直接迁移", "value": "该技能在目标领域的价值（高/中/低）"}}
  ],
  "need_enhancement": [
    {{"skill": "技能名", "current_level": "当前水平", "target_level": "需要的水平", "how_to_improve": "提升建议"}}
  ],
  "need_to_learn": [
    {{"skill": "技能名", "importance": "重要性（高/中）", "learning_resource": "学习资源建议", "estimated_time": "预计学习时间"}}
  ],
  "transition_difficulty": "转型难度评估（容易/中等/困难）",
  "suggested_timeline": "建议转型周期（如3-6个月）",
  "overall_advice": "综合转型建议（100字以内）"
}}"""

    llm = _get_tool_llm()
    resp = await llm.ainvoke([HumanMessage(content=prompt)])
    content = resp.content if hasattr(resp, 'content') else str(resp)

    try:
        return json.dumps(json.loads(_extract_json(content)), ensure_ascii=False)
    except Exception:
        return content


@tool
async def recommend_transition_roles(
    current_field: str,
    target_field: str,
    experience_level: str,
) -> str:
    """推荐从当前领域到目标领域的过渡岗位。
    调用时机：用户确定了转型方向后，需要知道有哪些过渡岗位可以平滑切入。
    current_field: 当前领域，如"Java后端"
    target_field: 目标领域，如"AI工程师"
    experience_level: 工作年限，如"3年"或"应届" """

    prompt = f"""你是一个职业规划师。请为以下转型路径推荐过渡岗位。

## 转型路径
当前领域：{current_field}
目标领域：{target_field}
经验：{experience_level}

## 输出要求
返回JSON数组（不要其他文字），每个岗位包含：
{{
  "role": "过渡岗位名称",
  "stage": "第几阶段（1-3）",
  "why_suitable": "为什么这个岗位适合作为过渡",
  "skill_overlap": "与该岗位重叠的现有技能",
  "new_skills_to_learn": "需要新学的技能",
  "typical_salary": "典型薪资范围",
  "job_search_keywords": "搜索该岗位的关键词"
}}

请推荐3-5个过渡岗位，按从最容易切入到最接近目标排序。"""

    llm = _get_tool_llm()
    resp = await llm.ainvoke([HumanMessage(content=prompt)])
    content = resp.content if hasattr(resp, 'content') else str(resp)

    try:
        roles = json.loads(_extract_json(content))
        return json.dumps(roles, ensure_ascii=False)
    except Exception:
        return content


# ═══════════════════════════════════════════
#  Mode 4: 薪资谈判教练 (Salary Negotiation Coach)
# ═══════════════════════════════════════════

@tool
async def get_salary_benchmark(
    position: str,
    city: str,
    experience_level: str,
) -> str:
    """获取某个岗位的薪资基准数据。
    调用时机：用户想了解某个岗位在特定城市的薪资范围时。
    position: 岗位名称，如"前端开发工程师"
    city: 城市，如"北京"
    experience_level: 经验级别，如"应届"、"3年"、"5年+" """

    query = f"{position} {city} {experience_level} 薪资 2025 2026"
    snippets = await _ddg_search(query, 5)

    prompt = f"""根据以下搜索结果和你的知识，给出该岗位的薪资基准。

## 岗位信息
岗位：{position}
城市：{city}
经验：{experience_level}

## 搜索结果
{snippets or '（无搜索结果，请基于训练数据估算）'}

## 输出要求
返回JSON格式（不要其他文字）：
{{
  "salary_range_monthly": "月薪范围（如15K-25K）",
  "median_salary": "中位数",
  "percentile_75": "75分位",
  "percentile_90": "90分位",
  "common_benefits": ["常见福利1", "常见福利2"],
  "negotiation_room": "通常谈判空间（如10%-20%）",
  "data_source_note": "数据来源说明"
}}"""

    llm = _get_tool_llm()
    resp = await llm.ainvoke([HumanMessage(content=prompt)])
    content = resp.content if hasattr(resp, 'content') else str(resp)

    try:
        return json.dumps(json.loads(_extract_json(content)), ensure_ascii=False)
    except Exception:
        return content


@tool
async def generate_negotiation_strategy(
    position: str,
    salary_data_json: str,
    personal_strengths: str,
) -> str:
    """生成个性化的薪资谈判策略和话术。
    调用时机：用户获得了薪资数据，准备进行薪资谈判时。
    position: 岗位名称
    salary_data_json: 薪资基准数据（JSON字符串）
    personal_strengths: 个人优势（如"3年相关经验、项目成果好、有竞赛奖项"）"""

    prompt = f"""你是一个资深薪资谈判教练。请为以下求职场景生成谈判策略。

## 场景
岗位：{position}
个人优势：{personal_strengths}
薪资数据：{salary_data_json}

## 输出要求
返回JSON格式（不要其他文字）：
{{
  "target_salary": "建议目标薪资（具体数字和理由）",
  "reservation_salary": "底线薪资",
  "opening_script": "开场话术（当HR问'你的期望薪资是多少'时怎么说）",
  "counter_offer_scripts": [
    {{"scenario": "HR说预算不够", "response": "如何回应"}},
    {{"scenario": "HR说先给低点以后会涨", "response": "如何回应"}},
    {{"scenario": "HR问你现在薪资", "response": "如何回应"}},
    {{"scenario": "HR说需要和领导商量", "response": "如何推动进程"}}
  ],
  "key_talking_points": ["关键谈判论点1", "关键谈判论点2", "关键谈判论点3"],
  "common_mistakes": ["常见谈判错误1", "常见谈判错误2"],
  "non_salary_negotiation": ["除了薪资还可以谈的福利"]
}}"""

    llm = _get_tool_llm()
    resp = await llm.ainvoke([HumanMessage(content=prompt)])
    content = resp.content if hasattr(resp, 'content') else str(resp)

    try:
        return json.dumps(json.loads(_extract_json(content)), ensure_ascii=False)
    except Exception:
        return content


# ═══════════════════════════════════════════
#  Shared utility
# ═══════════════════════════════════════════

async def _ddg_search(query: str, limit: int = 5) -> str:
    """DuckDuckGo HTML search, returns number-prefixed snippet lines. 3s timeout."""
    url = "https://html.duckduckgo.com/html/"
    try:
        async with httpx.AsyncClient(timeout=3.0) as client:
            resp = await client.get(
                url,
                params={"q": query},
                headers={"User-Agent": "OfferHunter/2.0 (career-tool)"},
            )
            snippets = re.findall(r'class="result__snippet"[^>]*>(.*?)</a>', resp.text)
            if not snippets:
                snippets = re.findall(r'class="result__snippet">(.+?)</a>', resp.text)
            lines = []
            for i, s in enumerate(snippets):
                text = re.sub(r"<[^>]*>", "", s).strip()
                if len(text) > 20:
                    lines.append(f"{i+1}. {text}")
                    if len(lines) >= limit:
                        break
            return "\n".join(lines)
    except Exception as e:
        logger.warning("DDG search failed: %s", e)
        return ""


def _extract_json(text: str) -> str:
    """Strip markdown fences and extract the outermost JSON object."""
    cleaned = text.strip()
    for prefix in ("```json", "```JSON", "```"):
        if cleaned.startswith(prefix):
            cleaned = cleaned[len(prefix):]
    if cleaned.endswith("```"):
        cleaned = cleaned[:-3]
    cleaned = cleaned.strip()
    start = cleaned.find("{")
    if start == -1:
        return cleaned  # not JSON object, return as-is
    depth = 0
    for i in range(start, len(cleaned)):
        if cleaned[i] == "{":
            depth += 1
        elif cleaned[i] == "}":
            depth -= 1
            if depth == 0:
                return cleaned[start:i + 1]
    return cleaned


# ═══════════════════════════════════════════
#  Tool lists per mode
# ═══════════════════════════════════════════

from services.agent_tools import parse_resume, parse_jd, analyze_match, research_market

INTERVIEW_TOOLS = [generate_interview_questions]
STRATEGIST_TOOLS = [search_job_listings, generate_job_search_plan, parse_resume, research_market]
PIVOT_TOOLS = [analyze_skill_transferability, recommend_transition_roles, parse_resume, research_market]
NEGOTIATION_TOOLS = [get_salary_benchmark, generate_negotiation_strategy]

# ═══════════════════════════════════════════
#  Composite tools — wrap multiple fine-grained tools to keep total count ≤8
#  (DeepSeek function calling may be unreliable with >8 tool definitions)
# ═══════════════════════════════════════════

@tool
async def start_mock_interview(
    jd_text: str = "",
    resume_text: str = "",
    config=None,
) -> str:
    """启动模拟面试：根据JD和简历生成8-10道面试题，然后逐题考核用户。
    调用时机：用户说"模拟面试"、"面试一下"、"帮我准备面试"等。
    jd_text: 岗位JD原文（可选，如果之前已通过parse_jd解析则可留空）
    resume_text: 简历原文（可选，如果之前已通过parse_resume解析则可留空）"""
    from services.agent_tools import _thread_cache as cache

    tid = _get_thread_id_from_config(config)
    jd = jd_text or cache.get(tid, {}).get("job_text", "")
    resume = resume_text or cache.get(tid, {}).get("resume_text", "")
    if not jd or not resume:
        return "需要先提供简历和JD才能生成面试题。请先让用户提供这两个信息。"

    result = await generate_interview_questions.ainvoke({"jd_text": jd, "resume_text": resume})
    return f"面试题已生成。请你作为面试官，按顺序逐题提问，每题等用户回答后评分(1-10)并给改进建议，然后进入下一题。\n\n{result}"


@tool
async def create_job_search_plan(
    target_position: str,
    target_city: str = "",
    experience_summary: str = "",
    salary_expectation: str = "",
    config=None,
) -> str:
    """制定求职策略和行动计划：搜索岗位、分析市场、生成分阶段求职方案。
    调用时机：用户说"找XX工作"、"求XX岗位"、"帮我规划求职"等。
    target_position: 目标岗位，如"Go后端开发"
    target_city: 目标城市，如"上海"
    experience_summary: 经验概述，如"2年前端开发"
    salary_expectation: 期望薪资，如"18-25K"（可选）"""

    # Step 1: Search listings
    listings = await search_job_listings.ainvoke({
        "keyword": target_position,
        "city": target_city,
        "limit": 5,
    })
    # Step 2: Market research
    mkt = await research_market.ainvoke({
        "job_info": f"{target_position} {target_city}",
        "job_preference": experience_summary,
    })
    # Step 3: Generate plan
    plan = await generate_job_search_plan.ainvoke({
        "target_position": target_position,
        "target_city": target_city or "未指定",
        "personal_summary": f"{experience_summary}，期望薪资{salary_expectation}" if salary_expectation else experience_summary,
        "market_context": mkt,
    })
    return f"求职策略制定完成。\n\n## 市场岗位\n{listings}\n\n## 市场分析\n{mkt}\n\n## 行动计划\n{plan}"


@tool
async def plan_career_pivot(
    current_field: str,
    target_field: str,
    experience_summary: str,
    config=None,
) -> str:
    """规划职业转型路径：分析可迁移技能、推荐过渡岗位、制定学习路线。
    调用时机：用户说"转行"、"转XX方向"、"从XX转到XX"等。
    current_field: 当前领域/技术栈，如"Java后端"
    target_field: 目标领域，如"AI/机器学习"
    experience_summary: 当前经验概述，含技能列表和工作年限"""

    # Step 1: Skill transferability
    skills = await analyze_skill_transferability.ainvoke({
        "current_skills_json": experience_summary,
        "target_field": target_field,
    })
    # Step 2: Transition roles
    roles = await recommend_transition_roles.ainvoke({
        "current_field": current_field,
        "target_field": target_field,
        "experience_level": experience_summary[:30],
    })
    # Step 3: Market for target field
    mkt = await research_market.ainvoke({
        "job_info": f"{target_field}",
        "job_preference": experience_summary[:50],
    })
    return f"职业转型分析完成。\n\n## 技能迁移\n{skills}\n\n## 过渡岗位\n{roles}\n\n## 目标市场\n{mkt}"


@tool
async def advise_salary_negotiation(
    position: str,
    city: str,
    experience_level: str,
    personal_strengths: str = "",
    config=None,
) -> str:
    """提供薪资谈判指导：查询薪资基准、生成谈判策略和话术。
    调用时机：用户问"薪资多少"、"怎么谈薪"、"帮我看看这个offer"等。
    position: 岗位名称
    city: 城市
    experience_level: 经验级别，如"应届"、"3年"
    personal_strengths: 个人优势和筹码，如"有3年经验、项目成果好"（可选）"""

    bench = await get_salary_benchmark.ainvoke({
        "position": position,
        "city": city,
        "experience_level": experience_level,
    })
    strat = await generate_negotiation_strategy.ainvoke({
        "position": position,
        "salary_data_json": bench,
        "personal_strengths": personal_strengths or f"{city}{experience_level}{position}",
    })
    return f"薪资谈判分析完成。\n\n## 薪资基准\n{bench}\n\n## 谈判策略\n{strat}"


def _get_thread_id_from_config(config):
    if config and isinstance(config, dict) and "configurable" in config:
        return config["configurable"].get("thread_id", "default")
    return "default"


# Export: 4 composite tools for the agent (core 4 + these 4 = 8 total)
ALL_ADVANCED_TOOLS = [
    start_mock_interview,
    create_job_search_plan,
    plan_career_pivot,
    advise_salary_negotiation,
]

# All fine-grained tools (for programmatic use or debugging)
ALL_FINEGRAINED_TOOLS = [
    generate_interview_questions,
    search_job_listings,
    generate_job_search_plan,
    analyze_skill_transferability,
    recommend_transition_roles,
    get_salary_benchmark,
    generate_negotiation_strategy,
]
