GAP_ANALYSIS_PROMPT = """你是一个职业发展顾问。请根据以下简历和岗位要求，分析能力差距并提供学习建议。

## 分析要求
1. 识别当前简历中缺失的关键能力/经验
2. 区分短期可弥补（1-2周）和长期需要积累的能力
3. 为每项差距提供具体的学习资源建议

## 输出格式
请返回JSON格式，注意 importance 必须使用英文值：
{
  "gap_analysis": [
    {
      "skill": "缺失的技能/经验名称",
      "importance": "high 或 medium 或 low（必须使用英文）",
      "timeline": "short_term 或 long_term",
      "resource": "学习资源建议（书籍/课程/实践项目方向）"
    }
  ]
}

## 简历信息
{resume_json}

## 岗位信息
{job_json}"""

GAP_ANALYSIS_INSTRUCTION = "请严格按照JSON格式返回，不要包含任何其他文字。"
