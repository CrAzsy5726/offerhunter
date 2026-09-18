# Offer 捕手 LangChain Edition — 技术文档

> 给二次开发者看的。包含架构、关键代码路径、如何添加新功能、常见问题排查。

---

## 1. 目录结构

```
Offer-Hunter-langchain/
├── main.py                     # FastAPI 入口，11 个路由
├── models.py                   # Pydantic 模型（Resume/Job/MatchReport/ChatRequest）
├── config.py                   # .env 加载
├── requirements.txt
├── .gitignore
│
├── agent/
│   ├── agent_factory.py        # 意图分类 + Agent 工厂 + 中间件 ★
│   └── middleware.py            # LoggingTimingMiddleware
│
├── handlers/                   # FastAPI 路由处理器
│   ├── chat_handler.py         # POST /api/chat — SSE 流式 ★ 核心入口
│   ├── resume_handler.py       # /api/resume/parse, /api/resume/upload
│   ├── job_handler.py         # /api/job/parse
│   ├── match_handler.py       # /api/match（传统一键匹配，保留）
│   ├── market_handler.py      # /api/market
│   ├── intent_handler.py      # /api/intent/check
│   └── export_handler.py      # /api/export
│
├── services/                   # 业务逻辑
│   ├── agent_tools.py          # 4 个核心 @tool ★
│   ├── agent_tools_advanced.py # 7 个高级 @tool ★
│   ├── resume_parser.py       # DeepSeek + MiMo
│   ├── jd_parser.py
│   ├── matcher.py              # 三层匹配 + 确定性评分公式
│   ├── optimizer.py
│   ├── gap_analyzer.py
│   ├── market_research.py     # DuckDuckGo + LLM
│   ├── intent_validator.py
│   ├── extract_json.py        # JSON 提取工具
│   └── llm_factory.py         # ChatOpenAI 工厂
│
├── prompts/                    # 所有 LLM 提示词
│   ├── agent_system.py        # 默认系统提示词
│   ├── agent_modes.py         # 5 种模式专用提示词
│   ├── resume_extraction.py
│   ├── jd_extraction.py
│   ├── matching.py
│   ├── optimization.py
│   ├── gap_analysis.py
│   └── market_research.py
│
└── static/
    └── index.html             # 单页前端（~1200 行）
```

---

## 2. 请求生命周期

```
用户输入 "帮我分析匹配度"
  │
  ▼
POST /api/chat {message, thread_id}
  │
  ├─ Round 1: classify_intent(msg) ── LLM 单 token 输出 "match" (0.5s)
  │            yield SSE: event=intent
  │
  ├─ set_intent("match") ── 存入 contextvars，供中间件读取
  │
  ├─ Round 2: get_agent() ── 返回单例 Agent (11 个 tool 全部注册)
  │     agent.invoke({"messages": [HumanMessage(msg)]}, config={"thread_id": tid})
  │       │
  │       ├─ intent_tool_filter 中间件 ── 根据 intent 过滤 tool
  │       │    "match" → [parse_resume, parse_jd, analyze_match, research_market]
  │       │
  │       ├─ LoggingTimingMiddleware ── 记录每次 LLM + tool 耗时
  │       │
  │       └─ InMemorySaver ── 多轮对话记忆
  │
  └─ SSE 流式: token → tool_start → tool_end → report → done
```

## 3. 关键文件路径

### 意图分类 (agent/agent_factory.py)

`classify_intent(msg)` — 用 1 token 的 LLM 调用判断用户意图。

意图映射：
| 关键词 | 意图 | 加载的工具 |
|--------|------|-----------|
| 分析/简历/JD | match | parse_resume, parse_jd, analyze_match, research_market |
| 面试/模拟 | interview | generate_interview_questions, research_market |
| 找/求职/工作 | strategy | search_job_listings, generate_job_search_plan, research_market |
| 转行/转型 | pivot | analyze_skill_transferability, recommend_transition_roles, research_market |
| 薪资/谈判 | salary | get_salary_benchmark, generate_negotiation_strategy, research_market |
| 其他 | other | research_market |

如果要添加新意图：修改 `INTENT_CLASSIFY_PROMPT`、`INTENT_LABELS`、`INTENT_TOOL_NAMES`。

### 工具定义 (services/agent_tools.py + agent_tools_advanced.py)

核心 4 工具：
```python
@tool
async def parse_resume(text: str, config) -> str:   # 简历解析
@tool
async def parse_jd(text: str, config) -> str:        # JD 解析
@tool
async def analyze_match(resume_text, job_text, config) -> str:  # 匹配分析
@tool
async def research_market(job_info, job_preference) -> str:     # 市场调研
```

高级 7 工具（按意图加载）：
```python
generate_interview_questions   # 生成 8-10 道面试题
search_job_listings            # DuckDuckGo 搜索岗位
generate_job_search_plan       # 生成求职计划
analyze_skill_transferability  # 技能迁移分析
recommend_transition_roles     # 推荐过渡岗位
get_salary_benchmark           # 薪资基准查询
generate_negotiation_strategy  # 谈判策略
```

添加新工具：用 `@tool` 装饰器，写清楚 docstring（LLM 靠它判断何时调用），加入对应的 `INTENT_TOOL_NAMES`。

### 前端 SSE 事件 (static/index.html)

| 事件 | 触发 | 前端处理 |
|------|------|---------|
| intent | 意图分类完成 | 显示模式标签 |
| token | LLM 流式文字块 | 追加到气泡（match 模式跳过，tool 执行时跳过） |
| tool_start | 工具开始 | 显示状态条 + 进度动画 |
| tool_end | 工具结束 | 隐藏状态条 |
| report | 匹配报告生成 | 切换到报告页 |
| done | 本轮结束 | 更新 thread_id |
| error | 异常 | 显示错误 |

### 评分公式 (services/matcher.py:35-44)

```python
c.overall_power = (c.project_depth + c.intern_quality + c.tech_stack_fit) // 3
report.overall_score = (
    c.project_depth * 25 + c.intern_quality * 20
    + c.tech_stack_fit * 35 + report.hidden_fit * 20
) // 100
```

---

## 4. 如何添加新功能

### 新增一个意图+工具

1. `services/agent_tools.py` 或 `agent_tools_advanced.py` — 用 `@tool` 装饰器创建工具函数
2. `agent/agent_factory.py` — 在对应的 `INTENT_TOOL_NAMES` 中加入工具名；如果是全新意图，加 prompt 映射
3. `prompts/agent_modes.py` — 如果需要专用系统提示词，加一个

### 新增报告页章节

1. `static/index.html` — `renderReport()` 函数中添加新的 card 和 HTML
2. 如果需要新数据，在 `analyze_match` 工具中收集，存入 `_thread_cache`

### 新增后端 API 端点

1. `handlers/` 下新建 `.py` 文件，定义 FastAPI router
2. `main.py` 中 `include_router`

### 修改前端样式

- 暗色主题 CSS 在 `<style>` 主块里（第 12-295 行）
- 暖色覆盖在 `<style id="warm-theme">`（第 296-299 行），用 `!important` 覆盖
- 主题切换逻辑在 `<script>` 最上面的 `setTheme/applyTheme/toggleThemeMenu`

---

## 5. 常见问题

### DeepSeek function calling 不调工具

- 确保 tool 数量 ≤ 5（DeepSeek 限制）
- 检查 `agent_factory.py` 中 `INTENT_TOOL_NAMES` 每个意图是否不超过 5 个
- 温度必须为 0

### 匹配报告不显示

- `handlers/chat_handler.py` 在 agent 完成后检查 `_thread_cache[thread_id]["report"]`
- `analyze_match` 工具必须成功执行并把结果存入 `_thread_cache`
- 前端 `handleSSEEvent` 的 `report` case 调用 `renderReport()`

### 文件上传失败

- PDF 依赖 `https://cdn.jsdelivr.net/npm/pdfjs-dist@5/build/pdf.min.mjs`
- 图片上传走 `/api/resume/upload` 或 base64
- 如果 CDN 挂了，用户可以用粘贴文本代替

### Nginx 配置换行

腾讯云 OrcaTerm 粘贴多行会换行，用 Python 写文件或 base64 解码方式写入。

---

## 6. 环境要求

- Python 3.12+
- DeepSeek API Key (`deepseek-chat`)
- MiMo API Key（可选，多模态简历解析用）
- Ubuntu 24.04 / systemd / Nginx

## 7. 启动命令

```bash
# 开发环境
cd Offer-Hunter-langchain
python main.py
# 访问 http://localhost:8080

# 生产环境
sudo systemctl restart offer-hunter
sudo systemctl status offer-hunter
journalctl -u offer-hunter -f
```
