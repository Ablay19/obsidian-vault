#!/bin/bash

set -e

echo "🚀 Deploying AI Manim Worker..."

# Check prerequisites
if ! command -v wrangler &> /dev/null; then
  echo "❌ Wrangler not found. Install with: npm install -g wrangler"
  exit 1
fi

if ! command -v docker &> /dev/null; then
  echo "❌ Docker not found. Install Docker for renderer deployment."
  exit 1
fi

# Deploy Worker
echo "📦 Deploying Worker..."
cd "$(dirname "$0")/.."

echo "Running wrangler deploy..."
npx wrangler deploy

if [ $? -eq 0 ]; then
  echo "✅ Worker deployed successfully"
else
  echo "❌ Worker deployment failed"
  exit 1
fi

# Deploy Renderer (if applicable)
if [ -f "manim-renderer/Dockerfile" ]; then
  echo "🎬 Deploying Manim Renderer..."
  cd manim-renderer

  # Check if container registry is configured
  if [ -n "$CONTAINER_REGISTRY" ]; then
    TAG="${CONTAINER_REGISTRY}/manim-renderer:latest"
    echo "Building and pushing image: $TAG"
    docker build -t "$TAG" .
    docker push "$TAG"

    if [ $? -eq 0 ]; then
      echo "✅ Renderer image deployed to $CONTAINER_REGISTRY"
    else
      echo "❌ Renderer image deployment failed"
      exit 1
    fi
  else
    echo "⚠️  No container registry configured. Skipping renderer deployment."
    echo "   Set CONTAINER_REGISTRY environment variable to enable."
  fi
fi

echo "🎉 Deployment complete!"
echo ""
echo "Worker URL: https://ai-manim-worker.abdoullahelvogani.workers.dev"
echo ""
echo "Next steps:"
echo "1. Set up Telegram webhook:"
echo "   curl -F 'url=https://ai-manim-worker.abdoullahelvogani.workers.dev/telegram/webhook' \\"
echo "     -H 'X-Telegram-Bot-Api-Secret-Token: YOUR_SECRET' \\"
echo "     https://api.telegram.org/bot\$TELEGRAM_BOT_TOKEN/setWebhook"
echo ""
echo "2. Test the webhook:"
echo "   Send a message to your bot: @YourBotName"
