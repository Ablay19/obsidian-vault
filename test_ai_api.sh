#!/bin/bash

# This script sends a message to the bot, which in turn will call the active AI provider.
# You need to observe the bot's response in Telegram to verify that the AI API call was successful.

# You can set the bot's token and chat ID here
# TELEGRAM_BOT_TOKEN="your-token"
# CHAT_ID="your-chat-id"

if [ -z "$TELEGRAM_BOT_TOKEN" ] || [ -z "$CHAT_ID" ]; then
    echo "Please set TELEGRAM_BOT_TOKEN and CHAT_ID environment variables."
    exit 1
fi

MESSAGE="Hello, this is a test message to check the AI API call."

curl -s -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/sendMessage" \
    -d "chat_id=$CHAT_ID" \
    -d "text=$MESSAGE"
