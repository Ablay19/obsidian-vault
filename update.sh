#!/bin/bash
set -e
echo "🔄 Updating bot..."
docker stop obsidian-bot 2>/dev/null || true
docker rm obsidian-bot 2>/dev/null || true
./quick-start.sh
echo "✅ Update complete!"
