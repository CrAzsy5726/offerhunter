package services

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"github.com/sashabaranov/go-openai"

	"offerhunter/internal/config"
	"offerhunter/internal/llm"
	"offerhunter/internal/models"
	"offerhunter/internal/prompts"
)

type ResumeParser struct{}

func (p *ResumeParser) Parse(ctx context.Context, text string) (*models.Resume, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("empty resume text")
	}

	resume := &models.Resume{}
	err := callJSON(
		ctx, llm.DeepSeek(), config.DeepSeekModel, 0,
		prompts.ResumeExtractionPrompt,
		text+"\n\n"+prompts.ResumeExtractionInstruction,
		resume,
	)
	if err != nil {
		return nil, err
	}
	resume.RawText = text
	return resume, nil
}

// ParseFromImages OCRs resume images using the MiMo multimodal LLM and
// returns the extracted plain text.
func (p *ResumeParser) ParseFromImages(ctx context.Context, images []string) (string, error) {
	if len(images) == 0 {
		return "", errors.New("no images provided")
	}

	system := prompts.ResumeOCRPrompt
	userText := "这是一份简历的" + strconv.Itoa(len(images)) + "页截图。请识别并提取所有文字信息，包括个人信息、教育背景、技能、项目经历、实习经历等。"

	parts := []openai.ChatMessagePart{
		{Type: openai.ChatMessagePartTypeText, Text: userText},
	}
	for _, img := range images {
		url := img
		if !strings.HasPrefix(img, "data:") {
			url = "data:image/jpeg;base64," + img
		}
		parts = append(parts, openai.ChatMessagePart{
			Type:     openai.ChatMessagePartTypeImageURL,
			ImageURL: &openai.ChatMessageImageURL{URL: url},
		})
	}

	resp, err := llm.MiMo().CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: config.MiMoModel,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: system},
			{Role: openai.ChatMessageRoleUser, MultiContent: parts},
		},
		Temperature: 0,
		MaxTokens:   8192,
	})
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", errors.New("empty MiMo response")
	}
	return resp.Choices[0].Message.Content, nil
}
