# Offer 捕手 LangChain Edition — 项目总结

## 项目概述

AI 求职助手，纯对话式交互。用户粘贴简历+JD、上传文件、或直接说"模拟面试""找XX工作""从Java转AI"，Agent 自动理解意图并调用对应工具完成任务。

基于 LangChain `create_agent` + DeepSeek function calling，11 个工具、6 种意图自动分类、SSE 流式输出、双主题（暖白抹茶绿/暗色靛紫）。

---

## 核心功能

| 功能 | 触发方式 | 使用的工具 |
|------|---------|-----------|
| 📄 简历+JD匹配分析 | 发简历+JD，说"帮我分析" | parse_resume → parse_jd → analyze_match → 报告页 |
| 🎤 模拟面试 | 说"模拟面试" | generate_interview_questions → 逐题提问评分 |
| 📋 求职策略 | 说"找XX工作" | search_job_listings + generate_job_search_plan + research_market |
| 🔄 职业转型 | 说"从Java转AI" | analyze_skill_transferability + recommend_transition_roles |
| 💰 薪资谈判 | 说"帮我看薪资" | get_salary_benchmark + generate_negotiation_strategy |
| 💬 闲聊引导 | 说"你好"等 | 自然聊天，结尾引导回求职话题 |

## 技术栈

- **后端**: FastAPI + LangChain (create_agent) + DeepSeek (chat + function calling)
- **前端**: 单页 HTML (~1200 行)，Server-Sent Events 流式输出
- **部署**: systemd + Nginx + Let's Encrypt SSL + Ubuntu 24.04
- **多模态**: MiMo（PDF/图片简历 OCR），仅后端解析时使用

## 关键架构决策

1. **两轮调用** — Round 1 LLM 分类意图(0.5s, 1 token)，Round 2 根据意图加载对应的 tool 子集执行。避免 11 个 tool 同时发给 DeepSeek 导致 function calling 不稳定
2. **单 Agent + 共享 InMemorySaver** — 所有意图共用一个 Agent，对话历史跨模式保留
3. **工具集隔离** — 每个意图只加载 2-4 个工具，DeepSeek 在 ≤5 个 tool 时稳定
4. **双主题注入/移除** — 暗色原生代码不动，暖色作为 `<style id="warm-theme">` 覆S盖层动态注入/移除，localStorage 持久化

## 部署信息

- 域名: `https://offerhunting.site`
- 服务器: 腾讯云 49.232.59.127
- 后端: systemd `offer-hunter` 服务，端口 8080
- 反代: Nginx 80→443，SSL 证书 Let's Encrypt certbot
- 部署脚本: `setup.sh`
