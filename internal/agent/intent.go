package agent

import (
	"context"
	"log"
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

// keywordIntent is a deterministic pre-check. The LLM classifier tends to
// misread the request when it sits at the end of a long message (resume+JD
// pasted before "帮我分析"), so we check the message tail for obvious
// intent phrases first.
func keywordIntent(message string) (string, bool) {
	r := []rune(strings.ToLower(message))
	if len(r) > 120 {
		r = r[len(r)-120:]
	}
	tail := string(r)

	has := func(keys ...string) bool {
		for _, k := range keys {
			if strings.Contains(tail, k) {
				return true
			}
		}
		return false
	}

	switch {
	case has("转行", "转型", "换方向", "跨领域", "换岗", "转岗"):
		return "pivot", true
	case has("模拟面试", "面试练习", "面试题", "出题", "面试一下", "准备面试"):
		return "interview", true
	case has("薪资", "薪酬", "谈薪", "谈判", "报价", "offer", "工资"):
		return "salary", true
	case has("找工作", "求职计划", "求职策略", "求职规划", "求职"):
		return "strategy", true
	case has("分析", "匹配", "合适", "测评", "匹配度"):
		return "match", true
	}
	return "", false
}

// ClassifyIntent returns the intent key and its Chinese label.
func ClassifyIntent(ctx context.Context, message string) (string, string) {
	if intent, ok := keywordIntent(message); ok {
		log.Printf("[intent] keyword hit: %s", intent)
		return intent, intentLabels[intent]
	}

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
		log.Printf("[intent] LLM classify error: %v", err)
		return "other", intentLabels["other"]
	}
	if len(resp.Choices) == 0 {
		return "other", intentLabels["other"]
	}

	raw := strings.ToLower(strings.TrimSpace(resp.Choices[0].Message.Content))
	log.Printf("[intent] LLM raw=%q", raw)
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
