# Quickstart: Mauritania Network Integration

**Feature**: Mauritania Network Integration
**Date**: January 17, 2025

## Overview

The Mauritania Network Integration provides a shell-like interface for Termux that enables project management through Mauritanian network provider services. Execute commands remotely using social media APIs, SM APOS Shipper, and NRT routing for optimal connectivity in regions with limited direct internet access.

## Prerequisites

### System Requirements
- **Android Device**: With Termux installed
- **Node.js**: 18.0.0 or higher (install via Termux pkg)
- **Social Media Accounts**: Facebook, WhatsApp, or Telegram
- **SM APOS Shipper Access**: Mauritanian network provider subscription

### Network Requirements
- **Mauritanian Mobile Network**: Mauritel, Chinguitel, or Mattel
- **Social Media Access**: Through mobile data or WiFi
- **SM APOS Shipper**: Enabled on your mobile plan

## Installation

### Termux Installation
```bash
# Update Termux packages
pkg update && pkg upgrade

# Install Node.js and dependencies
pkg install nodejs git curl

# Clone the project
git clone <repository>
cd <project-directory>

# Install the CLI package
npm install -g packages/mauritania-net-integration/
```

### Configuration
```bash
# Initialize configuration
mauritania-cli init

# Configure social media access
mauritania-cli config social-media setup

# Configure SM APOS Shipper
mauritania-cli config shipper setup

# Configure NRT routing
mauritania-cli config nrt setup
```

## Basic Usage

### Starting the CLI
```bash
# Start the CLI server
mauritania-cli start

# Or run in foreground for debugging
mauritania-cli start --foreground
```

### Social Media Command Interface
```bash
# Send commands via social media
echo "ls -la" | mauritania-cli send --platform whatsapp --recipient "+22212345678"

# Check command status
mauritania-cli status <command-id>

# View command results
mauritania-cli result <command-id>
```

### SM APOS Shipper Commands
```bash
# Execute commands via shipper service
mauritania-cli shipper exec "git status"

# List available shipper commands
mauritania-cli shipper commands

# Check shipper session status
mauritania-cli shipper status
```

### NRT Network Routing
```bash
# View available network routes
mauritania-cli routes list

# Test network connectivity
mauritania-cli routes test

# Force specific route
mauritania-cli routes use social-media
```

## Development Workflow

### Project Management Commands
```bash
# Git operations via social media
mauritania-cli send "git add . && git commit -m 'update'" --platform telegram

# NPM operations
mauritania-cli send "npm install && npm run build" --platform whatsapp

# File operations
mauritania-cli send "find . -name '*.log' -delete" --platform facebook
```

### Monitoring and Debugging
```bash
# View CLI logs
mauritania-cli logs

# Check network status
mauritania-cli network status

# View command queue
mauritania-cli queue list

# Clear failed commands
mauritania-cli queue clear --failed
```

## Configuration

### Social Media Setup
```json
{
  "socialMedia": {
    "platforms": {
      "whatsapp": {
        "apiKey": "your-whatsapp-api-key",
        "webhookUrl": "https://your-domain.com/webhook/whatsapp"
      },
      "telegram": {
        "botToken": "your-telegram-bot-token",
        "webhookUrl": "https://your-domain.com/webhook/telegram"
      },
      "facebook": {
        "pageAccessToken": "your-facebook-token",
        "webhookUrl": "https://your-domain.com/webhook/facebook"
      }
    },
    "rateLimits": {
      "maxRequestsPerHour": 30,
      "maxMessageLength": 4000
    }
  }
}
```

### SM APOS Shipper Configuration
```json
{
  "shipper": {
    "endpoint": "https://api.mauritania-shipper.mr",
    "credentials": {
      "username": "your-phone-number",
      "password": "your-shipper-password"
    },
    "sessionTimeout": 3600,
    "rateLimits": {
      "requestsPerHour": 100,
      "maxExecutionTime": 300
    }
  }
}
```

### NRT Routing Configuration
```json
{
  "nrt": {
    "enabled": true,
    "routes": [
      {
        "name": "social-media-primary",
        "type": "social_media",
        "priority": 1,
        "costWeight": 0.3,
        "reliabilityWeight": 0.7
      },
      {
        "name": "shipper-backup",
        "type": "sm_apos",
        "priority": 2,
        "costWeight": 0.6,
        "reliabilityWeight": 0.4
      }
    ],
    "failover": {
      "enabled": true,
      "timeout": 30,
      "maxRetries": 3
    }
  }
}
```

## Advanced Usage

### Batch Command Execution
```bash
# Create command batch file
cat > commands.txt << EOF
npm install
npm run lint
npm run test
npm run build
EOF

# Execute batch via social media
mauritania-cli send --batch commands.txt --platform whatsapp
```

### Custom Transport Plugins
```bash
# List available transports
mauritania-cli transport list

# Install custom transport
mauritania-cli transport install my-transport

# Configure custom transport
mauritania-cli transport config my-transport --param value
```

### Offline Mode
```bash
# Enable offline queuing
mauritania-cli offline enable

# View queued commands
mauritania-cli offline queue

# Force sync when online
mauritania-cli offline sync
```

### Security Configuration
```bash
# Configure command whitelisting
mauritania-cli security whitelist add "git *"
mauritania-cli security whitelist add "npm *"

# Enable authentication
mauritania-cli security auth enable

# Configure user permissions
mauritania-cli security users add +22212345678 --permissions "read,execute"
```

## Troubleshooting

### Common Issues

#### Social Media Connection Failed
```bash
# Test social media connectivity
mauritania-cli test social-media whatsapp

# Check API credentials
mauritania-cli config social-media validate

# View connection logs
mauritania-cli logs --filter social-media
```

#### SM APOS Shipper Authentication
```bash
# Test shipper connection
mauritania-cli test shipper

# Re-authenticate
mauritania-cli shipper login

# Check session status
mauritania-cli shipper session
```

#### NRT Routing Issues
```bash
# Test all routes
mauritania-cli routes test --all

# Reset route preferences
mauritania-cli routes reset

# Manual route selection
mauritania-cli routes use direct
```

#### Command Execution Problems
```bash
# Check command syntax
mauritania-cli validate "your-command"

# View execution logs
mauritania-cli logs --command-id <id>

# Retry failed command
mauritania-cli retry <command-id>
```

### Performance Optimization
```bash
# Enable compression
mauritania-cli config compression enable

# Optimize for mobile
mauritania-cli config mobile optimize

# Monitor performance
mauritania-cli monitor performance
```

## API Reference

### REST API Endpoints
- `POST /commands/execute` - Execute command
- `GET /commands/{id}/status` - Get command status
- `GET /commands/{id}/result` - Get command result
- `GET /queue` - List queued commands
- `GET /transports` - List available transports
- `POST /sessions` - Create shipper session

### CLI Commands
- `mauritania-cli send <command>` - Send command via transport
- `mauritania-cli status <id>` - Check command status
- `mauritania-cli result <id>` - Get command result
- `mauritania-cli queue list` - Show command queue
- `mauritania-cli routes list` - Show network routes

## Security Best Practices

### Authentication
- Use strong passwords for SM APOS Shipper
- Enable two-factor authentication when available
- Regularly rotate API keys and tokens

### Command Security
- Use command whitelisting to restrict allowed operations
- Validate all input parameters
- Log all command executions for audit trails

### Network Security
- Use HTTPS for all API communications
- Encrypt sensitive data in transit
- Monitor for unusual network activity

## Cost Optimization

### Social Media Usage
- Batch multiple commands to reduce API calls
- Use compression to minimize data transfer
- Schedule heavy operations during off-peak hours

### SM APOS Shipper
- Monitor usage costs regularly
- Use appropriate execution timeouts
- Optimize command frequency

### NRT Routing
- Configure cost weights based on your usage patterns
- Use cheaper routes for non-urgent operations
- Monitor route performance and adjust preferences

---

**Managing projects from anywhere in Mauritania! 🌍📱**