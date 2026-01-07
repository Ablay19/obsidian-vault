# Cloudflare Workers AI Setup Guide

## Quick Setup with Cloudflare Workers AI

### 1. Create Cloudflare Account
1. Sign up at [cloudflare.com](https://cloudflare.com)
2. Upgrade to Workers Paid plan ($5/month for AI access)

### 2. Deploy AI Proxy Worker
```bash
# Install Wrangler CLI
npm install -g wrangler

# Authenticate with Cloudflare
wrangler auth login

# Deploy the worker
cd workers/ai-proxy
npm install
wrangler deploy
```

### 3. Configure Environment Variables
Add to your `.env` file:

```bash
# Cloudflare Workers AI (Recommended Default)
CLOUDFLARE_WORKER_URL=https://your-worker.your-subdomain.workers.dev

# Set Cloudflare as default provider
ACTIVE_PROVIDER=Cloudflare

# Optional: Keep other providers as fallback
GEMINI_API_KEY=your_gemini_key
GROQ_API_KEY=your_groq_key
```

### 4. Worker Features
The AI Proxy Worker includes:
- **Multiple AI Provider Support**: Cloudflare, Gemini, Groq, Claude, GPT-4
- **Intelligent Caching**: Smart caching based on content type
- **Rate Limiting**: Prevents abuse per IP and provider
- **Cost Optimization**: Selects cheapest/fastest provider automatically
- **Analytics**: Tracks usage, costs, and performance
- **Fallback Logic**: Automatic failover between providers

### 5. Available Models
- **Cloudflare**: `@cf/meta/llama-3-8b-instruct` (Free, fast)
- **Gemini**: `gemini-1.5-flash`, `gemini-1.5-pro`
- **Groq**: `llama-3.1-8b`, `mixtral-8x7b`, `gemma-7b`
- **OpenRouter**: Access to 100+ models

### 6. Test Configuration
```bash
# Restart your bot
./bot

# Test via API
curl -X POST http://localhost:8080/api/v1/qa \
  -d "question=What is 2+2?"

# Check provider status
curl http://localhost:8080/api/v1/ai/providers
```

### 7. Monitor Worker
The worker provides analytics at:
- `https://your-worker.workers.dev/status` - Health check
- `https://your-worker.workers.dev/ai-test` - Test AI binding
- Cloudflare Dashboard → Analytics → Workers

## Benefits of Cloudflare Workers AI

✅ **Free Tier**: 100K requests/day for Llama 3 8B
✅ **Low Latency**: Edge computing, global network
✅ **High Availability**: Built-in redundancy and scaling
✅ **Cost Effective**: $5/month includes 10M AI requests
✅ **Privacy**: Data stays in Cloudflare's network
✅ **Easy Setup**: No infrastructure management

## Worker Configuration

The worker automatically:
- Caches identical prompts for 1-24 hours based on content type
- Rate limits to 60 requests/minute per IP
- Selects optimal provider based on cost, latency, availability
- Tracks usage analytics for monitoring
- Handles retries and fallbacks automatically

## Troubleshooting

1. **Worker not responding**: Check deployment logs with `wrangler tail`
2. **AI binding error**: Ensure Workers Paid plan is active
3. **High latency**: Check `x-response-time` header in responses
4. **Rate limited**: Wait 60 seconds or use different IP
5. **Cache issues**: Add cache-busting parameter to requests

Your bot will now use Cloudflare Workers AI as the default provider with automatic fallback to other configured providers!