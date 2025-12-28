#!/bin/bash
source .env

# Test API directly
echo "Testing Telegram API..."
curl -s "https://api.telegram.org/bot${TELEGRAM_BOT_TOKEN}/getMe" | jq .

# If that works, test with Go
echo ""
echo "Testing Go bot..."
go run main.go processor.go health.go stats.go dedup.go ai_ollama.go
