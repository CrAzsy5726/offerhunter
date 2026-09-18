package agent

import (
	"sync"

	"github.com/sashabaranov/go-openai"

	"offerhunter/internal/models"
)

// ThreadState holds per-conversation memory: chat history plus parsed
// artifacts (resume/job/report/interview questions) so tools can retrieve
// them without the LLM re-passing large text blobs between tool calls.
type ThreadState struct {
	messages           []openai.ChatCompletionMessage
	resumeText         string
	resume             *models.Resume
	jobText            string
	job                *models.Job
	report             *models.MatchReport
	interviewQuestions []models.InterviewQuestion
}

// Agent is the unified agent with one shared in-memory conversation store.
type Agent struct {
	mu      sync.RWMutex
	threads map[string]*ThreadState
}

func NewAgent() *Agent {
	return &Agent{threads: map[string]*ThreadState{}}
}

func (a *Agent) getOrCreate(threadID string) *ThreadState {
	a.mu.Lock()
	defer a.mu.Unlock()
	st, ok := a.threads[threadID]
	if !ok {
		st = &ThreadState{}
		a.threads[threadID] = st
	}
	return st
}
