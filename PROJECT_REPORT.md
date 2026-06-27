# Offer-Hunter LangChain Edition — 项目报告

**日期**: 2026-06-17  
**目录**: `C:\Users\63151\Desktop\v2\Offer-Hunter-langchain`  
**技术栈**: FastAPI + LangChain (create_agent) + DeepSeek (chat + function calling) + 单页 HTML 前端

---

## 1. 项目结构

```
Offer-Hunter-langchain/
├── main.py                          # FastAPI 入口，11 个路由
├── models.py                        # Pydantic 数据模型（Resume/Job/MatchReport/ChatRequest）
├── config.py                        # .env 加载
├── requirements.txt                 # 7 个依赖
├── .env                             # API Keys（DeepSeek + MiMo）
│
├── agent/
│   ├── agent_factory.py             # create_agent(4 tools + InMemorySaver)
│   └── middleware.py                # LoggingTimingMiddleware
│
├── handlers/                        # 8 个 FastAPI 路由处理器
│   ├── chat_handler.py              # POST /api/chat — SSE 流式 Agent 对话 ★
│   ├── resume_handler.py            # /api/resume/parse, /api/resume/upload
│   ├── job_handler.py              # /api/job/parse
│   ├── match_handler.py            # /api/match — 传统一键匹配（保留兼容）
│   ├── market_handler.py           # /api/market
│   ├── intent_handler.py           # /api/intent/check
│   └── export_handler.py           # /api/export — Markdown 报告导出
│
├── services/
│   ├── agent_tools.py               # 4 个 @tool ★ 当前 Agent 在用的唯一工具集
│   ├── agent_tools_advanced.py      # 11 个 @tool ★ 已编写但未接入 Agent
│   ├── resume_parser.py            # 简历解析（DeepSeek + MiMo 多模态）
│   ├── jd_parser.py                # JD 解析
│   ├── matcher.py                   # 三层次匹配引擎（含确定性评分公式）
│   ├── optimizer.py                # 简历优化建议
│   ├── gap_analyzer.py             # 技能差距分析
│   ├── market_research.py          # DuckDuckGo + LLM 市场调研
│   ├── intent_validator.py         # 意图检测
│   ├── extract_json.py             # JSON 提取工具
│   └── llm_factory.py              # ChatOpenAI 工厂
│
├── prompts/
│   ├── agent_system.py              # 当前 Agent 系统提示词 ★
│   ├── agent_modes.py               # 5 种模式提示词 ★ 未接入
│   ├── resume_extraction.py
│   ├── jd_extraction.py
│   ├── matching.py
│   ├── optimization.py
│   ├── gap_analysis.py
│   └── market_research.py
│
└── static/
    └── index.html                   # 单页前端（~1000 行）
```

## 2. API 端点（11 个）

| # | 方法 | 路径 | 用途 | 状态 |
|---|------|------|------|------|
| 1 | GET | `/` | 前端页面 | ✓ |
| 2 | GET | `/health` | 健康检查 | ✓ |
| 3 | POST | `/api/chat` | **SSE 流式 Agent 对话** | ✓ 核心 |
| 4 | POST | `/api/resume/parse` | 解析简历文本 | ✓ |
| 5 | POST | `/api/resume/upload` | 上传简历（文本/图片 base64） | ✓ |
| 6 | POST | `/api/job/parse` | 解析 JD | ✓ |
| 7 | POST | `/api/match` | 传统一键匹配 | ✓ 保留兼容 |
| 8 | POST | `/api/market` | 市场调研 | ✓ |
| 9 | POST | `/api/intent/check` | 意图检测 | ✓ |
| 10 | POST | `/api/export` | Markdown 报告导出 | ✓ |
| 11 | GET | `/static/...` | 静态文件 | ✓ |

## 3. Agent 架构（两轮调用）

```
用户消息
  │
  ├─ Round 1: classify_intent(msg) → LLM 分类 (max_tokens=10, ~0.5s)
  │           → "match" | "interview" | "strategy" | "pivot" | "salary" | "other"
  │
  └─ Round 2: get_agent(intent) → Agent(4 core + mode_tools, ≤6 tools total)
              正常流式执行，~5-15s
```

### Intent → Tool 映射

| Intent | 模式提示词 | 额外工具 | 总 tools |
|--------|----------|---------|---------|
| match | agent_system.py | — | 4 |
| interview | agent_modes.py INTERVIEW | generate_interview_questions | 5 |
| strategy | agent_modes.py STRATEGIST | search_job_listings, generate_job_search_plan | 6 |
| pivot | agent_modes.py PIVOT | analyze_skill_transferability, recommend_transition_roles | 6 |
| salary | agent_modes.py NEGOTIATION | get_salary_benchmark, generate_negotiation_strategy | 6 |
| other | agent_modes.py STARTER | — | 4 |

**意图分类准确率**: 12/12 (100%)

### SSE 事件（7 种）

| 事件 | 说明 |
|------|------|
| `intent` | Round 1 完成，告知前端当前模式 |
| `token` | 流式文本 token（tool 执行期间自动抑制） |
| `tool_start` | Tool 开始执行 |
| `tool_end` | Tool 执行完毕 |
| `report` | 匹配分析完成后推送完整 MatchReport |
| `done` | 本轮结束 |
| `error` | 出错 |

## 4. 当前 Agent 实际使用的 Tool（11 个，按意图动态加载 4~6 个）

### 核心 Tool（所有模式共用，4 个）

| Tool | 功能 | 内部调用链 |
|------|------|-----------|
| `parse_resume(text)` | 解析简历 | ResumeParser → DeepSeek |
| `parse_jd(text)` | 解析 JD | JDParser → DeepSeek |
| `analyze_match(resume, jd)` | 完整匹配分析 | ResumeParser → JDParser → Matcher(评分公式) → Optimizer → GapAnalyzer |
| `research_market(job, preference)` | 市场调研 | DuckDuckGo 搜索 + DeepSeek 合成 |

### 按需加载的高级 Tool（7 个）

| Tool | 模式 | 功能 |
|------|------|------|
| `generate_interview_questions` | interview | LLM 生成 8-10 道面试题 |
| `search_job_listings` | strategy | DuckDuckGo 搜索真实岗位 |
| `generate_job_search_plan` | strategy | LLM 生成分阶段求职计划 |
| `analyze_skill_transferability` | pivot | LLM 分析技能可迁移性 |
| `recommend_transition_roles` | pivot | LLM 推荐过渡岗位 |
| `get_salary_benchmark` | salary | DuckDuckGo + LLM 薪资基准查询 |
| `generate_negotiation_strategy` | salary | LLM 生成谈判策略和话术 |

**匹配评分公式（确定性）**:
```
overall_power = (project_depth + intern_quality + tech_stack_fit) / 3
overall_score = (project*25 + intern*20 + tech*35 + hidden_fit*20) / 100
```

## 5. 前端功能

- 纯对话 UI（暗色主题、玻璃态卡片、动画背景）
- 文件上传（PDF/图片/TXT，拖拽/粘贴/按钮）→ 客户端 pdf.js 渲染 → 后端 MiMo OCR
- SSE 流式显示 token（工具执行期间自动抑制 JSON 泄漏）
- 匹配完成后自动切到**报告页面**（总分、门槛、竞争力、隐性契合、优化建议、差距分析、市场参考）
- 报告页底部 **"继续对话"** 按钮 → 回到对话使用更多功能
- 导出 Markdown / 复制报告
- `thread_id` localStorage 持久化

## 6. 当前状态：全部功能可用

### ✅ 全部可用的场景
- 粘贴简历+JD → Agent 自动调 parse_resume + parse_jd + analyze_match → 报告页
- 问市场薪资 → Agent 调 research_market
- **模拟面试** → Round 1 分类 interview → Agent 调 generate_interview_questions → 逐题提问评分
- **求职策略** → Round 1 分类 strategy → Agent 调 search_job_listings + 生成计划
- **职业转型** → Round 1 分类 pivot → Agent 调 analyze_skill_transferability + recommend
- **薪资谈判** → Round 1 分类 salary → Agent 调 salary_benchmark + strategy
- 上传 PDF/图片简历（MiMo OCR）
- 多轮对话（同一 thread_id 内保留上下文）
- 报告页底部"继续对话"回到对话

## 7. 已知问题

| # | 严重度 | 问题 | 位置 |
|---|--------|------|------|
| 1 | 中 | 11 个高级 tool 未接入 Agent（DeepSeek tool 数量限制） | agent_factory.py, agent_tools_advanced.py |
| 2 | 中 | `prompts/agent_modes.py` 整文件未导入 | 全项目 |
| 3 | 低 | `ChatRequest.mode` 字段被 chat_handler 忽略 | models.py, chat_handler.py |
| 4 | 低 | `export_handler.py` 短期/长期标签互换 | line ~65 |
| 5 | 低 | `_thread_cache` 无限增长（无 TTL 淘汰） | agent_tools.py |
| 6 | 低 | 多处重复创建 ChatOpenAI 实例（handler 层 + service 层 + agent 层） | 多处 |
| 7 | 低 | 前端 `renderReport(resumeText, jobText)` 参数未使用 | index.html |
| 8 | 低 | 前端 `setCollectedInfo` / `resetCollectedInfo` 空函数残留 | index.html |

## 8. 文件统计

| 指标 | 数量 |
|------|------|
| Python 文件 | 35 |
| 已注册 API 端点 | 11 |
| 已定义 Tool（合计） | 15 |
| Agent 实际使用 Tool | 4 |
| 未接入 Tool | 11 |
| 前端代码行数 | ~1000 |