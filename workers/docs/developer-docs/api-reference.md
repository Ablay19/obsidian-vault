# API Reference

## 🌐 HTTP API Endpoints

The enhanced workers expose a comprehensive HTTP API for AI processing, monitoring, and management operations.

## 🤖 AI Processing Endpoints

### POST /

Generate AI responses using configured providers.

**Request:**
```http
POST /
Content-Type: application/json

{
  "prompt": "Explain quantum computing in simple terms",
  "model": "gpt-3.5-turbo",
  "maxTokens": 500,
  "temperature": 0.7,
  "provider": "openai"
}
```

**Parameters:**
- `prompt` (string, required): The input text to process
- `model` (string, optional): AI model to use (defaults to provider default)
- `maxTokens` (number, optional): Maximum tokens in response (default: 1000)
- `temperature` (number, optional): Creativity level 0.0-1.0 (default: 0.7)
- `provider` (string, optional): Preferred AI provider

**Response:**
```json
{
  "result": "Quantum computing uses quantum mechanics principles...",
  "provider": "openai",
  "model": "gpt-3.5-turbo",
  "tokens": {
    "input": 8,
    "output": 156,
    "total": 164
  },
  "processingTime": 1250,
  "cache": {
    "hit": false,
    "ttl": 3600
  }
}
```

**Error Response:**
```json
{
  "error": {
    "type": "rate_limit_exceeded",
    "message": "Too many requests. Please try again later.",
    "retryAfter": 60
  },
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### GET /models

List available AI models and providers.

**Response:**
```json
{
  "providers": {
    "openai": {
      "models": ["gpt-3.5-turbo", "gpt-4"],
      "status": "healthy",
      "latency": 850
    },
    "anthropic": {
      "models": ["claude-2", "claude-instant"],
      "status": "healthy",
      "latency": 720
    }
  },
  "defaultProvider": "openai"
}
```

## 📊 Analytics Endpoints

### GET /analytics

Get comprehensive analytics data.

**Query Parameters:**
- `range` (string): Time range (1h, 24h, 7d, 30d)
- `metrics` (string): Comma-separated metrics to include

**Response:**
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "period": "1h",
  "metrics": {
    "totalRequests": 1250,
    "successfulRequests": 1187,
    "failedRequests": 63,
    "errorRate": "5.04%",
    "averageResponseTime": 45.2,
    "medianResponseTime": 32.0,
    "p95ResponseTime": 120.5,
    "p99ResponseTime": 250.8
  },
  "cache": {
    "hitRate": "87.3%",
    "totalHits": 1092,
    "totalMisses": 158,
    "evictions": 23
  },
  "providers": {
    "openai": {
      "requests": 650,
      "successRate": "94.5%",
      "averageLatency": 1250
    }
  }
}
```

### GET /analytics/trends

Get performance trends over time.

**Response:**
```json
{
  "trends": {
    "responseTime": {
      "direction": "improving",
      "changePercent": -12.5,
      "period": "24h",
      "data": [
        { "timestamp": "2024-01-14T10:00:00Z", "value": 52.3 },
        { "timestamp": "2024-01-14T11:00:00Z", "value": 48.7 }
      ]
    },
    "errorRate": {
      "direction": "stable",
      "changePercent": 0.8,
      "period": "24h"
    }
  }
}
```

## 🏥 Health & Monitoring Endpoints

### GET /health

Get system health status.

**Response:**
```json
{
  "status": "healthy",
  "version": "2.0.0",
  "uptime": "2h 30m 15s",
  "lastHealthCheck": "2024-01-15T10:29:45Z",
  "resources": {
    "memory": {
      "current": "8.5MB",
      "peak": "12.3MB",
      "limit": "128MB",
      "utilization": "6.6%"
    },
    "cpu": {
      "average": "15.2%",
      "peak": "45.8%"
    }
  },
  "dependencies": {
    "cache": {
      "status": "healthy",
      "latency": 2.3
    },
    "rateLimiter": {
      "status": "healthy",
      "latency": 1.8
    }
  }
}
```

### GET /health/detailed

Get detailed health information including component status.

**Response:**
```json
{
  "overall": "healthy",
  "components": {
    "requestHandler": {
      "status": "healthy",
      "uptime": "2h 30m",
      "errors": 0,
      "lastError": null
    },
    "cache": {
      "status": "healthy",
      "size": 85,
      "hitRate": "87.3%",
      "evictions": 23
    },
    "rateLimiter": {
      "status": "healthy",
      "activeBuckets": 12,
      "blockedRequests": 5
    },
    "providers": {
      "openai": {
        "status": "healthy",
        "latency": 850,
        "successRate": "94.5%"
      }
    }
  },
  "system": {
    "workerInstances": 3,
    "memoryPressure": "low",
    "cpuPressure": "normal"
  }
}
```

## 💾 Cache Management Endpoints

### GET /cache/status

Get cache performance statistics.

**Response:**
```json
{
  "size": 85,
  "maxSize": 100,
  "hitRate": "87.3%",
  "totalHits": 1092,
  "totalMisses": 158,
  "evictions": 23,
  "memoryUsage": "6.8MB",
  "ttlDistribution": {
    "0-60s": 120,
    "60-300s": 245,
    "300-3600s": 890,
    "3600s+": 156
  },
  "efficiency": "87.3%"
}
```

### POST /cache/clear

Clear all cache entries.

**Request:**
```json
{
  "pattern": "*",  // Optional: clear specific patterns
  "confirm": true  // Required: confirmation flag
}
```

**Response:**
```json
{
  "cleared": 85,
  "pattern": "*",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### GET /cache/keys

List cache keys (admin/debug endpoint).

**Query Parameters:**
- `limit` (number): Maximum keys to return (default: 100)
- `pattern` (string): Key pattern to match

**Response:**
```json
{
  "keys": [
    {
      "key": "ai:response:abc123",
      "ttl": 2850,
      "size": 2048,
      "hits": 15
    }
  ],
  "total": 85,
  "limit": 100
}
```

## 🛡️ Rate Limiting Endpoints

### GET /rate-limits

Get current rate limiting status.

**Response:**
```json
{
  "global": {
    "requestsPerHour": 1000,
    "requestsPerDay": 10000,
    "currentHour": 450,
    "currentDay": 2100,
    "blockedThisHour": 5,
    "blockedToday": 12
  },
  "userLimits": {
    "activeBuckets": 12,
    "totalBlocked": 17,
    "resetTimes": {
      "nextHour": "2024-01-15T11:00:00Z",
      "nextDay": "2024-01-16T00:00:00Z"
    }
  }
}
```

### POST /rate-limits/reset

Reset rate limits for a specific identifier.

**Request:**
```json
{
  "identifier": "user_123",
  "type": "hour",  // "hour", "day", or "all"
  "confirm": true
}
```

**Response:**
```json
{
  "reset": true,
  "identifier": "user_123",
  "type": "hour",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

## ⚙️ Configuration Endpoints

### GET /config

Get current configuration.

**Response:**
```json
{
  "cache": {
    "size": 100,
    "ttl": 3600,
    "enabled": true
  },
  "rateLimit": {
    "perHour": 1000,
    "burstSize": 50,
    "algorithm": "token-bucket"
  },
  "analytics": {
    "enabled": true,
    "retentionDays": 7
  },
  "providers": {
    "openai": {
      "enabled": true,
      "priority": 1
    }
  }
}
```

### PATCH /config

Update configuration settings.

**Request:**
```json
{
  "cache": {
    "size": 200,
    "ttl": 7200
  },
  "rateLimit": {
    "perHour": 2000
  }
}
```

**Response:**
```json
{
  "updated": ["cache.size", "cache.ttl", "rateLimit.perHour"],
  "restartRequired": false,
  "effectiveAt": "2024-01-15T10:30:00Z"
}
```

## 🔍 Debugging Endpoints

### POST /debug/trace/start

Start request tracing.

**Response:**
```json
{
  "tracing": "enabled",
  "traceId": "trace_abc123",
  "duration": 300,  // seconds
  "maxRequests": 1000
}
```

### GET /debug/trace/{traceId}

Get trace results.

**Response:**
```json
{
  "traceId": "trace_abc123",
  "duration": 300,
  "requests": 45,
  "traces": [
    {
      "requestId": "req_001",
      "method": "POST",
      "url": "/api/chat",
      "startTime": "2024-01-15T10:30:00.000Z",
      "duration": 1250,
      "status": 200,
      "cacheHit": false,
      "provider": "openai",
      "steps": [
        { "step": "parse", "duration": 5 },
        { "step": "rate-limit", "duration": 2 },
        { "step": "cache-lookup", "duration": 8 },
        { "step": "ai-call", "duration": 1200 },
        { "step": "response-format", "duration": 35 }
      ]
    }
  ]
}
```

### POST /debug/profile/start

Start performance profiling.

**Response:**
```json
{
  "profiling": "enabled",
  "profileId": "profile_xyz789",
  "maxDuration": 300,
  "sampleRate": 100  // milliseconds
}
```

### GET /debug/profile/{profileId}

Get profiling results.

**Response:**
```json
{
  "profileId": "profile_xyz789",
  "duration": 300,
  "totalRequests": 1250,
  "profiles": [
    {
      "name": "request-processing",
      "count": 1250,
      "totalDuration": 56250,
      "avgDuration": 45,
      "minDuration": 12,
      "maxDuration": 2500,
      "memoryDelta": 2048000,
      "checkpoints": [
        { "name": "parse", "avgDuration": 3.2 },
        { "name": "cache", "avgDuration": 8.7 },
        { "name": "ai-call", "avgDuration": 28.5 }
      ]
    }
  ],
  "system": {
    "avgMemoryUsage": "8.5MB",
    "peakMemoryUsage": "12.3MB",
    "avgCpuUsage": "15.2%"
  }
}
```

## 🚨 Alert Management Endpoints

### GET /alerts

List active alerts.

**Response:**
```json
{
  "alerts": [
    {
      "id": "alert_001",
      "name": "high-error-rate",
      "condition": "error_rate > 0.05",
      "status": "active",
      "lastTriggered": "2024-01-15T10:25:00Z",
      "channels": ["slack", "email"],
      "cooldown": 300
    }
  ]
}
```

### POST /alerts

Create a new alert.

**Request:**
```json
{
  "name": "response-time-spike",
  "condition": "response_time_p95 > 1000",
  "channels": ["slack", "webhook"],
  "cooldown": 300,
  "description": "Alert when 95th percentile response time exceeds 1 second"
}
```

### DELETE /alerts/{alertId}

Delete an alert.

**Response:**
```json
{
  "deleted": true,
  "alertId": "alert_001"
}
```

## 📋 Metrics Export Endpoints

### GET /metrics

Export metrics in Prometheus format.

**Response:**
```
# HELP worker_requests_total Total number of requests processed
# TYPE worker_requests_total counter
worker_requests_total{status="success"} 1187
worker_requests_total{status="error"} 63

# HELP worker_response_time Response time in milliseconds
# TYPE worker_response_time histogram
worker_response_time_bucket{le="10"} 234
worker_response_time_bucket{le="50"} 892
worker_response_time_bucket{le="100"} 1056
worker_response_time_bucket{le="500"} 1187
worker_response_time_bucket{le="+Inf"} 1187
```

### GET /metrics/json

Export metrics in JSON format.

**Response:**
```json
{
  "counters": {
    "requests_total": {
      "value": 1250,
      "labels": {}
    }
  },
  "gauges": {
    "memory_usage_mb": {
      "value": 8.5
    },
    "cache_hit_rate": {
      "value": 87.3
    }
  },
  "histograms": {
    "response_time_ms": {
      "buckets": {
        "10": 234,
        "50": 892,
        "100": 1056,
        "500": 1187
      }
    }
  }
}
```

## 🔐 Authentication

### API Key Authentication

Include API key in request headers:

```http
Authorization: Bearer your-api-key
X-API-Key: your-api-key
```

### Rate Limiting Headers

Response includes rate limiting information:

```http
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 850
X-RateLimit-Reset: 1642156800
X-RateLimit-Retry-After: 3600
```

## 📊 Status Codes

| Code | Meaning | Description |
|------|---------|-------------|
| 200 | OK | Successful request |
| 201 | Created | Resource created |
| 400 | Bad Request | Invalid request parameters |
| 401 | Unauthorized | Authentication required |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource not found |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Server Error | Server error |
| 502 | Bad Gateway | Upstream service error |
| 503 | Service Unavailable | Service temporarily unavailable |

## 🔄 WebSocket API

### Real-time Monitoring

Connect to WebSocket for real-time updates:

```javascript
const ws = new WebSocket('wss://your-worker.workers.dev/metrics/stream');

// Listen for metrics updates
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Real-time metrics:', data);
};

// Subscribe to specific metrics
ws.send(JSON.stringify({
  type: 'subscribe',
  metrics: ['response_time', 'error_rate', 'cache_hits']
}));
```

### Streaming Responses

For long-running AI requests:

```javascript
const response = await fetch('/api/stream', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ prompt: 'Long story...' })
});

const reader = response.body.getReader();
while (true) {
  const { done, value } = await reader.read();
  if (done) break;

  const chunk = new TextDecoder().decode(value);
  console.log('Stream chunk:', chunk);
}
```

## 🎯 Best Practices

### Request Optimization
- Use appropriate `maxTokens` to control costs
- Cache frequently used prompts
- Implement exponential backoff for retries

### Monitoring Integration
- Set up alerts for critical metrics
- Monitor trends over time
- Use `/health` endpoint for load balancer health checks

### Error Handling
- Check `X-RateLimit-Retry-After` header for rate limits
- Implement proper retry logic for transient errors
- Monitor `error.type` in error responses

This API provides comprehensive control and monitoring capabilities for the enhanced workers system.