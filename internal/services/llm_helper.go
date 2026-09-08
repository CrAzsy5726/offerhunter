package services

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/sashabaranov/go-openai"
)

// callJSON invokes the LLM and unmarshals the response content into target.
// It first tries a direct unmarshal, then falls back to ExtractJSON to strip
// code fences / surrounding prose before unmarshalling.
func callJSON(
	ctx context.Context,
	cli *openai.Client,
	model string,
	temperature float32,
	system, user string,
	target any,
) error {
	resp, err := cli.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:       model,
		Messages:    []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: system},
			{Role: openai.ChatMessageRoleUser, Content: user},
		},
		Temperature: temperature,
		MaxTokens:   4096,
	})
	if err != nil {
		return err
	}
	if len(resp.Choices) == 0 {
		return fmt.Errorf("empty LLM response")
	}
	content := resp.Choices[0].Message.Content

	if json.Unmarshal([]byte(content), target) == nil {
		return nil
	}
	jsonStr := ExtractJSON(content)
	if jsonStr == "" {
		return fmt.Errorf("failed to parse LLM response as JSON")
	}
	return json.Unmarshal([]byte(jsonStr), target)
}
