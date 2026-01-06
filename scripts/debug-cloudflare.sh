#!/bin/bash

echo "🧪 Testing Cloudflare AI Provider Directly..."

# Set environment variables
export CLOUDFLARE_WORKER_URL="https://obsidian-bot-workers.abdoullahelvogani.workers.dev"
export ACTIVE_PROVIDER="Cloudflare"

echo "✅ Environment configured"
echo "🌐 Worker URL: $CLOUDFLARE_WORKER_URL"

# Test 1: Direct worker test
echo ""
echo "🔍 Testing worker health..."
curl -s "$CLOUDFLARE_WORKER_URL/health" && echo "✅ Health check passed" || echo "❌ Health check failed"

# Test 2: AI binding test
echo ""
echo "🤖 Testing AI binding..."
response=$(curl -s "$CLOUDFLARE_WORKER_URL/ai-test")
echo "Response: $response"

# Test 3: Simple AI request (if binding works)
echo ""
echo "💬 Testing AI generation..."
ai_response=$(curl -s -X POST "$CLOUDFLARE_WORKER_URL/ai/proxy/cloudflare" \
  -H "Content-Type: text/plain" \
  -d "What is 2+2?")
echo "AI Response: $ai_response"

# Test 4: Check with current bot
echo ""
echo "🚀 Testing with current bot..."
if pgrep -f "bot" > /dev/null; then
    echo "✅ Bot is running"
    
    # Test providers endpoint
    echo "📋 Checking provider status..."
    curl -s http://localhost:8080/api/v1/ai/providers | head -5
    
    # Test AI generation through bot
    echo "💭 Testing AI through bot..."
    test_response=$(curl -s -X POST http://localhost:8080/api/v1/qa \
      -H "Content-Type: application/x-www-form-urlencoded" \
      -d "question=What is 1+1?")
    echo "Bot AI Response: $test_response"
else
    echo "❌ Bot is not running. Start with: ./bot"
fi

echo ""
echo "🎯 Debugging Info:"
echo "  • Cloudflare Worker: $CLOUDFLARE_WORKER_URL"
echo "  • Active Provider: $ACTIVE_PROVIDER"
echo "  • Worker Status: $(curl -s -o /dev/null -w "%{http_code}" "$CLOUDFLARE_WORKER_URL/health")"