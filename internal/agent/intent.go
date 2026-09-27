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
	"other":     "通用助手",
}

// keywordIntent is a deterministic pre-check. The LLM classifier tends to
// misread the request when it sits at the end of a long message (resume+JD
// pasted before "帮我分析"), so we check the message tail for obvious
// intent phrases first.
//
// prevIntent is the intent of the previous turn in the same conversation, used
// to continue an in-progress flow (e.g. interview) when the user follows up
// with a bare reply and no explicit instruction.
func keywordIntent(message, prevIntent string) (string, bool) {
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

	// 1. Explicit intent phrases win over everything (user stated a goal).
	switch {
	case has("模拟面试", "面试练习", "面试题", "出题", "面试一下", "准备面试"):
		return "interview", true
	case has("分析", "匹配", "合适", "测评", "匹配度"):
		return "match", true
	}

	// 2. No explicit instruction: a bare reply continues an in-progress
	// interview (the assistant asked for resume/JD, the user answered).
	if prevIntent == "interview" {
		return "interview", true
	}

	// 3. No explicit instruction and no ongoing flow: pasted resume+JD
	// material is a cold-start match.
	if hasResumeAndJD(message) {
		return "match", true
	}

	return "", false
}

// jdSignals are structural markers of a pasted job description. Loose words
// like "工程师"/"待遇" are deliberately excluded — they also appear in short
// replies such as "前端工程师，重庆，应届生".
var jdSignals = []string{
	"招聘", "岗位职责", "任职要求", "职位描述", "工作职责",
	"薪资范围", "薪资待遇", "岗位要求", "职位要求", "岗位描述",
	"工作内容", "加分项", "优先考虑", "五险一金", "福利待遇",
}

// resumeSignals are structural markers of a pasted resume. Loose words like
// "应届"/"学校" are deliberately excluded for the same reason as above.
var resumeSignals = []string{
	"教育背景", "教育经历", "项目经历", "项目经验", "实习经历",
	"实习经验", "工作经历", "自我评价", "个人简介", "专业技能",
	"个人技能", "毕业院校", "所学专业", "求职意向", "期望岗位",
}

// minMaterialRunes is the length below which a message cannot plausibly be
// pasted resume/JD material. Real materials are documents, not one-liners.
const minMaterialRunes = 80

// hasResumeAndJD reports whether the message is pasted resume+JD material:
// it must carry structural markers from both sides AND be long enough to
// actually be a document.
func hasResumeAndJD(msg string) bool {
	if len([]rune(msg)) < minMaterialRunes {
		return false
	}
	lower := strings.ToLower(msg)

	hasJD := false
	for _, s := range jdSignals {
		if strings.Contains(lower, s) {
			hasJD = true
			break
		}
	}
	if !hasJD {
		return false
	}

	for _, s := range resumeSignals {
		if strings.Contains(lower, s) {
			return true
		}
	}
	return false
}

// ClassifyIntent returns the intent key and its Chinese label.
// prevIntent is the previous turn's intent, used to continue in-progress flows.
func ClassifyIntent(ctx context.Context, message, prevIntent string) (string, string) {
	preview := message
	if len([]rune(preview)) > 120 {
		preview = string([]rune(preview)[:120])
	}
	log.Printf("[intent] msg preview: %q, prevIntent: %q", preview, prevIntent)

	if intent, ok := keywordIntent(message, prevIntent); ok {
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
// "other" is intentionally empty: chitchat is plain chat, no tools.
var intentToolNames = map[string][]string{
	"match":     {"parse_resume", "parse_jd", "analyze_match"},
	"interview": {"parse_resume", "parse_jd", "generate_interview_questions"},
	"other":     {},
}

// unsupportedTopics are subjects the product deliberately does not cover.
// A prompt-level instruction is not enough — the LLM tends to answer anyway
// (and burn tokens fabricating data), so we intercept deterministically.
var unsupportedTopics = []string{
	"薪资", "工资", "薪酬", "年薪", "月薪", "待遇",
	"转行", "转型", "职业规划", "职业发展",
	"求职策略", "求职计划", "市场行情", "行业趋势", "就业前景", "行业前景",
}

const unsupportedReply = "抱歉，这个不在我的能力范围内。我可以帮你做**简历和 JD 的匹配分析**，" +
	"或者用你的简历做一次**模拟面试** —— 想先试试哪个？"

// UnsupportedTopicReply returns a canned reply when the message asks about a
// subject the product does not support, so no LLM call is made. It returns
// "" when the message is fine to handle normally.
func UnsupportedTopicReply(message string) string {
	lower := strings.ToLower(message)
	for _, t := range unsupportedTopics {
		if strings.Contains(lower, t) {
			return unsupportedReply
		}
	}
	return ""
}
