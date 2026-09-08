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

type JDParser struct{}

func (p *JDParser) Parse(ctx context.Context, text string) (*models.Job, error) {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, errors.New("empty JD text")
	}

	job := &models.Job{}
	err := callJSON(
		ctx, llm.DeepSeek(), config.DeepSeekModel, 0,
		prompts.JDExtractionPrompt,
		text+"\n\n"+prompts.JDExtractionInstruction,
		job,
	)
	if err != nil {
		return nil, err
	}
	job.Description = text
	return job, nil
}
