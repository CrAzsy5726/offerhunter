"""Resume parsing endpoints."""
from fastapi import APIRouter, HTTPException
from models import ResumeParseRequest, ResumeUploadRequest
from services.resume_parser import ResumeParser

router = APIRouter(prefix="/api/resume", tags=["resume"])


_resume_parser: ResumeParser | None = None


def get_resume_parser() -> ResumeParser:
    global _resume_parser
    if _resume_parser is None:
        _resume_parser = ResumeParser()
    return _resume_parser


@router.post("/parse")
async def parse_resume(req: ResumeParseRequest):
    try:
        resume = await get_resume_parser().parse(req.text)
        return {"resume": resume.model_dump()}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@router.post("/upload")
async def upload_resume(req: ResumeUploadRequest):
    try:
        parser = get_resume_parser()
        if req.images:
            resume = await parser.parse_from_images(req.images)
        elif req.text:
            resume = await parser.parse(req.text)
        else:
            raise HTTPException(status_code=400, detail="No text or images provided")
        return {"resume": resume.model_dump()}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))
