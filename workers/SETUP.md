# AI Manim Video Generator - Setup Guide

## Quick Start

### 1. Create Cloudflare KV Namespace

```bash
cd workers/ai-manim-worker

# Create KV namespace
npx wrangler kv:namespace create "SESSIONS"

# This will output JSON like:
# { binding: "SESSIONS", id: "xxxxxxxxxxxx", preview_id: "yyyyyyyyyy" }

# Update wrangler.toml with the IDs
```

Or use the helper script:
```bash
chmod +x scripts/create-kv.sh
./scripts/create-kv.sh
```

### 2. Set Worker Secrets

```bash
# Telegram bot token
npx wrangler secret put TELEGRAM_BOT_TOKEN --env staging

# Telegram webhook secret
npx wrangler secret put TELEGRAM_SECRET --env staging

# AI Provider keys (optional, defaults configured)
npx wrangler secret put GROQ_API_KEY --env staging
npx wrangler secret put HUGGINGFACE_API_KEY --env staging
```

### 3. Deploy Manim Renderer to Render.com

1. Create account at [render.com](https://render.com)
2. Connect your GitHub repository
3. Create a new Web Service:
   - Build Command: `pip install -r requirements.txt`
   - Start Command: `gunicorn src.app:app --bind 0.0.0.0:$PORT --workers 2 --timeout 300`
   - Environment Variables:
     - `PORT`: 8080
     - `MANIM_QUALITY`: medium
4. Copy the service URL

### 4. Configure Renderer URL

```bash
# Set the renderer URL as a secret
npx wrangler secret put MANIM_RENDERER_URL --env staging
# Enter your Render.com service URL (e.g., https://manim-renderer.onrender.com)
```

### 5. Test with Telegram

Message @WhatsAppToObsidian_bot on Telegram:
- Send `/start` for help
- Send a mathematical problem like "Explain the Pythagorean theorem"

## Architecture

```
Telegram → Cloudflare Worker → AI Provider → Render.com Renderer → Ephemeral URL → User
              ↓
         Cloudflare KV
         (Session storage)
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| TELEGRAM_BOT_TOKEN | Yes | Telegram bot token |
| TELEGRAM_SECRET | Yes | Webhook authentication secret |
| MANIM_RENDERER_URL | Yes | Render.com service URL |
| GROQ_API_KEY | No | Groq API key (fallback) |
| HUGGINGFACE_API_KEY | No | HuggingFace API key (fallback) |
| USE_MOCK_RENDERER | No | 'true' for testing without renderer |

## Development

```bash
# Start worker in development mode (uses mock renderer)
npx wrangler dev --env development

# Run tests
npm test

# Deploy to staging
npx wrangler deploy --env staging
```

## Files

- `workers/ai-manim-worker/` - Cloudflare Worker
- `workers/manim-renderer/` - Render.com Flask app
- `specs/006-ai-manim-video/` - Feature specification

## Troubleshooting

### Telegram webhook not receiving messages
1. Check webhook URL: `https://api.telegram.org/bot<TOKEN>/getWebhookInfo`
2. Verify secrets: `npx wrangler secret list --env staging`
3. Check worker logs: `npx wrangler tail --env staging`

### Video generation fails
1. Check renderer health: `curl https://<renderer-url>/health`
2. Verify MANIM_RENDERER_URL is set as secret
3. Check worker logs for AI generation errors

### Session not found
1. Create KV namespace: `npx wrangler kv:namespace create "SESSIONS"`
2. Update wrangler.toml with KV IDs
3. Redeploy worker
