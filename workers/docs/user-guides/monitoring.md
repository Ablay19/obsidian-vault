# Monitoring Dashboard Guide

## 📊 Overview

The enhanced workers provide comprehensive monitoring capabilities through built-in dashboards and APIs. This guide covers all monitoring features, how to access them, and how to interpret the data.

## 🎯 Available Dashboards

### 1. Main Analytics Dashboard

**URL**: `https://your-worker.workers.dev/analytics`

Provides real-time insights into:
- Request volume and patterns
- Error rates and types
- Response time distributions
- Cache performance metrics

### 2. Health Check Dashboard

**URL**: `https://your-worker.workers.dev/health`

Monitors system health:
- Service availability
- Memory usage
- CPU utilization
- Uptime statistics

### 3. Cache Performance Dashboard

**URL**: `https://your-worker.workers.dev/cache/status`

Shows cache effectiveness:
- Hit/miss ratios
- Memory usage
- Eviction statistics
- TTL distributions

### 4. Rate Limiting Dashboard

**URL**: `https://your-worker.workers.dev/rate-limits`

Displays rate limiting status:
- Active rate limits
- Blocked requests
- Reset timers
- Burst utilization

## 📈 Analytics Dashboard

### Real-Time Metrics

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
  }
}
```

### Cache Performance

```json
{
  "cache": {
    "hitRate": "87.3%",
    "totalHits": 1092,
    "totalMisses": 158,
    "memoryUsage": "6.8MB",
    "maxSize": "10MB",
    "evictions": 23,
    "ttlDistribution": {
      "0-60s": 120,
      "60-300s": 245,
      "300-3600s": 890,
      "3600s+": 156
    }
  }
}
```

### Provider Statistics

```json
{
  "providers": {
    "openai": {
      "requests": 650,
      "successRate": "94.5%",
      "averageLatency": 1250,
      "cost": 2.34,
      "errors": 35
    },
    "anthropic": {
      "requests": 600,
      "successRate": "96.2%",
      "averageLatency": 980,
      "cost": 4.12,
      "errors": 23
    }
  }
}
```

## 🔍 Health Check Dashboard

### System Health

```json
{
  "status": "healthy",
  "version": "2.0.0",
  "uptime": "2h 30m 15s",
  "lastHealthCheck": "2024-01-15T10:29:45Z",
  "nextHealthCheck": "2024-01-15T10:30:15Z"
}
```

### Resource Usage

```json
{
  "resources": {
    "memory": {
      "current": "8.5MB",
      "peak": "12.3MB",
      "limit": "128MB",
      "utilization": "6.6%"
    },
    "cpu": {
      "average": "15.2%",
      "peak": "45.8%",
      "limit": "100%"
    }
  }
}
```

### Service Dependencies

```json
{
  "dependencies": {
    "cache": {
      "status": "healthy",
      "latency": 2.3,
      "lastCheck": "2024-01-15T10:29:45Z"
    },
    "rateLimiter": {
      "status": "healthy",
      "latency": 1.8,
      "lastCheck": "2024-01-15T10:29:47Z"
    },
    "analytics": {
      "status": "healthy",
      "latency": 3.1,
      "lastCheck": "2024-01-15T10:29:46Z"
    }
  }
}
```

## 📊 Interpreting Metrics

### Response Time Distribution

| Percentile | Meaning | Target |
|------------|---------|--------|
| p50 (median) | Typical user experience | <50ms |
| p95 | Worst experience for 5% of users | <200ms |
| p99 | Outlier performance | <500ms |

### Error Rate Categories

| Rate | Severity | Action Required |
|------|----------|-----------------|
| <1% | Excellent | Monitor |
| 1-5% | Good | Investigate trends |
| 5-10% | Concerning | Immediate review |
| >10% | Critical | Immediate action |

### Cache Hit Rate Guidelines

| Hit Rate | Performance | Action |
|----------|-------------|--------|
| >90% | Excellent | Optimal |
| 80-90% | Good | Monitor |
| 70-80% | Fair | Tune cache |
| <70% | Poor | Review strategy |

## 🚨 Alert Configuration

### Setting Up Alerts

Configure alerts via API:

```bash
# Set error rate alert
curl -X POST https://your-worker.workers.dev/alerts \
  -H "Content-Type: application/json" \
  -d '{
    "type": "error_rate",
    "threshold": 0.05,
    "window": "5m",
    "channels": ["slack", "email"],
    "cooldown": "10m"
  }'

# Set response time alert
curl -X POST https://your-worker.workers.dev/alerts \
  -H "Content-Type: application/json" \
  -d '{
    "type": "response_time",
    "threshold": 1000,
    "percentile": "p95",
    "channels": ["webhook"],
    "cooldown": "5m"
  }'
```

### Alert Types

| Type | Description | Example Threshold |
|------|-------------|-------------------|
| `error_rate` | Request error percentage | 5% over 5 minutes |
| `response_time` | Response time exceeded | 1000ms p95 |
| `cache_hit_rate` | Cache effectiveness dropped | Below 70% |
| `memory_usage` | Memory usage high | Above 80% |
| `rate_limit_hits` | Rate limiting triggered | 10 hits per minute |

## 📋 Custom Dashboards

### Creating Custom Views

Build custom dashboards using the metrics API:

```javascript
// Fetch metrics for custom dashboard
async function getDashboardData() {
  const [analytics, health, cache] = await Promise.all([
    fetch('/analytics').then(r => r.json()),
    fetch('/health').then(r => r.json()),
    fetch('/cache/status').then(r => r.json())
  ]);

  return {
    timestamp: new Date().toISOString(),
    summary: {
      totalRequests: analytics.metrics.totalRequests,
      errorRate: analytics.metrics.errorRate,
      avgResponseTime: analytics.metrics.averageResponseTime,
      cacheHitRate: cache.hitRate,
      memoryUsage: health.resources.memory.utilization,
      uptime: health.uptime
    },
    performance: analytics.metrics,
    health: health,
    cache: cache
  };
}
```

### Real-Time Updates

Set up real-time monitoring:

```javascript
// WebSocket connection for real-time updates
const ws = new WebSocket('wss://your-worker.workers.dev/metrics/stream');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  updateDashboard(data);
};

function updateDashboard(data) {
  // Update charts and metrics
  updateCharts(data);
  updateAlerts(data);
  updateHealthIndicators(data);
}
```

## 📈 Historical Analysis

### Time Range Queries

```bash
# Last hour
curl "https://your-worker.workers.dev/analytics?range=1h"

# Last 24 hours
curl "https://your-worker.workers.dev/analytics?range=24h"

# Custom range
curl "https://your-worker.workers.dev/analytics?start=2024-01-14T00:00:00Z&end=2024-01-15T00:00:00Z"
```

### Trend Analysis

```json
{
  "trends": {
    "responseTime": {
      "direction": "improving",
      "changePercent": -12.5,
      "period": "24h"
    },
    "errorRate": {
      "direction": "stable",
      "changePercent": 0.8,
      "period": "24h"
    },
    "cacheHitRate": {
      "direction": "improving",
      "changePercent": 5.2,
      "period": "24h"
    }
  }
}
```

## 🔧 Monitoring Tools

### Built-in Tools

```bash
# Real-time log tailing
wrangler tail

# Performance profiling
curl https://your-worker.workers.dev/profile/start
curl https://your-worker.workers.dev/profile/stop

# Export metrics
curl https://your-worker.workers.dev/metrics/export?format=json
```

### External Monitoring

Integrate with external monitoring:

```javascript
// Prometheus metrics export
app.get('/metrics', (req, res) => {
  const metrics = generatePrometheusMetrics();
  res.set('Content-Type', 'text/plain');
  res.send(metrics);
});

// DataDog integration
const datadog = require('datadog-metrics');
datadog.init({ apiKey: process.env.DD_API_KEY });

function sendMetricsToDataDog(metrics) {
  datadog.gauge('worker.response_time', metrics.averageResponseTime);
  datadog.gauge('worker.error_rate', parseFloat(metrics.errorRate));
  datadog.gauge('worker.cache_hit_rate', parseFloat(metrics.cache.hitRate));
}
```

## 🚨 Incident Response

### Alert Response Workflow

1. **Receive Alert**
   - Check alert type and severity
   - Review current metrics
   - Assess impact

2. **Investigate**
   ```bash
   # Check recent logs
   wrangler tail --since 5m

   # Get detailed metrics
   curl https://your-worker.workers.dev/analytics?range=15m

   # Check health status
   curl https://your-worker.workers.dev/health
   ```

3. **Diagnose**
   - Identify root cause
   - Check provider status
   - Review recent deployments
   - Analyze traffic patterns

4. **Resolve**
   - Apply immediate fixes
   - Scale resources if needed
   - Update configurations
   - Communicate with stakeholders

### Common Issues

#### High Error Rates
```
Check:
- Provider API status
- Rate limiting
- Request validation
- Network connectivity

Actions:
- Enable fallback providers
- Adjust rate limits
- Update error handling
```

#### Slow Response Times
```
Check:
- Cache hit rate
- Provider latency
- Memory usage
- Concurrent requests

Actions:
- Tune cache settings
- Adjust concurrency
- Scale worker instances
```

#### Memory Issues
```
Check:
- Cache size
- Request processing
- Memory leaks
- Worker limits

Actions:
- Reduce cache size
- Optimize memory usage
- Restart workers
- Update memory limits
```

## 📊 Performance Optimization

### Using Monitoring Data

```javascript
// Automated optimization based on metrics
async function optimizeBasedOnMetrics() {
  const metrics = await fetch('/analytics').then(r => r.json());

  // Adjust cache size based on hit rate
  if (metrics.cache.hitRate < 70) {
    await fetch('/config/cache', {
      method: 'PATCH',
      body: JSON.stringify({ size: metrics.cache.size * 1.5 })
    });
  }

  // Scale rate limits based on usage
  if (metrics.rateLimit.blockedRequests > metrics.totalRequests * 0.1) {
    await fetch('/config/rate-limit', {
      method: 'PATCH',
      body: JSON.stringify({ perHour: metrics.rateLimit.perHour * 1.2 })
    });
  }
}
```

## 🎯 Best Practices

### Dashboard Usage
- Set up automated alerts for critical metrics
- Review dashboards daily for trends
- Use historical data for capacity planning
- Monitor both averages and percentiles

### Alert Management
- Configure alerts for business-critical metrics
- Set appropriate cooldown periods
- Use multiple notification channels
- Regularly review and adjust thresholds

### Performance Monitoring
- Track key performance indicators
- Monitor trends over time
- Compare performance across deployments
- Use monitoring data for optimization decisions

### Incident Prevention
- Monitor leading indicators
- Set up automated remediation
- Regular performance reviews
- Proactive capacity planning

## 📞 Support

For monitoring issues:
- Check the [troubleshooting guide](../user-guides/troubleshooting.md)
- Review [operations guide](../operations/monitoring.md)
- Contact the development team

Happy monitoring! 📊🚀