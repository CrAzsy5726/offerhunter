JD_EXTRACTION_PROMPT = """你是一个专业的招聘需求分析专家。请从以下岗位描述(JD)中提取结构化信息。

请返回JSON格式，包含以下字段：
{
  "title": "岗位名称",
  "company": "公司名称（如有）",
  "location": "工作地点",
  "salary": "薪资范围（如有）",
  "requirements": [
    {
      "category": "要求类别（学历/专业/技能/经验/其他）",
      "value": "具体要求内容",
      "is_hard": true/false,
      "priority": "high/medium/low"
    }
  ],
  "bonus_items": ["加分项1", "加分项2"],
  "hidden_prefs": ["隐性偏好1", "隐性偏好2"],
  "description": "岗位职责描述",
  "team_info": "团队信息（如有）",
  "growth_path": "成长路径（如有）"
}

要求：
1. 区分硬性要求和软性要求
2. 识别隐性偏好（如"有激情"可能暗示需要加班、"扁平化管理"可能暗示需要自驱力）
3. 提取所有技术栈要求
4. 如果信息不完整，标注出来

岗位描述文本："""

JD_EXTRACTION_INSTRUCTION = "请严格按照JSON格式返回，不要包含任何其他文字。"
