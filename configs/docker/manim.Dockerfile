# Dockerfile for Manim video rendering
FROM python:3.9-slim

# Install system dependencies for Manim
RUN apt-get update && apt-get install -y \
    ffmpeg \
    texlive-latex-base \
    texlive-fonts-recommended \
    texlive-extra-utils \
    texlive-latex-extra \
    texlive-fonts-extra \
    texlive-xetex \
    fonts-lmodern \
    && rm -rf /var/lib/apt/lists/*

# Install Manim Community
RUN pip install --no-cache-dir manim

# Create working directory
WORKDIR /workspace

# Create media directory for outputs
RUN mkdir -p /workspace/media

# Set Python path
ENV PYTHONPATH=/workspace

# Default command
CMD ["python", "-c", "import manim; print('Manim ready')"]