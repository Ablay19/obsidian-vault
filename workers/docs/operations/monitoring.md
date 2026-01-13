# Monitoring Setup

## 📊 Monitoring Overview

This guide covers the complete monitoring setup for enhanced Cloudflare Workers, including metrics collection, alerting, log aggregation, and dashboard creation. Proper monitoring is essential for maintaining performance, reliability, and security.

## 🎯 Monitoring Objectives

### Key Monitoring Goals
- **Performance Monitoring**: Track response times, throughput, and resource usage
- **Error Tracking**: Identify and diagnose errors and exceptions
- **Availability Monitoring**: Ensure services are operational and accessible
- **Business Metrics**: Track user activity, conversion rates, and feature usage
- **Security Monitoring**: Detect unusual patterns and potential threats

### Service Level Objectives (SLOs)

```yaml
performance_slos:
  response_time_p50: "<100ms"
  response_time_p95: "<500ms"
  response_time_p99: "<1000ms"
  error_rate: "<0.1%"
  availability: "99.9%"

reliability_slos:
  uptime: "99.9%"
  data_consistency: "100%"
  recovery_time: "<5min"

business_slos:
  user_satisfaction: ">4.5/5"
  feature_adoption: ">80%"
  cost_efficiency: ">90%"
```

## 🔧 Core Monitoring Components

### 1. Metrics Collection

#### Built-in Worker Metrics

```javascript
// ai-proxy/src/metrics.js

export class WorkerMetrics {
  constructor() {
    this.requestCount = 0;
    this.errorCount = 0;
    this.responseTimes = [];
    this.cacheHits = 0;
    this.cacheMisses = 0;
  }

  recordRequest(duration) {
    this.requestCount++;
    this.responseTimes.push(duration);
    this.responseTimes = this.responseTimes.slice(-1000); // Keep last 1000
  }

  recordError() {
    this.errorCount++;
  }

  recordCacheHit() {
    this.cacheHits++;
  }

  recordCacheMiss() {
    this.cacheMisses++;
  }

  getStats() {
    const sorted = [...this.responseTimes].sort((a, b) => a - b);
    return {
      requestCount: this.requestCount,
      errorCount: this.errorCount,
      errorRate: this.errorCount / this.requestCount,
      responseTimeP50: sorted[Math.floor(sorted.length * 0.5)],
      responseTimeP95: sorted[Math.floor(sorted.length * 0.95)],
      responseTimeP99: sorted[Math.floor(sorted.length * 0.99)],
      cacheHitRate: this.cacheHits / (this.cacheHits + this.cacheMisses)
    };
  }
}
```

#### Custom Metrics

```javascript
// Define custom metrics for specific features
export const customMetrics = {
  aiRequestLatency: new Map(),
  providerResponseTime: new Map(),
  rateLimitViolations: 0,
  costOptimizationSavings: 0,
  healthCheckFailures: 0
};

export function recordAIMetrics(provider, latency, cost) {
  if (!customMetrics.aiRequestLatency.has(provider)) {
    customMetrics.aiRequestLatency.set(provider, []);
  }
  
  customMetrics.aiRequestLatency.get(provider).push({
    timestamp: Date.now(),
    latency,
    cost
  });
}
```

### 2. Log Aggregation

#### Structured Logging

```javascript
// ai-proxy/src/logger.js

export class StructuredLogger {
  constructor(context) {
    this.context = context;
  }

  log(level, message, data = {}) {
    const logEntry = {
      timestamp: new Date().toISOString(),
      level,
      message,
      environment: this.context.env.ENVIRONMENT || 'development',
      requestId: this.context.requestId,
      ...data
    };

    if (level === 'error') {
      console.error(JSON.stringify(logEntry));
    } else if (level === 'warn') {
      console.warn(JSON.stringify(logEntry));
    } else {
      console.log(JSON.stringify(logEntry));
    }

    return logEntry;
  }

  info(message, data) {
    return this.log('info', message, data);
  }

  warn(message, data) {
    return this.log('warn', message, data);
  }

  error(message, data) {
    return this.log('error', message, data);
  }
}
```

#### Log Levels

```javascript
export const LOG_LEVELS = {
  DEBUG: 0,
  INFO: 1,
  WARN: 2,
  ERROR: 3,
  FATAL: 4
};

export function shouldLog(level, configuredLevel) {
  return LOG_LEVELS[level] >= LOG_LEVELS[configuredLevel];
}
```

### 3. Health Checks

#### Health Check Endpoint

```javascript
// ai-proxy/src/health-check.js

export async function healthCheck(context) {
  const checks = {
    timestamp: new Date().toISOString(),
    status: 'healthy',
    checks: {}
  };

  // Check cache health
  try {
    checks.checks.cache = {
      status: 'healthy',
      size: context.cache.size,
      hitRate: context.cache.getHitRate()
    };
  } catch (error) {
    checks.checks.cache = {
      status: 'unhealthy',
      error: error.message
    };
    checks.status = 'degraded';
  }

  // Check provider health
  try {
    const providerHealth = await context.providerManager.healthCheck();
    checks.checks.providers = providerHealth;
    
    if (providerHealth.healthy < providerHealth.total) {
      checks.status = 'degraded';
    }
  } catch (error) {
    checks.checks.providers = {
      status: 'unhealthy',
      error: error.message
    };
    checks.status = 'unhealthy';
  }

  // Check rate limiter
  try {
    checks.checks.rateLimiter = {
      status: 'healthy',
      activeUsers: context.rateLimiter.getUserCount()
    };
  } catch (error) {
    checks.checks.rateLimiter = {
      status: 'unhealthy',
      error: error.message
    };
    checks.status = 'degraded';
  }

  return checks;
}
```

### 4. Alerting

#### Alert Configuration

```javascript
// ai-proxy/src/alerting.js

export class AlertManager {
  constructor(config) {
    this.config = config;
    this.alertHistory = [];
    this.alertThresholds = config.thresholds || {
      errorRate: 0.01, // 1% error rate
      responseTimeP95: 1000, // 1 second
      availability: 99.9 // 99.9% uptime
    };
  }

  checkMetrics(metrics) {
    const alerts = [];

    // Check error rate
    if (metrics.errorRate > this.alertThresholds.errorRate) {
      alerts.push({
        severity: 'critical',
        metric: 'error_rate',
        value: metrics.errorRate,
        threshold: this.alertThresholds.errorRate,
        message: `Error rate exceeded threshold: ${metrics.errorRate} > ${this.alertThresholds.errorRate}`
      });
    }

    // Check response time
    if (metrics.responseTimeP95 > this.alertThresholds.responseTimeP95) {
      alerts.push({
        severity: 'warning',
        metric: 'response_time_p95',
        value: metrics.responseTimeP95,
        threshold: this.alertThresholds.responseTimeP95,
        message: `P95 response time exceeded threshold: ${metrics.responseTimeP95}ms > ${this.alertThresholds.responseTimeP95}ms`
      });
    }

    // Store alerts
    alerts.forEach(alert => {
      this.alertHistory.push({
        ...alert,
        timestamp: new Date().toISOString()
      });
    });

    return alerts;
  }

  async sendAlert(alert) {
    // Send to configured endpoints
    if (this.config.slackWebhook) {
      await this.sendSlackAlert(alert);
    }
    if (this.config.pagerDutyKey) {
      await this.sendPagerDutyAlert(alert);
    }
    if (this.config.emailRecipients) {
      await this.sendEmailAlert(alert);
    }
  }

  async sendSlackAlert(alert) {
    const payload = {
      attachments: [{
        color: alert.severity === 'critical' ? 'danger' : 'warning',
        title: 'Alert Triggered',
        text: alert.message,
        fields: [
          { title: 'Severity', value: alert.severity, short: true },
          { title: 'Metric', value: alert.metric, short: true },
          { title: 'Value', value: alert.value, short: true },
          { title: 'Threshold', value: alert.threshold, short: true }
        ],
        timestamp: new Date().toISOString()
      }]
    };

    await fetch(this.config.slackWebhook, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload)
    });
  }
}
```

## 📈 Dashboard Creation

### Cloudflare Analytics Dashboard

```javascript
// Custom dashboard configuration for Cloudflare
const dashboardConfig = {
  title: 'Workers Performance Dashboard',
  refreshInterval: '30s',
  widgets: [
    {
      type: 'line',
      title: 'Request Rate',
      queries: [
        {
          name: 'Requests',
          query: 'sum(rate(requests_total[5m]))'
        }
      ]
    },
    {
      type: 'line',
      title: 'Response Time',
      queries: [
        {
          name: 'P50',
          query: 'histogram_quantile(0.5, request_duration_seconds)'
        },
        {
          name: 'P95',
          query: 'histogram_quantile(0.95, request_duration_seconds)'
        },
        {
          name: 'P99',
          query: 'histogram_quantile(0.99, request_duration_seconds)'
        }
      ]
    },
    {
      type: 'gauge',
      title: 'Error Rate',
      queries: [
        {
          name: 'Errors',
          query: 'rate(errors_total[5m]) / rate(requests_total[5m])'
        }
      ]
    },
    {
      type: 'stat',
      title: 'Cache Hit Rate',
      queries: [
        {
          name: 'Hit Rate',
          query: 'cache_hits / (cache_hits + cache_misses)'
        }
      ]
    }
  ]
};
```

### Grafana Dashboard

```json
{
  "dashboard": {
    "title": "Workers Performance Dashboard",
    "panels": [
      {
        "title": "Request Rate",
        "targets": [
          {
            "expr": "sum(rate(worker_requests_total[5m]))",
            "legendFormat": "Requests/sec"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Response Time Percentiles",
        "targets": [
          {
            "expr": "histogram_quantile(0.50, worker_request_duration_seconds)",
            "legendFormat": "P50"
          },
          {
            "expr": "histogram_quantile(0.95, worker_request_duration_seconds)",
            "legendFormat": "P95"
          },
          {
            "expr": "histogram_quantile(0.99, worker_request_duration_seconds)",
            "legendFormat": "P99"
          }
        ],
        "type": "graph"
      },
      {
        "title": "Error Rate",
        "targets": [
          {
            "expr": "rate(worker_errors_total[5m]) / rate(worker_requests_total[5m]) * 100",
            "legendFormat": "Error Rate %"
          }
        ],
        "type": "stat"
      }
    ]
  }
}
```

## 🔔 Alerting Configuration

### Critical Alerts

```yaml
critical_alerts:
  - name: "High Error Rate"
    condition: "error_rate > 0.05"
    duration: "5m"
    severity: "critical"
    channels:
      - slack
      - pagerduty
      - email

  - name: "Service Down"
    condition: "availability < 99%"
    duration: "1m"
    severity: "critical"
    channels:
      - pagerduty
      - slack
      - sms

  - name: "Response Time Degraded"
    condition: "p99_response_time > 2000ms"
    duration: "10m"
    severity: "critical"
    channels:
      - slack
      - email
```

### Warning Alerts

```yaml
warning_alerts:
  - name: "Elevated Error Rate"
    condition: "error_rate > 0.01"
    duration: "10m"
    severity: "warning"
    channels:
      - slack

  - name: "Cache Hit Rate Low"
    condition: "cache_hit_rate < 0.7"
    duration: "15m"
    severity: "warning"
    channels:
      - slack

  - name: "High Memory Usage"
    condition: "memory_usage > 80%"
    duration: "5m"
    severity: "warning"
    channels:
      - slack
```

### Info Alerts

```yaml
info_alerts:
  - name: "Deployment Successful"
    condition: "deployment_status == 'success'"
    duration: "0s"
    severity: "info"
    channels:
      - slack

  - name: "Scheduled Maintenance"
    condition: "maintenance_mode == true"
    duration: "0s"
    severity: "info"
    channels:
      - slack
      - email
```

## 🔍 Monitoring Tools Integration

### 1. Cloudflare Analytics

```bash
# Enable Cloudflare Analytics
wrangler analytics enable

# View analytics dashboard
wrangler analytics dashboard

# Export analytics data
wrangler analytics export --start 2024-01-01 --end 2024-01-31 --format csv
```

### 2. Prometheus Integration

```javascript
// Expose metrics in Prometheus format
export async function metricsHandler(context) {
  const metrics = context.metrics.getStats();
  
  return new Response(`
# HELP worker_requests_total Total number of requests
# TYPE worker_requests_total counter
worker_requests_total ${metrics.requestCount}

# HELP worker_errors_total Total number of errors
# TYPE worker_errors_total counter
worker_errors_total ${metrics.errorCount}

# HELP worker_response_time_seconds Response time in seconds
# TYPE worker_response_time_seconds histogram
worker_response_time_seconds{quantile="0.5"} ${metrics.responseTimeP50 / 1000}
worker_response_time_seconds{quantile="0.95"} ${metrics.responseTimeP95 / 1000}
worker_response_time_seconds{quantile="0.99"} ${metrics.responseTimeP99 / 1000}

# HELP worker_cache_hits Total cache hits
# TYPE worker_cache_hits counter
worker_cache_hits ${metrics.cacheHits}

# HELP worker_cache_misses Total cache misses
# TYPE worker_cache_misses counter
worker_cache_misses ${metrics.cacheMisses}
`, {
    headers: { 'Content-Type': 'text/plain' }
  });
}
```

### 3. Log Forwarding

```javascript
// Forward logs to external services
export async function forwardLogs(logEntry, context) {
  const { LOG_FORWARDING_ENDPOINT, LOG_FORWARDING_API_KEY } = context.env;

  if (!LOG_FORWARDING_ENDPOINT) return;

  try {
    await fetch(LOG_FORWARDING_ENDPOINT, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': `Bearer ${LOG_FORWARDING_API_KEY}`
      },
      body: JSON.stringify(logEntry)
    });
  } catch (error) {
    console.error('Failed to forward logs:', error);
  }
}
```

## 📋 Monitoring Checklist

### Initial Setup
- [ ] Metrics collection enabled
- [ ] Structured logging configured
- [ ] Health check endpoint created
- [ ] Alert thresholds defined
- [ ] Dashboard configured

### Integration
- [ ] Cloudflare Analytics enabled
- [ ] External monitoring connected (Prometheus/Grafana)
- [ ] Log forwarding configured
- [ ] Alert notifications set up (Slack, PagerDuty, etc.)
- [ ] On-call rotation configured

### Validation
- [ ] Metrics being collected
- [ ] Logs being forwarded
- [ ] Alerts being triggered correctly
- [ ] Dashboards displaying data
- [ ] Notifications being received

### Maintenance
- [ ] Regular dashboard reviews
- [ ] Alert threshold tuning
- [ ] Log retention policies
- [ ] Monitoring documentation updated
- [ ] Team trained on monitoring tools

## 🚨 Incident Monitoring

### Real-Time Monitoring

```bash
# Monitor live logs
wrangler tail --env production

# Monitor specific endpoints
wrangler tail --env production --format=pretty | grep "POST /api/chat"

# Monitor errors only
wrangler tail --env production | grep "level\":\"error"
```

### Automated Monitoring Scripts

```bash
#!/bin/bash
# monitor-worker.sh

ENDPOINT="https://api.yourdomain.com"
THRESHOLD_MS=1000

while true; do
  START_TIME=$(date +%s%N)
  RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $ENDPOINT/health)
  END_TIME=$(date +%s%N)
  DURATION=$(( ($END_TIME - $START_TIME) / 1000000 ))

  if [ "$RESPONSE" != "200" ]; then
    echo "[$(date)] ERROR: Health check failed with status $RESPONSE"
    # Send alert
  fi

  if [ "$DURATION" -gt "$THRESHOLD_MS" ]; then
    echo "[$(date)] WARNING: Response time ${DURATION}ms exceeded threshold"
    # Send alert
  fi

  sleep 60
done
```

## 🔗 Related Documentation

- [Deployment Guide](./deployment.md)
- [Performance Tuning](./performance-tuning.md)
- [Incident Response](./incident-response.md)
- [Monitoring Dashboard](../user-guides/monitoring.md)
- [Troubleshooting Guide](../user-guides/troubleshooting.md)
