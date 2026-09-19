package agent

import (
	"sync"
	"time"

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
	lastAccess         time.Time
}

// Agent is the unified agent with one shared in-memory conversation store.
type Agent struct {
	mu      sync.RWMutex
	threads map[string]*ThreadState
	ttl     time.Duration
}

// NewAgent returns an Agent with an in-memory session store. Idle sessions
// older than ttl are evicted by a background goroutine.
func NewAgent() *Agent {
	return NewAgentWithTTL(30 * time.Minute)
}

// NewAgentWithTTL lets callers (or tests) override the idle-session TTL.
func NewAgentWithTTL(ttl time.Duration) *Agent {
	a := &Agent{
		threads: map[string]*ThreadState{},
		ttl:     ttl,
	}
	go a.cleanupLoop()
	return a
}

func (a *Agent) getOrCreate(threadID string) *ThreadState {
	a.mu.Lock()
	defer a.mu.Unlock()
	st, ok := a.threads[threadID]
	if !ok {
		st = &ThreadState{}
		a.threads[threadID] = st
	}
	st.lastAccess = time.Now()
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
	st.lastAccess = time.Now()
}

// cleanupLoop periodically evicts sessions idle longer than the TTL.
func (a *Agent) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		a.evictIdle()
	}
}

func (a *Agent) evictIdle() {
	now := time.Now()
	a.mu.Lock()
	defer a.mu.Unlock()
	for id, st := range a.threads {
		if now.Sub(st.lastAccess) > a.ttl {
			delete(a.threads, id)
		}
	}
}
