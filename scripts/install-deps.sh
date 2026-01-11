#!/bin/bash

# Install System Dependencies Script for Obsidian Bot
# This script installs required system dependencies for file processing

set -e

echo "🔧 Installing system dependencies for Obsidian Bot..."

# Update package manager
if command -v apt-get &> /dev/null; then
    echo "Using apt-get (Debian/Ubuntu)"
    sudo apt-get update
    sudo apt-get install -y \
        tesseract-ocr \
        tesseract-ocr-eng \
        poppler-utils \
        ghostscript \
        imagemagick \
        ffmpeg \
        libmagic1 \
        libmagic-dev \
        file

elif command -v yum &> /dev/null; then
    echo "Using yum (RHEL/CentOS/Fedora)"
    sudo yum update -y
    sudo yum install -y \
        tesseract \
        tesseract-langpack-eng \
        poppler-utils \
        ghostscript \
        ImageMagick \
        ffmpeg \
        file-devel \
        file

elif command -v apk &> /dev/null; then
    echo "Using apk (Alpine Linux)"
    sudo apk update
    sudo apk add \
        tesseract-ocr \
        tesseract-ocr-eng \
        poppler-utils \
        ghostscript \
        imagemagick \
        ffmpeg \
        file \
        file-dev

else
    echo "❌ Unsupported package manager. Please install manually:"
    echo "   - tesseract-ocr (OCR engine)"
    echo "   - poppler-utils (PDF tools, includes pdftotext)"
    echo "   - ghostscript (PostScript interpreter)"
    echo "   - imagemagick (Image processing)"
    echo "   - ffmpeg (Media processing)"
    echo "   - file (File type detection)"
    exit 1
fi

echo "✅ System dependencies installed successfully!"

# Verify installations
echo "🔍 Verifying installations..."

commands=("tesseract" "pdftotext" "gs" "convert" "ffmpeg" "file")
all_installed=true

for cmd in "${commands[@]}"; do
    if command -v "$cmd" &> /dev/null; then
        version=$($cmd --version 2>/dev/null | head -n1 || echo "unknown")
        echo "✅ $cmd: $version"
    else
        echo "❌ $cmd: not found"
        all_installed=false
    fi
done

if [ "$all_installed" = true ]; then
    echo "🎉 All dependencies verified successfully!"
else
    echo "⚠️  Some dependencies are missing. Please install them manually."
    exit 1
fi

echo "📝 Installation complete. You can now run the bot with full file processing capabilities."