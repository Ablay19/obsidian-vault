# Troubleshooting Guide

This guide helps you diagnose and resolve common issues with the Mauritania CLI.

## Quick Diagnosis

### 1. Check CLI Installation

```bash
# Verify installation
mauritania-cli --version

# Check binary location
which mauritania-cli

# Verify permissions
ls -la $(which mauritania-cli)
```

### 2. Check Configuration

```bash
# View current configuration
mauritania-cli config show

# Check config file location
ls -la ~/.mauritania-cli/config.toml

# Validate configuration
mauritania-cli config validate
```

### 3. Check System Status

```bash
# Overall system status
mauritania-cli status

# Network connectivity
curl -I google.com
mauritania-cli send "echo 'Network test'" --offline=false

# Check logs
mauritania-cli logs show --last 5
```

## Common Issues & Solutions

### Issue: "Exec format error"

**Symptoms:**
```
bash: ./mauritania-cli: cannot execute binary file: Exec format error
```

**Causes:**
- Wrong architecture binary downloaded
- Built for wrong platform

**Solutions:**

```bash
# Check your architecture
uname -m  # Should show aarch64 for ARM64

# Rebuild for correct platform (on same architecture)
go build -o mauritania-cli ./cmd/mauritania-cli

# Or download correct binary
# For Termux (ARM64 Android):
wget https://github.com/your-repo/mauritania-cli/releases/download/v1.0.0/mauritania-cli-termux

# For Linux ARM64:
wget https://github.com/your-repo/mauritania-cli/releases/download/v1.0.0/mauritania-cli-linux-arm64

# Make executable
chmod +x mauritania-cli*
```

### Issue: "Command not found"

**Symptoms:**
```
mauritania-cli: command not found
```

**Solutions:**

```bash
# Check if binary is in PATH
echo $PATH
which mauritania-cli

# Add to PATH (add to ~/.bashrc or ~/.zshrc)
export PATH="$HOME/bin:$PATH"

# Or move to system PATH
sudo mv mauritania-cli /usr/local/bin/

# Reload shell
source ~/.bashrc
```

### Issue: "Network offline" (but internet works)

**Symptoms:**
```
⚠️ Network offline - command will be queued for retry when connectivity returns
```
But `curl google.com` works fine.

**Causes:**
- DNS resolution failure in network monitor
- Firewall blocking test URLs
- Network monitor timing out

**Solutions:**

```bash
# Test DNS resolution manually
nslookup google.com
dig google.com

# Test the URLs used by network monitor
curl -I https://www.google.com
curl -I https://www.cloudflare.com
curl -I https://1.1.1.1

# Force online mode
mauritania-cli send "echo test" --offline=false

# Check network monitor logs
mauritania-cli logs show | grep network

# Modify test URLs in config (if needed)
nano ~/.mauritania-cli/config.toml
# Change test_urls array
```

### Issue: "Transport not available"

**Symptoms:**
```
Error: transport whatsapp not configured
```

**Solutions:**

```bash
# Check transport status
mauritania-cli status whatsapp

# Reconfigure transport
mauritania-cli config whatsapp setup

# Check configuration
mauritania-cli config show | grep whatsapp

# Verify credentials file exists
ls -la ~/.mauritania-cli/whatsapp/
```

### Issue: WhatsApp Authentication Fails

**Symptoms:**
```
WhatsApp authentication failed
QR code expired
```

**Solutions:**

```bash
# Delete old session and re-authenticate
rm -rf ~/.mauritania-cli/whatsapp/
mauritania-cli config whatsapp setup

# Ensure WhatsApp app is updated
# Make sure you're scanning with the correct phone

# Check for session conflicts
ps aux | grep whatsmeow
killall mauritania-cli  # If running
```

### Issue: Telegram Commands Not Received

**Symptoms:**
Commands sent but no responses received.

**Solutions:**

```bash
# Check bot token
mauritania-cli config show | grep telegram

# Verify bot is running
curl "https://api.telegram.org/bot<YOUR_TOKEN>/getMe"

# Check chat ID
mauritania-cli config telegram setup  # Re-run to get chat ID

# Send test message to bot
curl -X POST "https://api.telegram.org/bot<YOUR_TOKEN>/sendMessage" \
  -d "chat_id=<CHAT_ID>&text=Test"
```

### Issue: SM APOS Shipper Connection Fails

**Symptoms:**
```
Shipper authentication failed
Connection timeout
```

**Solutions:**

```bash
# Check shipper configuration
mauritania-cli config show | grep shipper

# Test API endpoint
curl -I https://api.shipper.mr/health

# Verify credentials
mauritania-cli config shipper setup  # Re-enter credentials

# Check network connectivity to shipper
ping api.shipper.mr
```

### Issue: Commands Timeout

**Symptoms:**
```
Command execution timed out
```

**Solutions:**

```bash
# Increase timeout for long-running commands
mauritania-cli send "sleep 60" --timeout 90

# Check if command actually completed
mauritania-cli history --last 1

# For very long commands, use background execution
mauritania-cli send "nohup long-running-command &" --timeout 3600
```

### Issue: High Memory Usage

**Symptoms:**
CLI using excessive memory.

**Solutions:**

```bash
# Reduce queue size
nano ~/.mauritania-cli/config.toml
# Set offline_queue_size = 500

# Clear old logs
mauritania-cli logs clear

# Restart CLI
pkill mauritania-cli
mauritania-cli monitor &
```

### Issue: Permission Denied

**Symptoms:**
```
bash: ./mauritania-cli: Permission denied
```

**Solutions:**

```bash
# Make binary executable
chmod +x mauritania-cli

# Check file permissions
ls -la mauritania-cli

# Fix ownership if needed
sudo chown $USER:$USER mauritania-cli
```

## Advanced Troubleshooting

### Debug Logging

Enable detailed logging for diagnosis:

```bash
# Set debug level
export MAURITANIA_CLI_LOG_LEVEL=debug

# Run command with debug output
mauritania-cli send "echo debug test"

# View debug logs
mauritania-cli logs show --level debug
```

### Network Debugging

```bash
# Test all network components
mauritania-cli status

# Manual network tests
ping 8.8.8.8
nslookup google.com
curl -v https://www.google.com

# Check proxy settings
echo $http_proxy $https_proxy

# Test with different DNS
echo "nameserver 1.1.1.1" > /tmp/resolv.conf
# Use with: mauritania-cli send "echo test" --offline=false
```

### Transport-Specific Debugging

#### WhatsApp Debug
```bash
# Check session database
ls -la ~/.mauritania-cli/whatsapp/
file ~/.mauritania-cli/whatsapp/whatsapp.db

# Test WhatsApp connection manually
mauritania-cli status whatsapp

# Clear WhatsApp session
rm -rf ~/.mauritania-cli/whatsapp/
```

#### Telegram Debug
```bash
# Test bot API
BOT_TOKEN=$(mauritania-cli config show | grep bot_token | cut -d'"' -f2)
curl "https://api.telegram.org/bot$BOT_TOKEN/getMe"

# Send test message
CHAT_ID=$(mauritania-cli config show | grep chat_id | cut -d'"' -f2)
curl -X POST "https://api.telegram.org/bot$BOT_TOKEN/sendMessage" \
  -d "chat_id=$CHAT_ID&text=Debug test"
```

#### Shipper Debug
```bash
# Test API connectivity
curl -H "Authorization: Bearer YOUR_API_KEY" \
  https://api.shipper.mr/health

# Check authentication
mauritania-cli config shipper test
```

### Performance Issues

```bash
# Monitor resource usage
top -p $(pgrep mauritania-cli)

# Check queue status
mauritania-cli queue stats

# Clear performance logs
mauritania-cli logs export --format json > perf_$(date +%s).json

# Optimize configuration
nano ~/.mauritania-cli/config.toml
# Reduce queue sizes, increase timeouts, adjust rate limits
```

### Database Issues

```bash
# Check database file
ls -la ~/.mauritania-cli/commands.db

# Verify database integrity
sqlite3 ~/.mauritania-cli/commands.db "PRAGMA integrity_check;"

# Clear old data
mauritania-cli queue clear
mauritania-cli history clear --older-than 30d

# Backup and restore
cp ~/.mauritania-cli/commands.db ~/.mauritania-cli/commands.db.backup
rm ~/.mauritania-cli/commands.db
mauritania-cli config init  # Recreates database
```

## Getting Help

### Log Collection

For support requests, collect diagnostic information:

```bash
# System information
uname -a
go version
mauritania-cli --version

# Configuration (redact sensitive data)
mauritania-cli config show

# Recent logs
mauritania-cli logs export --last 100 > debug_logs.json

# Queue status
mauritania-cli queue stats

# Transport status
mauritania-cli status
```

### Support Channels

- **GitHub Issues**: [Report bugs](https://github.com/your-repo/mauritania-cli/issues)
- **Discussions**: [Ask questions](https://github.com/your-repo/mauritania-cli/discussions)
- **Documentation**: [Read the docs](https://github.com/your-repo/mauritania-cli/wiki)

### Emergency Commands

If CLI becomes unresponsive:

```bash
# Kill all instances
pkill -f mauritania-cli

# Clear locks
rm -f ~/.mauritania-cli/*.lock

# Reset configuration
mv ~/.mauritania-cli/config.toml ~/.mauritania-cli/config.toml.backup
mauritania-cli config init

# Clean start
mauritania-cli status
```

## Prevention Tips

### Regular Maintenance
```bash
# Weekly cleanup
mauritania-cli logs clear --older-than 7d
mauritania-cli history clear --older-than 30d
mauritania-cli queue clear --completed

# Monthly optimization
sqlite3 ~/.mauritania-cli/commands.db "VACUUM;"
mauritania-cli config backup
```

### Monitoring Setup
```bash
# Create health check script
cat > health_check.sh << 'EOF'
#!/bin/bash
if ! mauritania-cli status > /dev/null 2>&1; then
    echo "CLI unhealthy, restarting..."
    pkill -f mauritania-cli
    sleep 2
    mauritania-cli monitor &
fi
EOF

# Run every 5 minutes
crontab -e
# */5 * * * * /path/to/health_check.sh
```

### Backup Strategy
```bash
# Daily backup
cat > backup.sh << 'EOF'
#!/bin/bash
DATE=$(date +%Y%m%d)
tar czf ~/.mauritania-cli/backup_$DATE.tar.gz ~/.mauritania-cli/
# Keep only last 7 backups
ls -t ~/.mauritania-cli/backup_*.tar.gz | tail -n +8 | xargs rm -f
EOF
```

This troubleshooting guide should help resolve most issues. If you encounter problems not covered here, please check the GitHub repository for updates or create an issue with your diagnostic information.