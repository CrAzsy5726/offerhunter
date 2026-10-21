MATCHING_PROMPT = """你是一个拥有10年经验的资深HR和技术面试官。请对以下简历和岗位进行多维度匹配分析。

## 分析维度

### 第一层：硬性门槛判定
检查以下硬性条件：
- 学历要求是否满足
- 专业要求是否满足
- 必备技能是否具备
- 工作经验是否符合

对每个条件给出：pass/fail/warn 状态，并说明原因。如果是warn，说明是否可以通过简历优化来弥补。

### 第二层：竞争力评分
从以下维度评估竞争力：
1. 项目经验深度（技术难度、个人贡献度、业务影响力）：0-100分
2. 实习含金量（公司知名度、岗位相关性、产出量化程度）：0-100分
3. 工具链熟练度（与岗位要求技术栈的重叠度）：0-100分
4. 整体竞争力（综合以上维度）：0-100分

每个维度需要：
- 给出具体分数
- 提供一句话评语
- 引用简历中的具体经历作为证据

### 第三层：隐性契合分析
分析以下方面：
1. 公司风格匹配（大厂vs创业、业务型vs技术型）
2. 成长路径匹配（当前阶段是否适合岗位的成长曲线）
3. 团队文化匹配（基于JD语言分析的团队风格）
4. 隐性契合度：0-100分

## 输出格式
请返回JSON格式：
{
  "threshold_check": {
    "passed": true/false,
    "items": [
      {
        "category": "类别",
        "requirement": "要求内容",
        "status": "pass/fail/warn",
        "reason": "原因说明",
        "fixable": true/false
      }
    ]
  },
  "competitiveness": {
    "project_depth": {"score": 85, "comment": "评语", "evidence": "证据"},
    "intern_quality": {"score": 70, "comment": "评语", "evidence": "证据"},
    "tech_stack_fit": {"score": 90, "comment": "评语", "evidence": "证据"}
  },
  "hidden_fit": {
    "company_style": "分析说明",
    "growth_path": "分析说明",
    "team_culture": "分析说明",
    "score": 75
  },
  "benchmark": {
    "pass_line": 65,
    "percentile": "top-30%",
    "comment": "该岗位通常匹配度65分以上可通过初筛，你的分数处于较高水平"
  },
  "summary": "整体匹配分析总结"
}

## 简历信息
{resume_json}

## 岗位信息
{job_json}"""

MATCHING_INSTRUCTION = "请严格按照JSON格式返回，不要包含任何其他文字。"
