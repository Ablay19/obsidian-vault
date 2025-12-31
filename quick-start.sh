#!/bin/bash
set -e

# Change to the script's directory
cd "$(dirname "$0")"

if [ ! -f .env ]; then
    echo "ERROR: .env file not found!"
    exit 1
fi

source .env

if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    echo "ERROR: TELEGRAM_BOT_TOKEN not set in .env"
    exit 1
fi

if ! command -v docker-compose &> /dev/null; then
    echo "ERROR: docker-compose not found. Please install it to run the bot."
    exit 1
fi

echo "🔨 Building and starting Obsidian Bot with Docker Compose..."
docker-compose up -d --build

echo ""
echo "✅ Bot started successfully!"
echo ""
echo "Commands:"
echo "  docker-compose logs -f         # View logs"
echo "  curl localhost:8080/health     # Health check"
echo "  docker-compose restart         # Restart"
echo "  docker-compose down            # Stop and remove containers"
echo ""
