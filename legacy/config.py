import os
from dotenv import load_dotenv

load_dotenv()

# DeepSeek API (Text LLM)
DEEPSEEK_API_KEY = os.getenv("DEEPSEEK_API_KEY", "")
DEEPSEEK_BASE_URL = os.getenv("DEEPSEEK_BASE_URL", "https://api.deepseek.com/v1")
DEEPSEEK_MODEL = os.getenv("DEEPSEEK_MODEL", "deepseek-chat")

# MiMo API (Multimodal LLM)
MIMO_API_KEY = os.getenv("MIMO_API_KEY", "")
MIMO_BASE_URL = os.getenv("MIMO_BASE_URL", "https://api.xiaomimimo.com/v1")
MIMO_MODEL = os.getenv("MIMO_MODEL", "mimo-v2.5")

# Server
SERVER_PORT = int(os.getenv("SERVER_PORT", "8080"))
SERVER_HOST = os.getenv("SERVER_HOST", "0.0.0.0")
