"""Offer-Hunter LangChain Edition — FastAPI server entry point."""
import logging
from pathlib import Path
from contextlib import asynccontextmanager
from fastapi import FastAPI
from fastapi.middleware.cors import CORSMiddleware
from fastapi.staticfiles import StaticFiles
from fastapi.responses import FileResponse
import config

from handlers import resume_handler, job_handler, match_handler, market_handler, intent_handler, export_handler, chat_handler

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

STATIC_DIR = Path(__file__).parent / "static"


@asynccontextmanager
async def lifespan(app: FastAPI):
    logger.info("Offer-Hunter LangChain Edition starting on %s:%s", config.SERVER_HOST, config.SERVER_PORT)
    yield


app = FastAPI(
    title="Offer-Hunter LangChain Edition",
    description="AI-powered job matching analysis using LangChain + DeepSeek/MiMo",
    version="1.0.0",
    lifespan=lifespan,
)

app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

app.include_router(resume_handler.router)
app.include_router(job_handler.router)
app.include_router(match_handler.router)
app.include_router(market_handler.router)
app.include_router(intent_handler.router)
app.include_router(export_handler.router)
app.include_router(chat_handler.router)


@app.get("/health")
async def health():
    return {"status": "ok", "service": "offer-hunter-langchain"}


@app.get("/")
async def index():
    return FileResponse(
        STATIC_DIR / "index.html",
        headers={
            "Cache-Control": "no-store, no-cache, must-revalidate, max-age=0",
            "Pragma": "no-cache",
            "Expires": "0",
        },
    )


if STATIC_DIR.exists():
    app.mount("/static", StaticFiles(directory=str(STATIC_DIR)), name="static")


if __name__ == "__main__":
    import uvicorn
    uvicorn.run(
        "main:app",
        host=config.SERVER_HOST,
        port=config.SERVER_PORT,
        reload=True,
    )
