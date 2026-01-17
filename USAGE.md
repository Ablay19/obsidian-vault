# Usage Examples & API Reference

This document provides comprehensive usage examples and API reference for the Mauritania CLI.

## Basic Usage Examples

### 1. Getting Started

```bash
# Check CLI status
mauritania-cli --version
mauritania-cli --help

# Initialize configuration
mauritania-cli config init

# View current configuration
mauritania-cli config show
```

### 2. Transport Setup

#### WhatsApp Setup
```bash
# Start WhatsApp authentication
mauritania-cli config whatsapp setup

# The CLI will display a QR code
# Scan it with WhatsApp on your phone
# Wait for "Authentication successful" message
```

#### Telegram Setup
```bash
# Configure Telegram bot
mauritania-cli config telegram setup

# Enter your bot token from @BotFather
# Send /start to your bot to get chat ID
# CLI will automatically detect and save chat ID
```

#### SM APOS Shipper Setup
```bash
# Configure secure command execution
mauritania-cli config shipper setup

# Enter your SM APOS credentials
# API endpoint, username, password
```

### 3. Basic Command Execution

```bash
# Send simple command (uses default transport)
mauritania-cli send "pwd"

# Specify transport explicitly
mauritania-cli send "ls -la" --platform whatsapp
mauritania-cli send "git status" --platform telegram

# Use secure execution
mauritania-cli send "cat /etc/passwd" --transport shipper

# Force offline mode (queue for later)
mauritania-cli send "npm install" --offline=true
```

## Advanced Usage Examples

### 4. Command Management

#### Priority Commands
```bash
# High priority command
mauritania-cli send "systemctl restart nginx" --priority high

# Urgent command (bypasses some rate limits)
mauritania-cli send "kill -9 1234" --priority urgent
```

#### Long-Running Commands
```bash
# Commands with custom timeout (default: 30 seconds)
mauritania-cli send "sleep 60 && echo 'Done'" --timeout 90

# Background compilation
mauritania-cli send "make -j4" --timeout 600
```

#### Command Pipelines
```bash
# Multiple commands (be careful with shell injection)
mauritania-cli send "git pull && npm install && npm run build"

# Conditional execution
mauritania-cli send "if [ -f config.yml ]; then echo 'Config exists'; else echo 'No config'; fi"
```

### 5. Queue Management

```bash
# View queued commands
mauritania-cli queue list

# Clear all queued commands
mauritania-cli queue clear

# Retry failed commands
mauritania-cli queue retry

# View queue statistics
mauritania-cli queue stats

# Remove specific command from queue
mauritania-cli queue remove <command-id>
```

### 6. Monitoring & Status

```bash
# Overall system status
mauritania-cli status

# Transport-specific status
mauritania-cli status whatsapp
mauritania-cli status telegram
mauritania-cli status shipper

# Real-time monitoring (Ctrl+C to exit)
mauritania-cli monitor

# Command history
mauritania-cli history
mauritania-cli history --last 10
mauritania-cli history --platform whatsapp
```

### 7. Logging & Debugging

```bash
# View recent logs
mauritania-cli logs show

# Filter logs by level
mauritania-cli logs show --level error

# Filter by transport
mauritania-cli logs show --transport whatsapp

# Export logs
mauritania-cli logs export --format json --output logs.json

# Enable debug logging
export MAURITANIA_CLI_LOG_LEVEL=debug
mauritania-cli send "echo debug test"
```

## API Reference

### CLI Commands

#### Global Flags
```bash
--config string      Config file path (default "~/.mauritania-cli/config.toml")
--verbose             Enable verbose output
--offline             Force offline mode
--help               Show help
--version            Show version
```

#### `mauritania-cli send [command]`
Send a command for execution.

**Flags:**
```bash
-p, --platform string    Target platform (whatsapp, telegram, facebook)
-t, --transport string    Transport type (social_media, shipper, nrt)
--priority string         Command priority (low, normal, high, urgent)
--timeout int             Command timeout in seconds (default 30)
--offline                Queue command for offline execution
```

**Examples:**
```bash
mauritania-cli send "ls -la"
mauritania-cli send "git status" --platform telegram --priority high
mauritania-cli send "docker build ." --timeout 600 --offline
```

#### `mauritania-cli config [subcommand]`
Manage configuration.

**Subcommands:**
```bash
init           Initialize default configuration
show           Display current configuration
edit           Edit configuration file
whatsapp setup Configure WhatsApp transport
telegram setup Configure Telegram transport
facebook setup Configure Facebook transport
shipper setup  Configure SM APOS Shipper transport
```

**Examples:**
```bash
mauritania-cli config init
mauritania-cli config show
mauritania-cli config whatsapp setup
```

#### `mauritania-cli queue [subcommand]`
Manage command queue.

**Subcommands:**
```bash
list           List queued commands
clear          Clear all queued commands
retry          Retry failed commands
stats          Show queue statistics
remove <id>    Remove specific command
```

#### `mauritania-cli status [transport]`
Show system status.

**Examples:**
```bash
mauritania-cli status           # Overall status
mauritania-cli status whatsapp  # WhatsApp status
mauritania-cli status telegram  # Telegram status
```

#### `mauritania-cli history [flags]`
Show command execution history.

**Flags:**
```bash
--last int         Show last N commands (default 10)
--platform string  Filter by platform
--status string    Filter by status (success, failed, queued)
--format string    Output format (table, json)
```

#### `mauritania-cli logs [subcommand]`
Manage logs.

**Subcommands:**
```bash
show           Show recent logs
export         Export logs to file
clear          Clear old logs
```

### Exit Codes

| Code | Description |
|------|-------------|
| 0    | Success |
| 1    | General error |
| 2    | Network error |
| 3    | Authentication error |
| 4    | Configuration error |
| 5    | Command validation error |
| 6    | Transport unavailable |
| 7    | Rate limit exceeded |

## Transport APIs

### WhatsApp Transport

**Features:**
- QR code authentication (no phone number required)
- Automatic reconnection
- Message splitting for long responses
- Rate limiting (1000/hour)

**Configuration:**
```toml
[transports.social_media.whatsapp]
database_path = "~/.mauritania-cli/whatsapp"
rate_limit = 1000
auto_connect = true
```

**API Methods:**
```go
type WhatsAppTransport struct {
    // SendMessage sends a message via WhatsApp
    SendMessage(recipient, message string) (*MessageResponse, error)

    // ReceiveMessages retrieves pending messages
    ReceiveMessages() ([]*IncomingMessage, error)

    // GetStatus returns transport status
    GetStatus() (*TransportStatus, error)

    // ValidateCredentials checks authentication
    ValidateCredentials() error

    // GetRateLimit returns rate limit status
    GetRateLimit() (*RateLimit, error)
}
```

### Telegram Transport

**Features:**
- Bot API integration
- Webhook support
- Channel and group messaging
- Rich text formatting

**Configuration:**
```toml
[transports.social_media.telegram]
bot_token = "1234567890:ABCdefGHIjklMNOpqrsTUVwxyz"
chat_id = "123456789"
webhook_url = "https://your-webhook.com"  # Optional
rate_limit = 30
```

### SM APOS Shipper Transport

**Features:**
- Encrypted command execution
- Secure network transport
- Session management
- Result encryption

**Configuration:**
```toml
[transports.shipper]
base_url = "https://api.shipper.mr"
api_key = "your_api_key"
timeout = 300
rate_limit = 10
```

## Configuration Reference

### Main Configuration File

Location: `~/.mauritania-cli/config.toml`

```toml
[general]
default_transport = "whatsapp"  # Default transport for commands
log_level = "info"             # Logging level (debug, info, warn, error)
offline_queue_size = 1000      # Maximum queued commands

[transports.social_media.whatsapp]
database_path = "~/.mauritania-cli/whatsapp"  # Session database
rate_limit = 1000                            # Messages per hour
auto_connect = true                          # Auto-connect on startup

[transports.social_media.telegram]
bot_token = "your_bot_token"    # From @BotFather
chat_id = "your_chat_id"        # Target chat ID
webhook_url = ""                # Optional webhook URL
rate_limit = 30                 # Messages per minute

[transports.social_media.facebook]
app_id = "your_app_id"          # Facebook App ID
app_secret = "your_app_secret"  # Facebook App Secret
access_token = "your_token"     # Page Access Token
page_id = "your_page_id"        # Facebook Page ID
rate_limit = 200                # Messages per hour

[transports.shipper]
base_url = "https://api.shipper.mr"  # SM APOS API endpoint
api_key = "your_api_key"             # API authentication key
timeout = 300                        # Command timeout (seconds)
rate_limit = 10                      # Commands per minute

[security]
enable_encryption = true             # Encrypt sensitive commands
allowed_commands = [                 # Whitelist of allowed commands
    "ls", "pwd", "git", "npm", "yarn", "echo",
    "cat", "head", "tail", "grep", "find", "ps"
]
max_command_length = 10000           # Maximum command length
require_approval = false             # Require manual approval for dangerous commands
```

## Environment Variables

```bash
# Configuration
MAURITANIA_CLI_CONFIG=/path/to/config.toml
MAURITANIA_CLI_LOG_LEVEL=debug

# Transport Credentials
MAURITANIA_TELEGRAM_BOT_TOKEN=your_token
MAURITANIA_SHIPPER_API_KEY=your_key
MAURITANIA_WHATSAPP_SESSION_PATH=/path/to/session

# Behavior
MAURITANIA_CLI_DEFAULT_TRANSPORT=telegram
MAURITANIA_CLI_OFFLINE_MODE=true
MAURITANIA_CLI_QUEUE_SIZE=500
```

## Error Codes & Troubleshooting

### Common Errors

#### "Transport not available"
```
Cause: Transport not configured or offline
Solution:
mauritania-cli config whatsapp setup
mauritania-cli status whatsapp
```

#### "Rate limit exceeded"
```
Cause: Too many messages sent
Solution:
mauritania-cli status whatsapp  # Check limits
wait 1 hour  # WhatsApp resets hourly
```

#### "Command failed: permission denied"
```
Cause: Command not in allowed list
Solution:
Edit ~/.mauritania-cli/config.toml
Add command to allowed_commands array
```

#### "Network offline"
```
Cause: No internet connectivity
Solution:
Check internet: curl google.com
Use offline mode: mauritania-cli send "cmd" --offline=true
Wait for connectivity, then: mauritania-cli queue retry
```

### Debug Commands

```bash
# Enable debug logging
export MAURITANIA_CLI_LOG_LEVEL=debug

# Test network connectivity
mauritania-cli send "curl -I google.com" --timeout 10

# Check transport health
mauritania-cli status
curl -X GET "http://localhost:3001/health"  # If server running

# View raw logs
tail -f ~/.mauritania-cli/logs/mauritania-cli.log

# Test transport directly
mauritania-cli config whatsapp test
```

## Performance Tuning

### Mobile Optimization
```toml
[general]
offline_queue_size = 500  # Smaller queue for mobile

[transports.social_media.whatsapp]
rate_limit = 500  # Conservative rate limiting

[security]
max_command_length = 5000  # Shorter commands on mobile
```

### Server Optimization
```toml
[general]
offline_queue_size = 5000  # Larger queue for server

[transports.shipper]
rate_limit = 100  # Higher limits for server
timeout = 600     # Longer timeouts
```

### Network Optimization
```toml
# For slow connections
[transports.shipper]
timeout = 120     # Longer timeouts
rate_limit = 5    # Conservative rate limiting
```

## Integration Examples

### CI/CD Pipeline
```yaml
# .github/workflows/deploy.yml
name: Deploy
on: [push]
jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
    - name: Deploy via Mauritania CLI
      run: |
        mauritania-cli send "git pull && npm install && npm run build" --platform telegram
```

### System Monitoring
```bash
# Create monitoring script
cat > monitor.sh << 'EOF'
#!/bin/bash
# System monitoring via WhatsApp

CPU=$(top -bn1 | grep "Cpu(s)" | sed "s/.*, *\([0-9.]*\)%* id.*/\1/" | awk '{print 100 - $1}')
MEM=$(free | grep Mem | awk '{printf "%.2f", $3/$2 * 100.0}')

mauritania-cli send "echo 'CPU: ${CPU}%, RAM: ${MEM}%'" --platform whatsapp
EOF

# Run every 5 minutes
crontab -e
# Add: */5 * * * * /path/to/monitor.sh
```

### Development Workflow
```bash
# Development shortcuts
alias deploy="mauritania-cli send 'git pull && npm run build && systemctl restart app'"
alias logs="mauritania-cli send 'tail -f /var/log/app.log'"
alias backup="mauritania-cli send 'tar czf backup.tar.gz /var/data' --transport shipper"

# Quick deployment
deploy

# Monitor logs
logs
```

## Advanced Features

### Custom Command Templates
```bash
# Save common commands
mauritania-cli template save "deploy" "git pull && npm install && npm run build && pm2 restart app"
mauritania-cli template save "logs" "pm2 logs --lines 50"
mauritania-cli template save "backup" "tar czf /tmp/backup.tar.gz /var/data"

# Execute templates
mauritania-cli template run deploy
```

### Batch Operations
```bash
# Execute multiple commands
mauritania-cli batch << EOF
git status
npm test
npm run lint
EOF

# Batch with different transports
mauritania-cli batch --platform telegram << EOF
echo "Starting deployment"
git pull
npm install
npm run build
echo "Deployment complete"
EOF
```

### Webhook Integration
```bash
# Start webhook server
mauritania-cli server --port 3001

# Configure external webhook
curl -X POST http://your-server:3001/webhooks/github \
  -H "Content-Type: application/json" \
  -d '{"action": "push", "repository": {"name": "myapp"}}'

# Automatic deployment on push
mauritania-cli send "git pull && docker-compose up -d" --platform telegram
```

This comprehensive guide covers all aspects of using the Mauritania CLI for remote development in low-connectivity environments! 🎯