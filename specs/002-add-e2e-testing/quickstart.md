# Quickstart: E2E Testing with Doppler Environment Variables

**Feature**: 002-add-e2e-testing
**Date**: January 18, 2026

## Overview

This guide shows how to set up and run end-to-end tests using Doppler for secure environment variable management. Get started in 5 steps.

## Prerequisites

- Doppler account (free tier available)
- Mauritania CLI project
- Go 1.21+ for testing
- Access to test credentials (API keys, tokens)

## Step 1: Install Doppler CLI

### Linux
```bash
curl -Ls --tlsv1.2 --proto "=https" --retry 3 https://cli.doppler.com/install.sh | sudo sh
```

### macOS
```bash
brew install dopplerhq/cli/doppler
```

### Verify Installation
```bash
doppler --version
```

## Step 2: Authenticate with Doppler

```bash
doppler login
```

Follow the browser login flow to authenticate.

## Step 3: Set Up Doppler Project

### Create Project and Environments
```bash
# Create project
doppler projects create mauritania-cli

# Create environments
doppler environments create dev --project mauritania-cli
doppler environments create staging --project mauritania-cli
doppler environments create e2e --project mauritania-cli
```

### Add Test Credentials
```bash
# Development environment
doppler secrets set TELEGRAM_BOT_TOKEN "your_bot_token" --project mauritania-cli --config dev
doppler secrets set WHATSAPP_API_KEY "your_whatsapp_key" --project mauritania-cli --config dev
doppler secrets set FACEBOOK_APP_ID "your_app_id" --project mauritania-cli --config dev
doppler secrets set FACEBOOK_APP_SECRET "your_app_secret" --project mauritania-cli --config dev

# E2E test environment (use test/sandbox credentials)
doppler secrets set TELEGRAM_BOT_TOKEN "test_bot_token" --project mauritania-cli --config e2e
doppler secrets set TEST_DATABASE_URL "sqlite://:memory:" --project mauritania-cli --config e2e
```

## Step 4: Set Up CI/CD Service Tokens

### Generate Service Token for CI/CD
```bash
# Create token for E2E testing
doppler service-tokens create e2e-testing --project mauritania-cli --config e2e
```

This outputs a token like `dp.st.e2e.xxx`. Save this securely.

### Configure CI/CD Environment
Set the service token as `DOPPLER_TOKEN` in your CI/CD pipeline:

```yaml
# GitHub Actions example
env:
  DOPPLER_TOKEN: ${{ secrets.DOPPLER_TOKEN }}

# GitLab CI example
variables:
  DOPPLER_TOKEN: $DOPPLER_TOKEN
```

## Step 5: Run E2E Tests with Doppler

### Local Development
```bash
# Run tests with Doppler environment injection
doppler run --project mauritania-cli --config dev -- go test ./tests/e2e/...

# Or save to .env file for local development
doppler secrets download --project mauritania-cli --config dev --format env > .env
```

### CI/CD Pipeline
```yaml
# GitHub Actions example
name: E2E Tests
on: [push, pull_request]

jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: dopplerhq/cli-action@v3
      - run: go test ./tests/e2e/...
        env:
          DOPPLER_TOKEN: ${{ secrets.DOPPLER_TOKEN }}
```

## Environment Variable Management

### Available Variables

The following environment variables are managed by Doppler:

| Variable | Description | Required |
|----------|-------------|----------|
| `TELEGRAM_BOT_TOKEN` | Telegram bot API token | Yes |
| `WHATSAPP_API_KEY` | WhatsApp Business API key | Yes |
| `FACEBOOK_APP_ID` | Facebook App ID | No |
| `FACEBOOK_APP_SECRET` | Facebook App Secret | No |
| `TEST_DATABASE_URL` | Test database connection | No |

### Fallback Configuration

If Doppler is unavailable, create a `.env` file:

```bash
# .env file for local development
TELEGRAM_BOT_TOKEN=your_local_token
WHATSAPP_API_KEY=your_local_key
TEST_DATABASE_URL=sqlite://./test.db
```

## Troubleshooting

### Doppler CLI Issues
```bash
# Check authentication
doppler me

# List projects
doppler projects

# Check environment variables
doppler secrets --project mauritania-cli --config dev
```

### Test Failures
- **Missing credentials**: Check Doppler configuration and token
- **Network timeouts**: Verify Doppler service availability
- **Permission denied**: Ensure proper Doppler project access

### CI/CD Issues
- **Token expired**: Regenerate service token
- **Environment mismatch**: Verify correct config selection
- **Rate limiting**: Doppler has API rate limits

## Advanced Usage

### Multiple Test Environments
```bash
# Run tests against different environments
doppler run --config dev -- go test ./tests/integration/...
doppler run --config staging -- go test ./tests/e2e/...
```

### Environment-Specific Configurations
```bash
# Override variables for specific tests
DOPPLER_CONFIG=dev go test -run TestSpecificFeature
```

### Local Development with .env
```bash
# Download environment to .env file
doppler secrets download --format env > .env

# Use with godotenv in tests
go test ./tests/e2e/...
```

## Security Best Practices

- **Never commit .env files** to version control
- **Use service tokens** for CI/CD, not personal tokens
- **Rotate tokens regularly** in production
- **Limit token scopes** to specific projects/configs
- **Monitor token usage** in Doppler dashboard

## Next Steps

- Configure additional test environments
- Set up automated test reporting
- Integrate with test coverage tools
- Add performance benchmarking tests

The E2E testing setup is now ready with secure credential management via Doppler! 🔐