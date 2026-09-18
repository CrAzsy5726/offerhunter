package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/sashabaranov/go-openai"

	"offerhunter/internal/config"
	"offerhunter/internal/llm"
	"offerhunter/internal/models"
)

// InterviewGrader grades a candidate's answer to an interview question.
type InterviewGrader struct{}

const interviewGradePrompt = `你是一个资深技术面试官。请对以下面试题的回答进行评分。

## 题目
{question}

## 题目类别
{category}

## 题目难度
{difficulty}

## 评分要点
{scoring_points}

## 候选人回答
{answer}

## 输出要求
返回JSON格式（不要其他文字）：
{
  "score": 8,
  "comment": "点评（指出回答的优点和不足，语言客观、鼓励为主，80字以内）",
  "reference_answer": "参考答案要点（结合评分要点给出，150字以内）"
}

score 为 1-10 的整数。`

func (g *InterviewGrader) Grade(ctx context.Context, q models.InterviewQuestion, answer string) (*models.InterviewGradeResult, error) {
	scoringPoints := strings.Join(q.ScoringPoints, "、")
	if scoringPoints == "" {
		scoringPoints = "无"
	}

	prompt := interviewGradePrompt
	prompt = strings.ReplaceAll(prompt, "{question}", q.Question)
	prompt = strings.ReplaceAll(prompt, "{category}", q.Category)
	prompt = strings.ReplaceAll(prompt, "{difficulty}", q.Difficulty)
	prompt = strings.ReplaceAll(prompt, "{scoring_points}", scoringPoints)
	prompt = strings.ReplaceAll(prompt, "{answer}", answer)

	resp, err := llm.DeepSeek().CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: config.DeepSeekModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleUser, Content: prompt},
		},
		Temperature: 0.3,
		MaxTokens:   1000,
	})
	if err != nil {
		return nil, err
	}
	if len(resp.Choices) == 0 {
		return nil, errors.New("empty LLM response")
	}

	content := resp.Choices[0].Message.Content
	result := &models.InterviewGradeResult{}
	if json.Unmarshal([]byte(content), result) != nil {
		jsonStr := ExtractJSON(content)
		if jsonStr == "" {
			return nil, errors.New("failed to parse LLM response as JSON")
		}
		if err := json.Unmarshal([]byte(jsonStr), result); err != nil {
			return nil, err
		}
	}
	return result, nil
}
