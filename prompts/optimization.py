OPTIMIZATION_PROMPT = """你是一个专业的简历优化顾问。请根据以下简历、岗位要求和匹配分析，提供具体的简历优化建议。

## 优化原则
1. 只做"翻译"，不做"编造" - 所有改动必须基于简历中的原始内容
2. 将学生经历"翻译"成招聘方能理解的语言
3. 突出与岗位最相关的经历和技能
4. 用量化数据增强说服力
5. 区分"有证据的建议"和"无证据的建议"：
   - has_evidence=true: 简历中已有相关经历，只是表述不到位，优化能显著提升匹配度
   - has_evidence=false: 简历中确实缺少该项能力或经验，需要先学习再写入简历

## 输出格式
按"有证据→无证据"顺序排列，同组内按优先级排序：
{
  "optimization_tips": [
    {
      "priority": 1,
      "category": "建议类别（技能/项目/实习/描述/其他）",
      "problem": "当前问题描述",
      "suggestion": "优化方向说明",
      "example": "具体的修改示例",
      "has_evidence": true
    }
  ]
}

## 简历信息
{resume_json}

## 岗位信息
{job_json}

## 匹配分析
{match_json}"""

OPTIMIZATION_INSTRUCTION = "请严格按照JSON格式返回，不要包含任何其他文字。"
