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

type GapAnalyzer struct{}

func (g *GapAnalyzer) Analyze(ctx context.Context, resume *models.Resume, job *models.Job) ([]models.GapItem, error) {
	resumeJSON, _ := json.Marshal(resume)
	jobJSON, _ := json.Marshal(job)

	prompt := prompts.GapAnalysisPrompt
	prompt = strings.ReplaceAll(prompt, "{resume_json}", string(resumeJSON))
	prompt = strings.ReplaceAll(prompt, "{job_json}", string(jobJSON))

	var wrapper struct {
		GapAnalysis []models.GapItem `json:"gap_analysis"`
	}
	err := callJSON(
		ctx, llm.DeepSeek(), config.DeepSeekModel, 0.3,
		prompt, prompts.GapAnalysisInstruction, &wrapper,
	)
	if err != nil {
		return nil, err
	}
	return wrapper.GapAnalysis, nil
}
