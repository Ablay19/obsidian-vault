# Doppler Setup Guide

This guide explains how to set up and configure Doppler for secrets management in the Obsidian Vault project.

## Overview

Doppler provides centralized secrets management for all environment variables, API keys, and sensitive configuration. This eliminates hardcoded credentials and ensures consistent configuration across all environments.

## Prerequisites

- Doppler CLI installed (`brew install dopplerhq/cli/doppler` on macOS)
- Doppler account with access to the Obsidian Vault workspace
- Git repository cloned locally

## Installation

### macOS
```bash
brew install dopplerhq/cli/doppler
```

### Linux
```bash
curl -L https://github.com/DopplerHQ/cli/releases/latest/download/doppler -o /usr/local/bin/doppler
chmod +x /usr/local/bin/doppler
```

### Verification
```bash
doppler --version
```

## Authentication

### Login
```bash
doppler login
```

### Token Setup
1. Navigate to [Doppler Dashboard](https://dashboard.doppler.com)
2. Select the Obsidian Vault workspace
3. Go to Settings → Access Tokens
4. Create a new token with appropriate permissions
5. Set the token as an environment variable:
   ```bash
   export DOPPLER_TOKEN=your_token_here
   ```

## Configuration

### Project Structure

```
obsidian-vault/
├── Doppler.toml          # Doppler configuration file
├── .doppler.yaml         # Local Doppler settings
├── apps/
│   ├── api-gateway/
│   ├── worker-whatsapp/
│   └── worker-ai-manim/
├── internal/
└── deploy/
```

### Doppler.toml Example

```toml
[project]
name = "obsidian-vault"

[configs]
sandbox = "sandbox"
staging = "staging"
production = "production"

[environments.sandbox]
prefix = "SANDBOX_"

[environments.staging]
prefix = "STAGING_"

[environments.production]
prefix = "PROD_"
```

### Environment Setup

#### Development (Local)
```bash
doppler setup --config sandbox
```

#### Staging
```bash
doppler setup --config staging
```

#### Production
```bash
doppler setup --config production
```

## Integration with Applications

### Go Applications

```go
import "github.com/DopplerHQ/go-sdk"

func init() {
    token := os.Getenv("DOPPLER_TOKEN")
    client := doppler.NewClient(token)
    
    secrets, err := client.GetSecrets()
    if err != nil {
        log.Fatal(err)
    }
    
    // Access secrets
    dbPassword := secrets["DB_PASSWORD"].Value
    apiKey := secrets["API_KEY"].Value
}
```

### Node.js Applications

```javascript
const Doppler = require('@dopplerhq/node-sdk');

async function init() {
    const client = new Doppler({ token: process.env.DOPPLER_TOKEN });
    const secrets = await client.getSecrets();
    
    const dbHost = secrets.DB_HOST;
    const apiKey = secrets.API_KEY;
}
```

### Cloudflare Workers

```typescript
export default {
    async fetch(request, env) {
        const apiKey = await env.DOPPLER.get('API_KEY');
        // Use the API key
    }
};
```

## Secret Categories

### Database Secrets
- `DB_HOST` - Database server hostname
- `DB_PORT` - Database port
- `DB_NAME` - Database name
- `DB_USER` - Database username
- `DB_PASSWORD` - Database password
- `DB_SSL_MODE` - SSL connection mode

### API Keys
- `OPENAI_API_KEY` - OpenAI API key
- `TELEGRAM_BOT_TOKEN` - Telegram bot token
- `CLOUDFLARE_API_TOKEN` - Cloudflare API token
- `ANTHROPIC_API_KEY` - Anthropic API key

### Service Configuration
- `REDIS_URL` - Redis connection string
- `R2_ENDPOINT` - Cloudflare R2 endpoint
- `R2_ACCESS_KEY_ID` - R2 access key
- `R2_SECRET_ACCESS_KEY` - R2 secret key

### Security
- `JWT_SECRET` - JWT signing secret
- `ENCRYPTION_KEY` - Data encryption key
- `SESSION_SECRET` - Session cookie secret

## Environment-Specific Values

### Sandbox Environment
| Variable | Value | Description |
|----------|-------|-------------|
| `DB_HOST` | localhost | Local database |
| `REDIS_URL` | redis://localhost:6379 | Local Redis |
| `LOG_LEVEL` | debug | Verbose logging |

### Staging Environment
| Variable | Value | Description |
|----------|-------|-------------|
| `DB_HOST` | staging-db.example.com | Staging database |
| `REDIS_URL` | redis://staging-redis:6379 | Staging Redis |
| `LOG_LEVEL` | info | Standard logging |

### Production Environment
| Variable | Value | Description |
|----------|-------|-------------|
| `DB_HOST` | prod-db.example.com | Production database |
| `REDIS_URL` | redis://prod-redis:6379 | Production Redis |
| `LOG_LEVEL` | warn | Minimal logging |

## CI/CD Integration

### GitHub Actions

```yaml
name: Deploy with Doppler

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Doppler
        uses: dopplerhq/cli-action@v1
        
      - name: Deploy
        run: |
          doppler run -- ./deploy.sh
        env:
          DOPPLER_TOKEN: ${{ secrets.DOPPLER_TOKEN }}
```

### Docker

```dockerfile
FROM golang:1.21-alpine

# Install Doppler CLI
RUN apk add --no-cache curl
RUN curl -L https://github.com/DopplerHQ/cli/releases/latest/download/doppler -o /usr/local/bin/doppler
RUN chmod +x /usr/local/bin/doppler

WORKDIR /app

# Copy Doppler configuration
COPY Doppler.toml .

# Fetch secrets at build time
COPY . .
RUN doppler run -- go build -o main

CMD ["./main"]
```

## Rotation Procedures

### API Key Rotation

1. Generate new API key in provider dashboard
2. Update Doppler secret with new value
3. Verify application health
4. Revoke old API key after 24-hour grace period

### Database Password Rotation

1. Update password in Doppler
2. Trigger application restart
3. Verify database connectivity
4. Update rotation policy for automated rotation

## Troubleshooting

### Common Issues

#### "Not authenticated"
```bash
# Check authentication
doppler me

# Re-authenticate
doppler logout
doppler login
```

#### "Config not found"
```bash
# List available configs
doppler configs

# Set active config
doppler setup --config sandbox
```

#### "Secrets not syncing"
```bash
# Refresh secrets cache
doppler secrets --refresh

# Check secrets for specific config
doppler secrets --config sandbox
```

### Debug Mode

```bash
DOPPLER_DEBUG=1 doppler run -- your-command
```

## Security Best Practices

1. **Never commit Doppler tokens** to version control
2. **Use short-lived tokens** for CI/CD (max 24 hours)
3. **Restrict secret access** by environment
4. **Enable audit logging** in Doppler dashboard
5. **Rotate secrets regularly** (minimum 90 days)
6. **Use secret scanning** to detect leaked credentials

## Support

- Doppler Documentation: https://docs.doppler.com
- Doppler Support: support@doppler.com
- Internal Slack: #infrastructure-secrets
