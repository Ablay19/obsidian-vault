# Cloudflare Workers Deployment Guide

Complete guide for deploying Obsidian Bot as a serverless application using Cloudflare Workers with AI fallback capabilities.

## 📋 Table of Contents

1. [Overview](#overview)
2. [Architecture](#architecture)
3. [Prerequisites](#prerequisites)
4. [Quick Start](#quick-start)
5. [Configuration](#configuration)
6. [Deployment](#deployment)
7. [Features](#features)
8. [Monitoring](#monitoring)
9. [Troubleshooting](#troubleshooting)

## 🎯 Overview

The Obsidian Bot Cloudflare Workers deployment provides:
- **Serverless Architecture**: No servers to manage or scale
- **Global Edge Network**: Automatic worldwide distribution
- **Built-in AI**: Cloudflare Workers AI with fallback providers
- **Cost-Effective**: Pay-per-request pricing model
- **High Availability**: 99.9%+ uptime SLA
- **Auto-Scaling**: Handles traffic spikes automatically

## 🏗️ Architecture

```
┌─────────────────────────────────────────────┐
│                 Telegram API                │
│                     ↑                    │
│                     │                    │
│              ┌────┴────┐              │
│              │  Cloudflare │              │
│              │   Workers   │              │
│              └─────┬─────┘              │
│                      │                    │
│           ┌────────┴────────┐           │
│           │  AI Services     │           │
│           │  • Built-in AI    │           │
│           │  • Gemini        │           │
│           │  • Groq          │           │
│           └─────────────────┘           │
│                      │                    │
│           ┌────────────────┐           │
│           │  Data Storage  │           │
│           │  • KV Store      │           │
│           │  • R2 Storage   │           │
│           │  • D1 Database  │           │
│           └────────────────┘           │
└─────────────────────────────────────────────┘
```

## 🚀 Prerequisites

### Required Tools
```bash
# Node.js and npm
node --version  # >= 18.0.0
npm --version

# Cloudflare Wrangler
npm install -g wrangler

# Git for version control
git --version

# Telegram Bot Token
export TELEGRAM_BOT_TOKEN="your-bot-token"
```

### Cloudflare Account Setup
1. **Create Account**: [Cloudflare Dashboard](https://dash.cloudflare.com)
2. **Enable Workers**: Go to Workers & Pages
3. **Configure AI**: Enable Workers AI in AI section
4. **Set Billing**: Add payment method for AI usage

### Required Services
- **Workers**: For serverless execution
- **KV Store**: For bot state and user data
- **R2 Storage** (optional): For file uploads
- **D1 Database** (optional): For persistent data
- **Workers AI** (optional): Built-in AI capabilities

## ⚡ Quick Start

### 1. One-Command Setup
```bash
# Install dependencies and deploy
./workers/setup-telegram-webhook.sh setup
```

### 2. Manual Setup
```bash
# Set Telegram token
export TELEGRAM_BOT_TOKEN="your-bot-token"

# Setup webhook
./workers/setup-telegram-webhook.sh setup

# Deploy workers
./workers/deploy.sh deploy production

# Test deployment
curl https://obsidian-bot-workers.your-username.workers.dev/health
```

### 3. Verify Deployment
```bash
# Check bot status
./workers/deploy.sh info

# Test webhook
./workers/setup-telegram-webhook.sh test

# Check AI providers
curl https://obsidian-bot-workers.your-username.workers.dev/ai-test
```

## ⚙️ Configuration

### 1. wrangler.toml

#### Basic Configuration
```toml
name = "obsidian-bot-workers"
main = "obsidian-bot-worker.js"
compatibility_date = "2024-01-01"
compatibility_flags = ["nodejs_compat"]

[env.production]
name = "obsidian-bot-workers-prod"
routes = [
  { pattern = "api.obsidian-bot.com/*", zone_name = "obsidian-bot.com" }
]
```

#### AI Services Configuration
```toml
[ai]
binding = "AI"

[env.production.vars]
ENVIRONMENT = "production"
LOG_LEVEL = "info"
FALLBACK_AI_ENABLED = "true"
```

#### Storage Configuration
```toml
[[env.production.kv_namespaces]]
binding = "BOT_STATE"
id = "your-kv-namespace-id"
preview_id = "your-preview-id"

[[env.production.r2_buckets]]
binding = "MEDIA_STORAGE"
bucket_name = "obsidian-bot-media"
```

### 2. Environment Variables

| Variable | Required | Description |
|-----------|-----------|------------|
| `TELEGRAM_BOT_TOKEN` | Yes | Telegram bot authentication token |
| `ENVIRONMENT` | No | Environment (production/staging/development) |
| `LOG_LEVEL` | No | Logging level (debug/info/warn/error) |
| `FALLBACK_AI_ENABLED` | No | Enable AI fallback system (true/false) |
| `WEBHOOK_SECRET` | Yes | Webhook verification secret |

### 3. AI Provider Configuration

#### Built-in Workers AI
```javascript
// Automatically available with AI binding
const response = await env.AI.run('@cf/meta/llama-3-8b-instruct', {
    prompt: 'Hello, world!'
});
```

#### External AI Providers
```javascript
// Gemini API
const response = await fetch('https://generativelanguage.googleapis.com/v1beta/models/gemini-pro:generateContent', {
    method: 'POST',
    headers: {
        'Authorization': `Bearer ${env.GEMINI_API_KEY}`,
        'Content-Type': 'application/json'
    },
    body: JSON.stringify({
        contents: [{ parts: [{ text: prompt }] }]
    })
});
```

## 🚢 Deployment

### 1. Production Deployment
```bash
# Deploy to production
./workers/deploy.sh deploy production

# Check deployment status
wrangler deployments list --env production

# Health check
curl https://api.obsidian-bot.com/health
```

### 2. Multi-Environment Deployment
```bash
# Deploy to all environments
for env in production staging development; do
    echo "Deploying to $env..."
    ./workers/deploy.sh deploy $env
done
```

### 3. Preview Deployment
```bash
# Start preview server
./workers/deploy.sh preview

# Test changes locally
curl -X POST http://localhost:8787/ai \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello"}'
```

### 4. Custom Domains
```bash
# Set up custom domain
wrangler custom-domains create api.obsidian-bot.com

# Update routes
wrangler routes create api.obsidian-bot.com/ai --zone-name obsidian-bot.com

# SSL/TLS automatically handled by Cloudflare
```

## 🌟 Features

### 1. Intelligent AI Fallback
The system automatically falls back between AI providers:

1. **Cloudflare Workers AI** (primary, built-in)
2. **Gemini API** (secondary, external)
3. **Groq API** (tertiary, external)

#### Fallback Logic
```javascript
async function processWithFallback(prompt) {
    const providers = [cloudflareAI, gemini, groq];
    
    for (const provider of providers) {
        try {
            const result = await provider.process(prompt);
            if (result.success) return result;
        } catch (error) {
            console.log(`${provider.name} failed: ${error}`);
            continue;
        }
    }
    
    return {
        success: false,
        error: 'All AI providers unavailable'
    };
}
```

### 2. File Processing
#### Image Analysis
- Automatic metadata extraction
- AI-powered image descriptions
- Size and format validation
- Processing status tracking

#### Document Handling
- PDF file reception and queuing
- Asynchronous processing workflow
- Cloud Storage integration
- Progress tracking and notifications

#### Text Processing
- Real-time AI responses
- Conversation context management
- Multi-language support
- Custom instruction handling

### 3. State Management
#### KV Store Integration
- User preferences and settings
- Conversation history
- Processing status tracking
- Provider configuration

#### Data Persistence
```javascript
class BotState {
    async getUserState(userId) {
        const state = await BOT_STATE.get(`user:${userId}`);
        return state ? JSON.parse(state) : null;
    }
    
    async setUserState(userId, state) {
        await BOT_STATE.put(`user:${userId}`, JSON.stringify(state), {
            expirationTtl: 86400 // 24 hours
        });
    }
}
```

### 4. Command System
#### Built-in Commands
- `/start` - Welcome message
- `/help` - Command help
- `/status` - Current status
- `/provider` - AI provider status
- `/health` - System health check

#### Custom Commands
- Easily extensible command system
- Argument parsing
- Permission checks
- Error handling

## 📊 Monitoring

### 1. Health Checks
```javascript
// Health endpoint
export default {
    async fetch(request, env) {
        if (request.url.includes('/health')) {
            const health = {
                status: 'ok',
                timestamp: new Date().toISOString(),
                services: await checkAllServices()
            };
            
            return new Response(JSON.stringify(health), {
                headers: { 'Content-Type': 'application/json' }
            });
        }
    }
};
```

### 2. Metrics Collection
```javascript
// Request metrics
const metrics = {
    requests: 0,
    errors: 0,
    ai_responses: 0,
    provider_failures: {}
};

// Analytics endpoint
export default {
    async fetch(request, env) {
        if (request.url.includes('/metrics')) {
            return new Response(JSON.stringify(metrics), {
                headers: { 'Content-Type': 'application/json' }
            });
        }
    }
};
```

### 3. Logging System
#### Structured Logging
```javascript
class Logger {
    static async log(message, level = 'info') {
        const logEntry = {
            timestamp: new Date().toISOString(),
            level,
            environment: env.ENVIRONMENT,
            message,
            source: 'cloudflare-worker'
        };

        console.log(JSON.stringify(logEntry));
        
        // Send to external logging service
        if (env.LOG_ENDPOINT) {
            await fetch(env.LOG_ENDPOINT, {
                method: 'POST',
                body: JSON.stringify(logEntry)
            });
        }
    }
}
```

### 4. Real-Time Monitoring
#### Cloudflare Analytics
```javascript
// Using Cloudflare Web Analytics
const analytics = {
    page_views: 0,
    api_requests: 0,
    error_rate: 0
};

// Track in request handler
analytics.api_requests++;
```

#### Custom Dashboards
```javascript
// Custom metrics dashboard
export default {
    async fetch(request, env) {
        if (request.url.includes('/dashboard')) {
            const dashboard = {
                total_requests: metrics.requests,
                success_rate: ((metrics.requests - metrics.errors) / metrics.requests * 100).toFixed(2),
                provider_health: await getProviderHealth(),
                active_conversations: await getActiveConversations()
            };
            
            return new Response(JSON.stringify(dashboard), {
                headers: { 'Content-Type': 'application/json' }
            });
        }
    }
};
```

## 🔧 Troubleshooting

### 1. Common Issues

#### Worker Not Responding
```bash
# Check deployment status
wrangler deployments list

# Check logs
wrangler tail --format=pretty

# Test health endpoint
curl -v https://api.obsidian-bot.com/health
```

#### Telegram Webhook Issues
```bash
# Check webhook configuration
curl -X POST https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getWebhookInfo

# Test webhook manually
curl -X POST https://api.obsidian-bot.com/webhook \
  -H "Content-Type: application/json" \
  -d '{"message": {"text": "test"}}'
```

#### AI Provider Issues
```bash
# Check AI availability
curl https://api.obsidian-bot.com/ai-test

# Test individual providers
curl -X POST https://api.obsidian-bot.com/ai \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello", "provider": "cloudflare"}'

# Test external providers
curl -X POST https://api.obsidian-bot.com/ai \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello", "provider": "gemini"}'
```

### 2. Debugging Commands

#### Local Development
```bash
# Start local development server
wrangler dev --local --port 8787

# Test with curl
curl -X POST http://localhost:8787/ai \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Test"}'
```

#### Remote Debugging
```bash
# Stream logs in real-time
wrangler tail --env production --format=json

# Check specific errors
wrangler tail --format=json | grep "ERROR"
```

### 3. Performance Optimization

#### Worker Optimization
```javascript
// Cache AI responses
const cache = caches.default;

export default {
    async fetch(request, env) {
        const cacheKey = new Request(request.url);
        const cached = await cache.match(cacheKey);
        
        if (cached) {
            return cached;
        }
        
        // Process request
        const response = await processRequest(request);
        
        // Cache response
        ctx.waitUntil(cache.put(cacheKey, response.clone()));
        
        return response;
    }
};
```

#### Memory Management
```javascript
// Limit concurrent processing
const CONCURRENT_LIMIT = 10;
let currentProcessing = 0;

export default {
    async fetch(request, env) {
        if (currentProcessing >= CONCURRENT_LIMIT) {
            return new Response('Rate limited', { status: 429 });
        }
        
        currentProcessing++;
        try {
            const response = await processRequest(request);
            return response;
        } finally {
            currentProcessing--;
        }
    }
};
```

## 📚 Additional Resources

### Documentation
- [Cloudflare Workers Documentation](https://developers.cloudflare.com/workers/)
- [Workers AI Documentation](https://developers.cloudflare.com/workers-ai/)
- [Wrangler CLI Documentation](https://developers.cloudflare.com/workers/wrangler/)
- [Telegram Bot API Documentation](https://core.telegram.org/bots/api)

### Tools and SDKs
- **Wrangler CLI**: `npm install -g wrangler`
- **Telegram Bot SDK**: `npm install telegram-bot-api`
- **Cloudflare SDK**: Built into Workers runtime

### Community and Support
- **Cloudflare Discord**: [https://discord.cloudflare.com](https://discord.cloudflare.com)
- **Telegram Bot Community**: [https://core.telegram.org/bots/faq](https://core.telegram.org/bots/faq)
- **Stack Overflow**: `cloudflare-workers` tag

## 🎯 Deployment Summary

The Cloudflare Workers deployment provides:

### ✅ **Benefits**
- **Zero Infrastructure**: No servers to manage
- **Global Edge**: Automatic worldwide distribution
- **Built-in CDN**: Static assets served from edge
- **Auto-Scaling**: Handles any traffic volume
- **Cost Predictable**: Pay-per-request pricing

### 🌟 **Features**
- **AI Fallback**: Intelligent provider switching
- **State Persistence**: KV Store for user data
- **File Processing**: Image and document handling
- **Real-time Analytics**: Built-in monitoring
- **Custom Domains**: Easy SSL/TLS setup

### 🚀 **Performance**
- **Sub-second Response**: Edge computing speeds
- **99.9%+ Uptime**: Cloudflare reliability
- **Auto-recovery**: Built-in failover mechanisms
- **Global Coverage**: 200+ edge locations

---

**Deploy your Obsidian Bot to Cloudflare Workers for a truly serverless, globally distributed, cost-effective AI-powered bot! 🌍**