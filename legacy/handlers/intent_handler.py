"""Intent check endpoint."""
from fastapi import APIRouter, HTTPException
from models import IntentCheckRequest
from services.intent_validator import IntentValidator

router = APIRouter(prefix="/api/intent", tags=["intent"])


_intent_validator: IntentValidator | None = None


def get_intent_validator() -> IntentValidator:
    global _intent_validator
    if _intent_validator is None:
        _intent_validator = IntentValidator()
    return _intent_validator


@router.post("/check")
async def check_intent(req: IntentCheckRequest):
    is_related, reason = await get_intent_validator().is_job_related(req.text)
    return {"is_related": is_related, "reason": reason}
