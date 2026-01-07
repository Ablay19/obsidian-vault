#!/bin/bash

# Load environment variables
if [ -f ".env" ]; then
    export $(grep -v '^#' .env | xargs)
fi

echo "🚀 Testing Cloudflare Workers AI Configuration..."

# Test 1: Environment variable check
echo "📋 Checking environment variables..."
if [ -z "$CLOUDFLARE_WORKER_URL" ]; then
    echo "❌ CLOUDFLARE_WORKER_URL not set"
    exit 1
else
    echo "✅ CLOUDFLARE_WORKER_URL: $CLOUDFLARE_WORKER_URL"
fi

# Test 2: URL format validation
if [[ "$CLOUDFLARE_WORKER_URL" =~ workers\.dev ]]; then
    echo "✅ URL format looks valid"
else
    echo "⚠️  URL might not be a Cloudflare Workers URL"
fi

# Test 3: Worker health check
echo "🏥 Testing worker health..."
if curl -s "$CLOUDFLARE_WORKER_URL/health" > /dev/null 2>&1; then
    echo "✅ Worker is responding"
else
    echo "⚠️  Worker not responding (might not be deployed)"
fi

# Test 4: Test AI binding
echo "🤖 Testing AI binding..."
response=$(curl -s "$CLOUDFLARE_WORKER_URL/ai-test" 2>/dev/null)
if echo "$response" | grep -q "hasAIBinding"; then
    echo "✅ AI binding check passed"
    has_ai=$(echo "$response" | grep -o '"hasAIBinding":[^,]*' | cut -d':' -f2)
    echo "   AI Binding Available: $has_ai"
else
    echo "❌ AI binding check failed"
fi

echo ""
echo "🎯 Next Steps:"
echo "1. If worker is not responding, deploy it:"
echo "   cd workers/ai-proxy && wrangler deploy"
echo ""
echo "2. Update your .env file with the correct worker URL"
echo "3. Restart the bot: ./bot"
echo "4. Test with: /ask What is Cloudflare Workers?"

exit 0