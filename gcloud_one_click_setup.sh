#!/bin/bash
# One-click setup script for Google Cloud Shell ephemeral mode
# Clone, setup, and run the advanced vision processing bot

set -e

echo "🚀 One-Click Vision Bot Setup for Google Cloud Shell"
echo "==================================================="

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

log() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; }

# Install Tesseract OCR
log "Installing Tesseract OCR..."
sudo apt-get install -y -qq tesseract-ocr tesseract-ocr-eng tesseract-ocr-fra tesseract-ocr-ara tesseract-ocr-deu tesseract-ocr-spa

# Install ImageMagick
log "Installing ImageMagick..."
sudo apt-get install -y -qq imagemagick

# Clone repository
log "Cloning repository..."
if [[ -d "obsidian-vault" ]]; then
    cd obsidian-vault && git pull origin main
else
    git clone https://github.com/Ablay19/obsidian-vault.git
    cd obsidian-vault
fi

# Setup project
log "Setting up project..."
go mod download

# Build the bot
log "Building the bot..."
go build -o bin/bot ./cmd/bot/

# Create environment file template
if [[ ! -f ".env" ]]; then
    cat > .env << 'EOF'
# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=your_telegram_bot_token_here

# AI Provider API Keys (configure at least one)
GEMINI_API_KEY=your_gemini_api_key_here
DEEPSEEK_API_KEY=your_deepseek_api_key_here
OPENAI_API_KEY=your_openai_api_key_here

# Optional: Vault Configuration
VAULT_ADDR=https://your-vault-server
VAULT_TOKEN=your_vault_token

# Optional: Database
DATABASE_URL=sqlite:///tmp/obsidian.db
EOF
    warning "Environment file created. Please edit .env with your API keys!"
fi

# Create necessary directories
mkdir -p vault/Inbox logs attachments

# Final instructions
echo ""
echo "╔══════════════════════════════════════════════════════════════════════════════╗"
echo "║                          🎉 SETUP COMPLETE! 🎉                              ║"
echo "╠══════════════════════════════════════════════════════════════════════════════╣"
echo "║                                                                              ║"
echo "║  Advanced Vision Processing Bot Ready!                                       ║"
echo "║                                                                              ║"
echo "║  Next Steps:                                                                 ║"
echo "║  1. Edit API keys: vim .env                                                  ║"
echo "║  2. Start the bot: ./bin/bot                                                 ║"
echo "║                                                                              ║"
echo "║  Features Available:                                                         ║"
echo "║  ✅ DeepSeek OCR Pipeline                                                    ║"
echo "║  ✅ Multimodal Vision Processing                                             ║"
echo "║  ✅ Progress Bars & Real-time Updates                                        ║"
echo "║  ✅ Enterprise Document Understanding                                        ║"
echo "║                                                                              ║"
echo "╚══════════════════════════════════════════════════════════════════════════════╝"
echo ""

success "Ready to process images with advanced AI vision! 🚀"
