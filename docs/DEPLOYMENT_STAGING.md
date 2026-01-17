# Staging Deployment Guide

This guide covers deploying Obsidian Vault to the staging environment.

## Overview

Staging environment mirrors production with:
- Real Cloudflare infrastructure
- Test data isolation
- Preview deployments for PRs
- Full feature parity with production

## Prerequisites

### Required Access

| Resource | Access Level | How to Obtain |
|----------|--------------|---------------|
| Cloudflare Account | Developer | Request via #cloud-infra Slack |
| Doppler Workspace | Editor | Request via #security Slack |
| GitHub Repository | Write | Team membership |
| Database | Read/Write | #database-access request |

### Tools Required

```bash
# Install Wrangler
npm install -g wrangler

# Install Doppler CLI
brew install dopplerhq/cli/doppler

# Verify access
wrangler whoami
doppler me
```

## Environment Configuration

### 1. Setup Doppler Secrets

```bash
# Select staging config
doppler setup --config staging

# Verify secrets are available
doppler secrets --config staging
```

### 2. Create Staging Environment File

Create `wrangler.staging.toml`:

```toml
name = "ai-manim-worker-staging"
main = "src/index.ts"
compatibility_date = "2024-09-23"
compatibility_flags = ["nodejs_compat"]

[vars]
LOG_LEVEL = "info"
ENVIRONMENT = "staging"
USE_MOCK_RENDERER = "false"

# Staging KV Namespace
[[kv_namespaces]]
binding = "SESSIONS"
id = "staging-kv-namespace-id"

[env.staging]
name = "ai-manim-worker-staging"
vars = { 
  LOG_LEVEL = "debug"
  USE_MOCK_RENDERER = "false"
}

[env.staging.kv_namespaces]
binding = "SESSIONS"
id = "staging-kv-namespace-id"
```

### 3. R2 Bucket Configuration

```bash
# Create staging bucket (one-time setup)
aws s3 mb s3://obsidian-vault-staging-videos \
  --endpoint-url=https://${CLOUDFLARE_ACCOUNT_ID}.r2.cloudflarestorage.com

# Configure lifecycle rule (30-day retention)
aws s3api put-bucket-lifecycle-configuration \
  --bucket obsidian-vault-staging-videos \
  --endpoint-url=https://${CLOUDFLARE_ACCOUNT_ID}.r2.cloudflarestorage.com \
  --LifecycleConfiguration '{"Rules":[{"ID":"VideoExpiry","Status":"Enabled","Expiration":{"Days":30},"Prefix":"videos/"}]}'
```

## Database Setup

### Staging Database

```bash
# Connect to staging database
psql -h staging-db.obsidianvault.com \
  -U obsidian_vault_staging \
  -d obsidian_vault_staging

# Run migrations
go run cmd/migrate/main.go --env staging up

# Seed test data (optional)
go run cmd/seed/main.go --env staging
```

### Test Data

The staging database includes:
- 100 test users
- 500 sample sessions
- 1000 job records
- Test media files

## Deployment Process

### 1. Pull Latest Code

```bash
git checkout main
git pull origin main

# Install dependencies
go mod download
cd workers/ai-manim-worker && npm install
```

### 2. Run Tests

```bash
# Run all tests
npm run test
go test ./... -short

# Run integration tests
npm run test:integration
go test ./... -tags=integration -v
```

### 3. Build Workers

```bash
# Build TypeScript workers
cd workers/ai-manim-worker
npm run build

cd ../worker-whatsapp
npm run build
```

### 4. Deploy to Staging

```bash
# Deploy workers
cd workers/ai-manim-worker
npx wrangler deploy --env staging --config wrangler.staging.toml

cd ../worker-whatsapp
npx wrangler deploy --env staging
```

### 5. Deploy Additional Services

```bash
# Deploy API Gateway
cd apps/api-gateway
npx wrangler deploy --env staging

# Deploy Manim Renderer (if using container)
kubectl apply -f deploy/k8s/staging/manim-renderer.yaml
```

### 6. Verify Deployment

```bash
# Check worker health
curl https://staging-ai-manim-worker.obsidianvault.workers.dev/health

# Check API health
curl https://staging-api.obsidianvault.com/health

# Check all services status
curl https://staging-api.obsidianvault.com/api/v1/status
```

## Preview Deployments

### Automatic Preview for PRs

Every PR automatically gets a preview deployment:

```
https://pr-{PR_NUMBER}.staging.workers.dev
```

### Access Preview Deployment

```bash
# Get preview URL for current PR
gh pr view --json url

# Deploy to specific preview
npx wrangler deploy --env preview --config wrangler.preview.toml
```

### Preview Verification Checklist

- [ ] Health check passes
- [ ] Database migrations applied
- [ ] R2 bucket accessible
- [ ] AI providers responding
- [ ] Webhooks configured
- [ ] Integration tests passing

## Staging-Specific Testing

### Load Testing

```bash
# Run k6 load test
k6 run scripts/load-test.js \
  -e ENV_URL=https://staging-api.obsidianvault.com \
  -e USERS=50 \
  -e DURATION=5m
```

### Integration Testing

```bash
# Run full integration test suite
npm run test:integration

# Test specific flows
go test ./tests/integration/... -v -run "TestWhatsAppFlow"
go test ./tests/integration/... -v -run "TestAIFlow"
go test ./tests/integration/... -v -run "TestManimFlow"
```

### End-to-End Testing

```bash
# Test complete user journey
npm run test:e2e

# Manual testing checklist
# 1. Submit WhatsApp message
# 2. Verify AI response
# 3. Request video generation
# 4. Verify video delivery
```

## Monitoring Staging

### Access Metrics

```bash
# Grafana dashboard
open https://staging-grafana.obsidianvault.com

# Key dashboards:
# - Staging Overview
# - API Performance
# - AI Service Metrics
# - Video Generation Stats
```

### Alert Configuration

Staging alerts (lower thresholds than production):

| Alert | Threshold | Channel |
|-------|-----------|---------|
| High Error Rate | > 1% | #staging-alerts |
| High Latency | P99 > 2s | #staging-alerts |
| AI Provider Down | Any provider | #staging-alerts |

### Log Access

```bash
# View recent logs
npx wrangler tail --env staging --no-timestamp

# Search for errors
npx wrangler tail --env staging 2>&1 | grep ERROR

# Export logs for analysis
npx wrangler tail --env staging --format json > staging-logs.json
```

## Rollback Procedure

### Quick Rollback

```bash
# List recent deployments
npx wrangler deploy list --env staging

# Deploy previous version
npx wrangler deploy --env staging --triggers previous

# Or rollback to specific version
npx wrangler deploy --env staging --triggers deployment-id-123
```

### Database Rollback

```bash
# Rollback last migration
go run cmd/migrate/main.go --env staging down 1

# Restore from backup (if needed)
pg_restore -h staging-db.obsidianvault.com \
  -U obsidian_vault_staging \
  -d obsidian_vault_staging \
  staging-backup.dump
```

### Verify Rollback

```bash
# Check deployment version
curl https://staging-api.obsidianvault.com/health | jq '.version'

# Run smoke tests
npm run test:smoke

# Notify team
# Post to #deployments: "Staging rolled back to version X.X.X"
```

## Data Management

### Staging Data Refresh

```bash
# Full refresh from production (sanitized)
./scripts/refresh-staging-data.sh

# Partial refresh (last 24 hours)
./scripts/refresh-staging-data.sh --hours=24

# Clear and reseed
./scripts/clear-staging-data.sh
./scripts/seed-staging-data.sh
```

### Backup Staging Data

```bash
# Create backup
pg_dump -h staging-db.obsidianvault.com \
  -U obsidian_vault_staging \
  -d obsidian_vault_staging \
  > staging-backup-$(date +%Y%m%d).dump

# Upload to backup storage
aws s3 cp staging-backup-$(date +%Y%m%d).dump \
  s3://obsidian-vault-backups/staging/
```

## Troubleshooting

### Common Issues

#### Deployment Failed

```bash
# Check deployment logs
npx wrangler deploy --env staging --debug 2>&1 | tail -50

# Verify secrets
doppler secrets --config staging

# Check KV namespace IDs
npx wrangler kv:namespace list | grep staging
```

#### Database Connection Issues

```bash
# Test database connection
pg_isready -h staging-db.obsidianvault.com -p 5432

# Check connection pool
psql -h staging-db.obsidianvault.com \
  -U obsidian_vault_staging \
  -c "SELECT count(*) FROM pg_stat_activity;"
```

#### AI Service Not Responding

```bash
# Test AI provider connectivity
curl -X POST "https://api.openai.com/v1/chat/completions" \
  -H "Authorization: Bearer $OPENAI_API_KEY" \
  -d '{"model":"gpt-4o","messages":[{"role":"user","content":"test"}]}'

# Check provider status pages
# OpenAI: https://status.openai.com
# Anthropic: https://status.anthropic.com
```

## Security Checklist

- [ ] Secrets managed via Doppler
- [ ] No hardcoded credentials in code
- [ ] Webhook URLs use HTTPS
- [ ] CORS configured correctly
- [ ] Rate limiting enabled
- [ ] Audit logging active

## Next Steps

After staging deployment:

1. Complete [Integration Testing](#staging-specific-testing)
2. Get [PR Review Approval](https://github.com/your-org/obsidian-vault/pull/XXX)
3. Prepare [Production Deployment](DEPLOYMENT_PRODUCTION.md)
4. Notify stakeholders of deployment status
