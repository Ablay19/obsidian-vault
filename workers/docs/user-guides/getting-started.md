# Getting Started with Enhanced Workers

## 🎯 Welcome

Welcome to the enhanced Cloudflare Workers implementation! This guide will help you get started with the new analytics, caching, and performance monitoring features.

## 📋 Prerequisites

Before you begin, ensure you have:

- **Node.js 18+** installed
- **npm** package manager
- **Cloudflare account** with Workers enabled
- **Wrangler CLI** installed (`npm install -g wrangler`)

## 🚀 Quick Setup

### 1. Clone and Install

```bash
# Navigate to the workers directory
cd workers

# Install dependencies
npm install
```

### 2. Configure Environment

Copy the environment template and fill in your values:

```bash
cp .env.example .env
```

Edit `.env` with your configuration:

```env
# Cloudflare Workers
CLOUDFLARE_API_TOKEN=your_api_token
CLOUDFLARE_ACCOUNT_ID=your_account_id

# AI Providers (optional - for enhanced features)
OPENAI_API_KEY=your_openai_key
ANTHROPIC_API_KEY=your_anthropic_key

# Configuration
CACHE_SIZE=100
RATE_LIMIT_PER_HOUR=1000
ENABLE_ANALYTICS=true
LOG_LEVEL=info
```

### 3. Test Locally

Run the enhanced functionality test:

```bash
npm run test:enhanced
```

You should see output like:
```
🧪 Testing Enhanced Workers Functionality...

📊 Testing Analytics System...
✅ Analytics Report: {
  totalRequests: 3,
  errorRate: '33.3%',
  avgResponseTime: '33.3ms',
  healthScore: 100
}

🎉 All Enhanced Worker Tests Completed Successfully!
```

### 4. Deploy to Production

```bash
# Deploy to Cloudflare
npm run deploy

# Check deployment status
wrangler tail
```

## 🎛️ Basic Usage

### Making Requests

The enhanced workers accept the same API format as before:

```javascript
// Basic AI request
const response = await fetch('https://your-worker.workers.dev/', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Authorization': 'Bearer your-token'
  },
  body: JSON.stringify({
    prompt: "Hello, how are you?",
    model: "gpt-3.5-turbo",
    maxTokens: 100
  })
});

const result = await response.json();
console.log(result);
```

### New Features Available

#### Analytics Monitoring

Check real-time performance metrics:

```bash
curl https://your-worker.workers.dev/analytics
```

Response:
```json
{
  "totalRequests": 1250,
  "errorRate": "0.8%",
  "avgResponseTime": "45.2ms",
  "cacheHitRate": "87.3%",
  "healthScore": 95,
  "uptime": 3600000
}
```

#### Health Check

Monitor worker health:

```bash
curl https://your-worker.workers.dev/health
```

Response:
```json
{
  "status": "healthy",
  "version": "2.0.0",
  "uptime": "2h 30m",
  "memoryUsage": "8.5MB",
  "activeConnections": 12
}
```

#### Cache Status

Check cache performance:

```bash
curl https://your-worker.workers.dev/cache/status
```

Response:
```json
{
  "size": 85,
  "maxSize": 100,
  "hitRate": "87.3%",
  "evictions": 23,
  "totalSets": 1240
}
```

## 🔧 Configuration Options

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `CACHE_SIZE` | 100 | Maximum cache entries |
| `RATE_LIMIT_PER_HOUR` | 100 | Requests per hour per user |
| `ENABLE_ANALYTICS` | true | Enable performance monitoring |
| `LOG_LEVEL` | info | Logging verbosity (debug/info/warn/error) |
| `MAX_MEMORY_MB` | 10 | Memory usage limit |

### Runtime Configuration

Update `wrangler.toml` for advanced settings:

```toml
[vars]
CACHE_SIZE = "100"
RATE_LIMIT_PER_HOUR = "1000"
ENABLE_ANALYTICS = "true"

# Performance tuning
MAX_CONCURRENCY = "50"
REQUEST_TIMEOUT_MS = "30000"
CACHE_TTL_SECONDS = "3600"
```

## 📊 Understanding the Dashboard

### Analytics Dashboard

The `/analytics` endpoint provides:

- **Request Metrics**: Total requests, error rates, response times
- **Cache Performance**: Hit rates, evictions, memory usage
- **Provider Statistics**: Success rates, latency by provider
- **System Health**: Memory usage, CPU load, uptime

### Cache Analytics

Monitor cache effectiveness:

- **Hit Rate**: Percentage of requests served from cache
- **Memory Usage**: Current cache size vs. maximum
- **Eviction Rate**: How often items are removed due to capacity
- **TTL Distribution**: Cache entry age distribution

### Rate Limiting Status

Track rate limiting:

- **Active Limits**: Current rate limit counters
- **Blocked Requests**: Requests denied due to limits
- **Reset Times**: When limits will reset
- **Bursting**: Burst capacity utilization

## 🚨 Troubleshooting

### Common Issues

#### High Error Rates
```
Symptom: Error rate >5%
Solution:
1. Check provider API keys
2. Review rate limits
3. Monitor provider status
4. Enable fallback providers
```

#### Slow Response Times
```
Symptom: Response time >100ms
Solution:
1. Check cache hit rate
2. Monitor provider latency
3. Adjust concurrency settings
4. Enable performance profiling
```

#### Memory Issues
```
Symptom: Memory usage >10MB
Solution:
1. Reduce cache size
2. Enable garbage collection
3. Monitor memory leaks
4. Adjust worker limits
```

### Getting Help

1. **Check Logs**: `wrangler tail` for real-time logs
2. **Review Metrics**: Use `/analytics` endpoint
3. **Test Locally**: `npm run test:enhanced`
4. **Check Documentation**: See troubleshooting guide

## 🎯 Next Steps

Now that you have the basics running:

1. [Configure advanced settings](./configuration.md)
2. [Set up monitoring alerts](./monitoring.md)
3. [Learn about performance tuning](./operations/performance-tuning.md)
4. [Explore developer extensions](./developer-docs/extension-guide.md)

## 📞 Support

Need help? Check:
- [Troubleshooting Guide](./troubleshooting.md)
- [Developer Documentation](../developer-docs/)
- [Operations Guide](../operations/)

Happy coding with your enhanced workers! 🚀