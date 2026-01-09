#!/bin/bash

# API Key Testing Script
# Usage: ./test_real_api_keys.sh

echo "🔍 Testing Real AI Provider API Keys..."
echo "========================================"

# Source the Doppler environment file
if [ -f ".env.doppler" ]; then
    export $(cat .env.doppler | grep -v '^#' | xargs)
    echo "✅ Loaded environment from .env.doppler"
else
    echo "❌ .env.doppler file not found"
    exit 1
fi

# Test Cloudflare Worker (should work)
echo "📡 Testing Cloudflare Worker..."
CLOUDFLARE_URL="$CLOUDFLARE_WORKER_URL"
if curl -s -X GET "$CLOUDFLARE_URL/health" | grep -q "OK"; then
    echo "✅ Cloudflare Worker: HEALTHY"
else
    echo "❌ Cloudflare Worker: FAILED"
fi

# Test Gemini API
echo "🧠 Testing Gemini API..."
if [ -n "$GEMINI_API_KEY" ] && [ "$GEMINI_API_KEY" != "ACTUAL_GEMINI_API_KEY_HERE" ]; then
    if curl -s -H "Content-Type: application/json" \
        -H "x-goog-api-key: $GEMINI_API_KEY" \
        -d '{"contents":[{"parts":[{"text":"Hello"}]}]}' \
        "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent" | grep -q "candidates\|content"; then
        echo "✅ Gemini API: HEALTHY"
    else
        echo "❌ Gemini API: FAILED (invalid key or network issue)"
    fi
else
    echo "⚠️  Gemini API: NOT CONFIGURED (placeholder key)"
fi

# Test Groq API
echo "⚡ Testing Groq API..."
if [ -n "$GROQ_API_KEY" ] && [ "$GROQ_API_KEY" != "ACTUAL_GROQ_API_KEY_HERE" ]; then
    if curl -s -H "Content-Type: application/json" \
        -H "Authorization: Bearer $GROQ_API_KEY" \
        -d '{"model":"llama-3.1-8b-instant","messages":[{"role":"user","content":"Hello"}],"max_tokens":10}' \
        "https://api.groq.com/v1/chat/completions" | grep -q "choices\|content"; then
        echo "✅ Groq API: HEALTHY"
    else
        echo "❌ Groq API: FAILED (invalid key or network issue)"
    fi
else
    echo "⚠️  Groq API: NOT CONFIGURED (placeholder key)"
fi

# Test OpenRouter API
echo "🛣️  Testing OpenRouter API..."
if [ -n "$OPENROUTER_API_KEY" ] && [ "$OPENROUTER_API_KEY" != "ACTUAL_OPENROUTER_API_KEY_HERE" ]; then
    if curl -s -H "Content-Type: application/json" \
        -H "Authorization: Bearer $OPENROUTER_API_KEY" \
        -H "HTTP-Referer: http://localhost:8080" \
        -H "X-Title: WhatsAppToObsidian" \
        -d '{"model":"openai/gpt-3.5-turbo","messages":[{"role":"user","content":"Hello"}],"max_tokens":10}' \
        "https://openrouter.ai/api/v1/chat/completions" | grep -q "choices\|content"; then
        echo "✅ OpenRouter API: HEALTHY"
    else
        echo "❌ OpenRouter API: FAILED (invalid key or network issue)"
    fi
else
    echo "⚠️  OpenRouter API: NOT CONFIGURED (placeholder key)"
fi

# Test HuggingFace API
echo "🤗 Testing HuggingFace API..."
if [ -n "$HUGGINGFACE_API_KEY" ] && [ "$HUGGINGFACE_API_KEY" != "ACTUAL_HUGGINGFACE_API_KEY_HERE" ]; then
    if curl -s -H "Content-Type: application/json" \
        -H "Authorization: Bearer $HUGGINGFACE_API_KEY" \
        -d '{"inputs":"Hello","parameters":{"max_tokens":10}}' \
        "https://api-inference.huggingface.co/models/MiniMaxAI/MiniMax-M2.1:novita" | grep -q "generated_text\|0"; then
        echo "✅ HuggingFace API: HEALTHY"
    else
        echo "❌ HuggingFace API: FAILED (invalid key or network issue)"
    fi
else
    echo "⚠️  HuggingFace API: NOT CONFIGURED (placeholder key)"
fi

echo "========================================"
echo "🎯 Test completed. Replace placeholder keys with real API keys to test all providers."
echo ""
echo "📝 To get API keys:"
echo "   Gemini: https://makersuite.google.com/app/apikey"
echo "   Groq:   https://console.groq.com/keys"
echo "   OpenRouter: https://openrouter.ai/keys"
echo "   HuggingFace: https://huggingface.co/settings/tokens"