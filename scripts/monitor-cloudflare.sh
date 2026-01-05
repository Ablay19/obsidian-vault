#!/bin/bash

# Cloudflare AI Monitoring Script

WORKER_URL="https://obsidian-bot-workers.abdoullahelvogani.workers.dev"
LOG_FILE="/var/log/obsidian-bot/ai-metrics.log"

echo "📊 Cloudflare AI Monitoring Dashboard"
echo "================================="

# Health check
echo -n "🏥 Health Status: "
if curl -s "$WORKER_URL/health" | grep -q "OK"; then
    echo "✅ HEALTHY"
else
    echo "❌ UNHEALTHY"
fi

# Response time test
echo -n "⚡ Response Time: "
start_time=$(date +%s%N)
response=$(curl -s -w "%{http_code}" "$WORKER_URL/ai/proxy/cloudflare" \
  -H "Content-Type: text/plain" \
  -d "test")
end_time=$(date +%s%N)
response_time=$(( (end_time - start_time) / 1000000 ))
echo "${response_time}ms"

# Usage check
echo -n "📈 Worker Status: "
status=$(curl -s "$WORKER_URL/status" | jq -r '.status // "unknown"')
echo "$status"

# Log metrics
echo "$(date): Health=$health_check, ResponseTime=$response_time, Status=$status" >> "$LOG_FILE"

echo ""
echo "📋 Quick Actions:"
echo "1. View detailed logs: wrangler tail"
echo "2. Check usage: curl $WORKER_URL/ai-test"  
echo "3. Deploy changes: wrangler deploy"
echo "4. Monitor costs: https://dash.cloudflare.com"