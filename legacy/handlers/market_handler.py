"""Market research endpoint."""
from fastapi import APIRouter, HTTPException
from models import MarketResearchRequest
from services.market_research import MarketResearchService

router = APIRouter(prefix="/api", tags=["market"])


_market_service: MarketResearchService | None = None


def get_market_service() -> MarketResearchService:
    global _market_service
    if _market_service is None:
        _market_service = MarketResearchService()
    return _market_service


@router.post("/market")
async def research_market(req: MarketResearchRequest):
    try:
        info = await get_market_service().research(req.job_info, req.job_preference)
        return {"market": info.model_dump()}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
