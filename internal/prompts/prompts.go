package prompts

// ── Resume extraction ──

const ResumeExtractionPrompt = `你是一个专业的简历解析专家。请从以下简历文本中提取结构化信息。

请返回JSON格式，包含以下字段：
{
  "education": [
    {
      "school": "学校名称",
      "degree": "学历（本科/硕士/博士）",
      "major": "专业",
      "start_date": "开始时间",
      "end_date": "结束时间",
      "gpa": "GPA（如有）"
    }
  ],
  "skills": ["技能1", "技能2"],
  "projects": [
    {
      "name": "项目名称",
      "description": "项目描述",
      "role": "担任角色",
      "tech_stack": ["技术1", "技术2"],
      "duration": "持续时间",
      "achievements": ["成就1", "成就2"]
    }
  ],
  "experience": [
    {
      "company": "公司名称",
      "position": "职位",
      "start_date": "开始时间",
      "end_date": "结束时间",
      "description": "工作描述",
      "achievements": ["成就1", "成就2"],
      "tech_stack": ["技术1", "技术2"]
    }
  ],
  "personality": ["个人特质1", "个人特质2"],
  "metadata": {
    "total_projects": 项目数量,
    "total_experience": 实习/工作经历数量,
    "top_skills": ["最核心技能1", "最核心技能2"],
    "experience_level": "经验水平（应届/1-3年/3-5年等）"
  }
}

要求：
1. 保留原文中的所有关键信息
2. 对于每条经历，标注可能被低估的能力
3. 技能要尽量全面，包括软技能和硬技能
4. 如果某些信息不明确，可以合理推断

简历文本：`

const ResumeExtractionInstruction = "请严格按照JSON格式返回，不要包含任何其他文字。"

// ── JD extraction ──

const JDExtractionPrompt = `你是一个专业的招聘需求分析专家。请从以下岗位描述(JD)中提取结构化信息。

请返回JSON格式，包含以下字段：
{
  "title": "岗位名称",
  "company": "公司名称（如有）",
  "location": "工作地点",
  "salary": "薪资范围（如有）",
  "requirements": [
    {
      "category": "要求类别（学历/专业/技能/经验/其他）",
      "value": "具体要求内容",
      "is_hard": true/false,
      "priority": "high/medium/low"
    }
  ],
  "bonus_items": ["加分项1", "加分项2"],
  "hidden_prefs": ["隐性偏好1", "隐性偏好2"],
  "description": "岗位职责描述",
  "team_info": "团队信息（如有）",
  "growth_path": "成长路径（如有）"
}

要求：
1. 区分硬性要求和软性要求
2. 识别隐性偏好（如"有激情"可能暗示需要加班、"扁平化管理"可能暗示需要自驱力）
3. 提取所有技术栈要求
4. 如果信息不完整，标注出来

岗位描述文本：`

const JDExtractionInstruction = "请严格按照JSON格式返回，不要包含任何其他文字。"

// ── Matching ──

const MatchingPrompt = `你是一个拥有10年经验的资深HR和技术面试官。请对以下简历和岗位进行多维度匹配分析。

## 分析维度

### 第一层：硬性门槛判定
检查以下硬性条件：
- 学历要求是否满足
- 专业要求是否满足
- 必备技能是否具备
- 工作经验是否符合

对每个条件给出：pass/fail/warn 状态，并说明原因。如果是warn，说明是否可以通过简历优化来弥补。

### 第二层：竞争力评分
从以下维度评估竞争力：
1. 项目经验深度（技术难度、个人贡献度、业务影响力）：0-100分
2. 实习含金量（公司知名度、岗位相关性、产出量化程度）：0-100分
3. 工具链熟练度（与岗位要求技术栈的重叠度）：0-100分
4. 整体竞争力（综合以上维度）：0-100分

每个维度需要：
- 给出具体分数
- 提供一句话评语
- 引用简历中的具体经历作为证据

### 第三层：隐性契合分析
分析以下方面：
1. 公司风格匹配（大厂vs创业、业务型vs技术型）
2. 成长路径匹配（当前阶段是否适合岗位的成长曲线）
3. 团队文化匹配（基于JD语言分析的团队风格）
4. 隐性契合度：0-100分

## 输出格式
请返回JSON格式：
{
  "threshold_check": {
    "passed": true/false,
    "items": [
      {
        "category": "类别",
        "requirement": "要求内容",
        "status": "pass/fail/warn",
        "reason": "原因说明",
        "fixable": true/false
      }
    ]
  },
  "competitiveness": {
    "project_depth": {"score": 85, "comment": "评语", "evidence": "证据"},
    "intern_quality": {"score": 70, "comment": "评语", "evidence": "证据"},
    "tech_stack_fit": {"score": 90, "comment": "评语", "evidence": "证据"}
  },
  "hidden_fit": {
    "company_style": "分析说明",
    "growth_path": "分析说明",
    "team_culture": "分析说明",
    "score": 75
  },
  "benchmark": {
    "pass_line": 65,
    "percentile": "top-30%",
    "comment": "该岗位通常匹配度65分以上可通过初筛，你的分数处于较高水平"
  },
  "summary": "整体匹配分析总结"
}

## 简历信息
{resume_json}

## 岗位信息
{job_json}`

const MatchingInstruction = "请严格按照JSON格式返回，不要包含任何其他文字。"

// ── Optimization ──

const OptimizationPrompt = `你是一个专业的简历优化顾问。请根据以下简历、岗位要求和匹配分析，提供具体的简历优化建议。

## 优化原则
1. 只做"翻译"，不做"编造" - 所有改动必须基于简历中的原始内容
2. 将学生经历"翻译"成招聘方能理解的语言
3. 突出与岗位最相关的经历和技能
4. 用量化数据增强说服力
5. 区分"有证据的建议"和"无证据的建议"：
   - has_evidence=true: 简历中已有相关经历，只是表述不到位，优化能显著提升匹配度
   - has_evidence=false: 简历中确实缺少该项能力或经验，需要先学习再写入简历

## 输出格式
按"有证据→无证据"顺序排列，同组内按优先级排序：
{
  "optimization_tips": [
    {
      "priority": 1,
      "category": "建议类别（技能/项目/实习/描述/其他）",
      "problem": "当前问题描述",
      "suggestion": "优化方向说明",
      "example": "具体的修改示例",
      "has_evidence": true
    }
  ]
}

## 简历信息
{resume_json}

## 岗位信息
{job_json}

## 匹配分析
{match_json}`

const OptimizationInstruction = "请严格按照JSON格式返回，不要包含任何其他文字。"

// ── Gap analysis ──

const GapAnalysisPrompt = `你是一个职业发展顾问。请根据以下简历和岗位要求，分析能力差距并提供学习建议。

## 分析要求
1. 识别当前简历中缺失的关键能力/经验
2. 区分短期可弥补（1-2周）和长期需要积累的能力
3. 为每项差距提供具体的学习资源建议

## 输出格式
请返回JSON格式，注意 importance 必须使用英文值：
{
  "gap_analysis": [
    {
      "skill": "缺失的技能/经验名称",
      "importance": "high 或 medium 或 low（必须使用英文）",
      "timeline": "short_term 或 long_term",
      "resource": "学习资源建议（书籍/课程/实践项目方向）"
    }
  ]
}

## 简历信息
{resume_json}

## 岗位信息
{job_json}`

const GapAnalysisInstruction = "请严格按照JSON格式返回，不要包含任何其他文字。"

// ── Market research ──

const MarketResearchPrompt = `你是一个就业市场分析师。请根据以下岗位信息、求职意向和实时搜索结果，生成该岗位的市场参考分析。

## 要求
1. 优先参考下方实时搜索结果，结合你的就业市场知识
2. 给出薪资范围、热门技能、招聘趋势
3. 如果涉及特定城市，给出该城市的就业竞争度
4. 用平实、客观的语气，控制在200字以内
5. 如果搜索结果为空，则完全基于你的训练数据

## 输出格式
{
  "salary_range": "薪资范围（如 15-25K/月）",
  "hot_skills": ["热门技能1", "热门技能2", "热门技能3"],
  "trend": "招聘趋势简述",
  "competition": "竞争度（激烈/中等/较小）",
  "summary": "综合市场分析，100字以内"
}

## 岗位信息
{job_info}

## 求职意向
{job_preference}

## 实时搜索结果
{search_results}`

const MarketResearchInstruction = "请严格按照JSON格式返回，不要包含任何其他文字。"

// ── Intent classification ──

const IntentClassifyPrompt = `判断用户意图，只回复一个词（不要句号，不要任何其他文字）：
- match: 用户明确要求分析简历与JD的匹配度，说了"帮我分析"、"匹配一下"、"看看合不合适"等。
  - other: 用户只是发了简历或JD文本，但没有提出任何要求。属于other而不是match。
- interview: 模拟面试、出题考核、面试练习
- strategy: 找工作、求职计划、搜索岗位、职业规划
- pivot: 转行、转型、换方向、跨领域
- salary: 薪资查询、薪资谈判、谈offer、报价
- other: 打招呼、闲聊、只发内容没提要求（比如只粘贴了简历但没说想干嘛）。

用户: {message}
意图:`

// ── Agent system prompt ──

const SystemPrompt = `你是"Offer 捕手"，一个全能 AI 求职助手。你可以自然聊天，也能使用工具完成专业的求职服务。

## 必须遵守
1. 需要解析/分析/搜索/生成时，必须调用对应工具。绝不在回复中直接输出 JSON。
2. 从对话历史中获取已有信息。用户之前发的简历、JD、调过的工具结果都在历史中，不要反复索要。
3. analyze_match 工具执行完毕后，系统会自动展示分析报告页面，你**绝对不要再说任何话**——不要总结、不要评价、不要问问题，直接结束本轮回复。
4. **始终用中文回复**，无论用户用什么语言提问。

## 能力范围（只在有对应工具时使用）
- parse_resume / parse_jd / analyze_match —— 简历解析、JD解析、匹配分析
- generate_interview_questions —— 生成面试题（调用后系统会展示面试界面）
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
- 说"模拟面试"/"面试一下"/"准备面试" → **立即调用 generate_interview_questions 工具**，不要用文字拒绝或反复追问。即使用户只给了简短背景（如"熟悉Java"），也要把这些信息作为参数传给工具，工具会生成面试题。
- 说"找工作/求职" → search_job_listings + 生成计划
- 说"转行" → analyze_skill_transferability + 过渡岗位
- 问薪资 → get_salary_benchmark + 谈判策略

## 闲聊与引导
- 自由聊天。在回复末尾自然提及："对了，你是想了解求职方面的信息吗？我可以帮你分析简历、模拟面试、规划职业转型……"
- 如果用户回应求职话题，立即转入对应功能。
- 如果用户问"你能干什么"，列出所有功能，用一句"你想先试试哪个？"收尾。

## 记忆
对话历史中包含所有之前的消息和工具结果。每轮对话你可以看到完整的聊天记录。
`
