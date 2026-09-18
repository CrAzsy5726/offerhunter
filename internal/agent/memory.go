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
	lastIntent         string
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

// LastIntent returns the intent of the previous turn in this thread, or ""
// for a fresh conversation.
func (a *Agent) LastIntent(threadID string) string {
	a.mu.RLock()
	defer a.mu.RUnlock()
	if st, ok := a.threads[threadID]; ok {
		return st.lastIntent
	}
	return ""
}

// RecordIntent persists the current turn's intent for context on the next turn.
func (a *Agent) RecordIntent(threadID, intent string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	st, ok := a.threads[threadID]
	if !ok {
		st = &ThreadState{}
		a.threads[threadID] = st
	}
	st.lastIntent = intent
}
