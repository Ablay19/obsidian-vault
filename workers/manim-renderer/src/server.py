from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import uvicorn
from .renderer import render_manim_video
from .logger import setup_logger

app = FastAPI(title="Manim Renderer API")
logger = setup_logger(__name__)

class RenderRequest(BaseModel):
    job_id: str
    problem_text: str
    manim_code: str
    callback_url: str | None = None

class RenderResponse(BaseModel):
    job_id: str
    status: str
    video_path: str | None = None
    video_size_bytes: int | None = None
    render_duration_seconds: float | None = None
    error: str | None = None

@app.post("/render")
async def submit_render(request: RenderRequest):
    logger.info(f"Received render request: {request.job_id}")

    try:
        result = await render_manim_video(
            job_id=request.job_id,
            problem_text=request.problem_text,
            manim_code=request.manim_code
        )

        response = RenderResponse(
            job_id=request.job_id,
            status="success",
            video_path=result.get("video_path"),
            video_size_bytes=result.get("video_size_bytes"),
            render_duration_seconds=result.get("render_duration_seconds")
        )

        logger.info(f"Render successful: {request.job_id}")

        if request.callback_url:
            await send_callback(request.callback_url, response.dict())

        return response

    except Exception as e:
        logger.error(f"Render failed: {request.job_id}", exc_info=True)

        response = RenderResponse(
            job_id=request.job_id,
            status="error",
            error=str(e)
        )

        if request.callback_url:
            await send_callback(request.callback_url, response.dict())

        raise HTTPException(status_code=500, detail=str(e))

@app.get("/status/{job_id}")
async def get_status(job_id: str):
    return {"job_id": job_id, "status": "processing"}

@app.get("/health")
async def health_check():
    return {"status": "healthy"}

async def send_callback(url: str, data: dict):
    import aiohttp
    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(url, json=data) as response:
                if response.status != 200:
                    logger.warning(f"Callback failed: {url} - {response.status}")
    except Exception as e:
        logger.error(f"Callback error: {url}", exc_info=True)

def start_server(host: str = "0.0.0.0", port: int = 8080):
    uvicorn.run(app, host=host, port=port)
