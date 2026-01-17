# Mauritania CLI - Network Integration for Remote Development

> **Empowering developers in low-connectivity regions with seamless remote development capabilities**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![Platform](https://img.shields.io/badge/Platform-Termux%20%7C%20Linux-orange.svg)](https://termux.dev/)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## 🌟 Overview

The **Mauritania CLI** is a revolutionary command-line interface designed specifically for developers in regions with limited internet connectivity. It enables seamless remote development by leveraging social media platforms and secure network providers as communication channels.

### 🎯 Key Features

- **📱 Social Media Integration** - Execute commands via WhatsApp, Telegram, and Facebook
- **🔒 Secure Remote Execution** - SM APOS Shipper integration for encrypted command execution
- **📶 Offline-First Design** - Intelligent queuing and retry when connectivity returns
- **🏗️ Mobile-Optimized** - Built specifically for Termux on Android devices
- **🔄 Multi-Transport Support** - Automatic failover between communication channels
- **⚡ Real-time Execution** - Live command output and status monitoring

### 🌍 Why Mauritania CLI?

Traditional remote development requires stable, high-speed internet connections. In many parts of the world, including Mauritania, this simply isn't available. The Mauritania CLI transforms everyday communication tools into powerful development interfaces, enabling:

- **Remote server management** without stable internet
- **Development workflows** using familiar messaging apps
- **Secure command execution** through trusted network providers
- **Collaborative development** in low-connectivity environments

## 🚀 Quick Start

### Prerequisites

- **Go 1.21+** installed
- **Termux** (recommended) or Linux environment
- **Active social media accounts** (WhatsApp, Telegram, or Facebook)

### Installation

#### Option 1: Pre-built Binary (Recommended)

```bash
# Download for your platform
# ARM64 Linux (Termux)
wget https://github.com/your-repo/mauritania-cli/releases/download/v1.0.0/mauritania-cli-linux-arm64
chmod +x mauritania-cli-linux-arm64
sudo mv mauritania-cli-linux-arm64 /usr/local/bin/mauritania-cli

# ARM64 Android (Termux direct)
wget https://github.com/your-repo/mauritania-cli/releases/download/v1.0.0/mauritania-cli-termux
chmod +x mauritania-cli-termux
```

#### Option 2: Build from Source

```bash
git clone https://github.com/your-repo/mauritania-cli.git
cd mauritania-cli

# Build for Termux
GOOS=android GOARCH=arm64 go build -o mauritania-cli-termux ./cmd/mauritania-cli

# Or build for Linux
GOOS=linux GOARCH=arm64 go build -o mauritania-cli ./cmd/mauritania-cli
```

### Basic Setup

```bash
# Initialize configuration
mauritania-cli config init

# Setup your preferred transport
mauritania-cli config telegram setup
# OR
mauritania-cli config whatsapp setup
# OR
mauritania-cli config facebook setup

# Configure SM APOS Shipper for secure execution
mauritania-cli config shipper setup
```

### First Command

```bash
# Send a simple command
mauritania-cli send "echo 'Hello from Mauritania CLI!'"

# Check command status
mauritania-cli status

# View queued commands
mauritania-cli queue list
```

## 📖 Usage Guide

### Command Structure

```bash
mauritania-cli [command] [subcommand] [flags]
```

### Core Commands

#### Send Commands
```bash
# Send command via default transport
mauritania-cli send "git status"

# Specify transport explicitly
mauritania-cli send "npm install" --transport telegram

# Force offline mode (queue for later)
mauritania-cli send "docker build ." --offline=true

# Set priority (normal, high, urgent)
mauritania-cli send "systemctl restart nginx" --priority high
```

#### Configuration Management
```bash
# Initialize configuration
mauritania-cli config init

# Setup transports
mauritania-cli config whatsapp setup
mauritania-cli config telegram setup
mauritania-cli config facebook setup
mauritania-cli config shipper setup

# View current configuration
mauritania-cli config show

# Edit configuration
mauritania-cli config edit
```

#### Queue Management
```bash
# List queued commands
mauritania-cli queue list

# Clear all queued commands
mauritania-cli queue clear

# Retry failed commands
mauritania-cli queue retry

# Show queue statistics
mauritania-cli queue stats
```

#### Status Monitoring
```bash
# Show overall system status
mauritania-cli status

# Monitor specific transport
mauritania-cli status whatsapp
mauritania-cli status telegram

# View command history
mauritania-cli history

# Real-time monitoring
mauritania-cli monitor
```

## 🔧 Transports

### WhatsApp Integration

**Features:**
- QR code authentication (no phone number required)
- Message threading support
- Automatic message splitting for long outputs
- Rate limiting (1000 messages/hour)

**Setup:**
```bash
mauritania-cli config whatsapp setup
# Follow QR code authentication prompts
```

**Usage:**
```bash
mauritania-cli send "ls -la" --platform whatsapp
```

### Telegram Bot Integration

**Features:**
- Bot-based communication
- Channel and group support
- Rich message formatting
- Webhook support for instant delivery

**Setup:**
```bash
# Create bot via @BotFather on Telegram
# Get your bot token
mauritania-cli config telegram setup
```

**Usage:**
```bash
mauritania-cli send "git log --oneline" --platform telegram
```

### Facebook Messenger Integration

**Features:**
- Page-based communication
- Rich media support
- Automated responses
- Webhook integration

**Setup:**
```bash
mauritania-cli config facebook setup
# Requires Facebook Developer App
```

### SM APOS Shipper Integration

**Features:**
- Encrypted command execution
- Secure network transport
- Result encryption
- Audit logging

**Setup:**
```bash
mauritania-cli config shipper setup
# Requires SM APOS credentials
```

**Usage:**
```bash
# Execute sensitive commands securely
mauritania-cli send "cat /etc/passwd" --transport shipper
```

## 🏗️ Architecture

### System Components

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   CLI Client    │────│  Transport Layer │────│  Remote Server  │
│                 │    │                  │    │                 │
│ • Command Input │    │ • WhatsApp       │    │ • Command Exec  │
│ • Queue Mgmt    │    │ • Telegram       │    │ • Result Return │
│ • Status Display│    │ • Facebook       │    │ • Error Handling│
└─────────────────┘    │ • SM APOS        │    └─────────────────┘
                       └──────────────────┘
```

### Data Flow

1. **Command Input** → CLI parses and validates
2. **Transport Selection** → Chooses optimal communication channel
3. **Message Encoding** → Encrypts for secure transport
4. **Network Transmission** → Sends via social media or secure channel
5. **Remote Execution** → Server executes command
6. **Result Return** → Encrypted results sent back
7. **Local Display** → Formatted output to user

### Security Model

- **Transport Encryption** - All communications encrypted
- **Command Validation** - Input sanitization and security checks
- **Access Control** - User authentication and authorization
- **Audit Logging** - Complete command execution history
- **Rate Limiting** - Prevents abuse and ensures fair usage

## 📊 Monitoring & Analytics

### Real-time Monitoring

```bash
# Start monitoring dashboard
mauritania-cli monitor

# View transport status
mauritania-cli status

# Command execution statistics
mauritania-cli analytics commands

# Network performance metrics
mauritania-cli analytics network
```

### Logging

All activities are logged to `~/.mauritania-cli/logs/`

```bash
# View recent logs
mauritania-cli logs show

# Filter logs by transport
mauritania-cli logs filter --transport whatsapp

# Export logs for analysis
mauritania-cli logs export --format json
```

## 🔧 Advanced Configuration

### Configuration File

Location: `~/.mauritania-cli/config.toml`

```toml
[general]
default_transport = "telegram"
offline_queue_size = 1000
log_level = "info"

[transports.social_media.whatsapp]
database_path = "~/.mauritania-cli/whatsapp"
rate_limit = 1000
auto_connect = true

[transports.social_media.telegram]
bot_token = "your_bot_token"
chat_id = "your_chat_id"
webhook_url = "https://your-webhook.com"

[transports.shipper]
endpoint = "https://api.shipper.mr"
api_key = "your_api_key"
timeout = 300

[security]
enable_encryption = true
allowed_commands = ["ls", "pwd", "git", "npm"]
max_command_length = 10000
```

### Environment Variables

```bash
# Override configuration
export MAURITANIA_CLI_DEFAULT_TRANSPORT=whatsapp
export MAURITANIA_CLI_OFFLINE_MODE=true
export MAURITANIA_CLI_LOG_LEVEL=debug

# Transport credentials
export MAURITANIA_TELEGRAM_BOT_TOKEN="your_token"
export MAURITANIA_SHIPPER_API_KEY="your_key"
```

## 🐛 Troubleshooting

### Common Issues

#### "Exec format error"
```bash
# Wrong architecture - rebuild for your platform
GOOS=linux GOARCH=arm64 go build -o mauritania-cli ./cmd/mauritania-cli
```

#### "Network offline" but internet works
```bash
# Check network detection
curl -I google.com
mauritania-cli status

# Force online mode
mauritania-cli send "echo test" --offline=false
```

#### Authentication failures
```bash
# Re-authenticate transports
mauritania-cli config whatsapp setup
mauritania-cli config telegram setup
```

#### Commands not executing
```bash
# Check queue status
mauritania-cli queue list

# View command logs
mauritania-cli logs show --last 10

# Test basic connectivity
mauritania-cli send "echo 'test'" --transport telegram
```

### Debug Mode

```bash
# Enable verbose logging
mauritania-cli --verbose send "ls -la"

# View debug logs
mauritania-cli logs show --level debug

# Test transport connectivity
mauritania-cli test whatsapp
mauritania-cli test telegram
```

## 📚 API Reference

### CLI Exit Codes

- `0` - Success
- `1` - General error
- `2` - Network error
- `3` - Authentication error
- `4` - Configuration error
- `5` - Command validation error

### Transport Status Codes

- `available` - Transport ready for use
- `connecting` - Establishing connection
- `offline` - Temporarily unavailable
- `error` - Permanent failure
- `rate_limited` - Exceeded rate limits

### Command Priorities

- `low` - Background tasks
- `normal` - Standard commands (default)
- `high` - Important operations
- `urgent` - Critical commands

## 🤝 Contributing

### Development Setup

```bash
# Clone repository
git clone https://github.com/your-repo/mauritania-cli.git
cd mauritania-cli

# Install dependencies
go mod download

# Run tests
go test ./...

# Build for multiple platforms
./scripts/build-all.sh

# Run linting
golangci-lint run
```

### Code Structure

```
cmd/mauritania-cli/           # Main CLI application
├── main.go                   # Application entry point
├── cmd/                      # CLI commands
│   ├── root.go              # Root command
│   ├── send.go              # Send command
│   ├── config.go            # Configuration commands
│   └── ...
└── internal/                 # Internal packages
    ├── models/              # Data models
    ├── transports/          # Transport implementations
    ├── services/            # Business logic
    └── utils/               # Utility functions
```

### Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./internal/transports/...

# Run with coverage
go test -cover ./...

# Integration tests
go test -tags=integration ./...
```

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- **Mauritanian Developer Community** - For inspiring this project
- **Termux Team** - For making mobile Linux development possible
- **WhatsMeow Library** - For WhatsApp Web API implementation
- **Go Community** - For the excellent tooling and ecosystem

## 📞 Support

- **Issues**: [GitHub Issues](https://github.com/your-repo/mauritania-cli/issues)
- **Discussions**: [GitHub Discussions](https://github.com/your-repo/mauritania-cli/discussions)
- **Documentation**: [Wiki](https://github.com/your-repo/mauritania-cli/wiki)

---

**Built with ❤️ for developers in low-connectivity regions**

*Transforming communication into computation* 🚀