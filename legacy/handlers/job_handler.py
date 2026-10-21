"""JD parsing endpoint."""
from fastapi import APIRouter, HTTPException
from models import JobParseRequest
from services.jd_parser import JDParser

router = APIRouter(prefix="/api/job", tags=["job"])


_jd_parser: JDParser | None = None


def get_jd_parser() -> JDParser:
    global _jd_parser
    if _jd_parser is None:
        _jd_parser = JDParser()
    return _jd_parser


@router.post("/parse")
async def parse_job(req: JobParseRequest):
    try:
        job = await get_jd_parser().parse(req.text)
        return {"job": job.model_dump()}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
