#!/bin/bash
echo "🔄 Restarting bot..."
docker restart obsidian-bot
sleep 2
./status.sh
