# Production Deployment Guide

This guide covers deploying Obsidian Vault to the production environment.

## ⚠️ Production Warning

**Before proceeding, ensure you have:**
- Completed staging deployment and testing
- Received approval from tech lead
- Prepared rollback plan
- Notified relevant stakeholders
- Prepared communication plan for users

## Overview

Production environment characteristics:
- Full production data
- Real monetary costs
- User-facing services
- 99.9% availability target
- 24/7 monitoring

## Prerequisites

### Required Access

| Resource | Access Level | Approver |
|----------|--------------|----------|
| Cloudflare Production | Admin | CTO |
| Doppler Production | Admin | Security |
| GitHub Production | Maintain | Tech Lead |
| PagerDuty | On-call | Engineering Manager |

### Approval Checklist

- [ ] Tech Lead approval obtained
- [ ] Security review completed
- [ ] Performance benchmarks passed
- [ ] Integration tests passing
- [ ] Load testing passed
- [ ] Rollback plan prepared
- [ ] Communication prepared

## Pre-Deployment Checklist

### 1. Code Freeze

```bash
# Ensure no unmerged changes in main
git checkout main
git pull origin main
git log --oneline -5

# Verify CI passing
gh run list --workflow "CI" --limit 1
```

### 2. Version Bump

```bash
# Update version
bump2version patch  # or minor/major

# Commit version change
git add .
git commit -m "chore: bump version to $(git describe --tags)"
git tag "v$(cat VERSION)"
```

### 3. Final Security Scan

```bash
# Run security audit
npm audit
go sec ./...

# Check for vulnerabilities
trivy fs .
```

### 4. Backup Production Data

```bash
# Create database backup
pg_dump -h prod-db.obsidianvault.com \
  -U obsidian_vault_prod \
  -d obsidian_vault_prod \
  > backup/prod-backup-$(date +%Y%m%d-%H%M%S).dump

# Upload to backup storage
aws s3 cp backup/prod-backup-*.dump \
  s3://obsidian-vault-backups/prod/ \
  --storage-class STANDARD_IA

# Verify backup
pg_restore --list prod-backup-*.dump | head -20
```

### 5. Notify Stakeholders

```bash
# Send deployment notification
./scripts/notify-deployment.sh \
  --environment production \
  --version $(cat VERSION) \
  --channel "#deployments" \
  --estimate "30 minutes"
```

## Environment Configuration

### 1. Production Secrets (Doppler)

```bash
# Verify production config
doppler configs get production

# Review critical secrets
doppler secrets get --config production --json \
  | jq '. | {OPENAI_API_KEY, DATABASE_URL, JWT_SECRET}'
```

### 2. Production wrangler.toml

```toml
name = "ai-manim-worker"
main = "src/index.ts"
compatibility_date = "2024-09-23"
compatibility_flags = ["nodejs_compat"]

[vars]
LOG_LEVEL = "info"
ENVIRONMENT = "production"
USE_MOCK_RENDERER = "false"

# Production KV Namespace
[[kv_namespaces]]
binding = "SESSIONS"
id = "prod-kv-namespace-id"

[env.production]
name = "ai-manim-worker"
vars = { 
  LOG_LEVEL = "warn"
  USE_MOCK_RENDERER = "false"
}

[env.production.kv_namespaces]
binding = "SESSIONS"
id = "prod-kv-namespace-id"

# Rate limiting
[triggers]
crons = ["*/5 * * * *"]  # Health check every 5 minutes
```

### 3. R2 Production Configuration

```bash
# Create production bucket
aws s3 mb s3://obsidian-vault-prod-videos \
  --endpoint-url=https://${CLOUDFLARE_ACCOUNT_ID}.r2.cloudflarestorage.com

# Configure strict lifecycle (24-hour retention)
aws s3api put-bucket-lifecycle-configuration \
  --bucket obsidian-vault-prod-videos \
  --endpoint-url=https://${CLOUDFLARE_ACCOUNT_ID}.r2.cloudflarestorage.com \
  --LifecycleConfiguration '{
    "Rules": [{
      "ID": "VideoExpiry24h",
      "Status": "Enabled",
      "Filter": {"Prefix": "videos/"},
      "Expiration": {"Days": 1},
      "NoncurrentVersionExpiration": {"NoncurrentDays": 1}
    }]
  }'

# Enable versioning
aws s3api put-bucket-versioning \
  --bucket obsidian-vault-prod-videos \
  --endpoint-url=https://${CLOUDFLARE_ACCOUNT_ID}.r2.cloudflarestorage.com \
  --VersioningConfiguration Status=Enabled
```

## Deployment Process

### 1. Deploy Workers

```bash
# Deploy AI Manim Worker
cd workers/ai-manim-worker
npx wrangler deploy --env production

# Deploy WhatsApp Worker
cd ../worker-whatsapp
npx wrangler deploy --env production

# Deploy API Gateway
cd ../../apps/api-gateway
npx wrangler deploy --env production
```

### 2. Deploy Kubernetes Services

```bash
# Deploy to Kubernetes
kubectl apply -f deploy/k8s/production/

# Verify deployments
kubectl get pods -n production
kubectl get svc -n production

# Wait for ready
kubectl rollout status deployment/api-gateway -n production
kubectl rollout status deployment/ai-worker -n production
kubectl rollout status deployment/manim-renderer -n production
```

### 3. Run Database Migrations

```bash
# Apply migrations with backup
go run cmd/migrate/main.go \
  --env production \
  --backup-before \
  up

# Verify migration
go run cmd/migrate/main.go --env production status
```

### 4. Verify Deployment

```bash
# Check all workers
curl https://ai-manim-worker.obsidianvault.workers.dev/health
curl https://worker-whatsapp.obsidianvault.workers.dev/health
curl https://api-gateway.obsidianvault.workers.dev/health

# Check API status
curl https://api.obsidianvault.com/api/v1/status | jq '{status, version, services}'

# Run smoke tests
npm run test:smoke -- --env production
```

### 5. Update DNS

```bash
# Update DNS records (Cloudflare)
# Already configured, just verify
dig api.obsidianvault.com
dig ai-manim-worker.obsidianvault.workers.dev

# Verify SSL certificate
echo | openssl s_client -servername api.obsidianvault.com -connect api.obsidianvault.com:443 | openssl x509 -noout -dates
```

## Post-Deployment Verification

### 1. Smoke Tests

```bash
#!/bin/bash
# smoke-tests.sh

set -e

echo "Running production smoke tests..."

# Test 1: Health check
echo "Testing health endpoint..."
HEALTH=$(curl -s https://api.obsidianvault.com/health)
if ! echo "$HEALTH" | jq -e '.status == "healthy"' 2>/dev/null; then
  echo "FAIL: Health check failed"
  exit 1
fi
echo "✓ Health check passed"

# Test 2: API status
echo "Testing API status..."
STATUS=$(curl -s https://api.obsidianvault.com/api/v1/status)
if ! echo "$STATUS" | jq -e '.status == "operational"' 2>/dev/null; then
  echo "FAIL: API status check failed"
  exit 1
fi
echo "✓ API status passed"

# Test 3: Database connectivity
echo "Testing database..."
if ! curl -s https://api.obsidianvault.com/api/v1/status | jq -e '.services.database == "healthy"' 2>/dev/null; then
  echo "FAIL: Database check failed"
  exit 1
fi
echo "✓ Database check passed"

# Test 4: AI service
echo "Testing AI service..."
curl -s -X POST https://api.obsidianvault.com/api/v1/ai/chat \
  -H "Content-Type: application/json" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"test"}]}' \
  > /dev/null
echo "✓ AI service passed"

echo ""
echo "All smoke tests passed! ✓"
```

```bash
chmod +x smoke-tests.sh
./smoke-tests.sh
```

### 2. Monitor Metrics

```bash
# Open Grafana dashboard
open https://grafana.obsidianvault.com/d/production-overview

# Key metrics to watch:
# - Request rate (should be stable)
# - Error rate (should be < 0.1%)
# - Latency P95 (should be < 500ms)
# - AI provider response times
```

### 3. Check Logs

```bash
# Check for errors
npx wrangler tail --env production 2>&1 | grep -i error | head -20

# Monitor rate of errors
npx wrangler tail --env production 2>&1 | grep -c ERROR

# Check for warnings
npx wrangler tail --env production 2>&1 | grep -c WARN
```

## Rollback Procedure

### 1. Immediate Rollback

```bash
# Rollback workers to previous version
npx wrangler deploy --env production --triggers previous

# Or deploy specific version
npx wrangler deploy --env production --triggers deployment-id-123
```

### 2. Database Rollback

```bash
# Rollback last migration
go run cmd/migrate/main.go --env production down 1

# Full database restore (if needed)
pg_restore -h prod-db.obsidianvault.com \
  -U obsidian_vault_prod \
  -d obsidian_vault_prod \
  --clean \
  backup/prod-backup-*.dump
```

### 3. Verify Rollback

```bash
# Check deployed version
curl https://api.obsidianvault.com/health | jq '.version'

# Run smoke tests
./smoke-tests.sh

# Notify team
./scripts/notify-deployment.sh \
  --environment production \
  --action rollback \
  --channel "#deployments,#incident-response"
```

### 4. Post-Incident Review

If rollback was needed:
- Document incident in #incident-reports
- Schedule post-mortem meeting
- Update runbooks
- Implement fix before next deployment

## Monitoring in Production

### Critical Alerts

| Alert | Threshold | Response |
|-------|-----------|----------|
| Error Rate | > 1% for 5min | Page on-call |
| Latency P99 | > 2s for 5min | Page on-call |
| AI Provider Down | Any provider | Auto-fallback |
| Database Unavailable | > 1min | Page on-call |

### Dashboard Links

- [Production Overview](https://grafana.obsidianvault.com/d/production-overview)
- [API Performance](https://grafana.obsidianvault.com/d/api-performance)
- [AI Service Metrics](https://grafana.obsidianvault.com/d/ai-service)
- [Cost Analysis](https://grafana.obsidianvault.com/d/cost-analysis)

### On-Call Schedule

Current on-call engineer:
- Check PagerDuty: https://obsidian-vault.pagerduty.com/schedules

## Cost Management

### Daily Cost Monitoring

```bash
# Check daily spend
curl https://api.obsidianvault.com/api/v1/status \
  | jq '.cost_today'

# Forecast monthly cost
curl https://api.obsidianvault.com/api/v1/status \
  | jq '.cost_forecast'
```

### Cost Alerts

| Alert | Threshold | Action |
|-------|-----------|--------|
| Daily Spend | > $100 | Notify #finops |
| Weekly Spend | > $700 | Review with team |
| Unexpected Spike | > 50% increase | Investigate immediately |

## Post-Deployment Tasks

### 1. Documentation Update

- [ ] Update API documentation
- [ ] Document new features
- [ ] Update runbooks
- [ ] Archive old documentation

### 2. Stakeholder Communication

```bash
# Send deployment summary
./scripts/send-deployment-summary.sh \
  --version $(cat VERSION) \
  --changes "Added video generation, improved AI responses" \
  --issues "None" \
  --next-steps "Monitor metrics for 24 hours"
```

### 3. Monitor for 24 Hours

- Check metrics every 4 hours
- Review error reports
- Monitor cost accumulation
- Watch for user reports

## Security Checklist

- [ ] All secrets via Doppler
- [ ] HTTPS enforced
- [ ] CORS configured
- [ ] Rate limiting enabled
- [ ] Audit logging active
- [ ] WAF rules enabled
- [ ] DDoS protection active
- [ ] SSL/TLS certificates valid

## Troubleshooting Common Issues

### High Error Rate

```bash
# Check recent errors
npx wrangler tail --env production 2>&1 | grep ERROR | head -50

# Check specific endpoint errors
curl https://api.obsidianvault.com/api/v1/status | jq '.error_breakdown'

# Restart affected workers
kubectl rollout restart deployment/api-gateway -n production
```

### High Latency

```bash
# Check latency breakdown
curl https://api.obsidianvault.com/api/v1/status | jq '.latency_breakdown'

# Check database query times
psql -h prod-db.obsidianvault.com \
  -U obsidian_vault_prod \
  -c "SELECT query, mean_time FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"

# Check AI provider response times
curl https://api.obsidianvault.com/api/v1/status | jq '.ai_provider_latency'
```

### R2 Upload Failures

```bash
# Check R2 connectivity
aws s3 ls s3://obsidian-vault-prod-videos \
  --endpoint-url=https://${CLOUDFLARE_ACCOUNT_ID}.r2.cloudflarestorage.com

# Check R2 quota
curl https://api.obsidianvault.com/api/v1/status | jq '.r2_usage'

# Verify credentials in Doppler
doppler secrets get --config production R2_ACCESS_KEY_ID
```

## Emergency Contacts

| Situation | Contact |
|-----------|---------|
| **Critical Outage** | PagerDuty: https://obsidian-vault.pagerduty.com |
| **Security Incident** | security@obsidianvault.com |
| **Database Emergency** | #database-emergency (Slack) |
| **Infrastructure** | infrastructure@obsidianvault.com |

## Post-Mortem Template

If any issues occurred:

```markdown
# Incident Report: [Title]

## Summary
[Description of incident]

## Timeline
- [Time] Incident detected
- [Time] Response initiated
- [Time] Root cause identified
- [Time] Fix deployed
- [Time] Incident resolved

## Root Cause
[Technical explanation]

## Impact
[Users affected, duration, data loss]

## Resolution
[What was done to fix]

## Lessons Learned
[What could be improved]

## Action Items
- [ ] [Task] - Owner - Due Date
```

---

## Summary Checklist

Before deployment:
- [ ] Code frozen
- [ ] Tests passing
- [ ] Security scan clean
- [ ] Backup created
- [ ] Stakeholders notified
- [ ] Rollback plan ready

After deployment:
- [ ] Smoke tests passing
- [ ] Metrics healthy
- [ ] No critical alerts
- [ ] 24-hour monitoring active
- [ ] Documentation updated
