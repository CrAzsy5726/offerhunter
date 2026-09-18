package agent

import (
	"context"
	"errors"

	"offerhunter/internal/models"
)

// GenerateInterview generates interview questions grounded in the given
// resume/JD, stores them (and the source materials) on the thread state for
// later conversation context, and records the intent for the next turn.
func (a *Agent) GenerateInterview(ctx context.Context, threadID, resumeText, jdText string) ([]models.InterviewQuestion, error) {
	if resumeText == "" && jdText == "" {
		return nil, errNoInterviewMaterial
	}

	st := a.getOrCreate(threadID)
	if resumeText != "" {
		st.resumeText = resumeText
	}
	if jdText != "" {
		st.jobText = jdText
	}

	questions, _, err := generateQuestions(ctx, jdText, resumeText, 8)
	if err != nil {
		return nil, err
	}
	st.interviewQuestions = questions

	a.RecordIntent(threadID, "interview")
	return questions, nil
}

var errNoInterviewMaterial = errors.New("请至少提供简历或JD中的一项")
