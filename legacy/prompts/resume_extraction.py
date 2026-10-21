RESUME_EXTRACTION_PROMPT = """你是一个专业的简历解析专家。请从以下简历文本中提取结构化信息。

请返回JSON格式，包含以下字段：
{
  "education": [
    {
      "school": "学校名称",
      "degree": "学历（本科/硕士/博士）",
      "major": "专业",
      "start_date": "开始时间",
      "end_date": "结束时间",
      "gpa": "GPA（如有）"
    }
  ],
  "skills": ["技能1", "技能2"],
  "projects": [
    {
      "name": "项目名称",
      "description": "项目描述",
      "role": "担任角色",
      "tech_stack": ["技术1", "技术2"],
      "duration": "持续时间",
      "achievements": ["成就1", "成就2"]
    }
  ],
  "experience": [
    {
      "company": "公司名称",
      "position": "职位",
      "start_date": "开始时间",
      "end_date": "结束时间",
      "description": "工作描述",
      "achievements": ["成就1", "成就2"],
      "tech_stack": ["技术1", "技术2"]
    }
  ],
  "personality": ["个人特质1", "个人特质2"],
  "metadata": {
    "total_projects": 项目数量,
    "total_experience": 实习/工作经历数量,
    "top_skills": ["最核心技能1", "最核心技能2"],
    "experience_level": "经验水平（应届/1-3年/3-5年等）"
  }
}

要求：
1. 保留原文中的所有关键信息
2. 对于每条经历，标注可能被低估的能力
3. 技能要尽量全面，包括软技能和硬技能
4. 如果某些信息不明确，可以合理推断

简历文本："""

RESUME_EXTRACTION_INSTRUCTION = "请严格按照JSON格式返回，不要包含任何其他文字。"
