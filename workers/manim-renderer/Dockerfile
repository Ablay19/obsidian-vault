FROM python:3.11-slim

LABEL maintainer="obsidian-vault"
LABEL description="Manim Video Renderer for AI Manim Worker"

WORKDIR /app

RUN apt-get update && apt-get install -y --no-install-recommends \
    ffmpeg \
    libcairo2-dev \
    libpq-dev \
    libffi-dev \
    && rm -rf /var/lib/apt/lists/*

RUN pip install --no-cache-dir manim==0.18.1 \
    && pip install --no-cache-dir flask gunicorn

COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

COPY src/ /app/src/

EXPOSE $PORT

ENV MANIM_COLOR=BLUE \
    MANIM_QUALITY=medium \
    MANIM_PREVIEW=false

CMD ["gunicorn", "--bind", "0.0.0.0:$PORT", "--workers", "2", "--timeout", "300", "src.app:app"]
