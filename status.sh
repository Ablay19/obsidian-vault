#!/bin/bash
echo "=== Obsidian Bot Status ==="
echo ""

if docker ps | grep -q obsidian-bot; then
    echo "✅ Bot is RUNNING"
    echo ""
    docker ps --format "table {{.Names}}\t{{.Status}}" | grep obsidian
    echo ""
    
    if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
        echo "✅ Health check: PASSING"
        curl -s http://localhost:8080/health
    else
        echo "❌ Health check: FAILING"
    fi
    
    echo ""
    echo "Recent logs:"
    docker logs obsidian-bot --tail=5
    
    echo ""
    if [ -f stats.json ]; then
        echo "Statistics:"
        cat stats.json
    fi
else
    echo "❌ Bot is NOT running"
    echo "Start with: ./quick-start.sh"
fi
