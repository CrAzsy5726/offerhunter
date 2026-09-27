package handlers

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"offerhunter/internal/agent"
	"offerhunter/internal/models"
)

var globalAgent = agent.NewAgent()

type chatRequest struct {
	Message  string `json:"message"`
	ThreadID string `json:"thread_id"`
	Mode     string `json:"mode"`
}

// Chat handles POST /api/chat with SSE streaming.
func Chat(c *gin.Context) {
	var req chatRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"detail": err.Error()})
		return
	}

	threadID := req.ThreadID
	if threadID == "" {
		threadID = newThreadID()
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Thread-ID", threadID)

	ctx := c.Request.Context()
	sink := &sseSink{w: c.Writer}

	prevIntent := globalAgent.LastIntent(threadID)
	intent, intentName := agent.ClassifyIntent(ctx, req.Message, prevIntent)
	globalAgent.RecordIntent(threadID, intent)
	writeSSE(c.Writer, "intent", map[string]string{"intent": intent, "name": intentName})

	// Unsupported topics get a canned reply — no LLM call, no token spend.
	if intent == "other" {
		if reply := agent.UnsupportedTopicReply(req.Message); reply != "" {
			writeSSE(c.Writer, "token", map[string]string{"content": reply})
			writeSSE(c.Writer, "done", map[string]string{"thread_id": threadID})
			return
		}
	}

	if err := globalAgent.Run(ctx, req.Message, threadID, intent, sink); err != nil {
		writeSSE(c.Writer, "error", map[string]string{"message": err.Error()})
	}

	writeSSE(c.Writer, "done", map[string]string{"thread_id": threadID})
}

type sseSink struct {
	w gin.ResponseWriter
}

func (s *sseSink) ToolStart(name string, input map[string]string) {
	writeSSE(s.w, "tool_start", map[string]any{"name": name, "input": input})
}

func (s *sseSink) ToolEnd(name string, output string) {
	preview := output
	if len([]rune(preview)) > 300 {
		preview = string([]rune(preview)[:300])
	}
	writeSSE(s.w, "tool_end", map[string]any{"name": name, "output_preview": preview})
}

func (s *sseSink) Report(report *models.MatchReport) {
	writeSSE(s.w, "report", map[string]any{"report": report})
}

func (s *sseSink) Interview(questions []models.InterviewQuestion) {
	writeSSE(s.w, "interview", map[string]any{"questions": questions})
}

func (s *sseSink) Token(text string) {
	writeSSE(s.w, "token", map[string]any{"content": text})
}

func writeSSE(w gin.ResponseWriter, event string, data any) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(data)
	jsonStr := strings.TrimSpace(buf.String())
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, jsonStr)
	w.Flush()
}

func newThreadID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
