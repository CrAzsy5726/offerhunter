# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Important: the backend is Go

The active backend is **Go**, under `cmd/` and `internal/`. The previous **Python/LangChain** implementation is archived in `legacy/` for reference only — it is **not the entry point** and is not used. All new work goes in Go. The frontend (`static/index.html`) is shared.

## Commands

```bash
# Build (binary written to ./offer-hunter)
go build -o offer-hunter ./cmd/server

# Run locally (reads .env from the working directory; see config section)
go run ./cmd/server

# Test everything / single package / single test
go test ./...
go test ./internal/agent
go test ./internal/agent -run TestKeywordIntent

# Vet
go vet ./...

# Cross-compile for Linux (used by deploy.sh)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o offer-hunter-linux ./cmd/server
```

`deploy.sh` / `undeploy.sh` handle production (systemd `offer-hunter` service on `/opt/offer-hunter`, port 8080, Nginx reverse proxy). They must be run as root on the Linux server, not locally.

## Configuration

There is **no config library**. `internal/config/config.go` hand-parses a `.env` file in the working directory into `envMap`, and process environment variables take precedence over `.env` values at lookup time. Keys: `DEEPSEEK_API_KEY` / `DEEPSEEK_BASE_URL` / `DEEPSEEK_MODEL` (text LLM), `MIMO_API_KEY` / `MIMO_BASE_URL` / `MIMO_MODEL` (multimodal OCR LLM), `SERVER_HOST` / `SERVER_PORT`. Copy `.env.example` to `.env` to run. `cmd/server/main.go` fatals if `DEEPSEEK_API_KEY` is missing and only warns if `MIMO_API_KEY` is missing (MiMo powers PDF/image resume OCR).

## Architecture

Package layout (all under `internal/`, Go module `offerhunter`):

- `internal/server/router.go` — Gin router; all routes under `/api`, static frontend served from `./static`.
- `internal/handlers/` — HTTP handlers. `chat.go` is the core SSE streaming entry point.
- `internal/agent/` — the **self-built tool-calling loop** that replaces LangChain `create_agent`. This is the heart of the app.
- `internal/services/` — business logic (resume/JD parsers, matcher, optimizer, gap analyzer, market research, interview grader).
- `internal/llm/` — singleton OpenAI-compatible clients (`DeepSeek()` for text, `MiMo()` for multimodal) built lazily with `sync.Once`.
- `internal/models/` — plain structs with JSON tags (Resume, Job, MatchReport, InterviewQuestion, etc.).
- `internal/prompts/` — all LLM prompt strings as `const`.

### The agent loop (`internal/agent/loop.go`)

`Agent.Run` implements one turn: append the user message, then loop calling DeepSeek with the intent-filtered tool set until the model returns a final answer (no tool calls). Tool calls are recorded verbatim in the message history (assistant + `role:"tool"` messages). The `Sink` interface (`ToolStart`, `ToolEnd`, `Report`, `Interview`, `Token`) lets the handler translate loop events into SSE events. `analyze_match` and `generate_interview_questions` push structured payloads (report / interview questions) to the sink **immediately** after execution, before the model continues talking.

Tools are registered as a flat `allTools` slice + `toolExec` map. The 4 core tools are in `tools.go`, the 7 advanced tools in `advanced_tools.go` (appended via `init()`). Each tool has a JSON-Schema `Parameters` and a Chinese `Description` docstring — the model relies on the description to decide when to call.

### Intent classification (`internal/agent/intent.go`)

Two-layer strategy:

1. **Deterministic keyword pre-check** (`keywordIntent`) — checks the message **tail** (last ~120 runes) for explicit intent phrases, plus `hasResumeAndJD` to detect a bare resume+JD paste. The tail check exists because the LLM classifier misreads intent when a request sits at the end of a long resume+JD message. `prevIntent` continues in-progress flows (e.g. a bare resume+JD during an interview stays `interview`).
2. **LLM fallback** — if no keyword hit, classify with a 1-token (`MaxTokens: 10`, temperature 0) LLM call.

The six intents and their allowed tools (each **≤ 5 tools** — a DeepSeek function-calling stability constraint; keep it that way when adding tools):

| intent | tools |
|--------|-------|
| `match` | parse_resume, parse_jd, analyze_match, research_market |
| `interview` | parse_resume, parse_jd, generate_interview_questions, research_market |
| `strategy` | search_job_listings, generate_job_search_plan, research_market |
| `pivot` | analyze_skill_transferability, recommend_transition_roles, research_market |
| `salary` | get_salary_benchmark, generate_negotiation_strategy, research_market |
| `other` | research_market |

### Session memory (`internal/agent/memory.go`)

`Agent` holds an in-memory `map[string]*ThreadState` guarded by `sync.RWMutex`. Each `ThreadState` stores the chat history plus parsed artifacts (`resumeText`/`resume`/`jobText`/`job`/`report`/`interviewQuestions`) so tools can retrieve them without the LLM re-passing large text blobs. Sessions idle longer than a 30-minute TTL are evicted by a background goroutine. `LastIntent` / `RecordIntent` persist each turn's intent for the next turn.

### Request lifecycle

```
POST /api/chat {message, thread_id}
  → ClassifyIntent(message, prevIntent)      // keyword pre-check → LLM fallback
  → RecordIntent(threadID, intent)
  → SSE "intent" event
  → Agent.Run(message, threadID, intent, sink)  // tool-calling loop
  → SSE "done" event
```

SSE events consumed by the frontend: `intent`, `token`, `tool_start`, `tool_end`, `report`, `interview`, `done`, `error`. `thread_id` is generated server-side when absent and returned via both the `X-Thread-ID` header and the `done` event.

### LLM conventions

- Agent loop and intent classification: `Temperature: 0`.
- Tool-internal LLM calls (`toolLLMCall` in `advanced_tools.go` and `callJSON` in `services/llm_helper.go`): `Temperature: 0.3`.
- LLM JSON output is parsed defensively — `services/extract_json.go` strips markdown fences and extracts the outermost `{...}` object; `callJSON` tries direct unmarshal first, then falls back to `ExtractJSON`.
- The match score is computed **deterministically** in `services/matcher.go` after the LLM returns component scores: `overall_power = (project_depth + intern_quality + tech_stack_fit) / 3` and `overall_score = (project*25 + intern*20 + tech*35 + hidden_fit*20) / 100`.

### Frontend

`static/index.html` is a single-page app (vanilla JS, no framework) that renders the chat UI, the match report page, and the interview page. It parses the SSE stream above and keeps `thread_id` in `localStorage`. Warm/dark themes are handled by dynamically injecting/removing a `<style id="warm-theme">` overlay rather than editing the dark-theme CSS.

## Adding a feature (intent + tool)

1. Add the tool to `internal/agent/tools.go` or `advanced_tools.go` (`Tool` def + `ToolFunc` in `toolExec`/`advancedToolExec`); write a clear Chinese description since the model uses it to decide invocation.
2. Add the tool name to the right slice in `intentToolNames` (`intent.go`), keeping each intent ≤ 5 tools.
3. If a new intent, also extend `intentLabels`, `keywordIntent`, and `prompts.IntentClassifyPrompt`.
4. If it produces a structured report/interview payload the frontend needs, push it through the `Sink` in `loop.go` and handle the SSE event in `static/index.html`.
