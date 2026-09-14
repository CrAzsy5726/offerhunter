package agent

import (
	"context"
	"strings"

	"github.com/sashabaranov/go-openai"

	"offerhunter/internal/config"
	"offerhunter/internal/llm"
	"offerhunter/internal/prompts"
)

var intentLabels = map[string]string{
	"match":     "简历分析",
	"interview": "模拟面试",
	"strategy":  "求职策略",
	"pivot":     "职业转型",
	"salary":    "薪资谈判",
	"other":     "通用助手",
}

// ClassifyIntent returns the intent key and its Chinese label.
func ClassifyIntent(ctx context.Context, message string) (string, string) {
	short := message
	if len([]rune(message)) > 500 {
		r := []rune(message)
		short = string(r[:200]) + " ... " + string(r[len(r)-300:])
	}

	prompt := strings.ReplaceAll(prompts.IntentClassifyPrompt, "{message}", short)

	resp, err := llm.DeepSeek().CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: config.DeepSeekModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: 0,
		MaxTokens:   10,
	})
	if err != nil {
		return "other", intentLabels["other"]
	}
	if len(resp.Choices) == 0 {
		return "other", intentLabels["other"]
	}

	raw := strings.ToLower(strings.TrimSpace(resp.Choices[0].Message.Content))
	for key, label := range intentLabels {
		if strings.Contains(raw, key) {
			return key, label
		}
	}
	return "other", intentLabels["other"]
}

// intentToolNames maps an intent to the tool names it is allowed to call.
var intentToolNames = map[string][]string{
	"match":     {"parse_resume", "parse_jd", "analyze_match", "research_market"},
	"interview": {"generate_interview_questions", "research_market"},
	"strategy":  {"search_job_listings", "generate_job_search_plan", "research_market"},
	"pivot":     {"analyze_skill_transferability", "recommend_transition_roles", "research_market"},
	"salary":    {"get_salary_benchmark", "generate_negotiation_strategy", "research_market"},
	"other":     {"research_market"},
}
