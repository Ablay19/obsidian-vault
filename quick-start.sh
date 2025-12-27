#!/bin/bash
set -e

cd ~/obsidian-automation

if [ ! -f .env ]; then
    echo "ERROR: .env file not found!"
    exit 1
fi

source .env

if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    echo "ERROR: TELEGRAM_BOT_TOKEN not set in .env"
    exit 1
fi

echo "🔨 Building Docker image..."
docker build -t obsidian-bot . || exit 1

echo "🛑 Stopping old container..."
docker stop obsidian-bot 2>/dev/null || true
docker rm obsidian-bot 2>/dev/null || true

echo "🚀 Starting new container..."
docker run -d \
  --name obsidian-bot \
  --restart unless-stopped \
  -e TELEGRAM_BOT_TOKEN="$TELEGRAM_BOT_TOKEN" \
  -e TZ=Africa/Tunis \
  -v "$(pwd)/vault:/app/vault" \
  -v "$(pwd)/attachments:/app/attachments" \
  -v "$(pwd)/stats.json:/app/stats.json" \
  -p 8080:8080 \
  --memory="512m" \
  --log-driver json-file \
  --log-opt max-size=10m \
  --log-opt max-file=3 \
  obsidian-bot

echo ""
echo "✅ Bot started successfully!"
echo ""
echo "Commands:"
echo "  docker logs -f obsidian-bot    # View logs"
echo "  curl localhost:8080/health     # Health check"
echo "  docker restart obsidian-bot    # Restart"
echo ""
