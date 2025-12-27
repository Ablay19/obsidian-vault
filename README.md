# Obsidian Automation Bot

Telegram bot that receives WhatsApp images/PDFs, extracts text via OCR, classifies content, and creates Obsidian notes.

## Quick Start

```bash
# 1. Set your bot token
nano .env  # Add TELEGRAM_BOT_TOKEN=...

# 2. Start the bot
./quick-start.sh

# 3. Send images/PDFs from Telegram
```

## Commands

```bash
./status.sh          # Check bot status
./logs.sh            # View live logs
./restart.sh         # Restart bot
./view-stats.sh      # View statistics
./update.sh          # Rebuild after code changes
./backup-vault.sh    # Backup vault to git
./stop-bot.sh        # Stop bot
```

## Architecture

```
WhatsApp → Forward to Telegram Bot → Docker Container → Obsidian Vault
```

## Features

- ✅ OCR text extraction (Tesseract)
- ✅ PDF text extraction
- ✅ Auto-classification (physics/math/chemistry/admin)
- ✅ Language detection (EN/FR/AR)
- ✅ Duplicate detection (SHA256)
- ✅ Statistics tracking
- ✅ Health monitoring (port 8080)
- ✅ Docker containerized
- ✅ Auto-restart on Cloud Shell wake

## File Structure

```
obsidian-automation/
├── main.go              # Bot handler
├── processor.go         # OCR & classification
├── health.go            # Health endpoint
├── stats.go             # Statistics
├── dedup.go             # Duplicate detection
├── Dockerfile           # Container definition
├── .env                 # Bot token (gitignored)
├── vault/               # Obsidian vault
│   ├── Inbox/          # New notes
│   └── Attachments/    # Files
└── attachments/         # Raw files
```

## Monitoring

- Health: `curl http://localhost:8080/health`
- Logs: `docker logs -f obsidian-bot`
- Stats: `./view-stats.sh`

## Setup Git Backup

```bash
cd vault
git init
git remote add origin https://github.com/YOUR_USERNAME/obsidian-vault.git
git add .
git commit -m "Initial commit"
git push -u origin main
```

## Auto-start on Cloud Shell Boot

Add to `~/.bashrc`:

```bash
if [ -d ~/obsidian-automation ] && ! docker ps | grep -q obsidian-bot; then
    echo "🚀 Starting Obsidian bot..."
    cd ~/obsidian-automation && ./quick-start.sh 2>&1 | grep "✅"
fi
```
