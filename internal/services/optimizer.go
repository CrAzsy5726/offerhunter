package services

import (
	"context"
	"encoding/json"
	"strings"

	"offerhunter/internal/config"
	"offerhunter/internal/llm"
	"offerhunter/internal/models"
	"offerhunter/internal/prompts"
)

type Optimizer struct{}

func (o *Optimizer) GenerateTips(ctx context.Context, resume *models.Resume, job *models.Job, report *models.MatchReport) ([]models.OptimizationTip, error) {
	resumeJSON, _ := json.Marshal(resume)
	jobJSON, _ := json.Marshal(job)
	matchJSON, _ := json.Marshal(report)

	prompt := prompts.OptimizationPrompt
	prompt = strings.ReplaceAll(prompt, "{resume_json}", string(resumeJSON))
	prompt = strings.ReplaceAll(prompt, "{job_json}", string(jobJSON))
	prompt = strings.ReplaceAll(prompt, "{match_json}", string(matchJSON))

	var wrapper struct {
		OptimizationTips []models.OptimizationTip `json:"optimization_tips"`
	}
	err := callJSON(
		ctx, llm.DeepSeek(), config.DeepSeekModel, 0.3,
		prompt, prompts.OptimizationInstruction, &wrapper,
	)
	if err != nil {
		return nil, err
	}
	return wrapper.OptimizationTips, nil
}
