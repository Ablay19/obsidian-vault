#!/usr/bin/env python3
"""
Manim Rendering Script

This script handles the rendering of Manim animations for the AI Manim Video Generator.
It accepts a job ID and Manim code, renders the animation, and uploads the result.
"""

import argparse
import json
import logging
import os
import subprocess
import sys
import tempfile
import time
from datetime import datetime
from pathlib import Path
from typing import Optional

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger('renderer')

MANIM_VERSION = "0.18.1"
MAX_RENDER_TIME_SECONDS = 300
MAX_VIDEO_SIZE_BYTES = 50 * 1024 * 1024


class ManimRenderer:
    def __init__(self, output_dir: str = "/tmp/manim_output"):
        self.output_dir = Path(output_dir)
        self.output_dir.mkdir(parents=True, exist_ok=True)
        self._check_dependencies()
    
    def _check_dependencies(self):
        try:
            result = subprocess.run(
                ["manim", "--version"],
                capture_output=True,
                text=True,
                timeout=10
            )
            if result.returncode == 0:
                logger.info(f"Manim version: {result.stdout.strip()}")
            else:
                logger.warning("Manim not found, using mock renderer")
        except (subprocess.TimeoutExpired, FileNotFoundError):
            logger.warning("Manim not available, using mock renderer")
    
    def render(self, job_id: str, manim_code: str, 
               quality: str = "medium", 
               format: str = "mp4") -> dict:
        start_time = time.time()
        
        logger.info(f"Starting render for job {job_id}")
        logger.info(f"Quality: {quality}, Format: {format}")
        
        try:
            with tempfile.TemporaryDirectory() as temp_dir:
                code_path = Path(temp_dir) / "scene.py"
                code_path.write_text(manim_code)
                
                output_path = self.output_dir / job_id
                output_path.mkdir(parents=True, exist_ok=True)
                
                manim_cmd = [
                    "manim",
                    str(code_path),
                    "--output_dir", str(output_path),
                    "--quality", quality,
                    "-p",
                ]
                
                result = subprocess.run(
                    manim_cmd,
                    capture_output=True,
                    text=True,
                    timeout=MAX_RENDER_TIME_SECONDS
                )
                
                render_duration = time.time() - start_time
                
                if result.returncode == 0:
                    video_files = list(output_path.glob(f"*.{format}"))
                    if video_files:
                        video_path = str(video_files[0])
                        video_size = os.path.getsize(video_path)
                        
                        if video_size > MAX_VIDEO_SIZE_BYTES:
                            logger.warning(f"Video too large ({video_size} bytes), compressing...")
                            video_path = self._compress_video(video_path, job_id)
                            video_size = os.path.getsize(video_path)
                        
                        logger.info(f"Render completed in {render_duration:.2f}s")
                        logger.info(f"Video size: {video_size} bytes")
                        
                        return {
                            "job_id": job_id,
                            "status": "success",
                            "video_path": video_path,
                            "video_size_bytes": video_size,
                            "render_duration_seconds": render_duration
                        }
                    else:
                        return {
                            "job_id": job_id,
                            "status": "error",
                            "error": "No video file generated",
                            "render_duration_seconds": render_duration
                        }
                else:
                    error_msg = result.stderr or result.stdout or "Unknown error"
                    logger.error(f"Manim render failed: {error_msg}")
                    return {
                        "job_id": job_id,
                        "status": "error",
                        "error": error_msg[:500],
                        "render_duration_seconds": render_duration
                    }
                    
        except subprocess.TimeoutExpired:
            logger.error(f"Render timed out after {MAX_RENDER_TIME_SECONDS}s")
            return {
                "job_id": job_id,
                "status": "timeout",
                "error": f"Render exceeded {MAX_RENDER_TIME_SECONDS}s timeout",
                "render_duration_seconds": MAX_RENDER_TIME_SECONDS
            }
        except Exception as e:
            logger.exception(f"Unexpected error during render: {e}")
            return {
                "job_id": job_id,
                "status": "error",
                "error": str(e)[:500],
                "render_duration_seconds": time.time() - start_time
            }
    
    def _compress_video(self, video_path: str, job_id: str) -> str:
        compressed_path = video_path.replace(f".mp4", "_compressed.mp4")
        try:
            subprocess.run([
                "ffmpeg", "-i", video_path,
                "-vcodec", "libx264",
                "-crf", "28",
                "-preset", "fast",
                compressed_path
            ], capture_output=True, timeout=120)
            if os.path.exists(compressed_path):
                os.remove(video_path)
                return compressed_path
        except Exception as e:
            logger.warning(f"Compression failed: {e}")
        return video_path
    
    def cleanup(self, job_id: str):
        job_dir = self.output_dir / job_id
        if job_dir.exists():
            import shutil
            shutil.rmtree(job_dir)
            logger.info(f"Cleaned up job {job_id}")


def main():
    parser = argparse.ArgumentParser(description="Manim Renderer")
    parser.add_argument("--job-id", required=True, help="Job ID")
    parser.add_argument("--manim-code", required=True, help="Path to Manim code file")
    parser.add_argument("--quality", default="medium", choices=["low", "medium", "high", "ultra"])
    parser.add_argument("--format", default="mp4", choices=["mp4", "webm"])
    parser.add_argument("--output-dir", default="/tmp/manim_output")
    parser.add_argument("--callback-url", help="URL to call when render completes")
    
    args = parser.parse_args()
    
    renderer = ManimRenderer(output_dir=args.output_dir)
    
    manim_code = Path(args.manim_code).read_text()
    
    result = renderer.render(
        job_id=args.job_id,
        manim_code=manim_code,
        quality=args.quality,
        format=args.format
    )
    
    print(json.dumps(result))
    
    if args.callback_url and result["status"] == "success":
        try:
            import requests
            requests.post(args.callback_url, json=result, timeout=10)
        except Exception as e:
            logger.error(f"Callback failed: {e}")


if __name__ == "__main__":
    main()
