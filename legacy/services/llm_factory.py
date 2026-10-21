"""LangChain LLM factory for DeepSeek (text) and MiMo (multimodal)."""
from langchain_openai import ChatOpenAI
import config


def create_deepseek_llm(temperature: float = 0) -> ChatOpenAI:
    """Create a ChatOpenAI instance configured for DeepSeek."""
    return ChatOpenAI(
        model=config.DEEPSEEK_MODEL,
        api_key=config.DEEPSEEK_API_KEY,
        base_url=config.DEEPSEEK_BASE_URL,
        temperature=temperature,
        max_tokens=4096,
    )


def create_mimo_llm(temperature: float = 0) -> ChatOpenAI:
    """Create a ChatOpenAI instance configured for MiMo (multimodal)."""
    return ChatOpenAI(
        model=config.MIMO_MODEL,
        api_key=config.MIMO_API_KEY,
        base_url=config.MIMO_BASE_URL,
        temperature=temperature,
        max_tokens=8192,
    )
