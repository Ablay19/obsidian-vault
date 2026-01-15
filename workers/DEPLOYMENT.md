# AI Manim Video Generator - Deployment Guide

## Overview
This guide covers the complete deployment of the AI Manim Video Generator feature, including the Cloudflare Worker and Manim Renderer service.

## Architecture

```
Telegram User → Telegram API → Cloudflare Worker → AI Provider → Manim Renderer → Video → User
```

## Components

### 1. Cloudflare Worker (`workers/ai-manim-worker/`)
- Telegram webhook handler
- Video submission endpoint
- Health check endpoints

### 2. Manim Renderer (`workers/manim-renderer/`)
- Flask-based video rendering service
- Docker container with Manim v0.18.1
- Job queue management

## Deployment

### Cloudflare Worker

```bash
# Deploy to staging
cd workers/ai-manim-worker
npx wrangler deploy --env staging

# Deploy to production
npx wrangler deploy --env production
```

### Telegram Webhook Configuration

The webhook is automatically configured when you deploy. To verify:

```bash
curl https://api.telegram.org/bot<TOKEN>/getWebhookInfo
```

### Manim Renderer

#### Option 1: Railway Deployment

```bash
cd workers/manim-renderer

# Install Railway CLI
npm install -g @railway/cli

# Login and deploy
railway login
railway init
railway up
```

#### Option 2: Docker Local Testing

```bash
cd workers/manim-renderer
docker-compose up -d
```

## Environment Variables

### Worker (`.env.staging`)
```env
TELEGRAM_BOT_TOKEN=your-bot-token
TELEGRAM_SECRET=your-webhook-secret
LOG_LEVEL=info
```

### Renderer (`.env`)
```env
PORT=8080
MANIM_QUALITY=medium
```

## Testing

### Unit Tests
```bash
cd workers/ai-manim-worker
npm test
```

### Integration Tests
```bash
npm run test:integration
```

### Telegram Integration
Send a message to your bot with a mathematical problem:
```
Explain the Pythagorean theorem
```

## URLs

- **Staging Worker**: https://ai-manim-worker-staging.abdoullahelvogani.workers.dev
- **Health Check**: https://ai-manim-worker-staging.abdoullahelvogani.workers.dev/health
- **Webhook**: https://ai-manim-worker-staging.abdoullahelvogani.workers.dev/telegram/webhook

## Troubleshooting

### Webhook Not Receiving Messages
1. Verify webhook is set: `curl https://api.telegram.org/bot<TOKEN>/getWebhookInfo`
2. Check worker logs: `npx wrangler tail --name ai-manim-worker-staging`
3. Ensure secrets are set: `npx wrangler secret list --name ai-manim-worker-staging`

### Video Rendering Fails
1. Check renderer health: `curl http://localhost:8080/health`
2. Verify Manim installation: `docker exec manim-renderer manim --version`

## Git History
```
9626569 feat: Complete Telegram webhook and integration testing
c1d2b86 feat: Add Manim renderer service and Docker container
86dfc44 feat: Add video handler for render pipeline
a2ca4f4 feat: Add Telegram webhook handler and implementation tasks
ccc7f3c feat: Complete 006-ai-manim-video implementation plan
```

## Next Steps
1. [ ] Configure Telegram bot commands via BotFather
2. [ ] Deploy Manim renderer to Railway
3. [ ] Set up Cloudflare R2 for video storage
4. [ ] Add AI code generation integration
5. [ ] Configure production secrets
