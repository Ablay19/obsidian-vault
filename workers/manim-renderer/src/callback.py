import aiohttp
import logging
from typing import Dict

logger = logging.getLogger(__name__)

WORKER_CALLBACK_TIMEOUT = 30

async def notify_worker_completion(job_id: str, status: str, data: Dict):
    """Send completion notification to worker via webhook"""
    callback_url = data.get('callback_url')

    if not callback_url:
        logger.warning(f"No callback URL for job {job_id}")
        return

    payload = {
        'job_id': job_id,
        'status': status,
        'video_path': data.get('video_path'),
        'video_size_bytes': data.get('video_size_bytes'),
        'render_duration_seconds': data.get('render_duration_seconds'),
        'error': data.get('error')
    }

    try:
        async with aiohttp.ClientSession(timeout=aiohttp.ClientTimeout(total=WORKER_CALLBACK_TIMEOUT)) as session:
            async with session.post(callback_url, json=payload) as response:
                if response.status == 200:
                    logger.info(f"Callback sent successfully for job {job_id}")
                else:
                    logger.error(f"Callback failed for job {job_id}: {response.status}")
    except Exception as e:
        logger.error(f"Failed to send callback for job {job_id}: {e}", exc_info=True)


async def send_heartbeat(worker_url: str):
    """Send heartbeat to worker to indicate renderer is alive"""
    try:
        async with aiohttp.ClientSession() as session:
            async with session.post(f"{worker_url}/heartbeat") as response:
                logger.info(f"Heartbeat sent: {response.status}")
    except Exception as e:
        logger.error(f"Heartbeat failed: {e}", exc_info=True)
