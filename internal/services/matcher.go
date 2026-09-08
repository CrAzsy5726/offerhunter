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

type Matcher struct{}

func (m *Matcher) Analyze(ctx context.Context, resume *models.Resume, job *models.Job) (*models.MatchReport, error) {
	resumeJSON, _ := json.Marshal(resume)
	jobJSON, _ := json.Marshal(job)

	prompt := prompts.MatchingPrompt
	prompt = strings.ReplaceAll(prompt, "{resume_json}", string(resumeJSON))
	prompt = strings.ReplaceAll(prompt, "{job_json}", string(jobJSON))

	report := &models.MatchReport{}
	err := callJSON(
		ctx, llm.DeepSeek(), config.DeepSeekModel, 0,
		prompt, prompts.MatchingInstruction, report,
	)
	if err != nil {
		return nil, err
	}

	// Deterministic weighted scoring (same formula as Python version)
	c := &report.Competitiveness
	c.OverallPower = (c.ProjectDepth.Score + c.InternQuality.Score + c.TechStackFit.Score) / 3
	w := c.ProjectDepth.Score*25 + c.InternQuality.Score*20 + c.TechStackFit.Score*35 + report.HiddenFit.Score*20
	report.OverallScore = w / 100

	return report, nil
}
