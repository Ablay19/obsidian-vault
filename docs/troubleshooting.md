# Troubleshooting Guide

This comprehensive guide helps diagnose and resolve issues with the Obsidian Vault system.

## Table of Contents

1. [Quick Diagnosis](#quick-diagnosis)
2. [Environment-Specific Issues](#environment-specific-issues)
3. [Service-Specific Issues](#service-specific-issues)
4. [Performance Problems](#performance-problems)
5. [Error Messages](#error-messages)
6. [Recovery Procedures](#recovery-procedures)
7. [Debugging Tools](#debugging-tools)

---

## Quick Diagnosis

### Health Check Commands

```bash
# Check API Gateway health
curl https://api.obsidianvault.com/health

# Check Worker status
curl https://ai-manim-worker.obsidianvault.workers.dev/health

# Check database connectivity
curl https://api.obsidianvault.com/api/v1/status

# Check all services
curl https://api.obsidianvault.com/api/v1/status | jq '.services'
```

### System Status Indicators

| Status | Meaning | Action |
|--------|---------|--------|
| `healthy` | All systems operational | No action needed |
| `degraded` | Some issues, partial functionality | Monitor closely |
| `unhealthy` | Major issues, limited functionality | Immediate action |

### Common Symptoms Quick Reference

| Symptom | Likely Cause | Quick Fix |
|---------|--------------|-----------|
| No response to messages | Worker not deployed | Redeploy worker |
| Slow AI responses | Provider issue or high load | Check provider status |
| Video generation fails | R2 upload or renderer error | Check R2 credentials |
| Database timeouts | Connection pool exhausted | Increase pool size |
| Rate limit errors | Too many requests | Implement backoff |

---

## Environment-Specific Issues

### Development Environment

#### Issue: Local Worker Not Starting

**Symptoms**
```
Error: Could not resolve module "@cloudflare/workers-types"
```

**Solution**
```bash
# Install dependencies
cd workers/ai-manim-worker
npm install

# Verify installation
ls node_modules/@cloudflare/workers-types
```

#### Issue: Database Connection Failed

**Symptoms**
```
dial tcp [::1]:5432: connect: connection refused
```

**Solution**
```bash
# Start PostgreSQL
pg_ctl -D /usr/local/var/postgres start

# Or check if running
ps aux | grep postgres

# Verify connection
psql -h localhost -U postgres -d obsidian_vault
```

#### Issue: R2 Access Denied

**Symptoms**
```
Access to R2 bucket denied: 403 Forbidden
```

**Solution**
```bash
# Verify R2 credentials in .env
cat .env | grep R2

# Test R2 connection
aws --endpoint-url=https://<account>.r2.cloudflarestorage.com \
    s3 ls s3://your-bucket-name \
    --access-key-id=<access-key> \
    --secret-access-key=<secret>
```

---

### Staging Environment

#### Issue: Preview Deployments Not Working

**Symptoms**
- Wrangler deploy fails
- Preview URL returns 404

**Solution**
```bash
# Check wrangler configuration
cat wrangler.toml | grep name

# Verify account connection
wrangler whoami

# Force redeploy
npx wrangler deploy --env staging
```

#### Issue: Staging Data Contamination

**Symptoms**
- Production data appearing in staging
- Inconsistent test results

**Solution**
```bash
# Isolate staging data
export DATABASE_URL="postgresql://user:pass@staging-db:5432/staging"
export SESSIONS_KV_ID="staging-kv-id"

# Clear cached data
wrangler kv:namespace list | grep staging
wrangler kv:delete --namespace-id=<staging-id> --key="*"
```

---

### Production Environment

#### Issue: High Error Rate

**Symptoms**
- >5% error rate on health check
- Increased 5xx responses

**Solution**
```bash
# Check recent errors
curl https://api.obsidianvault.com/health | jq '.error_rate'

# View worker logs
wrangler tail --environment production

# Check specific error patterns
wrangler tail 2>&1 | grep ERROR
```

#### Issue: Slow Response Times

**Symptoms**
- P99 latency > 2 seconds
- User complaints of timeouts

**Solution**
```bash
# Check latency metrics
curl https://api.obsidianvault.com/health | jq '.latency'

# Check provider response times
# Review AI service latency breakdown

# Check for bottlenecks
wrangler tail --environment production | head -100
```

---

## Service-Specific Issues

### WhatsApp Service

#### Webhook Verification Failed

**Error**
```
Webhook verification failed: wrong verification token
```

**Debug Steps**
```bash
# 1. Check verification token in code
grep -r "VERIFY_TOKEN" internal/whatsapp/

# 2. Compare with Meta dashboard
# Meta Business Dashboard > WhatsApp > Webhooks > Edit

# 3. Test webhook verification
curl -X GET "https://graph.facebook.com/v18.0/YOUR_PHONE_NUMBER_ID?fields=webhook_url&access_token=TOKEN"
```

**Resolution**
1. Ensure tokens match exactly (case-sensitive)
2. Check for whitespace in environment variable
3. Restart service after token update

#### Messages Not Delivered

**Error**
```
Message send failed: Too many messages sent
```

**Debug Steps**
```bash
# Check rate limit status
curl -H "Authorization: Bearer TOKEN" \
  https://api.obsidianvault.com/whatsapp/status

# Review sending rate
# Meta Business Manager > Account > WhatsApp > Activity
```

**Resolution**
1. Implement rate limiting
2. Queue messages with delays
3. Request limit increase from Meta

---

### AI Service

#### All Providers Failing

**Error**
```
All AI providers failed: connection refused to all endpoints
```

**Debug Steps**
```bash
# 1. Check internet connectivity
curl -I https://api.openai.com

# 2. Test each provider individually
curl -X POST "https://api.openai.com/v1/chat/completions" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"test"}]}'

# 3. Check provider status pages
# - status.openai.com
# - status.anthropic.com
```

**Resolution**
1. Check firewall rules
2. Verify API keys are valid
3. Implement fallback to cached responses

#### Unexpected High Costs

**Debug Steps**
```bash
# Check token usage
curl -H "Authorization: Bearer TOKEN" \
  https://api.obsidianvault.com/ai/usage | jq '.daily_cost'

# Review by user
curl -H "Authorization: Bearer TOKEN" \
  https://api.obsidianvault.com/ai/usage?user_id=XXX
```

**Resolution**
1. Implement per-user spending limits
2. Use cheaper models for simple queries
3. Add caching layer

---

### Manim Service

#### Render Timeout

**Error**
```
Render job timeout: exceeded 5 minutes
```

**Debug Steps**
```bash
# Check job status
curl https://ai-manim-worker.workers.dev/api/v1/jobs/{job_id}

# Check renderer logs
curl https://renderer.obsidianvault.com/logs/{job_id}
```

**Resolution**
1. Reduce animation complexity
2. Lower quality settings
3. Increase timeout for complex animations

#### Video Upload Failed

**Error**
```
R2 upload failed: signature does not match
```

**Debug Steps**
```bash
# Verify R2 credentials
cat .env | grep R2

# Test direct upload
aws s3 cp test.mp4 s3://bucket/key \
  --endpoint-url=https://r2.cloudflarestorage.com

# Check time sync on renderer
date
```

**Resolution**
1. Regenerate R2 credentials
2. Synchronize time on renderer server
3. Check R2 CORS configuration

---

## Performance Problems

### High Memory Usage

**Diagnosis**
```bash
# Check worker memory
wrangler tail --environment production | grep "memory"

# Profile memory usage
curl https://ai-manim-worker.workers.dev/debug/memory
```

**Solutions**
1. Implement streaming responses
2. Reduce batch sizes
3. Add memory limits to requests

### Slow Database Queries

**Diagnosis**
```bash
# Enable slow query logging
psql -c "ALTER SYSTEM SET log_min_duration_statement = 1000;"
psql -c "SELECT * FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"
```

**Solutions**
1. Add missing indexes
2. Optimize query patterns
3. Implement query caching

### Cache Inefficiency

**Diagnosis**
```bash
# Check cache hit rate
curl https://api.obsidianvault.com/health | jq '.cache_hit_rate'

# Review cache keys
redis-cli keys "*" | wc -l
```

**Solutions**
1. Increase cache TTL
2. Implement cache warming
3. Use LRU eviction

---

## Error Messages

### HTTP Error Codes

| Code | Meaning | Common Cause |
|------|---------|--------------|
| 400 | Bad Request | Invalid JSON, missing fields |
| 401 | Unauthorized | Invalid/missing token |
| 403 | Forbidden | Insufficient permissions |
| 404 | Not Found | Resource doesn't exist |
| 429 | Too Many Requests | Rate limit exceeded |
| 500 | Internal Error | Unhandled exception |
| 502 | Bad Gateway | Upstream service error |
| 503 | Service Unavailable | Overload or maintenance |

### Error Response Format

```json
{
  "error": {
    "code": "INVALID_REQUEST",
    "message": "Description of the error",
    "details": {
      "field": "specific_field",
      "reason": "why_it_failed"
    }
  },
  "request_id": "req_abc123",
  "timestamp": "2026-01-17T06:00:00Z"
}
```

### Common Error Codes

| Code | Resolution |
|------|------------|
| `PROVIDER_ERROR` | Check AI provider status, retry later |
| `RENDER_FAILED` | Check renderer logs, simplify animation |
| `UPLOAD_FAILED` | Verify R2 credentials, check storage quota |
| `RATE_LIMITED` | Implement backoff, request limit increase |
| `SESSION_EXPIRED` | User needs to re-authenticate |
| `VALIDATION_ERROR` | Fix request payload according to API spec |

---

## Recovery Procedures

### Full System Recovery

```bash
#!/bin/bash
# emergency-recovery.sh

echo "Starting full system recovery..."

# 1. Backup current state
echo "Backing up current data..."
./scripts/backup.sh

# 2. Rollback to last known good version
echo "Rolling back workers..."
npx wrangler deploy --env production --triggers backup

# 3. Flush caches
echo "Flushing caches..."
redis-cli FLUSHALL
wrangler kv:namespace list | grep production | xargs -I {} sh -c 'wrangler kv:delete --namespace-id={} --key="*"'

# 4. Restart services
echo "Restarting services..."
kubectl rollout restart deployment/api-gateway
kubectl rollout restart deployment/worker-ai

# 5. Verify health
echo "Verifying health..."
curl -f https://api.obsidianvault.com/health

echo "Recovery complete!"
```

### Database Recovery

```bash
#!/bin/bash
# db-recovery.sh

# Point-in-time recovery
pg_restore -h localhost -U postgres -d obsidian_vault \
  --point-in-time="2026-01-17T06:00:00Z" \
  backup.dump

# Verify data integrity
psql -c "SELECT count(*) FROM users;"
psql -c "SELECT count(*) FROM sessions;"
psql -c "SELECT count(*) FROM jobs;"
```

### Worker Rollback

```bash
# List previous deployments
npx wrangler deploy list --environment production

# Deploy previous version
npx wrangler deploy --env production --triggers previous

# Or specify a specific version
npx wrangler deploy --env production --triggers deployment-id-123
```

---

## Debugging Tools

### Log Analysis

```bash
# Tail real-time logs
wrangler tail --environment production

# Search for errors
wrangler tail 2>&1 | grep ERROR

# Filter by job ID
wrangler tail 2>&1 | grep "job_abc123"

# Export logs for analysis
wrangler tail --environment production --format json > logs.json
```

### Metrics Dashboard

Access Grafana dashboards at:
- Development: http://localhost:3000
- Staging: https://staging-grafana.obsidianvault.com
- Production: https://grafana.obsidianvault.com

**Key Dashboards**
- `obsidian-vault-overview` - System health
- `ai-service-metrics` - AI provider performance
- `manim-service` - Video generation stats
- `cost-analysis` - Resource costs

### Request Tracing

```bash
# Add trace ID to request
curl -H "X-Trace-ID: my-trace-123" \
  https://api.obsidianvault.com/api/v1/endpoint

# Search traces in Jaeger
# http://jaeger.obsidianvault.com
# Search by: my-trace-123
```

### Performance Profiling

```bash
# CPU profiling
curl https://api.obsidianvault.com/debug/pprof/profile > cpu.prof

# Memory profiling
curl https://api.obsidianvault.com/debug/pprof/heap > mem.prof

# Analyze with go tool
go tool pprof -http=:8080 cpu.prof
```

---

## Emergency Contacts

| Issue | Contact |
|-------|---------|
| **Production Outage** | on-call@obsidianvault.com |
| **Security Incident** | security@obsidianvault.com |
| **Billing Questions** | billing@obsidianvault.com |
| **Infrastructure** | infrastructure@obsidianvault.com |

### Runbooks

| Situation | Runbook |
|-----------|---------|
| High error rate | [Runbook: High Error Rate](runbooks/high-error-rate.md) |
| Database issues | [Runbook: Database Problems](runbooks/database.md) |
| AI service down | [Runbook: AI Service Outage](runbooks/ai-outage.md) |
| R2 storage full | [Runbook: Storage Issues](runbooks/storage.md) |
| Security breach | [Runbook: Security Incident](runbooks/security.md) |

---

## Preventive Measures

### Monitoring Checklist

- [ ] Check health endpoints every 5 minutes
- [ ] Review error rates daily
- [ ] Monitor costs weekly
- [ ] Test backup restoration monthly
- [ ] Review security logs weekly
- [ ] Update dependencies monthly

### Regular Maintenance

| Task | Frequency | Owner |
|------|-----------|-------|
| Rotate API keys | Quarterly | Security |
| Review access logs | Weekly | Ops |
| Update dependencies | Monthly | Dev |
| Test disaster recovery | Monthly | Ops |
| Capacity planning review | Quarterly | Architecture |
| Cost optimization review | Monthly | FinOps |
