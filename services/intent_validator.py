"""Intent validation service using LangChain + DeepSeek."""
import re
from langchain_core.messages import SystemMessage, HumanMessage
from services.llm_factory import create_deepseek_llm


class IntentValidator:
    _OFF_TOPIC_PATTERNS = [
        re.compile(p) for p in [
            r"^.{0,3}$",
            r"^(你好|hi|hello|hey|喂|在吗)$",
            r"^今天天气",
            r"^你是[谁什么]",
            r"^[多大了\|几岁]",
            r"^讲个(笑话|故事)",
            r"^(帮我)?写(一篇|一段)作文",
        ]
    ]

    def __init__(self):
        self._llm = create_deepseek_llm(temperature=0)

    async def is_job_related(self, text: str) -> tuple[bool, str]:
        text = text.strip()
        lower = text.lower()

        # Quick pre-filter
        if self._is_clearly_off_topic(lower):
            return False, "您的输入似乎与求职无关。请告诉我您想找什么岗位，或粘贴简历和JD。"

        # LLM classification
        system_prompt = "你是一个求职助手。判断用户的输入是否与求职相关（如简历、职位描述、求职意向、技能询问等）。只回复 \"yes\" 或 \"no\"。"
        messages = [
            SystemMessage(content=system_prompt),
            HumanMessage(content=text),
        ]
        try:
            response = await self._llm.ainvoke(messages)
            content = response.content if hasattr(response, 'content') else str(response)
            if content.strip().lower().startswith("yes"):
                return True, ""
        except Exception:
            return True, ""

        return False, "您的输入似乎与求职无关。我可以帮您分析简历、比较岗位匹配度、提供优化建议。请描述您的求职需求吧。"

    def _is_clearly_off_topic(self, text: str) -> bool:
        return any(p.match(text) for p in self._OFF_TOPIC_PATTERNS)
