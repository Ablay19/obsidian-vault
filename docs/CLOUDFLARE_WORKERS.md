# Cloudflare Workers Deployment Documentation

This document covers the deployment and configuration of all Cloudflare Workers in the obsidian-vault project.

## Overview

The project includes the following Cloudflare Workers:

| Worker | Location | Purpose | URL |
|--------|----------|---------|-----|
| **ai-worker** | `workers/ai-worker/` | AI processing worker | `https://ai-worker.abdoullahelvogani.workers.dev` |
| **image-worker** | `workers/image-worker/` | Image processing worker | `https://image-worker.abdoullahelvogani.workers.dev` |
| **obsidian-bot-workers** | `workers/` | Main bot API with AI proxy | `https://obsidian-bot-workers.abdoullahelvogani.workers.dev` |

## Prerequisites

### Required Tools
- Node.js 20+
- Wrangler CLI (`npm install -g wrangler`)
- Cloudflare account

### Authentication
```bash
# Login to Cloudflare
wrangler auth login

# Verify authentication
wrangler whoami
```

## Secrets Management

### Using Doppler (Recommended)

The project uses Doppler for secrets management. Configure your environment:

```bash
# Install Doppler CLI
npm install -g @dopplerhq/cli

# Login to Doppler
doppler login

# Set project config
export DOPPLER_CONFIG=dev
export DOPPLER_ENVIRONMENT=dev
export DOPPLER_PROJECT=bot
```

### Required Secrets

All workers require the following secrets to be set in Doppler:

```bash
# Required for obsidian-bot-workers
TELEGRAM_BOT_TOKEN=your-telegram-bot-token
GEMINI_API_KEYS=your-gemini-api-keys
GROQ_API_KEY=your-groq-api-key
OPENAI_API_KEY=your-openai-api-key
HUGGINGFACE_API_KEY=your-huggingface-api-key
OPENROUTER_API_KEY=your-openrouter-api-key
```

### Syncing Secrets to Workers

```bash
# Sync secrets to ai-worker
doppler secrets sync --project ai-worker

# Sync secrets to image-worker  
doppler secrets sync --project image-worker

# Sync secrets to obsidian-bot-workers
doppler secrets sync --project obsidian-bot-workers
```

### Manual Secret Management (Alternative)

If not using Doppler, set secrets directly:

```bash
# For obsidian-bot-workers
echo "$TELEGRAM_BOT_TOKEN" | wrangler secret put TELEGRAM_BOT_TOKEN --name obsidian-bot-workers
echo "$GEMINI_API_KEYS" | wrangler secret put GEMINI_API_KEYS --name obsidian-bot-workers
echo "$GROQ_API_KEY" | wrangler secret put GROQ_API_KEY --name obsidian-bot-workers
```

## Deployment

### Quick Deploy All Workers

```bash
# Deploy all workers
cd workers

# Deploy ai-worker
cd ai-worker && npm run deploy

# Deploy image-worker
cd image-worker && npm run deploy

# Deploy main bot worker
cd .. && wrangler deploy
```

### Deploy by Environment

```bash
# Deploy to development
npm run deploy -- --env development

# Deploy to staging
npm run deploy -- --env staging

# Deploy to production
npm run deploy -- --env production
```

### Using Deployment Script

```bash
# Deploy to production
./workers/deploy.sh deploy production

# Deploy to staging
./workers/deploy.sh deploy staging

# Deploy all environments
./workers/deploy.sh deploy-all

# Preview deployment (local dev)
./workers/deploy.sh preview
```

## Worker Details

### 1. AI Worker (`workers/ai-worker/`)

**Purpose:** Handles AI-related requests at the edge

**Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/workers` | List workers via API Gateway |
| GET | `/api/v1/workers/:id` | Get specific worker |

**Environment Variables:**
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `API_GATEWAY_URL`: URL of the API Gateway service

**Commands:**
```bash
# Install dependencies
cd workers/ai-worker && npm install

# Local development
npm run dev

# Deploy
npm run deploy

# Run tests
npm test
```

### 2. Image Worker (`workers/image-worker/`)

**Purpose:** Handles image processing at the edge

**Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| POST | `/process` | Process uploaded image |
| GET | `/api/v1/images` | List processed images |
| GET | `/api/v1/images/:id` | Get specific image |

**Environment Variables:**
- `LOG_LEVEL`: Logging level
- `API_GATEWAY_URL`: URL of the API Gateway service
- `AI_GATEWAY_URL`: URL of the AI Gateway service

**Commands:**
```bash
cd workers/image-worker
npm install
npm run dev     # Local development
npm run deploy  # Deploy to Cloudflare
```

### 3. Obsidian Bot Workers (`workers/`)

**Purpose:** Main bot API with Telegram handling and AI proxy capabilities

**Endpoints:**
| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check with AI provider status |
| POST | `/ai` | AI request proxy |
| GET | `/metrics` | Worker metrics |

**Features:**
- Multi-provider AI support (GPT-4, Claude, Gemini, Groq, Cloudflare)
- Caching via KV
- Rate limiting
- Cost optimization
- Analytics
- Fallback providers

**Environment Variables:**
- `ENVIRONMENT`: deployment environment
- `LOG_LEVEL`: logging level
- `FALLBACK_AI_ENABLED`: enable fallback providers
- `AI_ENABLED`: enable AI features

**Secrets Required:**
- `TELEGRAM_BOT_TOKEN`: Telegram bot token
- `GEMINI_API_KEYS`: Gemini API keys (comma-separated)
- `GROQ_API_KEY`: Groq API key
- `OPENAI_API_KEY`: OpenAI API key
- `HUGGINGFACE_API_KEY`: HuggingFace API key
- `OPENROUTER_API_KEY`: OpenRouter API key

## CI/CD Deployment

### GitHub Actions

The project includes GitHub Actions workflows for automatic deployment:

- `.github/workflows/ai-worker-deploy.yml`: Deploys ai-worker
- `.github/workflows/ai-worker-ci.yml`: CI pipeline for ai-worker

#### Deployment Triggers

| Branch | Environment | Trigger |
|--------|-------------|---------|
| `develop` | Staging | Auto on push |
| `main` | Production | Auto on push |

#### Required GitHub Secrets

Add these secrets to your GitHub repository:

1. `CLOUDFLARE_API_TOKEN`: Cloudflare API token with Workers write permissions

### Manual Rollback

```bash
# List deployments
wrangler deployments list --name obsidian-bot-workers

# Rollback to specific version
wrangler rollback <version-id> --name obsidian-bot-workers
```

## Local Development

### Start All Workers Locally

```bash
# Terminal 1: AI Worker
cd workers/ai-worker && npm run dev

# Terminal 2: Image Worker
cd workers/image-worker && npm run dev

# Terminal 3: Main Bot Worker
cd workers && npx wrangler dev
```

### Using Docker Compose

```bash
# Start local development environment
docker-compose -f .docker-compose.dev.yml up -d
```

## Monitoring & Debugging

### View Logs

```bash
# Tail logs for a worker
wrangler tail obsidian-bot-workers

# Tail all workers
wrangler tail --name ai-worker
```

### Health Checks

```bash
# Check all workers
curl https://ai-worker.abdoullahelvogani.workers.dev/health
curl https://image-worker.abdoullahelvogani.workers.dev/health
curl https://obsidian-bot-workers.abdoullahelvogani.workers.dev/health
```

### Metrics

```bash
# Get worker metrics
curl https://obsidian-bot-workers.abdoullahelvogani.workers.dev/metrics
```

## Troubleshooting

### Common Issues

#### 1. Secret Not Found
```bash
# Verify secrets are set
wrangler secret list --name obsidian-bot-workers

# Resync from Doppler
doppler secrets sync --project obsidian-bot-workers
```

#### 2. KV Namespace Not Found
```bash
# Create KV namespace
wrangler kv:namespace create "BOT_STATE"

# Add binding to wrangler.toml
```

#### 3. Deployment Fails
```bash
# Check wrangler version
wrangler --version

# Update wrangler
npm install -g wrangler@latest
```

### Debug Mode

Set `LOG_LEVEL=debug` in your environment variables for detailed logging.

## Architecture

```
                    ┌─────────────────────┐
                    │   Telegram Bot      │
                    │   (Main Server)     │
                    └──────────┬──────────┘
                               │
                               ▼
┌─────────────────────────────────────────────────────────────────┐
│                    Cloudflare Workers                           │
├─────────────────┬─────────────────┬─────────────────────────────┤
│   ai-worker     │  image-worker   │   obsidian-bot-workers      │
│  (AI at edge)   │ (Images at edge)│   (API + AI Proxy)          │
└─────────────────┴─────────────────┴─────────────────────────────┘
                               │
                               ▼
                    ┌─────────────────────┐
                    │   AI Providers      │
                    │  (Cloudflare AI,    │
                    │   Gemini, Groq,     │
                    │   OpenAI, etc.)     │
                    └─────────────────────┘
```

## Configuration Files

| File | Purpose |
|------|---------|
| `workers/ai-worker/wrangler.toml` | AI Worker configuration |
| `workers/image-worker/wrangler.toml` | Image Worker configuration |
| `workers/wrangler.toml` | Main bot worker configuration |
| `.env` | Local environment variables |
| `.docker-compose.dev.yml` | Docker development environment |

## Support

For issues with:
- **Cloudflare Workers**: https://developers.cloudflare.com/workers/
- **Wrangler CLI**: https://developers.cloudflare.com/workers/wrangler/
- **Doppler**: https://docs.doppler.com/

## Version History

| Version | Date | Changes |
|---------|------|---------|
| 1.0.0 | 2026-01-14 | Initial deployment |

## Quick Reference

```bash
# Deploy ai-worker
cd workers/ai-worker && npm run deploy

# Deploy image-worker
cd workers/image-worker && npm run deploy

# Deploy main bot worker
cd workers && wrangler deploy

# Set secrets (Doppler)
doppler secrets sync --project obsidian-bot-workers

# View logs
wrangler tail obsidian-bot-workers

# Check health
curl https://obsidian-bot-workers.abdoullahelvogani.workers.dev/health
```
