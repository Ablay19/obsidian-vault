# Installation & Setup Guide

This guide covers the installation and initial setup of the Mauritania CLI for various platforms and use cases.

## Prerequisites

### System Requirements

- **Operating System**: Android (Termux) or Linux
- **Architecture**: ARM64 (aarch64) or AMD64 (x86_64)
- **Memory**: Minimum 512MB RAM, 1GB recommended
- **Storage**: 50MB free space for installation
- **Network**: Internet connection (can work offline after setup)

### Required Accounts

Choose at least one communication transport:

#### WhatsApp (Recommended)
- WhatsApp account on mobile device
- No additional setup required (uses QR code authentication)

#### Telegram
- Telegram account
- Bot token from [@BotFather](https://t.me/botfather)
- Chat ID for command responses

#### Facebook Messenger
- Facebook Developer account
- Facebook Page
- Messenger API access

#### SM APOS Shipper (Optional - for secure execution)
- SM APOS account
- API credentials
- Active subscription

## Installation Methods

### Method 1: Pre-built Binaries (Recommended)

#### For Termux (Android)
```bash
# Download ARM64 Android binary
wget https://github.com/your-repo/mauritania-cli/releases/download/v1.0.0/mauritania-cli-termux

# Make executable and install
chmod +x mauritania-cli-termux
sudo mv mauritania-cli-termux /data/data/com.termux/files/usr/bin/mauritania-cli

# Verify installation
mauritania-cli --version
```

#### For Linux ARM64
```bash
# Download statically linked binary
wget https://github.com/your-repo/mauritania-cli/releases/download/v1.0.0/mauritania-cli-linux-arm64

# Install globally
chmod +x mauritania-cli-linux-arm64
sudo mv mauritania-cli-linux-arm64 /usr/local/bin/mauritania-cli

# Verify installation
mauritania-cli --version
```

#### For Linux AMD64
```bash
# Download AMD64 binary
wget https://github.com/your-repo/mauritania-cli/releases/download/v1.0.0/mauritania-cli-linux-amd64

# Install and verify
chmod +x mauritania-cli-linux-amd64
sudo mv mauritania-cli-linux-amd64 /usr/local/bin/mauritania-cli
mauritania-cli --version
```

### Method 2: Build from Source

#### Prerequisites
```bash
# Install Go 1.21+
# On Ubuntu/Debian:
sudo apt update && sudo apt install golang

# On Termux:
pkg install golang

# Verify installation
go version
```

#### Clone and Build
```bash
# Clone repository
git clone https://github.com/your-repo/mauritania-cli.git
cd mauritania-cli

# Build for current platform
go build -o mauritania-cli ./cmd/mauritania-cli

# Or build for specific platforms
# Android ARM64 (Termux)
GOOS=android GOARCH=arm64 go build -o mauritania-cli-termux ./cmd/mauritania-cli

# Linux ARM64
GOOS=linux GOARCH=arm64 go build -o mauritania-cli-linux-arm64 ./cmd/mauritania-cli

# Install
sudo mv mauritania-cli /usr/local/bin/
```

### Method 3: Docker Installation

```bash
# Build Docker image
docker build -t mauritania-cli .

# Run container
docker run -it --rm \
  -v ~/.mauritania-cli:/app/data \
  mauritania-cli --help
```

## Initial Configuration

### 1. Initialize Configuration

```bash
# Initialize default configuration
mauritania-cli config init

# This creates ~/.mauritania-cli/config.toml
# with default settings
```

### 2. Choose Primary Transport

#### WhatsApp Setup (Easiest)
```bash
# Start WhatsApp setup
mauritania-cli config whatsapp setup

# Follow prompts:
# 1. Install WhatsMeow dependencies (if building from source)
# 2. Scan QR code with WhatsApp mobile app
# 3. Wait for authentication confirmation
```

#### Telegram Setup (Alternative)
```bash
# Create Telegram bot first
# 1. Message @BotFather on Telegram
# 2. Send /newbot and follow instructions
# 3. Save the bot token

# Configure CLI
mauritania-cli config telegram setup

# Enter your bot token when prompted
# The CLI will guide you through chat ID setup
```

### 3. Configure SM APOS Shipper (Optional)

```bash
# Setup secure command execution
mauritania-cli config shipper setup

# Enter your SM APOS credentials:
# - API endpoint
# - Username
# - Password/API key
```

### 4. Test Configuration

```bash
# Test transport connectivity
mauritania-cli status

# Send test command
mauritania-cli send "echo 'Mauritania CLI is working!'"

# Check command execution
mauritania-cli history --last 1
```

## Advanced Configuration

### Configuration File Location

**Default location**: `~/.mauritania-cli/config.toml`

**Custom location**:
```bash
export MAURITANIA_CLI_CONFIG=/path/to/config.toml
mauritania-cli config init
```

### Manual Configuration

Edit `~/.mauritania-cli/config.toml`:

```toml
[general]
default_transport = "whatsapp"  # whatsapp, telegram, facebook
log_level = "info"             # debug, info, warn, error
offline_queue_size = 1000      # Max queued commands

[transports.social_media.whatsapp]
database_path = "~/.mauritania-cli/whatsapp"
rate_limit = 1000
auto_connect = true

[transports.social_media.telegram]
bot_token = "your_bot_token_here"
chat_id = "your_chat_id_here"
webhook_url = ""  # Optional

[transports.social_media.facebook]
app_id = "your_facebook_app_id"
app_secret = "your_app_secret"
access_token = "your_page_access_token"
page_id = "your_page_id"

[transports.shipper]
base_url = "https://api.shipper.mr"
api_key = "your_shipper_api_key"
timeout = 300

[security]
enable_encryption = true
allowed_commands = ["ls", "pwd", "git", "npm", "echo", "cat", "head", "tail"]
max_command_length = 10000
require_approval = false
```

### Environment Variables

Override configuration with environment variables:

```bash
# Transport settings
export MAURITANIA_CLI_DEFAULT_TRANSPORT=telegram
export MAURITANIA_TELEGRAM_BOT_TOKEN="your_token"
export MAURITANIA_SHIPPER_API_KEY="your_key"

# Behavior settings
export MAURITANIA_CLI_LOG_LEVEL=debug
export MAURITANIA_CLI_OFFLINE_MODE=true

# Run with custom settings
mauritania-cli send "ls -la"
```

## Platform-Specific Setup

### Termux (Android) Setup

#### 1. Install Termux
```bash
# Install from F-Droid or Google Play
# Grant storage permissions
termux-setup-storage
```

#### 2. Install Dependencies
```bash
# Update packages
pkg update && pkg upgrade

# Install essential tools
pkg install git curl wget openssh

# Install Go (for building from source)
pkg install golang
```

#### 3. Install Mauritania CLI
```bash
# Download binary
wget https://github.com/your-repo/mauritania-cli/releases/download/v1.0.0/mauritania-cli-termux
chmod +x mauritania-cli-termux

# Move to PATH
mkdir -p ~/bin
mv mauritania-cli-termux ~/bin/
export PATH="$HOME/bin:$PATH"

# Add to shell profile
echo 'export PATH="$HOME/bin:$PATH"' >> ~/.bashrc
```

#### 4. Configure for Mobile Use
```bash
# Create config optimized for mobile
mauritania-cli config init

# Edit config for mobile settings
nano ~/.mauritania-cli/config.toml

# Add mobile-specific settings:
# - Lower timeout values
# - Smaller queue sizes
# - Battery optimization
```

### Linux Server Setup

#### 1. System Dependencies
```bash
# Ubuntu/Debian
sudo apt update
sudo apt install curl wget git build-essential

# CentOS/RHEL
sudo yum install curl wget git gcc
```

#### 2. Install Go
```bash
# Download and install Go
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```

#### 3. Install Mauritania CLI
```bash
# Build from source
git clone https://github.com/your-repo/mauritania-cli.git
cd mauritania-cli
go build -o mauritania-cli ./cmd/mauritania-cli
sudo mv mauritania-cli /usr/local/bin/
```

### Docker Setup

#### 1. Build Image
```bash
git clone https://github.com/your-repo/mauritania-cli.git
cd mauritania-cli

# Build Docker image
docker build -t mauritania-cli .

# Or use the provided Dockerfile
docker build -f Dockerfile.termux -t mauritania-cli-termux .
```

#### 2. Run Container
```bash
# Run with config persistence
docker run -it --rm \
  -v ~/.mauritania-cli:/app/.mauritania-cli \
  -e MAURITANIA_CLI_CONFIG=/app/.mauritania-cli/config.toml \
  mauritania-cli config init

# Run CLI commands
docker run -it --rm \
  -v ~/.mauritania-cli:/app/.mauritania-cli \
  mauritania-cli send "echo 'Hello from Docker!'"
```

## Verification & Testing

### Basic Functionality Test
```bash
# Test CLI installation
mauritania-cli --version
mauritania-cli --help

# Test configuration
mauritania-cli config show

# Test transport status
mauritania-cli status
```

### Network Connectivity Test
```bash
# Test internet connectivity
curl -I google.com

# Test CLI network detection
mauritania-cli send "echo 'Network test'" --offline=false
```

### Transport Tests
```bash
# Test WhatsApp (if configured)
mauritania-cli send "echo 'WhatsApp test'" --platform whatsapp

# Test Telegram (if configured)
mauritania-cli send "echo 'Telegram test'" --platform telegram

# Test SM APOS Shipper (if configured)
mauritania-cli send "echo 'Shipper test'" --transport shipper
```

## Troubleshooting Installation

### "Exec format error"
```
Cause: Wrong architecture binary
Solution: Rebuild for your platform
GOOS=linux GOARCH=arm64 go build -o mauritania-cli ./cmd/mauritania-cli
```

### "Permission denied"
```
Cause: Binary not executable
Solution: chmod +x mauritania-cli
```

### "Command not found"
```
Cause: Binary not in PATH
Solution: sudo mv mauritania-cli /usr/local/bin/
```

### "No space left on device" (Termux)
```
Cause: Insufficient storage
Solution:
- Clear Termux cache: pkg clean
- Remove unused packages: pkg uninstall <package>
- Use smaller build: go build -ldflags="-s -w" -o mauritania-cli ./cmd/mauritania-cli
```

### Go Build Errors
```
Cause: Missing dependencies or wrong Go version
Solution:
- Check Go version: go version (need 1.21+)
- Clean module cache: go clean -modcache
- Reinitialize modules: rm go.sum && go mod tidy
```

## Post-Installation Tasks

### 1. Set Up Backups
```bash
# Backup configuration
cp ~/.mauritania-cli/config.toml ~/.mauritania-cli/config.toml.backup

# Set up automatic config sync (optional)
# Use git to version control your config
cd ~/.mauritania-cli
git init
git add config.toml
git commit -m "Initial configuration"
```

### 2. Configure Aliases (Optional)
```bash
# Add to ~/.bashrc or ~/.zshrc
alias mc='mauritania-cli'
alias mcs='mauritania-cli send'
alias mcl='mauritania-cli logs show'

# Reload shell
source ~/.bashrc
```

### 3. Set Up Monitoring (Optional)
```bash
# Create systemd service for monitoring (Linux)
sudo tee /etc/systemd/system/mauritania-cli.service > /dev/null <<EOF
[Unit]
Description=Mauritania CLI Monitor
After=network.target

[Service]
Type=simple
User=your-user
ExecStart=/usr/local/bin/mauritania-cli monitor
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl enable mauritania-cli
sudo systemctl start mauritania-cli
```

## Updating

### Update Binary
```bash
# Stop any running instances
pkill -f mauritania-cli

# Download new version
wget https://github.com/your-repo/mauritania-cli/releases/download/v1.1.0/mauritania-cli-termux
chmod +x mauritania-cli-termux

# Replace old binary
sudo mv mauritania-cli-termux /usr/local/bin/mauritania-cli

# Restart services if applicable
sudo systemctl restart mauritania-cli
```

### Update from Source
```bash
cd /path/to/mauritania-cli-source
git pull
go build -o mauritania-cli ./cmd/mauritania-cli
sudo mv mauritania-cli /usr/local/bin/
```

## Security Considerations

### File Permissions
```bash
# Secure configuration directory
chmod 700 ~/.mauritania-cli
chmod 600 ~/.mauritania-cli/config.toml

# Secure database files
chmod 600 ~/.mauritania-cli/*.db
```

### Network Security
- Use HTTPS for webhook URLs
- Rotate API keys regularly
- Monitor for suspicious activity in logs
- Enable command encryption for sensitive operations

### Access Control
- Limit allowed commands in configuration
- Use strong authentication for all transports
- Regularly audit command execution logs
- Implement user-specific permissions

Now your Mauritania CLI is ready for remote development! 🚀