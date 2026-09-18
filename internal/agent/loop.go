package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"

	"github.com/sashabaranov/go-openai"

	"offerhunter/internal/config"
	"offerhunter/internal/llm"
	"offerhunter/internal/models"
)

var errNoChoices = errors.New("empty LLM response")

func errUnknownTool(name string) error {
	return fmt.Errorf("unknown tool: %s", name)
}

// Sink receives streaming events from the agent loop; the handler translates
// them into SSE events for the frontend.
type Sink interface {
	ToolStart(name string, input map[string]string)
	ToolEnd(name string, output string)
	Report(report *models.MatchReport)
	Interview(questions []models.InterviewQuestion)
	Token(text string)
}

// Run executes one agent turn: append the user message, then loop the
// tool-calling cycle until the model produces a final answer (or errors).
func (a *Agent) Run(ctx context.Context, message, threadID, intent string, sink Sink) error {
	st := a.getOrCreate(threadID)
	st.messages = append(st.messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	allowed := intentToolNames[intent]
	if allowed == nil {
		allowed = intentToolNames["other"]
	}

	for {
		tools := filterTools(allowed)

		var toolNames []string
		for _, t := range tools {
			toolNames = append(toolNames, t.Function.Name)
		}
		log.Printf("[agent] sending tools (intent=%s): %v", intent, toolNames)

		resp, err := llm.DeepSeek().CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:       config.DeepSeekModel,
			Messages:    st.messages,
			Temperature: 0,
			MaxTokens:   4096,
			Tools:       tools,
			ToolChoice:  "auto",
		})
		if err != nil {
			return err
		}
		if len(resp.Choices) == 0 {
			return errNoChoices
		}

		msg := resp.Choices[0].Message

		// Final answer: no tool calls.
		if len(msg.ToolCalls) == 0 {
			st.messages = append(st.messages, msg)
			if sink != nil && msg.Content != "" {
				sink.Token(msg.Content)
			}
			return nil
		}

		// Assistant message carrying tool calls must be recorded verbatim.
		st.messages = append(st.messages, msg)

		for _, tc := range msg.ToolCalls {
			name := tc.Function.Name
			args := parseArgs(tc.Function.Arguments)

			log.Printf("[agent] tool call: %s args=%v", name, args)

			if sink != nil {
				sink.ToolStart(name, args)
			}

			result, execErr := a.execTool(ctx, name, args, st)
			if execErr != nil {
				log.Printf("[agent] tool %s error: %v", name, execErr)
				result = "工具执行失败: " + execErr.Error()
			}

			if sink != nil {
				sink.ToolEnd(name, result)
			}

			// Push report right after analyze_match, before the model talks.
			if name == "analyze_match" && st.report != nil && sink != nil {
				sink.Report(st.report)
			}

			// Push structured interview questions after generation.
			if name == "generate_interview_questions" && len(st.interviewQuestions) > 0 && sink != nil {
				sink.Interview(st.interviewQuestions)
			}

			st.messages = append(st.messages, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				Content:    result,
				ToolCallID: tc.ID,
			})
		}
		// Continue looping; the model may call more tools or finish.
	}
}

func filterTools(allowed []string) []openai.Tool {
	allowedSet := make(map[string]bool, len(allowed))
	for _, n := range allowed {
		allowedSet[n] = true
	}
	var tools []openai.Tool
	for _, t := range allTools {
		if allowedSet[t.Name] {
			tools = append(tools, openai.Tool{
				Type: openai.ToolTypeFunction,
				Function: &openai.FunctionDefinition{
					Name:        t.Name,
					Description: t.Description,
					Parameters:  t.Parameters,
				},
			})
		}
	}
	return tools
}

func (a *Agent) execTool(ctx context.Context, name string, args map[string]string, st *ThreadState) (string, error) {
	fn, ok := toolExec[name]
	if !ok {
		return "", errUnknownTool(name)
	}
	return fn(ctx, args, st)
}

func parseArgs(raw string) map[string]string {
	args := map[string]string{}
	if raw == "" {
		return args
	}
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return args
	}
	for k, v := range m {
		switch vv := v.(type) {
		case string:
			args[k] = vv
		default:
			b, _ := json.Marshal(vv)
			args[k] = string(b)
		}
	}
	return args
}
