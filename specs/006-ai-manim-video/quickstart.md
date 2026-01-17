# Quickstart: AI Manim Video Generator

**Feature**: 006-ai-manim-video
**Date**: January 17, 2026
**Platforms**: Telegram, WhatsApp, Web Dashboard

---

## Overview

The AI Manim Video Generator creates mathematical animation videos from user input across multiple platforms. Users can submit problems for AI-powered code generation or provide Manim code directly.

**Supported Platforms:**
- **Telegram**: Bot-based messaging with problem submission
- **WhatsApp**: Business API integration with direct code detection
- **Web Dashboard**: Direct code submission with real-time monitoring

---

## Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| Node.js | 18+ | TypeScript workers |
| npm | 9+ | Package management |
| Python | 3.11+ | Manim rendering |
| Docker | 24+ | Container rendering |
| wrangler | 3.x | Cloudflare Workers CLI |
| git | 2.x | Version control |

---

## Installation

### 1. Clone and Setup

```bash
# Clone repository
git clone https://github.com/Ablay19/obsidian-vault.git
cd obsidian-vault

# Checkout feature branch
git checkout 006-ai-manim-video
```

### 2. Install Worker Dependencies

```bash
cd workers/ai-manim-worker
npm install
```

### 3. Install Renderer Dependencies

```bash
cd workers/manim-renderer
docker build -t manim-renderer:latest .
```

### 4. Configure Environment

```bash
# Copy example environment
cp .env.example .env

# Edit with your values
nano .env
```

Required environment variables:

```bash
# Telegram (get from @BotFather)
TELEGRAM_BOT_TOKEN=your-bot-token
TELEGRAM_SECRET=random-string-for-webhook-validation

# Cloudflare (get from dashboard)
CLOUDFLARE_ACCOUNT_ID=your-account-id
CLOUDFLARE_API_TOKEN=your-api-token

# R2 Storage (for video storage)
R2_ACCESS_KEY_ID=your-r2-access-key
R2_SECRET_ACCESS_KEY=your-r2-secret-key
R2_BUCKET_NAME=manim-videos
R2_ACCOUNT_ID=your-cloudflare-account-id

# AI Providers (free tiers)
# Cloudflare Workers AI - no additional config needed
# Groq - get from https://console.groq.com
GROQ_API_KEY=your-groq-api-key

# HuggingFace - get from https://huggingface.co/settings/tokens
HF_TOKEN=your-hf-token

# Optional: WhatsApp Business API
WHATSAPP_API_KEY=your_whatsapp_api_key
WHATSAPP_API_SECRET=your_whatsapp_api_secret
WHATSAPP_PHONE_NUMBER_ID=your_phone_number_id
WHATSAPP_WEBHOOK_VERIFY_TOKEN=your_webhook_verify_token

# Development settings
LOG_LEVEL=debug
USE_MOCK_RENDERER=true  # Skip Docker for faster development
```

---

## Development

### Run Worker Locally

```bash
cd workers/ai-manim-worker

# Start with hot reload
npm run dev

# Or use wrangler directly
npx wrangler dev
```

The worker will be available at `http://localhost:8787`.

### Test Telegram Webhook

Use Telegram's setWebhook API:

```bash
curl -F "url=https://your-worker.workers.dev/webhook/telegram" \
     -H "X-Telegram-Bot-Api-Secret-Token: your-secret" \
     https://api.telegram.org/bot<YOUR_BOT_TOKEN>/setWebhook
```

### Test Manim Renderer Locally

```bash
cd workers/manim-renderer

# Test with sample problem
echo '"Solve: x^2 - 4 = 0"' | python test_renderer.py
```

---

## Deployment

### Deploy Worker to Cloudflare

```bash
cd workers/ai-manim-worker

# Deploy to production
npm run deploy

# Or with wrangler
npx wrangler deploy
```

### Deploy Renderer Container

```bash
cd workers/manim-renderer

# Push to Cloudflare Container
# (or deploy to your container registry)
```

### Configure Telegram Webhook

```bash
# Set production webhook
curl -F "url=https://ai-manim-worker.abdoullahelvogani.workers.dev/webhook/telegram" \
     -H "X-Telegram-Bot-Api-Secret-Token: $TELEGRAM_SECRET" \
     https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/setWebhook
```

---

## Testing

### Run Unit Tests

```bash
cd workers/ai-manim-worker
npm test
```

### Run Integration Tests

```bash
# Start worker
npm run dev &

# Run integration tests
npm run test:integration
```

### Test Coverage

```bash
npm run test:coverage
```

Expected coverage:
- Unit tests: 90%+
- Integration tests: All critical paths

---

## Project Structure

```
workers/ai-manim-worker/
├── src/
│   ├── index.ts           # Entry point
│   ├── handlers/
│   │   ├── telegram.ts    # Telegram webhook
│   │   ├── ai.ts          # AI code generation
│   │   └── video.ts       # Video pipeline
│   ├── services/
│   │   ├── session.ts     # Session management
│   │   ├── fallback.ts    # AI provider fallback
│   │   └── renderer.ts    # Manim rendering
│   ├── types/
│   │   └── index.ts       # TypeScript types
│   └── utils/
│       └── logger.ts      # Structured logging
├── tests/
│   ├── unit/
│   │   ├── session.test.ts
│   │   └── fallback.test.ts
│   └── integration/
│       └── telegram.test.ts
├── wrangler.toml
├── package.json
└── tsconfig.json

workers/manim-renderer/
├── src/
│   ├── renderer.py
│   └── cleanup.py
├── Dockerfile
└── requirements.txt
```

---

## Manual Testing

### Web Dashboard
1. Open `http://localhost:8787/dashboard`
2. Submit a mathematical problem or Manim code
3. Monitor job status and access video when ready

### Telegram Bot (if configured)
1. Start a chat with your bot
2. Send a math problem: "Plot y = x² from -5 to 5"
3. Receive video link when processing completes

### WhatsApp (if configured)
1. Send message to your WhatsApp Business number
2. Submit problem or Manim code
3. Receive video link via WhatsApp

### Direct Code API
```bash
# Submit Manim code directly
curl -X POST http://localhost:8787/api/v1/code \
  -H "Content-Type: application/json" \
  -d '{"code": "from manim import *\nclass MyScene(Scene):\n    def construct(self):\n        circle = Circle()\n        self.play(Create(circle))"}'

# Validate code syntax
curl "http://localhost:8787/api/v1/code?code=from manim import *"
```

---

## Troubleshooting

### Worker Won't Start

```bash
# Check environment variables
cat .env

# Check wrangler config
npx wrangler validate

# Check Cloudflare authentication
npx wrangler whoami
```

### Telegram Webhook Not Working

```bash
# Verify webhook is set
curl https://api.telegram.org/bot<TOKEN>/getWebhookInfo

# Check worker logs
npx wrangler tail

# Verify secret token matches
echo $TELEGRAM_SECRET
```

### Video Generation Fails

```bash
# Check Manim container logs
docker logs manim-renderer

# Verify R2 credentials
npx wrangler secret list

# Test AI providers manually
curl -X POST https://cloudflare.com/ai/generate \
  -H "Authorization: Bearer $CF_API_TOKEN" \
  -d '{"prompt": "print(1+1)"}'
```

---

## Next Steps

1. **Implement Worker**: Build handlers and services
2. **Implement Renderer**: Create Manim Docker image
3. **Write Tests**: Follow TDD - tests first
4. **Deploy Staging**: Test with limited users
5. **Deploy Production**: Full rollout

---

## Additional Resources

- [Cloudflare Workers Documentation](https://developers.cloudflare.com/workers/)
- [Telegram Bot API](https://core.telegram.org/bots/api)
- [Manim Documentation](https://docs.manim.org/)
- [Wrangler CLI](https://developers.cloudflare.com/workers/wrangler/)
