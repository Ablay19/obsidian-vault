# Google Cloud Shell Setup for Advanced Vision Processing

## Overview

This setup script configures Google Cloud Shell with all necessary tools and dependencies for running the advanced multimodal vision processing system with DeepSeek OCR capabilities.

## Quick Start

```bash
# Clone and run the setup script
git clone https://github.com/Ablay19/obsidian-vault.git
cd obsidian-vault
./setup_gcloud_vision.sh
```

## What Gets Installed

### Core Dependencies
- **Go 1.21.5+** - Primary programming language
- **Tesseract OCR** - Multi-language text recognition
- **ImageMagick** - Image preprocessing and manipulation
- **Node.js 18+** - For worker processes and tooling

### Vision Processing Tools
- **DeepSeek OCR Pipeline** - Advanced document understanding
- **Multimodal Fusion Engine** - Vision-text integration
- **AI Provider Integrations** - DeepSeek, Gemini, OpenAI support

### Development Tools
- **Git** - Version control
- **Build Essentials** - Compilation tools
- **Python3 & pip** - Additional scripting support
- **Google Cloud SDK** - Cloud integration

## Configuration Required

After running the setup script, you need to configure:

### 1. Environment Variables (.env)
```bash
# Required API Keys
DEEPSEEK_API_KEY=your_deepseek_api_key
GEMINI_API_KEY=your_gemini_api_key
OPENAI_API_KEY=your_openai_api_key

# Optional
TELEGRAM_BOT_TOKEN=your_telegram_bot_token
VAULT_ADDR=https://your-vault-server
```

### 2. Vision Settings (config.yaml)
```yaml
vision:
  enabled: true
  encoder_model: "deepseek"  # deepseek, gemini, openai
  fusion_method: "cross_attention"
  min_confidence: 0.6
```

## Usage Commands

Once setup is complete, use these convenient aliases:

```bash
# Build the project
gobuild

# Run the bot
gorun

# View logs
gologs

# Check vision status
govision

# Update and rebuild
goupdate
```

## Testing Setup

Run the included test script to verify everything works:

```bash
./test_vision_setup.sh
```

## Troubleshooting

### Common Issues

1. **Permission Denied**
   ```bash
   chmod +x setup_gcloud_vision.sh
   ```

2. **Go Not Found After Setup**
   ```bash
   source ~/.bashrc
   ```

3. **Build Failures**
   ```bash
   go mod tidy
   go mod download
   ```

4. **API Key Issues**
   - Check .env file exists and has correct keys
   - Verify API keys are valid and have proper permissions

### System Requirements

- **RAM**: 4GB minimum (8GB recommended)
- **Disk**: 10GB free space
- **Network**: Stable internet for AI API calls

## Advanced Configuration

### Custom Vision Models
The system supports custom vision encoders. Add new encoders in `internal/vision/` and update the processor.

### Performance Tuning
Adjust vision settings in `config.yaml`:
- Increase `max_image_size` for higher quality processing
- Lower `min_confidence` for more lenient acceptance
- Change `fusion_method` based on your use case

### Monitoring
Use the built-in dashboard at `http://localhost:8080` for system monitoring and metrics.

---

**Ready to transform document processing with advanced AI vision capabilities! 🚀**