#!/bin/bash

# This script tests direct API calls to Gemini and Groq.
# IMPORTANT: You need to export your API keys before running this script:
# export GEMINI_API_KEY="YOUR_GEMINI_API_KEY"
# export GROQ_API_KEY="YOUR_GROQ_API_KEY"

# --- Gemini API Test ---
echo "Testing Gemini API..."
if [ -z "$GEMINI_API_KEY" ]; then
    echo "GEMINI_API_KEY environment variable is not set. Please export it before running this script."
else
    curl -s -X POST "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.5-flash:generateContent?key=$GEMINI_API_KEY" \
        -H "Content-Type: application/json" \
        -d '{
            "contents": [{"parts": [{"text": "Explain the importance of fast language models"}]}]}
' | jq .
fi

echo ""
echo "--------------------"
echo ""

# --- Groq API Test ---
echo "Testing Groq API..."
if [ -z "$GROQ_API_KEY" ]; then
    echo "GROQ_API_KEY environment variable is not set. Please export it before running this script."
else
    curl -s -X POST "https://api.groq.com/openai/v1/chat/completions" \
        -H "Authorization: Bearer $GROQ_API_KEY" \
        -H "Content-Type: application/json" \
        -d '{
            "messages": [
                {
                    "role": "user",
                    "content": "Explain the importance of fast language models"
                }
            ],
            "model": "llama-3.1-8b-instant"
        }' | jq .
fi
