# Offer 捕手（Offer Hunter）

AI 求职助手：**简历匹配分析** + **AI 模拟面试**。

上传你的简历和岗位 JD，系统基于大模型生成多维度的匹配分析报告，或针对性地生成面试题并逐题评分。后端采用 **Go 自研 tool-calling Agent**（未使用 LangChain），前端为原生 JS 单页应用，通过 SSE 流式通信。

---

## 功能特性

- **简历分析**：上传简历 + 粘贴 JD，生成完整匹配报告——硬性门槛判定、竞争力评分（项目深度 / 实习含金量 / 技术栈匹配）、隐性契合分析、简历优化建议、能力差距分析，支持导出 Markdown。
- **模拟面试**：根据简历 + JD 生成针对性的面试题，逐题作答、AI 评分（1-10 分 + 点评 + 参考要点），结束后给出总评。
- **匿名反馈板**：全匿名提交建议、回复、点赞（同意），按赞同数排序，数据持久化到本地 JSON 文件，无需登录。

---

## 技术栈

- **后端**：Go + Gin，自研 tool-calling Agent 循环
- **前端**：原生 HTML / CSS / JavaScript（无框架）
- **通信**：SSE（Server-Sent Events）流式输出
- **大模型**：DeepSeek（文本 LLM）+ MiMo（多模态，用于 PDF/图片简历 OCR）
- **部署**：单二进制 + systemd + Nginx

---

## 架构亮点

1. **自研 LLM Agent tool-calling 循环**：从零实现「拼消息 → 调 LLM → 执行工具 → 回填结果 → 循环直到最终答案」的完整机制，替代 LangChain 的 `create_agent`。
2. **意图识别双层设计**：确定性关键词预判 + 大模型兜底，兼顾速度、成本与覆盖率。
3. **工具裁剪（tool pruning）**：按用户意图动态收敛工具集（11 个工具 → 每次只暴露 4-5 个），解决 DeepSeek 多工具场景下调用不稳定的问题。
4. **SSE 事件流**：后端逐事件推送 `intent` / `token` / `tool_start` / `tool_end` / `report` / `interview` / `done`，前端渐进渲染。

---

## 快速开始

```bash
# 克隆仓库
git clone https://github.com/<你的用户名>/offer-hunter.git
cd offer-hunter

# 配置环境变量（填入真实的 DeepSeek / MiMo API Key）
cp .env.example .env
# 编辑 .env

# 编译并运行
go build -o offer-hunter ./cmd/server
./offer-hunter

# 浏览器打开 http://localhost:8080
```

> 需要 Go 1.26+。

---

## 配置说明（.env）

| 变量 | 说明 |
|------|------|
| `DEEPSEEK_API_KEY` | DeepSeek 文本大模型 API Key（必填） |
| `DEEPSEEK_BASE_URL` | DeepSeek API 地址，默认 `https://api.deepseek.com/v1` |
| `DEEPSEEK_MODEL` | 文本模型名，默认 `deepseek-flash` |
| `MIMO_API_KEY` | MiMo 多模态模型 API Key（用于 PDF/图片简历 OCR，缺失则 OCR 不可用） |
| `MIMO_BASE_URL` | MiMo API 地址 |
| `MIMO_MODEL` | 多模态模型名 |
| `SERVER_PORT` | 服务端口，默认 `8080` |
| `SERVER_HOST` | 监听地址，默认 `0.0.0.0` |

---

## 目录结构

```
cmd/server/          程序入口
internal/
  agent/             自研 tool-calling 循环 + 意图识别 + 会话记忆
  handlers/          HTTP handlers（chat / feedback / match / interview 等）
  services/          业务逻辑（简历解析、JD 解析、匹配、面试评分）
  llm/               DeepSeek / MiMo 客户端封装
  prompts/           所有 prompt 常量
  models/            数据结构
  config/            .env 配置读取
  server/            Gin 路由注册
static/              前端单页（含本地 pdf.js）
legacy/              Python 旧版实现（仅供参考，不再维护）
deploy.sh            服务器部署脚本（交叉编译 + systemd）
undeploy.sh          卸载脚本
```

---

## 部署

项目通过 `deploy.sh` 一键部署到 Linux 服务器（交叉编译纯静态二进制 + systemd 服务 + 前端静态文件）：

```bash
# 在服务器上以 root 执行
bash deploy.sh
```

脚本会自动：
1. 交叉编译 Linux 二进制（`CGO_ENABLED=0`，无 glibc 依赖）
2. 部署前端静态文件
3. 配置 systemd 服务并启动

卸载用 `bash undeploy.sh`。详细说明见脚本内注释。
