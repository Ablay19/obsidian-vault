# 🚀 Obsidian Bot - Production Deployment Guide

A powerful Telegram bot that processes images and PDFs into organized Obsidian notes using AI.

## ✨ Features

- **Multi-Strategy OCR** - 7+ processing algorithms for reliable text extraction
- **AI-Powered Analysis** - Automatic categorization, summarization, and insights
- **Interactive Commands** - Keyboard-based controls with progress indicators
- **Batch Processing** - Simultaneous processing with real-time progress
- **Google Gemini & DeepSeek Integration** - Free, powerful AI providers with excellent image processing
- **Turso Database** - Cloud database with automatic migrations
- **Auto-Documented Commands** - Commands automatically appear in Telegram menu

## 🛠️ Quick Start

### 1. Automatic Setup
```bash
# Run the automatic environment setup
./scripts/setup-environment.sh
```

This will:
- ✅ Prompt for required configuration
- ✅ Create `.env` file with all settings
- ✅ Set up required directories
- ✅ Generate debug templates
- ✅ Configure Google Gemini (free AI)

### 2. Verify Deployment Readiness
```bash
# Check if everything is configured correctly
./scripts/check-deployment.sh
```

### 3. Start the Bot
```bash
# Start in production mode
GIN_MODE=release ./bot
```

## 📋 Manual Configuration (Alternative)

If you prefer manual setup:

### Required Environment Variables

Create a `.env` file with:

```bash
# Required
TELEGRAM_BOT_TOKEN=your_bot_token_from_botfather
TURSO_DATABASE_URL=your_turso_database_url

# AI Provider (at least one required)
GEMINI_API_KEY=your_google_gemini_api_key

# Optional
TURSO_AUTH_TOKEN=your_turso_auth_token
SESSION_SECRET=auto_generated_secure_string
REDIS_ADDR=localhost:6379
```

### Get API Keys

1. **Telegram Bot Token**: Message `@BotFather` on Telegram
2. **Google Gemini**: https://makersuite.google.com/app/apikey (FREE)
3. **DeepSeek**: https://platform.deepseek.com/ (FREE - Excellent for image processing)
4. **Turso Database**: https://turso.tech (Free tier available)

## 🎯 Bot Commands

The bot automatically registers these commands with Telegram:

### File Processing
- `/process` - Process single file with AI analysis
- `/batch` - Process all pending files simultaneously
- `/reprocess` - Reprocess the last processed file

### AI Configuration
- `/setprovider` - Change AI provider (Gemini, Groq, etc.)
- `/mode` - Select processing mode (Fast/Quality/Conservative/Experimental)
- `/bots` - Choose bot instance (for multi-bot deployments)

### Information
- `/help` - Show all available commands and features
- `/stats` - Show usage statistics
- `/last` - Show the last created note

## 🔧 Advanced Configuration

### config.yaml

The bot uses `config.yaml` for advanced settings:

```yaml
providers:
  gemini:
    model: gemini-pro
  google:
    model: gemini-1.5-flash

provider_profiles:
  gemini:
    input_cost_per_token: 0.000001
    latency_ms_threshold: 500
  google:
    input_cost_per_token: 0.000001
    latency_ms_threshold: 300

switching_rules:
  default_provider: gemini
  latency_target: 5000
```

## 🧪 Testing & Debugging

### Test Files
The bot includes ready test files:
```bash
# See available test files
./debug/test_bot.sh files

# Get testing scenarios
./debug/test_bot.sh scenarios
```

### Test Images Available:
- `debug/templates/images/test_ocr.png` - Simple text OCR
- `debug/templates/images/test_invoice.png` - Invoice with tables
- `debug/templates/images/test_paper.png` - Research paper

## 📊 Architecture

### Processing Pipeline
1. **File Upload** → Telegram → Bot receives file
2. **Multi-Strategy OCR** → 7 algorithms try different approaches
3. **Quality Scoring** → Best result selected automatically
4. **AI Analysis** → Categorization, summarization, topics, questions
5. **Note Creation** → Markdown file in `vault/Inbox/`
6. **User Notification** → Success with file link

### Database Schema
- **processed_files**: File processing history and metadata
- **chat_history**: Conversation logs
- **vector_documents**: RAG embeddings for semantic search

## 🔒 Security

- **Environment Variables**: All secrets stored in `.env`
- **Session Management**: Secure encrypted sessions
- **API Key Protection**: Keys never logged or exposed
- **Database Security**: Turso provides enterprise security

## 🚀 Production Deployment

### System Requirements
- Go 1.21+
- Tesseract OCR
- Poppler utils (pdftotext)
- ImageMagick
- Pandoc (optional, for PDF generation)

### Docker Deployment
```bash
# Build container
docker build -t obsidian-bot .

# Run with environment
docker run -p 8080:8080 --env-file .env obsidian-bot
```

### Cloud Deployment Options
- **Railway**: Automatic deployments from Git
- **Render**: Docker-based deployments
- **Fly.io**: Global deployment with volumes
- **Kubernetes**: Helm charts available

## 📈 Monitoring & Analytics

The bot includes built-in monitoring:
- **Real-time metrics** via dashboard
- **Usage statistics** per command/provider
- **Performance tracking** for AI calls
- **Error logging** with detailed context

Access the dashboard at: `http://localhost:8080` (when running)

## 🆘 Troubleshooting

### Common Issues

**Bot not responding:**
```bash
# Check if running
./debug/test_bot.sh status

# Check logs
tail -f /tmp/bot.log
```

**OCR quality poor:**
- The bot automatically tries multiple strategies
- Check that Tesseract and ImageMagick are installed
- Use test images: `./debug/test_bot.sh files`

**Database connection failed:**
```bash
# Reset local database
./debug/reset_db.sh

# Check deployment
./scripts/check-deployment.sh
```

## 🤝 Contributing

### Development Setup
```bash
# Clone and setup
git clone <repository>
cd obsidian-vault

# Automatic setup
./scripts/setup-environment.sh

# Development mode
go run ./cmd/bot
```

### Testing
```bash
# Run all tests
go test ./...

# Integration tests
go test ./tests/integration/...

# Debug testing
./debug/test_bot.sh scenarios
```

## 📄 License

This project is open source. See LICENSE file for details.

---

## 🎉 Ready to Deploy!

Your Obsidian bot is now fully configured with:

✅ **Turso Database** - Cloud database restored  
✅ **Google Gemini** - Free AI provider configured  
✅ **No Hardcoded Values** - Everything environment-based  
✅ **Auto-Documented Commands** - Telegram menu automatically populated  
✅ **Multi-Strategy Processing** - Reliable OCR with fallbacks  
✅ **Interactive Controls** - Keyboard-based user interface  
✅ **Production Ready** - Monitoring, logging, error handling  

**🚀 Deploy with confidence!**