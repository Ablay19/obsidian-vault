# Configuration Guide

## 🎛️ Configuration Overview

The enhanced workers support extensive configuration options for performance tuning, security, monitoring, and feature control. This guide covers all available configuration methods and best practices.

## 📁 Configuration Files

### 1. Environment Variables (.env)

Create a `.env` file in the workers directory:

```env
# ============================================
# Enhanced Workers Configuration
# ============================================

# Cloudflare Workers
CLOUDFLARE_API_TOKEN=your_cf_api_token
CLOUDFLARE_ACCOUNT_ID=your_account_id

# AI Providers (for enhanced features)
OPENAI_API_KEY=sk-your-openai-key
ANTHROPIC_API_KEY=sk-ant-your-anthropic-key
HUGGINGFACE_API_KEY=hf_your-huggingface-key
REPLICATE_API_KEY=r8_your-replicate-key

# ============================================
# Core Performance Settings
# ============================================

# Cache Configuration
CACHE_SIZE=100                    # Maximum cache entries
CACHE_TTL_SECONDS=3600           # Default cache TTL (1 hour)
CACHE_CLEANUP_INTERVAL=300000    # Cleanup interval (5 minutes)
ENABLE_SMART_CACHE=true          # Enable LRU cache management

# Rate Limiting
RATE_LIMIT_PER_HOUR=1000         # Requests per hour per user
RATE_LIMIT_PER_DAY=10000         # Requests per day per user
RATE_LIMIT_BURST_SIZE=50         # Burst allowance
ENABLE_RATE_LIMITING=true        # Enable rate limiting

# Performance Tuning
MAX_CONCURRENCY=50               # Maximum concurrent requests
REQUEST_TIMEOUT_MS=30000         # Request timeout (30 seconds)
MAX_MEMORY_MB=10                 # Memory usage limit
ENABLE_PERFORMANCE_PROFILING=true # Enable profiling

# ============================================
# Analytics & Monitoring
# ============================================

# Analytics Configuration
ENABLE_ANALYTICS=true            # Enable analytics collection
ANALYTICS_RETENTION_DAYS=7      # Analytics data retention
ANALYTICS_BATCH_SIZE=100        # Batch size for analytics
ANALYTICS_FLUSH_INTERVAL=60000  # Flush interval (1 minute)

# Health Monitoring
HEALTH_CHECK_INTERVAL=30000      # Health check interval (30 seconds)
HEALTH_CHECK_TIMEOUT=5000       # Health check timeout (5 seconds)
ENABLE_HEALTH_MONITORING=true   # Enable health monitoring

# Logging
LOG_LEVEL=info                  # Log level (debug/info/warn/error)
LOG_FORMAT=json                 # Log format (json/text)
LOG_RETENTION_DAYS=30           # Log retention period
ENABLE_STRUCTURED_LOGGING=true  # Enable structured logging

# ============================================
# Provider Management
# ============================================

# Provider Settings
ENABLE_PROVIDER_LOAD_BALANCING=true  # Enable load balancing
PROVIDER_HEALTH_CHECK_INTERVAL=60000 # Health check interval
PROVIDER_FAILOVER_TIMEOUT=10000     # Failover timeout
ENABLE_COST_OPTIMIZATION=true       # Enable cost optimization

# Fallback Configuration
ENABLE_FALLBACK_PROVIDERS=true      # Enable fallback providers
FALLBACK_RETRY_ATTEMPTS=3          # Number of retry attempts
FALLBACK_RETRY_DELAY=1000          # Delay between retries (ms)

# ============================================
# Security & Privacy
# ============================================

# Security Settings
ENABLE_REQUEST_VALIDATION=true     # Validate incoming requests
ENABLE_RESPONSE_SANITIZATION=true  # Sanitize responses
MAX_REQUEST_SIZE_KB=1024          # Maximum request size
ENABLE_CORS=true                  # Enable CORS

# Privacy Settings
ENABLE_IP_ANONYMIZATION=true      # Anonymize IP addresses
ENABLE_USER_AGENT_MASKING=true    # Mask user agent strings
DATA_RETENTION_DAYS=30            # User data retention

# ============================================
# Advanced Features
# ============================================

# Experimental Features
ENABLE_EXPERIMENTAL_CACHE=true    # Enable experimental cache features
ENABLE_PREDICTIVE_LOAD_BALANCING=false # Enable predictive balancing
ENABLE_AI_OPTIMIZATION=true       # Enable AI-driven optimization

# Debug Settings
ENABLE_DEBUG_MODE=false           # Enable debug mode
DEBUG_REQUEST_TRACING=true        # Enable request tracing
DEBUG_PERFORMANCE_METRICS=true    # Enable performance metrics
```

### 2. Wrangler Configuration (wrangler.toml)

Update your `wrangler.toml` file:

```toml
name = "obsidian-bot-workers"
main = "ai-proxy/src/index.js"
compatibility_date = "2024-01-01"

# ============================================
# Environment Variables
# ============================================
[vars]
CACHE_SIZE = "100"
RATE_LIMIT_PER_HOUR = "1000"
ENABLE_ANALYTICS = "true"
LOG_LEVEL = "info"

# ============================================
# KV Namespaces
# ============================================
[[kv_namespaces]]
binding = "CACHE_KV"
id = "your_cache_kv_id"
preview_id = "your_cache_kv_preview_id"

[[kv_namespaces]]
binding = "ANALYTICS_KV"
id = "your_analytics_kv_id"
preview_id = "your_analytics_kv_preview_id"

# ============================================
# Durable Objects
# ============================================
[[durable_objects.bindings]]
name = "RATE_LIMITER"
class_name = "RateLimiter"

[[durable_objects.bindings]]
name = "CACHE_MANAGER"
class_name = "CacheManager"

# ============================================
# Environment-Specific Overrides
# ============================================
[env.production.vars]
RATE_LIMIT_PER_HOUR = "5000"
ENABLE_DEBUG_MODE = "false"
LOG_LEVEL = "warn"

[env.staging.vars]
RATE_LIMIT_PER_HOUR = "1000"
ENABLE_DEBUG_MODE = "true"
LOG_LEVEL = "debug"

# ============================================
# Build Configuration
# ============================================
[build]
command = "npm run build"
cwd = "."

[build.upload]
format = "service-worker"
```

### 3. Runtime Configuration

Some settings can be updated at runtime via API:

```bash
# Update cache settings
curl -X POST https://your-worker.workers.dev/config/cache \
  -H "Content-Type: application/json" \
  -d '{"size": 200, "ttl": 7200}'

# Update rate limiting
curl -X POST https://your-worker.workers.dev/config/rate-limit \
  -H "Content-Type: application/json" \
  -d '{"perHour": 2000, "burstSize": 100}'

# Update analytics settings
curl -X POST https://your-worker.workers.dev/config/analytics \
  -H "Content-Type: application/json" \
  -d '{"enabled": true, "retentionDays": 14}'
```

## ⚙️ Configuration Categories

### Cache Configuration

| Setting | Description | Default | Recommended |
|---------|-------------|---------|-------------|
| `CACHE_SIZE` | Maximum cache entries | 100 | 100-1000 |
| `CACHE_TTL_SECONDS` | Default TTL | 3600 | 1800-7200 |
| `ENABLE_SMART_CACHE` | Enable LRU management | true | true |
| `CACHE_CLEANUP_INTERVAL` | Cleanup frequency | 300000 | 300000 |

### Rate Limiting Configuration

| Setting | Description | Default | Recommended |
|---------|-------------|---------|-------------|
| `RATE_LIMIT_PER_HOUR` | Hourly limit per user | 1000 | 100-10000 |
| `RATE_LIMIT_PER_DAY` | Daily limit per user | 10000 | 1000-100000 |
| `RATE_LIMIT_BURST_SIZE` | Burst allowance | 50 | 10-200 |
| `RATE_LIMIT_ALGORITHM` | Algorithm to use | token-bucket | token-bucket |

### Analytics Configuration

| Setting | Description | Default | Recommended |
|---------|-------------|---------|-------------|
| `ENABLE_ANALYTICS` | Enable analytics | true | true |
| `ANALYTICS_RETENTION_DAYS` | Data retention | 7 | 7-30 |
| `ANALYTICS_BATCH_SIZE` | Batch size | 100 | 50-200 |
| `ANALYTICS_FLUSH_INTERVAL` | Flush frequency | 60000 | 30000-120000 |

### Performance Configuration

| Setting | Description | Default | Recommended |
|---------|-------------|---------|-------------|
| `MAX_CONCURRENCY` | Max concurrent requests | 50 | 10-100 |
| `REQUEST_TIMEOUT_MS` | Request timeout | 30000 | 10000-60000 |
| `MAX_MEMORY_MB` | Memory limit | 10 | 5-20 |
| `ENABLE_PROFILING` | Enable profiling | true | true |

## 🔧 Advanced Configuration

### Provider-Specific Settings

Configure individual AI providers:

```javascript
// In your worker code
const providerConfig = {
  openai: {
    priority: 1,
    costPerToken: 0.002,
    maxTokens: 4096,
    timeout: 30000
  },
  anthropic: {
    priority: 2,
    costPerToken: 0.008,
    maxTokens: 8192,
    timeout: 45000
  }
};
```

### Custom Routing Rules

Define custom routing logic:

```javascript
const routingRules = [
  {
    condition: (request) => request.userAgent.includes('mobile'),
    provider: 'anthropic', // Use Anthropic for mobile requests
    reason: 'Better mobile performance'
  },
  {
    condition: (request) => request.prompt.length > 1000,
    provider: 'openai',
    reason: 'Better for long prompts'
  }
];
```

### Monitoring Alerts

Configure alert thresholds:

```javascript
const alertConfig = {
  errorRate: {
    threshold: 0.05, // 5% error rate
    cooldown: 300000, // 5 minutes between alerts
    channels: ['slack', 'email']
  },
  responseTime: {
    threshold: 1000, // 1 second
    cooldown: 60000, // 1 minute between alerts
    channels: ['webhook']
  }
};
```

## 📊 Configuration Validation

### Validate Configuration

Run configuration validation:

```bash
npm run validate-config
```

### Check Configuration Status

Get current configuration:

```bash
curl https://your-worker.workers.dev/config/status
```

Response:
```json
{
  "status": "valid",
  "version": "2.0.0",
  "lastUpdated": "2024-01-15T10:30:00Z",
  "activeFeatures": [
    "analytics",
    "smart-cache",
    "rate-limiting",
    "health-monitoring"
  ]
}
```

## 🚨 Configuration Troubleshooting

### Common Issues

#### Configuration Not Loading
```
Symptom: Settings not taking effect
Solution:
1. Check .env file syntax
2. Verify wrangler.toml format
3. Restart local development server
4. Redeploy to production
```

#### Performance Issues
```
Symptom: Slow response times after config change
Solution:
1. Check CACHE_SIZE isn't too large
2. Verify MAX_CONCURRENCY settings
3. Monitor memory usage
4. Adjust REQUEST_TIMEOUT_MS
```

#### Rate Limiting Too Restrictive
```
Symptom: Too many requests blocked
Solution:
1. Increase RATE_LIMIT_PER_HOUR
2. Adjust RATE_LIMIT_BURST_SIZE
3. Change algorithm to sliding-window
4. Monitor actual usage patterns
```

## 🔄 Configuration Updates

### Hot Reload

For supported settings, update without redeployment:

```bash
# Update cache settings (hot reload)
curl -X PATCH https://your-worker.workers.dev/config \
  -H "Content-Type: application/json" \
  -d '{"cache": {"size": 200, "ttl": 7200}}'

# Update rate limits (hot reload)
curl -X PATCH https://your-worker.workers.dev/config \
  -H "Content-Type: application/json" \
  -d '{"rateLimit": {"perHour": 2000}}'
```

### Rolling Updates

For major configuration changes:

```bash
# Create new deployment with updated config
npm run deploy:new-version

# Gradually route traffic to new version
wrangler deployments tail

# Monitor performance during rollout
curl https://your-worker.workers.dev/analytics
```

## 📋 Best Practices

### Security
- Store API keys securely, never in code
- Use environment variables for sensitive data
- Enable request validation and sanitization
- Regularly rotate credentials

### Performance
- Start with conservative cache sizes
- Monitor memory usage closely
- Use appropriate rate limiting for your use case
- Enable analytics to understand usage patterns

### Monitoring
- Set up alerts for critical thresholds
- Monitor error rates and response times
- Keep analytics retention periods reasonable
- Use structured logging for better debugging

### Scalability
- Design for horizontal scaling
- Use appropriate concurrency limits
- Monitor resource usage across deployments
- Plan for traffic spikes with burst allowances

## 🎯 Next Steps

After configuring your workers:

1. [Test the configuration](./testing.md)
2. [Set up monitoring](./monitoring.md)
3. [Deploy to production](../operations/deployment.md)
4. [Tune performance](../operations/performance-tuning.md)