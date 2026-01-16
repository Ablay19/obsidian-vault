import aiohttp
import logging
from typing import Dict, Optional

logger = logging.getLogger(__name__)

DEFAULT_TIMEOUT = 300
MAX_RETRIES = 3
RETRY_DELAY = 1

class WorkerClient:
    """Client for communicating with Worker from Renderer"""

    def __init__(self, worker_url: str):
        self.worker_url = worker_url.rstrip('/')

    async def send_status_update(
        self,
        job_id: str,
        status: str,
        data: Optional[Dict] = None
    ):
        """Send status update to worker"""
        url = f"{self.worker_url}/api/v1/jobs/{job_id}/status"
        payload = {
            'job_id': job_id,
            'status': status,
            'timestamp': data.get('timestamp') if data else None,
            **(data or {})
        }

        for attempt in range(MAX_RETRIES):
            try:
                async with aiohttp.ClientSession() as session:
                    async with session.post(
                        url,
                        json=payload,
                        timeout=aiohttp.ClientTimeout(total=DEFAULT_TIMEOUT)
                    ) as response:
                        if response.status == 200:
                            logger.info(f"Status update sent: {job_id} -> {status}")
                            return True
                        else:
                            logger.warning(f"Status update failed: {job_id} - {response.status}")
                            return False
            except Exception as e:
                logger.warning(f"Status update attempt {attempt + 1} failed: {e}")

                if attempt < MAX_RETRIES - 1:
                    import asyncio
                    await asyncio.sleep(RETRY_DELAY)

        logger.error(f"All status update attempts failed: {job_id}")
        return False

    async def send_progress_update(
        self,
        job_id: str,
        progress: float,
        message: str
    ):
        """Send progress update to worker"""
        url = f"{self.worker_url}/api/v1/jobs/{job_id}/progress"
        payload = {
            'job_id': job_id,
            'progress': progress,
            'message': message,
            'timestamp': None
        }

        try:
            async with aiohttp.ClientSession() as session:
                async with session.post(url, json=payload) as response:
                    return response.status == 200
        except Exception as e:
            logger.error(f"Progress update failed: {job_id} - {e}", exc_info=True)
            return False
