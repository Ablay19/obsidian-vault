#!/bin/bash

# Quick Bot Testing Script
# Usage: ./debug/test_bot.sh [command]

echo "🤖 Obsidian Bot Quick Test"
echo "==========================="

case "${1:-help}" in
    "files")
        echo "📁 Available test files:"
        find debug/templates -type f -name "*.png" -o -name "*.pdf" -o -name "*.md" | head -10
        ;;
    "commands")
        echo "📝 Test commands:"
        echo "  /help - Show all commands"
        echo "  /setprovider - Interactive provider selection"
        echo "  /mode - Select processing mode"
        echo "  /bots - Choose bot instance"
        echo "  /process - Process single file"
        echo "  /batch - Process all pending files"
        echo "  /stats - Show usage statistics"
        ;;
    "scenarios")
        echo "🎯 Test scenarios:"
        echo "  1. Send test_ocr.png → /process"
        echo "  2. Send multiple images → /batch"
        echo "  3. Try /setprovider for interactive selection"
        echo "  4. Use /mode for processing modes"
        echo "  5. Try /bots for bot instance selection"
        ;;
    "status")
        echo "🔍 Bot status:"
        if pgrep -f "obsidian.*bot" > /dev/null; then
            echo "✅ Bot is running"
            ps aux | grep -E "obsidian.*bot" | grep -v grep
        else
            echo "❌ Bot is not running"
        fi
        ;;
    "vault")
        echo "📂 Vault contents:"
        ls -la vault/Inbox/ 2>/dev/null || echo "Vault/Inbox not found"
        ;;
    "help"|*)
        echo "Usage: $0 [command]"
        echo ""
        echo "Commands:"
        echo "  files     - List available test files"
        echo "  commands  - Show test commands"
        echo "  scenarios - Show test scenarios"
        echo "  status    - Check if bot is running"
        echo "  vault     - Check vault contents"
        echo "  help      - Show this help"
        ;;
esac