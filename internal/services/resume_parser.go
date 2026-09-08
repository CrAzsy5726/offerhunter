package services

import (
	"context"
	"errors"
	"strings"

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
