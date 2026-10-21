"""Logging and timing middleware for the Offer-Hunter agent."""
import time
import logging
from typing import Any
from langchain.agents.middleware import AgentMiddleware

logger = logging.getLogger(__name__)


class LoggingTimingMiddleware(AgentMiddleware):
    """Logs every LLM call and tool invocation with timing."""

    async def awrap_model_call(self, request: Any, handler: Any) -> Any:
        start = time.perf_counter()
        msg_count = len(request.state.get("messages", [])) if hasattr(request, "state") else 0
        tool_names = [t.name for t in request.tools] if hasattr(request, 'tools') else []
        logger.info("[Agent] LLM call starting (history: %d messages, tools: %s)", msg_count, tool_names)
        try:
            response = await handler(request)
            elapsed = time.perf_counter() - start
            logger.info("[Agent] LLM call completed in %.2fs", elapsed)
            return response
        except Exception:
            elapsed = time.perf_counter() - start
            logger.error("[Agent] LLM call failed after %.2fs", elapsed)
            raise

    async def awrap_tool_call(self, request: Any, handler: Any) -> Any:
        tc = request.tool_call if hasattr(request, "tool_call") else {}
        tool_name = tc.get("name", "unknown") if isinstance(tc, dict) else getattr(tc, "name", "unknown")
        start = time.perf_counter()
        logger.info("[Agent] Tool call starting: %s", tool_name)
        try:
            response = await handler(request)
            elapsed = time.perf_counter() - start
            logger.info("[Agent] Tool call '%s' completed in %.2fs", tool_name, elapsed)
            return response
        except Exception:
            elapsed = time.perf_counter() - start
            logger.error("[Agent] Tool call '%s' failed after %.2fs", tool_name, elapsed)
            raise
