#!/bin/bash

# Debug Templates for Obsidian Bot Testing
# This script provides ready-to-use test files and commands for debugging

set -e

echo "🧪 Obsidian Bot Debug Templates"
echo "================================"

# Create test directory
mkdir -p debug/templates/images
mkdir -p debug/templates/pdfs
mkdir -p debug/templates/messages

echo "📁 Created debug directories"

# Create a simple test image with text (if ImageMagick is available)
if command -v convert &> /dev/null; then
    echo "🎨 Creating test images..."
    convert -size 800x600 xc:white -pointsize 72 -fill black -gravity center -annotate +0+0 "TEST IMAGE\nSample Document\nFor OCR Testing" debug/templates/images/test_ocr.png
    convert -size 600x400 xc:lightblue -pointsize 48 -fill darkblue -gravity center -annotate +0+0 "INVOICE\nAmount: \$123.45\nDate: 2024-01-11" debug/templates/images/test_invoice.png
    convert -size 500x700 xc:lightgray -pointsize 36 -fill black -gravity north -annotate +0+50 "RESEARCH PAPER\n\nAbstract: This is a sample research paper for testing document processing capabilities.\n\nKeywords: AI, OCR, Document Processing\n\nConclusion: The system successfully processes various document types." debug/templates/images/test_paper.png
    echo "✅ Created test images"
else
    echo "⚠️  ImageMagick not available, skipping image creation"
fi

# Create sample PDF content (skip PDF generation for now)
echo "📄 Creating test PDF templates (content only)..."

# Create markdown content for PDFs (can be converted later when dependencies are available)
cat > debug/templates/sample_invoice.md << 'EOF'
# Invoice #12345

**Date:** January 11, 2024
**Due Date:** January 25, 2024

## Bill To:
John Doe
123 Main Street
Anytown, USA 12345

## Items:

| Description | Quantity | Unit Price | Total |
|-------------|----------|------------|-------|
| Consulting Services | 10 hours | $50.00 | $500.00 |
| Document Processing | 5 pages | $10.00 | $50.00 |

**Subtotal:** $550.00
**Tax (10%):** $55.00
**Total:** $605.00

Thank you for your business!
EOF

cat > debug/templates/sample_report.md << 'EOF'
# Monthly Sales Report

## Executive Summary

This report analyzes sales performance for January 2024. Overall sales increased by 15% compared to the previous month, driven primarily by digital product sales.

## Key Metrics

- **Total Revenue:** $125,000
- **Units Sold:** 2,340
- **Average Order Value:** $53.42
- **Conversion Rate:** 3.2%

## Top Performing Products

1. **AI Processing Tool** - $45,000 (36% of total)
2. **Document Scanner** - $28,000 (22% of total)
3. **Data Analytics Suite** - $32,000 (26% of total)

## Recommendations

1. Increase marketing spend on top-performing products
2. Expand AI processing tool feature set
3. Optimize conversion funnel for mobile users

## Next Steps

- Schedule follow-up meeting with sales team
- Update product roadmap based on customer feedback
- Prepare Q1 strategy presentation
EOF

echo "✅ Created test PDF templates (Markdown format)"

# Create test message templates
cat > debug/templates/messages/test_commands.txt << 'EOF'
Test Commands for Obsidian Bot:

1. Basic Commands:
   /help - Show all available commands
   /stats - Show bot usage statistics

2. Provider Management:
   /setprovider - Change AI provider (interactive)
   /mode - Select processing mode
   /bots - Choose bot instance

3. File Processing:
   Send image or PDF, then:
   /process - Process single file
   /batch - Process all pending files

4. History:
   /last - Show last created note
   /reprocess - Reprocess last file

5. Conversation:
   /clear - Clear conversation history

Test Messages:
- "Hello" - Basic AI response
- "What is OCR?" - Technical question
- "Summarize this document" - With attached file
EOF

cat > debug/templates/messages/test_scenarios.txt << 'EOF'
Test Scenarios:

1. Single Image Processing:
   - Send test_ocr.png
   - Use /process
   - Verify note creation in vault/Inbox/

2. Batch Processing:
   - Send multiple images
   - Use /batch
   - Check progress indicators
   - Verify all notes created

3. PDF Processing:
   - Send test_report.pdf
   - Use /process
   - Check text extraction quality

4. Error Handling:
   - Send unsupported file type
   - Check error messages
   - Verify graceful failure

5. Interactive Commands:
   - /setprovider - Test provider selection
   - /mode - Test processing mode selection
   - /bots - Test bot instance selection

6. Conversation Features:
   - Multiple messages
   - /clear command
   - Context persistence
EOF

echo "📝 Created test message templates"

# Create a quick test script
cat > debug/test_bot.sh << 'EOF'
#!/bin/bash

# Quick Bot Testing Script
# Usage: ./debug/test_bot.sh [command]

echo "🤖 Obsidian Bot Quick Test"
echo "==========================="

case "${1:-help}" in
    "files")
        echo "📁 Available test files:"
        find debug/templates -type f -name "*.png" -o -name "*.pdf" | head -10
        ;;
    "commands")
        echo "📝 Test commands:"
        cat debug/templates/messages/test_commands.txt
        ;;
    "scenarios")
        echo "🎯 Test scenarios:"
        cat debug/templates/messages/test_scenarios.txt
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
    "logs")
        echo "📋 Recent logs:"
        tail -20 /tmp/bot.log 2>/dev/null || echo "No log file found"
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
        echo "  logs      - Show recent bot logs"
        echo "  vault     - Check vault contents"
        echo "  help      - Show this help"
        ;;
esac
EOF

chmod +x debug/test_bot.sh

echo "🛠️  Created test script: debug/test_bot.sh"
echo ""
echo "🎉 Debug templates ready!"
echo ""
echo "Quick start:"
echo "  ./debug/test_bot.sh files     # See available test files"
echo "  ./debug/test_bot.sh commands  # See test commands"
echo "  ./debug/test_bot.sh scenarios # See test scenarios"
echo ""
echo "Test files location: debug/templates/"
echo "  - Images: debug/templates/images/"
echo "  - PDFs: debug/templates/pdfs/"
echo "  - Messages: debug/templates/messages/"