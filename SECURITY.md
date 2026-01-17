# Security Overview

This document outlines the security measures and best practices implemented in the Mauritania CLI.

## Security Model

### Core Principles

1. **Defense in Depth** - Multiple security layers protect against various attack vectors
2. **Least Privilege** - Commands execute with minimal required permissions
3. **Secure by Default** - Conservative security settings enabled by default
4. **Audit Everything** - Comprehensive logging of all operations

### Threat Model

**Primary Threats:**
- Command injection attacks
- Unauthorized command execution
- Credential theft
- Network interception
- Resource abuse

**Secondary Threats:**
- Denial of service
- Data exfiltration
- Session hijacking
- Malware execution

## Security Features

### 1. Command Security

#### Input Validation
```go
// Command sanitization
func (csv *CommandSecurityValidator) SanitizeCommand(command string) string {
    // Remove dangerous characters
    sanitized := strings.ReplaceAll(command, "\x00", "")

    // Validate against patterns
    for _, pattern := range dangerousPatterns {
        if strings.Contains(strings.ToLower(command), pattern) {
            return "", fmt.Errorf("dangerous pattern detected: %s", pattern)
        }
    }

    return sanitized, nil
}
```

#### Command Whitelisting
```toml
[security]
allowed_commands = [
    "ls", "pwd", "git", "npm", "yarn",
    "echo", "cat", "head", "tail", "grep"
]
max_command_length = 10000
```

#### Injection Prevention
- Null byte filtering
- Shell metacharacter validation
- Path traversal protection
- Command chaining restrictions

### 2. Transport Security

#### WhatsApp Security
- QR code-based authentication (no password storage)
- End-to-end encrypted sessions
- Automatic session cleanup
- Rate limiting (1000 messages/hour)

#### Telegram Security
- Bot token authentication
- Webhook signature verification (when enabled)
- Chat ID validation
- Rate limiting (30 messages/minute)

#### SM APOS Shipper Security
- API key authentication
- TLS 1.3 encryption
- Command encryption in transit
- Session-based authorization

### 3. Data Protection

#### Encryption at Rest
```go
// Sensitive data encryption
func (ce *CommandEncryption) EncryptCommand(command string, key string) (string, error) {
    // AES-256-GCM encryption
    block, err := aes.NewCipher(derivedKey)
    if err != nil {
        return "", err
    }

    aesGCM, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }

    // Encrypt with authentication
    ciphertext := aesGCM.Seal(nil, nonce, []byte(command), nil)
    return base64.StdEncoding.EncodeToString(append(nonce, ciphertext...)), nil
}
```

#### Credential Storage
- Encrypted API keys and tokens
- Secure file permissions (600)
- Isolated configuration directory
- Automatic credential rotation

#### Session Management
- Automatic session expiration (24 hours)
- Secure session cleanup
- Memory-only sensitive data
- Session audit logging

### 4. Network Security

#### Transport Layer Security
- HTTPS-only API communications
- Certificate validation
- TLS 1.3 support
- Secure webhook endpoints

#### Network Monitoring
- Connectivity verification
- Certificate expiry monitoring
- DNS security validation
- Network anomaly detection

### 5. Access Control

#### User Authentication
```go
// Multi-level authentication
func (cas *CommandAuthService) AuthenticateCommand(cmd *models.Command, senderID, platform string) error {
    // Validate sender identity
    if err := cas.validateSender(senderID, platform); err != nil {
        return fmt.Errorf("sender validation failed: %w", err)
    }

    // Check command permissions
    if err := cas.checkCommandPermissions(cmd, senderID); err != nil {
        return fmt.Errorf("permission denied: %w", err)
    }

    // Validate command content
    if err := cas.validateCommandContent(cmd); err != nil {
        return fmt.Errorf("content validation failed: %w", err)
    }

    return nil
}
```

#### Role-Based Permissions
- Platform-specific access control
- Command-specific permissions
- Time-based access restrictions
- Geographic restrictions (future)

### 6. Audit & Monitoring

#### Comprehensive Logging
```go
// Security event logging
func (cas *CommandAuthService) LogAuthAttempt(cmd *models.Command, senderID, platform string, success bool, reason string) {
    level := "INFO"
    if !success {
        level = "WARN"
    }

    log.WithFields(log.Fields{
        "sender":   senderID,
        "platform": platform,
        "command":  cmd.Command,
        "success":  success,
        "reason":   reason,
        "timestamp": time.Now(),
    }).Log(level, "Authentication attempt")
}
```

#### Security Metrics
- Failed authentication attempts
- Command execution statistics
- Rate limit violations
- Security incident alerts

## Security Best Practices

### For Administrators

#### 1. Secure Installation
```bash
# Secure file permissions
chmod 700 ~/.mauritania-cli
chmod 600 ~/.mauritania-cli/config.toml
chmod 600 ~/.mauritania-cli/commands.db

# Secure binary
sudo chown root:root /usr/local/bin/mauritania-cli
sudo chmod 755 /usr/local/bin/mauritania-cli
```

#### 2. Configuration Hardening
```toml
[security]
enable_encryption = true
require_approval = true  # For dangerous commands
max_sessions_per_user = 3
session_timeout_minutes = 480  # 8 hours

[transports]
# Use secure endpoints only
webhook_verify_signatures = true
force_https = true
```

#### 3. Network Security
```bash
# Configure firewall
sudo ufw allow 3001/tcp  # CLI API port
sudo ufw --force enable

# Use VPN for remote access
# Configure TLS certificates
```

### For Users

#### 1. Safe Command Practices
```bash
# Use read-only commands first
mauritania-cli send "ls -la"
mauritania-cli send "pwd"

# Avoid dangerous patterns
# ❌ DON'T do this:
mauritania-cli send "rm -rf /"
mauritania-cli send "sudo apt install malware"

# ✅ DO this instead:
mauritania-cli send "ls /tmp"
mauritania-cli send "git status"
```

#### 2. Credential Management
```bash
# Rotate API keys regularly
mauritania-cli config whatsapp setup  # Re-authenticate
mauritania-cli config telegram setup  # New bot token

# Don't share session information
# Use strong, unique credentials
```

#### 3. Network Awareness
```bash
# Verify connection security
mauritania-cli status

# Use encrypted transports when possible
mauritania-cli send "sensitive-command" --transport shipper

# Monitor for suspicious activity
mauritania-cli logs show | grep -i "failed\|error"
```

## Security Assessment

### Penetration Testing

#### Automated Security Tests
```bash
# Run security test suite
go test ./... -tags=security

# Vulnerability scanning
gosec ./...

# Dependency checking
nancy ./go.sum
```

#### Manual Security Review
- Code review for security vulnerabilities
- Dependency analysis for known CVEs
- Configuration review for secure defaults
- Network traffic analysis

### Compliance Considerations

#### SOC 2 Compliance
- Access logging and monitoring
- Data encryption at rest and in transit
- Regular security assessments
- Incident response procedures

#### GDPR Compliance
- Minimal data collection
- User consent for data processing
- Right to data deletion
- Data processing transparency

## Incident Response

### Security Incident Procedure

1. **Detection**
   ```bash
   # Monitor for anomalies
   mauritania-cli logs show | grep -i "security\|failed"

   # Check system status
   mauritania-cli status
   ```

2. **Containment**
   ```bash
   # Disable compromised transport
   mauritania-cli config whatsapp disable

   # Clear suspicious sessions
   mauritania-cli session clear --all

   # Enable emergency mode
   mauritania-cli security lockdown
   ```

3. **Investigation**
   ```bash
   # Export security logs
   mauritania-cli logs export --security-events --last-24h > incident_logs.json

   # Analyze command patterns
   mauritania-cli analytics security --last-24h
   ```

4. **Recovery**
   ```bash
   # Rotate all credentials
   mauritania-cli config rotate-keys

   # Restore from backup
   mauritania-cli config restore /path/to/backup

   # Resume normal operations
   mauritania-cli security unlock
   ```

### Emergency Commands

```bash
# Complete system lockdown
mauritania-cli security emergency-stop

# Wipe all sensitive data
mauritania-cli security wipe --credentials

# Generate security report
mauritania-cli security audit-report

# Restore from secure backup
mauritania-cli security restore --secure
```

## Security Updates

### Regular Maintenance

#### Weekly Tasks
```bash
# Update dependencies
go mod tidy
go mod download

# Security scan
gosec ./...
nancy ./go.sum

# Rotate logs
mauritania-cli logs rotate
```

#### Monthly Tasks
```bash
# Full security audit
mauritania-cli security audit

# Update configurations
mauritania-cli config update-security

# Review access logs
mauritania-cli analytics access-review
```

#### Quarterly Tasks
```bash
# Penetration testing
# External security assessment
# Policy review and updates
```

## Future Security Enhancements

### Planned Features
- **Zero-trust architecture** with continuous verification
- **Hardware security modules** for key storage
- **Advanced threat detection** with ML
- **Automated security patching**
- **Multi-factor authentication** for admin operations

### Research Areas
- **Post-quantum cryptography** for future-proofing
- **Blockchain-based audit trails** for immutable logging
- **AI-powered anomaly detection**
- **Secure multi-party computation** for collaborative commands

This security overview ensures the Mauritania CLI provides robust protection while enabling remote development in challenging network environments. 🔒🛡️