#!/bin/bash

# Direct Cloudflare AI Integration Test

echo "🧪 Testing Cloudflare AI Integration..."

# Set environment variables directly
export CLOUDFLARE_WORKER_URL="https://obsidian-bot-workers.abdoullahelvogani.workers.dev"
export TURSO_DATABASE_URL="file:./dev.db"
export SESSION_SECRET="change-me-to-something-very-secure"

echo "✅ Environment variables set"
echo "🌐 Cloudflare Worker URL: $CLOUDFLARE_WORKER_URL"

# Start the bot
timeout 10s go run ./cmd/bot/main.go &
BOT_PID=$!

# Wait for startup
sleep 3

# Check if process is still running
if ps -p $BOT_PID > /dev/null; then
    echo "✅ Bot is running successfully (PID: $BOT_PID)"
    
    # Test health endpoint
    echo "🏥 Testing health endpoint..."
    for i in {1..5}; do
        if curl -s http://localhost:8080/health > /dev/null; then
            echo "✅ Health check passed (attempt $i)"
            break
        else
            echo "⏳ Waiting for server to start... (attempt $i/5)"
            sleep 1
        fi
    done
    
    # Test AI integration
    echo "🤖 Testing Cloudflare AI integration..."
    curl -X POST http://localhost:8080/api/ai/generate \
      -H "Content-Type: application/json" \
      -d '{"prompt": "Hello from test script!"}' \
      -s > /tmp/ai-test.json
    
    if grep -q "content" /tmp/ai-test.json; then
        echo "✅ AI integration test passed"
    else
        echo "❌ AI integration test failed"
        cat /tmp/ai-test.json
    fi
    
else
    echo "❌ Bot failed to start"
fi

# Cleanup
kill $BOT_PID 2>/dev/null
rm -f /tmp/ai-test.json

echo ""
echo "🎯 Test Results:"
echo "Environment Variables: ✅"
echo "Bot Startup: $([ $? -eq 0 ] && echo '✅' || echo '❌')"
echo "Health Check: $([ -f /tmp/health-passed ] && echo '✅' || echo '❌')"
echo "AI Integration: $([ -f /tmp/ai-test.json ] && echo '✅' || echo '❌')"