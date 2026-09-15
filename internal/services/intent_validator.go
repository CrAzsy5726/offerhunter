package services

import (
	"context"
	"regexp"
	"strings"

	"github.com/sashabaranov/go-openai"

	"offerhunter/internal/config"
	"offerhunter/internal/llm"
)

// IntentValidator classifies whether user input is job-related, mirroring
// services/intent_validator.py.
type IntentValidator struct{}

var offTopicPatterns = []*regexp.Regexp{
	regexp.MustCompile(`^.{0,3}$`),
	regexp.MustCompile(`^(你好|hi|hello|hey|喂|在吗)$`),
	regexp.MustCompile(`^今天天气`),
	regexp.MustCompile(`^你是[谁什么]`),
	regexp.MustCompile(`^[多大了几岁|]`),
	regexp.MustCompile(`^讲个(笑话|故事)`),
	regexp.MustCompile(`^(帮我)?写(一篇|一段)作文`),
}

const intentValidatorPrompt = `你是一个求职助手。判断用户的输入是否与求职相关（如简历、职位描述、求职意向、技能询问等）。只回复 "yes" 或 "no"。`

// IsJobRelated returns whether the text is job-related and a reason string
// (empty when related).
func (v *IntentValidator) IsJobRelated(ctx context.Context, text string) (bool, string) {
	text = strings.TrimSpace(text)
	lower := strings.ToLower(text)

	if isClearlyOffTopic(lower) {
		return false, "您的输入似乎与求职无关。请告诉我您想找什么岗位，或粘贴简历和JD。"
	}

	resp, err := llm.DeepSeek().CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: config.DeepSeekModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: intentValidatorPrompt},
			{Role: openai.ChatMessageRoleUser, Content: text},
		},
		Temperature: 0,
		MaxTokens:   10,
	})
	if err != nil {
		return true, ""
	}
	if len(resp.Choices) == 0 {
		return true, ""
	}
	if strings.HasPrefix(strings.ToLower(strings.TrimSpace(resp.Choices[0].Message.Content)), "yes") {
		return true, ""
	}
	return false, "您的输入似乎与求职无关。我可以帮您分析简历、比较岗位匹配度、提供优化建议。请描述您的求职需求吧。"
}

func isClearlyOffTopic(text string) bool {
	for _, re := range offTopicPatterns {
		if re.MatchString(text) {
			return true
		}
	}
	return false
}
