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
)

func intParam(desc string) map[string]any {
	return map[string]any{"type": "integer", "description": desc}
}

// advancedTools holds the tools for the interview flow. New tools are
// registered here — init() appends them to the shared allTools/toolExec maps.
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
}

var advancedToolExec = map[string]ToolFunc{
	"generate_interview_questions": execGenerateInterviewQuestions,
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
	if jdText == "" && resumeText == "" {
		return "需要至少提供JD或简历中的一项来生成面试题。请询问用户提供面试岗位信息或背景。", nil
	}

	count := 8
	if c := args["count"]; c != "" {
		if v, err := parseInt(c); err == nil && v > 0 {
			count = v
		}
	}

	questions, raw, err := generateQuestions(ctx, jdText, resumeText, count)
	if err != nil {
		return "", err
	}
	st.interviewQuestions = questions
	return raw, nil
}

// generateQuestions builds an interview-questions prompt, calls the tool LLM,
// and parses the response into structured InterviewQuestion objects.
// Returns the questions, the raw normalized text, and any error.
func generateQuestions(ctx context.Context, jdText, resumeText string, count int) ([]models.InterviewQuestion, string, error) {
	prompt := strings.ReplaceAll(interviewQuestionsPrompt, "{count}", itoa(count))
	prompt = strings.ReplaceAll(prompt, "{jd_text}", truncateRunes(jdText, 3000))
	prompt = strings.ReplaceAll(prompt, "{resume_text}", truncateRunes(resumeText, 3000))

	content, err := toolLLMCall(ctx, prompt)
	if err != nil {
		return nil, "", err
	}

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
	if err := json.Unmarshal([]byte(cleaned), &questions); err != nil || len(questions) == 0 {
		return nil, "", errors.New("failed to parse interview questions")
	}

	return questions, normalizeJSON(content), nil
}

