# Troubleshooting Guide

## 🔍 Quick Diagnosis

Use this guide to quickly identify and resolve common issues with the enhanced workers.

## 🚀 Getting Started Checks

### 1. Basic Health Check

```bash
# Check if worker is responding
curl https://your-worker.workers.dev/health
```

Expected response:
```json
{
  "status": "healthy",
  "version": "2.0.0",
  "uptime": "2h 30m",
  "memoryUsage": "8.5MB"
}
```

### 2. Test Basic Functionality

```bash
# Test a simple request
curl -X POST https://your-worker.workers.dev/ \
  -H "Content-Type: application/json" \
  -d '{"prompt": "Hello", "maxTokens": 10}'
```

### 3. Check Logs

```bash
# View recent logs
wrangler tail --since 5m
```

## 🔧 Common Issues & Solutions

### Issue: Worker Not Responding

**Symptoms:**
- 500/502/504 errors
- Connection timeouts
- Worker appears down

**Solutions:**

1. **Check Worker Status**
   ```bash
   wrangler deployments list
   ```

2. **Restart Worker**
   ```bash
   wrangler deployments rollback <deployment-id>
   ```

3. **Check Resource Limits**
   ```bash
   curl https://your-worker.workers.dev/health
   # Look for memory usage > 80%
   ```

4. **Verify Configuration**
   ```bash
   wrangler tail | grep "ERROR\|WARN"
   ```

### Issue: High Error Rates

**Symptoms:**
- Error rate > 5%
- Many 4xx/5xx responses
- Analytics show increasing errors

**Solutions:**

1. **Check Provider Status**
   ```bash
   curl https://your-worker.workers.dev/providers/status
   ```

2. **Review Rate Limits**
   ```bash
   curl https://your-worker.workers.dev/rate-limits
   # Check if requests are being blocked
   ```

3. **Validate API Keys**
   ```bash
   # Check logs for authentication errors
   wrangler tail | grep "auth\|token\|key"
   ```

4. **Enable Fallback Providers**
   ```bash
   curl -X POST https://your-worker.workers.dev/config \
     -H "Content-Type: application/json" \
     -d '{"enableFallbackProviders": true}'
   ```

### Issue: Slow Response Times

**Symptoms:**
- Response time > 200ms
- Users reporting slow performance
- Analytics show increasing latency

**Solutions:**

1. **Check Cache Performance**
   ```bash
   curl https://your-worker.workers.dev/cache/status
   # Look for hit rate < 70%
   ```

2. **Monitor Provider Latency**
   ```bash
   curl https://your-worker.workers.dev/analytics
   # Check provider-specific latency
   ```

3. **Adjust Concurrency Settings**
   ```bash
   curl -X PATCH https://your-worker.workers.dev/config \
     -H "Content-Type: application/json" \
     -d '{"maxConcurrency": 25}'
   ```

4. **Enable Performance Profiling**
   ```bash
   curl https://your-worker.workers.dev/profile/start
   # Wait for some requests
   curl https://your-worker.workers.dev/profile/stop
   ```

### Issue: Memory Issues

**Symptoms:**
- Memory usage > 10MB
- Out of memory errors
- Worker restarts

**Solutions:**

1. **Reduce Cache Size**
   ```bash
   curl -X PATCH https://your-worker.workers.dev/config \
     -H "Content-Type: application/json" \
     -d '{"cache": {"size": 50}}'
   ```

2. **Enable Memory Monitoring**
   ```bash
   curl https://your-worker.workers.dev/health
   # Check memory utilization
   ```

3. **Optimize Request Processing**
   ```bash
   # Check for memory leaks in logs
   wrangler tail | grep "memory\|leak"
   ```

### Issue: Rate Limiting Too Aggressive

**Symptoms:**
- Too many requests blocked
- Users complaining about limits
- Legitimate requests being rejected

**Solutions:**

1. **Increase Rate Limits**
   ```bash
   curl -X PATCH https://your-worker.workers.dev/config \
     -H "Content-Type: application/json" \
     -d '{"rateLimit": {"perHour": 2000, "burstSize": 100}}'
   ```

2. **Change Algorithm**
   ```bash
   # Switch to sliding window for smoother limiting
   curl -X PATCH https://your-worker.workers.dev/config \
     -H "Content-Type: application/json" \
     -d '{"rateLimit": {"algorithm": "sliding-window"}}'
   ```

3. **Monitor Usage Patterns**
   ```bash
   curl https://your-worker.workers.dev/analytics
   # Check request patterns
   ```

### Issue: Cache Not Working

**Symptoms:**
- Low cache hit rate
- All requests going to providers
- No performance improvement

**Solutions:**

1. **Check Cache Configuration**
   ```bash
   curl https://your-worker.workers.dev/cache/status
   # Verify cache is enabled and configured
   ```

2. **Increase Cache TTL**
   ```bash
   curl -X PATCH https://your-worker.workers.dev/config \
     -H "Content-Type: application/json" \
     -d '{"cache": {"ttl": 7200}}'
   ```

3. **Enable Smart Cache**
   ```bash
   curl -X PATCH https://your-worker.workers.dev/config \
     -H "Content-Type: application/json" \
     -d '{"cache": {"smartCache": true}}'
   ```

## 🛠️ Advanced Troubleshooting

### Debug Mode

Enable debug logging:

```bash
# Enable debug mode
curl -X PATCH https://your-worker.workers.dev/config \
  -H "Content-Type: application/json" \
  -d '{"debug": {"enabled": true, "level": "debug"}}'

# Check debug logs
wrangler tail | grep "DEBUG"
```

### Request Tracing

Trace specific requests:

```bash
# Enable request tracing
curl -X POST https://your-worker.workers.dev/trace/start

# Make a test request
curl -X POST https://your-worker.workers.dev/ \
  -H "Content-Type: application/json" \
  -H "X-Trace-Id: test-123" \
  -d '{"prompt": "Test request"}'

# Get trace results
curl https://your-worker.workers.dev/trace/test-123
```

### Performance Profiling

Profile performance issues:

```bash
# Start profiling
curl https://your-worker.workers.dev/profile/start

# Run load test or normal traffic for 1-2 minutes
# (simulate normal usage)

# Stop profiling and get results
curl https://your-worker.workers.dev/profile/stop

# Export detailed profile
curl https://your-worker.workers.dev/profile/export?format=json
```

## 🔍 Diagnostic Commands

### System Diagnostics

```bash
# Full system status
curl https://your-worker.workers.dev/diagnostics

# Memory analysis
curl https://your-worker.workers.dev/diagnostics/memory

# Performance metrics
curl https://your-worker.workers.dev/diagnostics/performance
```

### Log Analysis

```bash
# Search for specific errors
wrangler tail | grep "ERROR" | tail -10

# Check rate limiting events
wrangler tail | grep "rate.limit" | tail -10

# Monitor cache events
wrangler tail | grep "cache" | tail -10
```

### Configuration Validation

```bash
# Validate current configuration
curl https://your-worker.workers.dev/config/validate

# List all configuration issues
curl https://your-worker.workers.dev/config/issues
```

## 🚨 Emergency Procedures

### Complete System Reset

If all else fails:

```bash
# Stop all traffic (if possible)
# Backup current configuration
curl https://your-worker.workers.dev/config > config-backup.json

# Reset to defaults
curl -X POST https://your-worker.workers.dev/reset

# Restore configuration
curl -X POST https://your-worker.workers.dev/config \
  -H "Content-Type: application/json" \
  -d @config-backup.json

# Restart services
wrangler deploy
```

### Emergency Contact

For critical issues:
1. Check status page for known incidents
2. Contact on-call engineer via Slack
3. Escalate to infrastructure team if needed
4. Update incident tracking system

## 📊 Monitoring During Issues

### Real-time Monitoring

```bash
# Monitor key metrics in real-time
watch -n 5 'curl -s https://your-worker.workers.dev/health | jq .'
```

### Alert Setup

```bash
# Set up emergency alerts
curl -X POST https://your-worker.workers.dev/alerts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "emergency-alert",
    "condition": "error_rate > 0.1",
    "channels": ["slack-emergency", "sms"],
    "cooldown": "60s"
  }'
```

## 🛡️ Prevention Best Practices

### Regular Maintenance

1. **Daily Checks**
   ```bash
   # Automate daily health checks
   curl https://your-worker.workers.dev/health
   curl https://your-worker.workers.dev/cache/status
   ```

2. **Weekly Reviews**
   ```bash
   # Review weekly metrics
   curl "https://your-worker.workers.dev/analytics?range=7d"
   ```

3. **Monthly Optimization**
   ```bash
   # Adjust configurations based on usage patterns
   curl https://your-worker.workers.dev/analytics/trends
   ```

### Proactive Monitoring

1. **Set Up Alerts**
   ```bash
   curl -X POST https://your-worker.workers.dev/alerts \
     -H "Content-Type: application/json" \
     -d @alerts-config.json
   ```

2. **Configure Dashboards**
   - Set up monitoring dashboards
   - Configure automated reports
   - Enable real-time notifications

3. **Performance Baselines**
   - Establish normal performance ranges
   - Monitor for deviations
   - Set up automated remediation

## 📞 Getting Help

### Self-Service Resources

1. **Check Documentation**
   - [Configuration Guide](./configuration.md)
   - [Monitoring Guide](./monitoring.md)
   - [Developer Documentation](../developer-docs/)

2. **Run Diagnostics**
   ```bash
   curl https://your-worker.workers.dev/diagnostics/full
   ```

3. **Search Logs**
   ```bash
   wrangler tail | grep -i "error\|warn\|fail" | tail -20
   ```

### Contact Support

For complex issues requiring assistance:

1. **Gather Information**
   ```bash
   # Collect diagnostic data
   {
     echo "=== Health Check ==="
     curl https://your-worker.workers.dev/health
     echo -e "\n=== Analytics ==="
     curl https://your-worker.workers.dev/analytics
     echo -e "\n=== Recent Logs ==="
     wrangler tail --since 10m
   } > diagnostic-report.txt
   ```

2. **Create Support Ticket**
   - Include diagnostic report
   - Describe the issue and steps to reproduce
   - Note any recent changes or deployments
   - Specify impact and urgency

3. **Escalation Path**
   - Tier 1: Documentation and self-service
   - Tier 2: Developer support (24h response)
   - Tier 3: Infrastructure team (4h response)
   - Critical: On-call engineer (immediate)

## 🎯 Quick Reference

| Issue | Quick Check | Solution |
|-------|-------------|----------|
| Worker down | `curl /health` | Restart deployment |
| High errors | `curl /analytics` | Check providers |
| Slow responses | `curl /cache/status` | Tune cache |
| Memory issues | `curl /health` | Reduce cache size |
| Rate limiting | `curl /rate-limits` | Adjust limits |

Remember: Most issues can be resolved with configuration changes. Always test changes in staging first! 🚀