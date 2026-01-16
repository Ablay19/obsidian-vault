import logging
import sys
from typing import Optional

WORKER_LOG_FORMAT = '{"level":"%(levelname)s","message":"%(message)s","component":"%(name)s","ts":"%(asctime)s"}'
RENDERER_LOG_FORMAT = '{"level":"%(levelname)s","message":"%(message)s","component":"%(name)s","ts":"%(asctime)s"}'

def setup_logger(
    name: str,
    level: str = "INFO",
    component: Optional[str] = None
) -> logging.Logger:
    """Create structured logger matching Worker's format"""
    logger = logging.getLogger(name)

    if logger.handlers:
        return logger

    handler = logging.StreamHandler(sys.stdout)
    formatter = logging.Formatter(
        RENDERER_LOG_FORMAT,
        datefmt='%Y-%m-%dT%H:%M:%S.%fZ'
    )
    handler.setFormatter(formatter)
    logger.addHandler(handler)

    log_level = getattr(logging, level.upper(), logging.INFO)
    logger.setLevel(log_level)

    return logger


class LoggerAdapter(logging.LoggerAdapter):
    """Adapter to add component field to logs"""

    def __init__(self, logger: logging.Logger, component: str):
        super().__init__(logger, {})
        self.component = component

    def process(self, msg, kwargs):
        msg = f"[{self.component}] {msg}"
        return msg, kwargs
