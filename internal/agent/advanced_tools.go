package agent

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"

	"github.com/sashabaranov/go-openai"

	"offerhunter/internal/config"
	"offerhunter/internal/llm"
	"offerhunter/internal/models"
	"offerhunter/internal/services"
)

func intParam(desc string) map[string]any {
	return map[string]any{"type": "integer", "description": desc}
}

// advancedTools holds the 7 fine-grained tools for interview / strategy /
// pivot / salary intents.
var advancedTools = []Tool{
	{
		Name: "generate_interview_questions",
		Description: "生成模拟面试的面试题目（JSON数组）。" +
			"调用时机：用户说\"模拟面试\"\"面试一下\"\"准备面试\"时，**必须立即调用此工具**，不要用文字询问或拒绝。" +
			"即使信息简短（如用户只说\"熟悉Java\"），也要把已知信息作为参数传入，工具会生成通用面试题。" +
			"jd_text: 岗位描述（没有就传空字符串）" +
			"resume_text: 简历或背景描述（没有就传空字符串）" +
			"count: 题目数量，默认8道",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"jd_text":     strParam("岗位描述原文（可空）"),
				"resume_text": strParam("简历原文或背景（可空）"),
				"count":       intParam("题目数量，默认8道"),
			},
		},
	},
	{
		Name: "search_job_listings",
		Description: "搜索真实在招岗位信息。搜索失败时直接告知用户\"搜索暂时不可用\"，绝对不要重试。" +
			"调用时机：用户需要了解当前市场上的在招岗位时。" +
			"keyword: 岗位关键词，如\"Go后端开发\"" +
			"city: 城市名，如\"北京\"，可留空表示不限" +
			"limit: 返回数量，默认5条",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"keyword": strParam("岗位关键词"),
				"city":    strParam("城市名（可留空）"),
				"limit":   intParam("返回数量，默认5条"),
			},
		},
	},
	{
		Name: "generate_job_search_plan",
		Description: "生成结构化的求职行动计划。" +
			"调用时机：已经收集了用户的背景、目标岗位和市场数据后，需要制定完整计划时。" +
			"target_position: 目标岗位名称" +
			"target_city: 目标城市" +
			"personal_summary: 个人背景摘要（技能、经验、竞争优势等）" +
			"market_context: 市场上下文（薪资范围、竞争度、趋势等）",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"target_position":  strParam("目标岗位名称"),
				"target_city":      strParam("目标城市"),
				"personal_summary": strParam("个人背景摘要"),
				"market_context":   strParam("市场上下文"),
			},
		},
	},
	{
		Name: "analyze_skill_transferability",
		Description: "分析从当前技能到目标领域的可迁移性。" +
			"调用时机：用户想从当前领域转向另一个领域时，需要评估技能匹配度。" +
			"current_skills_json: 当前技能的JSON列表或描述，如 '[\"Java\",\"Spring Boot\",\"MySQL\"]' 或技能描述文本" +
			"target_field: 目标领域，如\"AI/机器学习\"、\"前端开发\"、\"产品经理\"",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"current_skills_json": strParam("当前技能的JSON列表或描述"),
				"target_field":        strParam("目标领域"),
			},
		},
	},
	{
		Name: "recommend_transition_roles",
		Description: "推荐从当前领域到目标领域的过渡岗位。" +
			"调用时机：用户确定了转型方向后，需要知道有哪些过渡岗位可以平滑切入。" +
			"current_field: 当前领域，如\"Java后端\"" +
			"target_field: 目标领域，如\"AI工程师\"" +
			"experience_level: 工作年限，如\"3年\"或\"应届\"",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"current_field":    strParam("当前领域"),
				"target_field":     strParam("目标领域"),
				"experience_level": strParam("工作年限"),
			},
		},
	},
	{
		Name: "get_salary_benchmark",
		Description: "获取某个岗位的薪资基准数据。" +
			"调用时机：用户想了解某个岗位在特定城市的薪资范围时。" +
			"position: 岗位名称，如\"前端开发工程师\"" +
			"city: 城市，如\"北京\"" +
			"experience_level: 经验级别，如\"应届\"、\"3年\"、\"5年+\"",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"position":         strParam("岗位名称"),
				"city":             strParam("城市"),
				"experience_level": strParam("经验级别"),
			},
		},
	},
	{
		Name: "generate_negotiation_strategy",
		Description: "生成个性化的薪资谈判策略和话术。" +
			"调用时机：用户获得了薪资数据，准备进行薪资谈判时。" +
			"position: 岗位名称" +
			"salary_data_json: 薪资基准数据（JSON字符串）" +
			"personal_strengths: 个人优势（如\"3年相关经验、项目成果好、有竞赛奖项\"）",
		Parameters: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"position":           strParam("岗位名称"),
				"salary_data_json":   strParam("薪资基准数据（JSON字符串）"),
				"personal_strengths": strParam("个人优势"),
			},
		},
	},
}

var advancedToolExec = map[string]ToolFunc{
	"generate_interview_questions":   execGenerateInterviewQuestions,
	"search_job_listings":            execSearchJobListings,
	"generate_job_search_plan":       execGenerateJobSearchPlan,
	"analyze_skill_transferability":  execAnalyzeSkillTransferability,
	"recommend_transition_roles":     execRecommendTransitionRoles,
	"get_salary_benchmark":           execGetSalaryBenchmark,
	"generate_negotiation_strategy":  execGenerateNegotiationStrategy,
}

func init() {
	allTools = append(allTools, advancedTools...)
	for name, fn := range advancedToolExec {
		toolExec[name] = fn
	}
}

// toolLLMCall invokes the shared tool LLM (temperature 0.3) with a single
// user prompt and returns the raw text content.
func toolLLMCall(ctx context.Context, user string) (string, error) {
	resp, err := llm.DeepSeek().CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: config.DeepSeekModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		Temperature: 0.3,
		MaxTokens:   4096,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("empty LLM response")
	}
	return resp.Choices[0].Message.Content, nil
}

// normalizeJSON strips markdown fences and, if the content is valid JSON,
// re-serializes it compactly; otherwise returns the original content.
func normalizeJSON(content string) string {
	cleaned := strings.TrimSpace(content)
	for _, prefix := range []string{"```json", "```JSON", "```"} {
		if strings.HasPrefix(cleaned, prefix) {
			cleaned = cleaned[len(prefix):]
		}
	}
	if strings.HasSuffix(cleaned, "```") {
		cleaned = cleaned[:len(cleaned)-3]
	}
	cleaned = strings.TrimSpace(cleaned)

	var v any
	if err := json.Unmarshal([]byte(cleaned), &v); err == nil {
		b, _ := json.Marshal(v)
		return string(b)
	}
	return content
}

func truncateRunes(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(strings.TrimSpace(s))
}

func itoa(n int) string {
	return strconv.Itoa(n)
}

const interviewQuestionsPrompt = `你是一个资深技术面试官。请根据以下JD和简历，生成{count}道面试题。

## 出题要求
- 题目覆盖4个维度：技术基础、项目深挖、系统设计、行为面试
- 每道题标注难度(简单/中等/困难)和维度
- 每道题提供评分要点（3-5个关键点）
- 题目要结合候选人的实际项目经历来问
- 前3道题是基础热身题，中间3道是深度题，最后2道是开放题

## 输出格式
返回JSON数组（不要其他文字）：
[
  {
    "id": 1,
    "category": "技术基础",
    "difficulty": "简单",
    "question": "题目内容",
    "scoring_points": ["要点1", "要点2", "要点3"],
    "reference_answer_hint": "参考答案的方向提示"
  }
]

## JD内容
{jd_text}

## 简历内容
{resume_text}`

func execGenerateInterviewQuestions(ctx context.Context, args map[string]string, st *ThreadState) (string, error) {
	jdText := args["jd_text"]
	resumeText := args["resume_text"]

	// Fall back to the parsed resume/JD stored on the thread state, so the
	// questions are grounded in what the user actually uploaded/pasted.
	if jdText == "" {
		jdText = st.jobText
	}
	if resumeText == "" {
		resumeText = st.resumeText
	}
	if jdText == "" {
		jdText = "（未提供岗位JD，请生成通用技术岗位面试题）"
	}
	if resumeText == "" {
		resumeText = "（未提供简历，请生成通用面试题）"
	}

	count := 8
	if c := args["count"]; c != "" {
		if v, err := parseInt(c); err == nil && v > 0 {
			count = v
		}
	}

	prompt := strings.ReplaceAll(interviewQuestionsPrompt, "{count}", itoa(count))
	prompt = strings.ReplaceAll(prompt, "{jd_text}", truncateRunes(jdText, 3000))
	prompt = strings.ReplaceAll(prompt, "{resume_text}", truncateRunes(resumeText, 3000))

	content, err := toolLLMCall(ctx, prompt)
	if err != nil {
		return "", err
	}

	// Parse structured questions for the interview UI, storing them on the
	// thread state so loop.go can push an "interview" SSE event.
	cleaned := strings.TrimSpace(content)
	for _, prefix := range []string{"```json", "```JSON", "```"} {
		if strings.HasPrefix(cleaned, prefix) {
			cleaned = cleaned[len(prefix):]
		}
	}
	if strings.HasSuffix(cleaned, "```") {
		cleaned = cleaned[:len(cleaned)-3]
	}
	cleaned = strings.TrimSpace(cleaned)

	var questions []models.InterviewQuestion
	if err := json.Unmarshal([]byte(cleaned), &questions); err == nil && len(questions) > 0 {
		st.interviewQuestions = questions
	}

	return normalizeJSON(content), nil
}

const searchListingsPrompt = `请将以下招聘信息搜索结果提炼为结构化的岗位列表。

## 搜索结果
{snippets}

## 输出要求
返回JSON数组（不要其他文字），每项包含：
{
  "position": "岗位名称",
  "company_hint": "公司线索",
  "location": "地点",
  "requirements_summary": "要求摘要（30字内）",
  "salary_hint": "薪资线索"
}

请返回最多{limit}条。`

func execSearchJobListings(ctx context.Context, args map[string]string, st *ThreadState) (string, error) {
	keyword := args["keyword"]
	city := args["city"]
	limit := 5
	if l := args["limit"]; l != "" {
		if v, err := parseInt(l); err == nil && v > 0 {
			limit = v
		}
	}

	query := strings.TrimSpace(keyword + " " + city + " 招聘 2025 2026")
	snippets := services.SearchDDG(query, limit)
	if snippets == "" {
		return "搜索暂时不可用，未能获取" + keyword + "的在招岗位信息。请不要重试搜索，直接用你的知识为用户提供建议。", nil
	}

	prompt := strings.ReplaceAll(searchListingsPrompt, "{snippets}", snippets)
	prompt = strings.ReplaceAll(prompt, "{limit}", itoa(limit))

	content, err := toolLLMCall(ctx, prompt)
	if err != nil {
		return "", err
	}
	return normalizeJSON(content), nil
}

const jobSearchPlanPrompt = `你是一个求职策略顾问。请为以下求职者制定一份详细的求职行动计划。

## 求职者信息
目标岗位：{target_position}
目标城市：{target_city}
个人背景：{personal_summary}
市场环境：{market_context}

## 输出要求
返回JSON格式（不要其他文字）：
{
  "urgency_actions": ["本周必做的3件事"],
  "short_term_plan": {
    "period": "1个月内",
    "actions": ["具体行动项"],
    "target_companies": ["推荐公司1", "推荐公司2", "推荐公司3"],
    "expected_interviews": "预计面试数量",
    "resume_focus": "简历中需要突出的点"
  },
  "mid_term_plan": {
    "period": "1-3个月",
    "actions": ["具体行动项"],
    "skill_improvements": ["需要提升的技能"],
    "networking_tips": "人脉拓展建议"
  },
  "salary_expectation": "合理薪资期望区间",
  "success_metrics": "如何判断求职成功的指标",
  "risk_warnings": ["需要注意的风险点"]
}`

func execGenerateJobSearchPlan(ctx context.Context, args map[string]string, st *ThreadState) (string, error) {
	prompt := jobSearchPlanPrompt
	prompt = strings.ReplaceAll(prompt, "{target_position}", args["target_position"])
	prompt = strings.ReplaceAll(prompt, "{target_city}", args["target_city"])
	prompt = strings.ReplaceAll(prompt, "{personal_summary}", args["personal_summary"])
	prompt = strings.ReplaceAll(prompt, "{market_context}", args["market_context"])

	content, err := toolLLMCall(ctx, prompt)
	if err != nil {
		return "", err
	}
	return normalizeJSON(content), nil
}

const skillTransferPrompt = `你是一个职业转型顾问。请分析以下技能向目标领域的可迁移性。

## 当前技能
{current_skills_json}

## 目标领域
{target_field}

## 输出要求
返回JSON格式（不要其他文字）：
{
  "transferability_score": 60,
  "directly_transferable": [
    {"skill": "技能名", "reason": "为什么可以直接迁移", "value": "该技能在目标领域的价值（高/中/低）"}
  ],
  "need_enhancement": [
    {"skill": "技能名", "current_level": "当前水平", "target_level": "需要的水平", "how_to_improve": "提升建议"}
  ],
  "need_to_learn": [
    {"skill": "技能名", "importance": "重要性（高/中）", "learning_resource": "学习资源建议", "estimated_time": "预计学习时间"}
  ],
  "transition_difficulty": "转型难度评估（容易/中等/困难）",
  "suggested_timeline": "建议转型周期（如3-6个月）",
  "overall_advice": "综合转型建议（100字以内）"
}`

func execAnalyzeSkillTransferability(ctx context.Context, args map[string]string, st *ThreadState) (string, error) {
	prompt := strings.ReplaceAll(skillTransferPrompt, "{current_skills_json}", args["current_skills_json"])
	prompt = strings.ReplaceAll(prompt, "{target_field}", args["target_field"])

	content, err := toolLLMCall(ctx, prompt)
	if err != nil {
		return "", err
	}
	return normalizeJSON(content), nil
}

const transitionRolesPrompt = `你是一个职业规划师。请为以下转型路径推荐过渡岗位。

## 转型路径
当前领域：{current_field}
目标领域：{target_field}
经验：{experience_level}

## 输出要求
返回JSON数组（不要其他文字），每个岗位包含：
{
  "role": "过渡岗位名称",
  "stage": "第几阶段（1-3）",
  "why_suitable": "为什么这个岗位适合作为过渡",
  "skill_overlap": "与该岗位重叠的现有技能",
  "new_skills_to_learn": "需要新学的技能",
  "typical_salary": "典型薪资范围",
  "job_search_keywords": "搜索该岗位的关键词"
}

请推荐3-5个过渡岗位，按从最容易切入到最接近目标排序。`

func execRecommendTransitionRoles(ctx context.Context, args map[string]string, st *ThreadState) (string, error) {
	prompt := strings.ReplaceAll(transitionRolesPrompt, "{current_field}", args["current_field"])
	prompt = strings.ReplaceAll(prompt, "{target_field}", args["target_field"])
	prompt = strings.ReplaceAll(prompt, "{experience_level}", args["experience_level"])

	content, err := toolLLMCall(ctx, prompt)
	if err != nil {
		return "", err
	}
	return normalizeJSON(content), nil
}

const salaryBenchmarkPrompt = `根据以下搜索结果和你的知识，给出该岗位的薪资基准。

## 岗位信息
岗位：{position}
城市：{city}
经验：{experience_level}

## 搜索结果
{snippets}

## 输出要求
返回JSON格式（不要其他文字）：
{
  "salary_range_monthly": "月薪范围（如15K-25K）",
  "median_salary": "中位数",
  "percentile_75": "75分位",
  "percentile_90": "90分位",
  "common_benefits": ["常见福利1", "常见福利2"],
  "negotiation_room": "通常谈判空间（如10%-20%）",
  "data_source_note": "数据来源说明"
}`

func execGetSalaryBenchmark(ctx context.Context, args map[string]string, st *ThreadState) (string, error) {
	position := args["position"]
	city := args["city"]
	level := args["experience_level"]

	query := strings.TrimSpace(position + " " + city + " " + level + " 薪资 2025 2026")
	snippets := services.SearchDDG(query, 5)
	if snippets == "" {
		snippets = "（无搜索结果，请基于训练数据估算）"
	}

	prompt := strings.ReplaceAll(salaryBenchmarkPrompt, "{position}", position)
	prompt = strings.ReplaceAll(prompt, "{city}", city)
	prompt = strings.ReplaceAll(prompt, "{experience_level}", level)
	prompt = strings.ReplaceAll(prompt, "{snippets}", snippets)

	content, err := toolLLMCall(ctx, prompt)
	if err != nil {
		return "", err
	}
	return normalizeJSON(content), nil
}

const negotiationPrompt = `你是一个资深薪资谈判教练。请为以下求职场景生成谈判策略。

## 场景
岗位：{position}
个人优势：{personal_strengths}
薪资数据：{salary_data_json}

## 输出要求
返回JSON格式（不要其他文字）：
{
  "target_salary": "建议目标薪资（具体数字和理由）",
  "reservation_salary": "底线薪资",
  "opening_script": "开场话术（当HR问'你的期望薪资是多少'时怎么说）",
  "counter_offer_scripts": [
    {"scenario": "HR说预算不够", "response": "如何回应"},
    {"scenario": "HR说先给低点以后会涨", "response": "如何回应"},
    {"scenario": "HR问你现在薪资", "response": "如何回应"},
    {"scenario": "HR说需要和领导商量", "response": "如何推动进程"}
  ],
  "key_talking_points": ["关键谈判论点1", "关键谈判论点2", "关键谈判论点3"],
  "common_mistakes": ["常见谈判错误1", "常见谈判错误2"],
  "non_salary_negotiation": ["除了薪资还可以谈的福利"]
}`

func execGenerateNegotiationStrategy(ctx context.Context, args map[string]string, st *ThreadState) (string, error) {
	prompt := strings.ReplaceAll(negotiationPrompt, "{position}", args["position"])
	prompt = strings.ReplaceAll(prompt, "{personal_strengths}", args["personal_strengths"])
	prompt = strings.ReplaceAll(prompt, "{salary_data_json}", args["salary_data_json"])

	content, err := toolLLMCall(ctx, prompt)
	if err != nil {
		return "", err
	}
	return normalizeJSON(content), nil
}
