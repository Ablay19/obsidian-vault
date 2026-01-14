# Enhanced Workers Configuration Guide

This guide covers the configuration options for the Enhanced Cloudflare Workers AI Proxy, including environment variables, caching settings, analytics configuration, rate limiting options, and performance tuning parameters.

## 🎯 Configuration Overview

The workers support comprehensive configuration through environment variables and `wrangler.toml` settings. Configurations are divided into core functionality, performance tuning, and monitoring categories.

---

## 🔧 Core Configuration

### Environment Variables

#### Cloudflare Workers
```bash
# Required for deployment and API access
CLOUDFLARE_API_TOKEN=your_cloudflare_api_token
CLOUDFLARE_ACCOUNT_ID=your_account_id
WORKER_NAME=your-worker-name

# Optional: Custom domain configuration
CUSTOM_DOMAIN=your.custom.domain
WORKER_URL=https://your-worker.yourname.workers.dev
```

#### AI Provider Configuration
```bash
# Primary AI providers (at least one required)
GEMINI_API_KEY=your_gemini_api_key
DEEPSEEK_API_KEY=your_deepseek_api_key
GROQ_API_KEY=your_groq_api_key

# Secondary providers (optional)
OPENAI_API_KEY=your_openai_api_key
HUGGINGFACE_API_KEY=your_huggingface_api_key
OPENROUTER_API_KEY=your_openrouter_api_key

# Provider preferences
DEFAULT_PROVIDER=gemini
FALLBACK_PROVIDERS=deepseek,groq,openai
```

#### Application Settings
```bash
# Environment and logging
ENVIRONMENT=production
LOG_LEVEL=info
ENABLE_DEBUG_MODE=false

# CORS and security
ALLOWED_ORIGINS=https://yourdomain.com,https://anotherdomain.com
ENABLE_CORS=true
API_KEY_AUTH_ENABLED=false
```

---

## ⚡ Performance Configuration

### Caching Settings
```toml
# wrangler.toml
[vars]
CACHE_ENABLED = "true"
CACHE_SIZE = "100"              # Maximum cached responses
CACHE_TTL_SECONDS = "3600"      # 1 hour default TTL
CACHE_CLEANUP_INTERVAL = "300"  # Cleanup every 5 minutes

# Advanced caching options
CACHE_COMPRESSION = "true"
CACHE_ENCRYPTION = "false"
CACHE_METRICS_ENABLED = "true"
```

### Rate Limiting Configuration
```toml
[vars]
RATE_LIMIT_ENABLED = "true"
RATE_LIMIT_PER_MINUTE = "60"    # Requests per minute per IP
RATE_LIMIT_PER_HOUR = "1000"    # Requests per hour per IP
RATE_LIMIT_BURST = "10"         # Burst allowance

# Rate limiting algorithms
RATE_LIMIT_ALGORITHM = "token_bucket"  # token_bucket, sliding_window, fixed_window
RATE_LIMIT_WINDOW_SIZE = "60"   # Window size in seconds
```

### Load Balancing & Routing
```toml
[vars]
LOAD_BALANCING_ENABLED = "true"
PROVIDER_TIMEOUT_MS = "30000"   # 30 second timeout
RETRY_ATTEMPTS = "3"
RETRY_BACKOFF_MS = "1000"

# Cost optimization
COST_OPTIMIZATION_ENABLED = "true"
MAX_COST_PER_REQUEST = "0.01"   # Maximum cost in USD
PREFERRED_PROVIDERS = "gemini,deepseek,groq"
```

---

## 📊 Analytics & Monitoring

### Analytics Configuration
```toml
[vars]
ANALYTICS_ENABLED = "true"
ANALYTICS_RETENTION_DAYS = "30"
ANALYTICS_SAMPLING_RATE = "1.0"  # 100% sampling

# Metrics collection
COLLECT_RESPONSE_TIME = "true"
COLLECT_ERROR_RATES = "true"
COLLECT_CACHE_STATS = "true"
COLLECT_PROVIDER_STATS = "true"
```

### Monitoring Settings
```toml
[vars]
HEALTH_CHECK_ENABLED = "true"
METRICS_ENDPOINT_ENABLED = "true"
PERFORMANCE_PROFILING = "true"

# Alert thresholds
ALERT_ERROR_RATE_THRESHOLD = "0.05"    # 5% error rate
ALERT_RESPONSE_TIME_THRESHOLD = "5000" # 5 seconds
ALERT_CACHE_MISS_THRESHOLD = "0.8"     # 80% miss rate
```

### Logging Configuration
```toml
[vars]
LOG_REQUESTS = "true"
LOG_RESPONSES = "false"        # Don't log full responses for privacy
LOG_ERRORS = "true"
LOG_PERFORMANCE = "true"

# Log levels per component
LOG_LEVEL_CACHE = "info"
LOG_LEVEL_ANALYTICS = "debug"
LOG_LEVEL_ROUTING = "info"
```

---

## 🔄 Provider-Specific Configuration

### Gemini Configuration
```toml
[vars]
GEMINI_MODEL = "gemini-pro"
GEMINI_MAX_TOKENS = "4096"
GEMINI_TEMPERATURE = "0.7"
GEMINI_TIMEOUT_MS = "25000"
```

### DeepSeek Configuration
```toml
[vars]
DEEPSEEK_MODEL = "deepseek-chat"
DEEPSEEK_MAX_TOKENS = "4096"
DEEPSEEK_TEMPERATURE = "0.7"
DEEPSEEK_TIMEOUT_MS = "20000"
```

### Groq Configuration
```toml
[vars]
GROQ_MODEL = "mixtral-8x7b-32768"
GROQ_MAX_TOKENS = "4096"
GROQ_TEMPERATURE = "0.7"
GROQ_TIMEOUT_MS = "15000"
```

---

## 🚀 Performance Tuning

### Memory & Resource Management
```toml
[vars]
MAX_MEMORY_MB = "128"
CPU_TIME_LIMIT_MS = "5000"
SUBREQUEST_LIMIT = "50"

# Optimization flags
ENABLE_STREAMING = "true"
ENABLE_COMPRESSION = "true"
ENABLE_CONNECTION_POOLING = "true"
```

### Advanced Tuning
```toml
[vars]
# Request processing
MAX_REQUEST_SIZE_KB = "1024"
REQUEST_TIMEOUT_MS = "30000"
CONNECTION_TIMEOUT_MS = "10000"

# Cache optimization
CACHE_WARMUP_ENABLED = "true"
CACHE_PREFETCH_ENABLED = "false"
CACHE_INVALIDATION_STRATEGY = "ttl"  # ttl, lru, manual

# Performance monitoring
PROFILING_ENABLED = "true"
PROFILING_SAMPLE_RATE = "0.1"  # 10% sampling
MEMORY_PROFILING = "true"
```

---

## 🔒 Security Configuration

### Authentication & Authorization
```toml
[vars]
API_KEY_REQUIRED = "false"
API_KEY_HEADER = "X-API-Key"
JWT_ENABLED = "false"

# Rate limiting per API key
API_KEY_RATE_LIMIT_ENABLED = "true"
API_KEY_RATE_LIMIT_PER_HOUR = "10000"
```

### Data Protection
```toml
[vars]
ENCRYPT_CACHED_DATA = "false"
LOG_SENSITIVE_DATA = "false"
REDACT_REQUESTS_IN_LOGS = "true"

# Compliance settings
GDPR_COMPLIANT = "true"
DATA_RETENTION_DAYS = "30"
AUDIT_LOGGING = "true"
```

---

## 📋 Configuration Validation

The workers automatically validate configuration on startup. Invalid configurations will prevent deployment and log detailed error messages.

### Required Configurations
- At least one AI provider API key
- Cloudflare API token for deployment
- Basic rate limiting settings

### Recommended Configurations
- Analytics and monitoring enabled
- Caching configured for performance
- Multiple fallback providers
- Reasonable rate limits

---

## 🔄 Runtime Configuration Updates

Some configurations can be updated at runtime without redeployment:

### Hot-Reloadable Settings
- Rate limiting thresholds
- Cache TTL settings
- Provider preferences
- Log levels

### Settings Requiring Redeployment
- Core provider API keys
- Cloudflare account settings
- Worker routing rules
- Security policies

---

## 📊 Configuration Examples

### Development Configuration
```toml
[vars]
ENVIRONMENT = "development"
LOG_LEVEL = "debug"
CACHE_ENABLED = "false"
ANALYTICS_ENABLED = "false"
RATE_LIMIT_PER_MINUTE = "1000"
```

### Production Configuration
```toml
[vars]
ENVIRONMENT = "production"
LOG_LEVEL = "info"
CACHE_ENABLED = "true"
ANALYTICS_ENABLED = "true"
RATE_LIMIT_PER_MINUTE = "60"
RATE_LIMIT_PER_HOUR = "1000"
```

### High-Traffic Configuration
```toml
[vars]
CACHE_SIZE = "500"
RATE_LIMIT_PER_MINUTE = "120"
RATE_LIMIT_BURST = "20"
LOAD_BALANCING_ENABLED = "true"
COST_OPTIMIZATION_ENABLED = "true"
ANALYTICS_SAMPLING_RATE = "0.5"
```

---

**For detailed API documentation and advanced configuration options, see the [Developer Documentation](./workers/docs/developer-docs/).**

## Active Technologies
- Go 1.21+ (backend), JavaScript/Node.js (workers) + Go modules, npm/yarn packages, REST APIs (005-architecture-separation)

## Recent Changes
- 005-architecture-separation: Added Go 1.21+ (backend), JavaScript/Node.js (workers) + Go modules, npm/yarn packages, REST APIs
