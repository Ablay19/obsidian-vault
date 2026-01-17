#!/usr/bin/env python3
"""
Video Cleanup Script

This script handles the cleanup of rendered videos from the local filesystem
and optionally triggers deletion from cloud storage (R2).
"""

import argparse
import json
import logging
import os
import sys
from pathlib import Path
from typing import Optional

logging.basicConfig(
    level=logging.INFO,
    format='%(asctime)s - %(name)s - %(levelname)s - %(message)s'
)
logger = logging.getLogger('cleanup')


class VideoCleanupService:
    def __init__(self, output_dir: str = "/tmp/manim_output"):
        self.output_dir = Path(output_dir)
    
    def cleanup_job(self, job_id: str, delete_from_r2: bool = False) -> dict:
        result = {
            "job_id": job_id,
            "deleted": False,
            "files_deleted": 0,
            "bytes_freed": 0,
            "r2_deleted": False,
            "error": None
        }
        
        try:
            job_dir = self.output_dir / job_id
            if not job_dir.exists():
                logger.warning(f"Job directory not found: {job_dir}")
                result["deleted"] = True
                return result
            
            bytes_freed = 0
            files_deleted = 0
            
            for file_path in job_dir.rglob("*"):
                if file_path.is_file():
                    file_size = file_path.stat().st_size
                    os.remove(file_path)
                    bytes_freed += file_size
                    files_deleted += 1
                    logger.debug(f"Deleted: {file_path}")
            
            job_dir.rmdir()
            
            result["deleted"] = True
            result["files_deleted"] = files_deleted
            result["bytes_freed"] = bytes_freed
            logger.info(f"Cleaned up job {job_id}: {files_deleted} files, {bytes_freed} bytes")
            
            if delete_from_r2:
                try:
                    self._delete_from_r2(job_id)
                    result["r2_deleted"] = True
                except Exception as e:
                    logger.error(f"Failed to delete from R2: {e}")
                    result["error"] = f"R2 deletion failed: {str(e)}"
            
        except Exception as e:
            logger.exception(f"Cleanup failed for job {job_id}: {e}")
            result["error"] = str(e)
        
        return result
    
    def _delete_from_r2(self, job_id: str):
        r2_config = self._get_r2_config()
        if not r2_config:
            logger.warning("R2 configuration not found, skipping R2 cleanup")
            return
        
        try:
            import boto3
            s3_client = boto3.client(
                's3',
                endpoint_url=r2_config["endpoint"],
                aws_access_key_id=r2_config["access_key"],
                aws_secret_access_key=r2_config["secret_key"]
            )
            
            video_key = f"videos/{job_id}.mp4"
            s3_client.delete_object(
                Bucket=r2_config["bucket"],
                Key=video_key
            )
            logger.info(f"Deleted from R2: {video_key}")
            
        except ImportError:
            logger.warning("boto3 not available, cannot delete from R2")
        except Exception as e:
            logger.error(f"R2 deletion error: {e}")
            raise
    
    def _get_r2_config(self) -> Optional[dict]:
        return {
            "endpoint": os.environ.get("R2_ENDPOINT"),
            "access_key": os.environ.get("R2_ACCESS_KEY_ID"),
            "secret_key": os.environ.get("R2_SECRET_ACCESS_KEY"),
            "bucket": os.environ.get("R2_BUCKET_NAME"),
        }
    
    def cleanup_expired(self, max_age_hours: int = 24) -> dict:
        result = {
            "jobs_cleaned": 0,
            "bytes_freed": 0,
            "errors": []
        }
        
        if not self.output_dir.exists():
            return result
        
        from datetime import datetime, timedelta
        cutoff_time = datetime.now() - timedelta(hours=max_age_hours)
        
        for job_dir in self.output_dir.iterdir():
            if not job_dir.is_dir():
                continue
            
            try:
                mtime = datetime.fromtimestamp(job_dir.stat().st_mtime)
                if mtime < cutoff_time:
                    cleanup_result = self.cleanup_job(job_dir.name)
                    if cleanup_result["deleted"]:
                        result["jobs_cleaned"] += 1
                        result["bytes_freed"] += cleanup_result["bytes_freed"]
            except Exception as e:
                result["errors"].append({"job": job_dir.name, "error": str(e)})
        
        logger.info(f"Expired cleanup: {result['jobs_cleaned']} jobs, {result['bytes_freed']} bytes")
        return result


def main():
    parser = argparse.ArgumentParser(description="Video Cleanup Service")
    parser.add_argument("--job-id", required=True, help="Job ID to clean up")
    parser.add_argument("--output-dir", default="/tmp/manim_output")
    parser.add_argument("--delete-from-r2", action="store_true", help="Also delete from R2")
    parser.add_argument("--cleanup-expired", action="store_true", help="Clean up all expired videos")
    parser.add_argument("--max-age-hours", type=int, default=24, help="Max age in hours for expired cleanup")
    
    args = parser.parse_args()
    
    service = VideoCleanupService(output_dir=args.output_dir)
    
    if args.cleanup_expired:
        result = service.cleanup_expired(max_age_hours=args.max_age_hours)
    else:
        result = service.cleanup_job(args.job_id, delete_from_r2=args.delete_from_r2)
    
    print(json.dumps(result))
    return 0 if result.get("error") is None else 1


if __name__ == "__main__":
    sys.exit(main())
