# Offer 捕手 — AI 求职助手（Go 后端）

## 项目简介

一个基于大模型的 AI 求职助手，提供简历与 JD 匹配分析、模拟面试、求职策略、职业转型、薪资谈判五大能力。后端采用 **Go + Gin 自研 Agent 工具调用框架**（替代 LangChain），对接 DeepSeek / MiMo 大模型，通过 SSE 流式输出与前端交互，前端为单页应用（原生 JS，无框架）。

## 核心亮点

- **自研 LLM Agent 框架**：从零实现 tool-calling 循环（意图分类 → 工具过滤 → 并行工具调用 → 结果回填 → 流式输出），替代 LangChain `create_agent`，支持 11 个工具、6 类意图的自动路由
- **SSE 流式响应**：服务端通过 `http.Flusher` 逐事件推送（`intent`/`token`/`tool_start`/`tool_end`/`report`/`interview`/`done`），前端毫秒级渐进渲染，避免长时间白屏
- **多模态文件解析**：接入 MiMo 视觉大模型，支持 PDF/图片简历的 OCR 提取（前端 pdf.js 渲染 → base64 → 后端多模态调用）
- **结构化输出与容错**：LLM 返回 JSON 的清洗（剥离 markdown 代码块、括号配平提取），配合 Pydantic 等价的结构体校验，保证鲁棒性
- **会话记忆与上下文延续**：基于 `thread_id` 的并发安全内存会话（`sync.RWMutex`），跨轮记忆简历/JD/报告/面试进度

## 技术栈

`Go 1.26` · `Gin` · `go-openai` · `SSE` · `原生 JS/HTML/CSS` · `DeepSeek API` · `MiMo 多模态 API`

## 项目职责

- 主导后端技术栈从 Python（FastAPI + LangChain）迁移至 Go，功能与前端协议零改动，后端代码 3000+ 行、17 次迭代
- 设计并实现自研 tool-calling 循环，解决 DeepSeek 多工具并发调用不稳定的问题（按意图动态裁剪工具集至 ≤5 个）
- 设计报告页 / 面试页独立界面（淡入淡出、进度条、逐题评分），替代纯聊天气泡交互
- 实现意图识别的「确定性关键词预判 + LLM 兜底」双层策略，解决长文本场景下意图误判问题
