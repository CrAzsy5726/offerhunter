"""SSE streaming chat endpoint - classify intent, then run unified agent."""
import json
import uuid
import logging
from fastapi import APIRouter
from fastapi.responses import StreamingResponse
from langchain_core.messages import HumanMessage

from models import ChatRequest
from services.agent_tools import _thread_cache

router = APIRouter(prefix="/api", tags=["agent-chat"])
logger = logging.getLogger(__name__)


@router.post("/chat")
async def chat(req: ChatRequest):
    thread_id = req.thread_id or str(uuid.uuid4())

    async def event_stream():
        from agent.agent_factory import classify_intent, set_intent, get_agent

        # Phase 1: classify intent (single-token LLM call, ~0.5s)
        intent, intent_name = await classify_intent(req.message)
        yield _sse("intent", {"intent": intent, "name": intent_name})

        # Store intent for middleware to read
        set_intent(intent)

        # Phase 2: unified agent (tool-filtering middleware narrows tools per intent)
        agent = get_agent()
        config = {"configurable": {"thread_id": thread_id}}

        try:
            async for event in agent.astream_events(
                {"messages": [HumanMessage(content=req.message)]},
                config=config,
                version="v2",
            ):
                evt = event["event"]

                if evt == "on_chat_model_stream":
                    chunk = event["data"].get("chunk")
                    if chunk and hasattr(chunk, "content") and chunk.content:
                        token_text = chunk.content
                        if isinstance(token_text, list):
                            token_text = "".join(
                                p.get("text", "") if isinstance(p, dict) else str(p)
                                for p in token_text
                            )
                        if token_text:
                            yield _sse("token", {"content": token_text})

                elif evt == "on_tool_start":
                    tool_name = event.get("name", "unknown")
                    tool_input = event["data"].get("input", {})
                    safe_input = {k: v for k, v in tool_input.items() if k != "config"}
                    safe_input = {
                        k: (v[:100] + "..." if isinstance(v, str) and len(v) > 100 else v)
                        for k, v in safe_input.items()
                    }
                    yield _sse("tool_start", {"name": tool_name, "input": safe_input})

                elif evt == "on_tool_end":
                    tool_name = event.get("name", "unknown")
                    output = event["data"].get("output", "")
                    yield _sse("tool_end", {
                        "name": tool_name,
                        "output_preview": str(output)[:300] if output else "",
                    })

            # Check for match report
            report = _thread_cache.get(thread_id, {}).get("report")
            if report:
                yield _sse("report", {"report": report.model_dump()})

            yield _sse("done", {"thread_id": thread_id})

        except Exception as exc:
            logger.exception("Agent error in thread %s", thread_id)
            yield _sse("error", {"message": str(exc)})
            yield _sse("done", {"thread_id": thread_id})

    return StreamingResponse(
        event_stream(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "Connection": "keep-alive",
            "X-Thread-ID": thread_id,
        },
    )


def _sse(event: str, data: dict) -> str:
    return f"event: {event}\ndata: {json.dumps(data, ensure_ascii=False)}\n\n"
