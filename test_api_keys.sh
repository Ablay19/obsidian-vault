#!/bin/bash

echo "🔍 Testing AI Provider API Keys..."
echo "=================================="

# Test Cloudflare Worker (should work)
echo "📡 Testing Cloudflare Worker..."
CLOUDFLARE_URL="https://obsidian-bot-workers.abdoullahelvogani.workers.dev"
if curl -s -X POST "$CLOUDFLARE_URL" \
  -H "Content-Type: application/json" \
  -d '{"message": "test", "stream": false}' | grep -q "response\|content"; then
  echo "✅ Cloudflare Worker: HEALTHY"
else
  echo "❌ Cloudflare Worker: FAILED"
fi

# Test Gemini API
echo "🧠 Testing Gemini API..."
GEMINI_KEY=$(grep GEMINI_API_KEY .env | cut -d'=' -f2 | tr -d '"')
if [ -n "$GEMINI_KEY" ] && [ "$GEMINI_KEY" != "your-gemini-api-key" ]; then
  if curl -s -H "Content-Type: application/json" \
    -H "x-goog-api-key: $GEMINI_KEY" \
    -d '{"contents":[{"parts":[{"text":"Hello"}]}]}' \
    "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash-latest:generateContent" | grep -q "candidates\|content"; then
    echo "✅ Gemini API: HEALTHY"
  else
    echo "❌ Gemini API: FAILED (invalid key or network issue)"
  fi
else
  echo "⚠️  Gemini API: NOT CONFIGURED"
fi

# Test Groq API
echo "⚡ Testing Groq API..."
GROQ_KEY=$(grep GROQ_API_KEY .env | cut -d'=' -f2 | tr -d '"')
if [ -n "$GROQ_KEY" ] && [ "$GROQ_KEY" != "your-groq-api-key" ]; then
  if curl -s -H "Content-Type: application/json" \
    -H "Authorization: Bearer $GROQ_KEY" \
    -d '{"model":"llama-3.1-8b-instant","messages":[{"role":"user","content":"Hello"}],"max_tokens":10}' \
    "https://api.groq.com/v1/chat/completions" | grep -q "choices\|content"; then
    echo "✅ Groq API: HEALTHY"
  else
    echo "❌ Groq API: FAILED (invalid key or network issue)"
  fi
else
  echo "⚠️  Groq API: NOT CONFIGURED"
fi

# Test OpenRouter API
echo "🛣️  Testing OpenRouter API..."
OPENROUTER_KEY=$(grep OPENROUTER_API_KEY .env | cut -d'=' -f2 | tr -d '"')
if [ -n "$OPENROUTER_KEY" ] && [ "$OPENROUTER_KEY" != "your-openrouter-api-key" ]; then
  if curl -s -H "Content-Type: application/json" \
    -H "Authorization: Bearer $OPENROUTER_KEY" \
    -H "HTTP-Referer: http://localhost:8080" \
    -H "X-Title: WhatsAppToObsidian" \
    -d '{"model":"openai/gpt-3.5-turbo","messages":[{"role":"user","content":"Hello"}],"max_tokens":10}' \
    "https://openrouter.ai/api/v1/chat/completions" | grep -q "choices\|content"; then
    echo "✅ OpenRouter API: HEALTHY"
  else
    echo "❌ OpenRouter API: FAILED (invalid key or network issue)"
  fi
else
  echo "⚠️  OpenRouter API: NOT CONFIGURED"
fi

# Test HuggingFace API
echo "🤗 Testing HuggingFace API..."
HF_KEY=$(grep HUGGINGFACE_API_KEY .env | cut -d'=' -f2 | tr -d '"')
if [ -n "$HF_KEY" ] && [ "$HF_KEY" != "your-huggingface-api-key" ]; then
  if curl -s -H "Content-Type: application/json" \
    -H "Authorization: Bearer $HF_KEY" \
    -d '{"inputs":"Hello","parameters":{"max_tokens":10}}' \
    "https://api-inference.huggingface.co/models/MiniMaxAI/MiniMax-M2.1:novita" | grep -q "generated_text\|0"; then
    echo "✅ HuggingFace API: HEALTHY"
  else
    echo "❌ HuggingFace API: FAILED (invalid key or network issue)"
  fi
else
  echo "⚠️  HuggingFace API: NOT CONFIGURED"
fi

echo "=================================="
echo "🎯 Test completed. Check results above."