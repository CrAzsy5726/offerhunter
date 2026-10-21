"""Resume parsing service using LangChain + DeepSeek/MiMo."""
import json
from langchain_core.messages import SystemMessage, HumanMessage
from models import Resume
from prompts.resume_extraction import RESUME_EXTRACTION_PROMPT, RESUME_EXTRACTION_INSTRUCTION
from services.extract_json import extract_json
from services.llm_factory import create_deepseek_llm, create_mimo_llm


class ResumeParser:
    def __init__(self):
        self._text_llm = create_deepseek_llm(temperature=0)
        self._vision_llm = create_mimo_llm(temperature=0)

    async def parse(self, text: str) -> Resume:
        text = text.strip()
        if not text:
            raise ValueError("empty resume text")

        messages = [
            SystemMessage(content=RESUME_EXTRACTION_PROMPT),
            HumanMessage(content=text + "\n\n" + RESUME_EXTRACTION_INSTRUCTION),
        ]
        response = await self._text_llm.ainvoke(messages)
        content = response.content if hasattr(response, 'content') else str(response)

        resume = self._parse_response(content)
        resume.raw_text = text
        return resume

    async def parse_from_images(self, images: list[str]) -> Resume:
        """Parse resume from base64-encoded images (multimodal)."""
        system_content = (
            RESUME_EXTRACTION_PROMPT
            + "\n\n请从这个简历图片/文档中提取所有文字信息。注意这是多页简历的扫描/截图，请完整提取每一页的内容。"
        )
        user_text = (
            f"这是一份简历的{len(images)}页截图。请识别并提取所有文字信息，包括个人信息、教育背景、技能、项目经历、实习经历等。"
            + RESUME_EXTRACTION_INSTRUCTION
        )

        # Build multimodal user content
        user_content = [{"type": "text", "text": user_text}]
        for img in images:
            url = img if img.startswith("data:") else f"data:image/jpeg;base64,{img}"
            user_content.append({"type": "image_url", "image_url": {"url": url}})

        messages = [
            SystemMessage(content=system_content),
            HumanMessage(content=user_content),
        ]
        response = await self._vision_llm.ainvoke(messages)
        content = response.content if hasattr(response, 'content') else str(response)

        resume = self._parse_response(content)
        resume.raw_text = content
        return resume

    def _parse_response(self, content: str) -> Resume:
        try:
            return Resume.model_validate_json(content)
        except Exception:
            json_str = extract_json(content)
            if not json_str:
                raise ValueError(f"failed to parse LLM response as JSON")
            return Resume.model_validate_json(json_str)
