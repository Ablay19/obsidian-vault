# Deployment Guide

## 🚀 Deployment Overview

This guide covers the complete deployment process for enhanced Cloudflare Workers, from initial setup to production deployment. It includes CI/CD integration, environment management, and rollback procedures.

## 📋 Prerequisites

Before deploying, ensure you have:

- **Cloudflare Account**: Active account with Workers enabled
- **Wrangler CLI**: Installed and configured (`npm install -g wrangler`)
- **API Token**: Cloudflare API token with Workers permissions
- **Environment Variables**: All required environment variables configured
- **Node.js**: Version 16+ installed
- **Git Repository**: For CI/CD integration

## 🔧 Initial Setup

### 1. Install Wrangler CLI

```bash
# Install Wrangler globally
npm install -g wrangler

# Authenticate with Cloudflare
wrangler login

# Verify authentication
wrangler whoami
```

### 2. Configure Project

```bash
# Navigate to workers directory
cd workers

# Initialize Wrangler (if not already done)
wrangler init

# Update wrangler.toml with your configuration
```

Example `wrangler.toml`:

```toml
name = "ai-proxy-worker"
main = "ai-proxy/src/index.js"
compatibility_date = "2024-01-01"

[env.production]
name = "ai-proxy-worker-prod"
routes = [
  { pattern = "https://api.yourdomain.com/*", zone_name = "yourdomain.com" }
]

[env.staging]
name = "ai-proxy-worker-staging"
routes = [
  { pattern = "https://staging-api.yourdomain.com/*", zone_name = "yourdomain.com" }
]

[env.development]
name = "ai-proxy-worker-dev"
```

### 3. Set Environment Variables

```bash
# Create secrets for production
wrangler secret put OPENAI_API_KEY --env production
wrangler secret put ANTHROPIC_API_KEY --env production
wrangler secret put HUGGINGFACE_API_KEY --env production

# Create secrets for staging
wrangler secret put OPENAI_API_KEY --env staging
wrangler secret put ANTHROPIC_API_KEY --env staging
```

## 🌍 Environment Configuration

### Development Environment

```bash
# Deploy to development
wrangler deploy --env development

# View development logs
wrangler tail --env development

# Test development endpoint
curl https://ai-proxy-worker-dev.your-subdomain.workers.dev/health
```

### Staging Environment

```bash
# Deploy to staging
wrangler deploy --env staging

# Run smoke tests on staging
npm run test:staging

# Verify staging deployment
curl https://staging-api.yourdomain.com/health
```

### Production Environment

```bash
# Deploy to production (requires confirmation)
wrangler deploy --env production

# Run production health checks
npm run test:production

# Monitor production logs
wrangler tail --env production
```

## 🔄 CI/CD Integration

### GitHub Actions

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy Workers

on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '18'
          cache: 'npm'
      
      - name: Install dependencies
        run: npm ci
      
      - name: Run tests
        run: npm test
      
      - name: Run linter
        run: npm run lint
      
      - name: Type check
        run: npm run type-check

  deploy-staging:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/develop'
    steps:
      - uses: actions/checkout@v3
      
      - name: Deploy to Staging
        uses: cloudflare/wrangler-action@v2
        with:
          apiToken: ${{ secrets.CLOUDFLARE_API_TOKEN }}
          accountId: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
          command: deploy --env staging

      - name: Run smoke tests
        run: npm run test:staging

  deploy-production:
    needs: test
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v3
      
      - name: Deploy to Production
        uses: cloudflare/wrangler-action@v2
        with:
          apiToken: ${{ secrets.CLOUDFLARE_API_TOKEN }}
          accountId: ${{ secrets.CLOUDFLARE_ACCOUNT_ID }}
          command: deploy --env production

      - name: Run production tests
        run: npm run test:production

      - name: Notify deployment
        uses: 8398a7/action-slack@v3
        with:
          status: ${{ job.status }}
          text: 'Production deployment completed'
        env:
          SLACK_WEBHOOK_URL: ${{ secrets.SLACK_WEBHOOK_URL }}
```

### GitLab CI

Create `.gitlab-ci.yml`:

```yaml
stages:
  - test
  - deploy-staging
  - deploy-production

test:
  stage: test
  image: node:18
  script:
    - npm ci
    - npm test
    - npm run lint

deploy-staging:
  stage: deploy-staging
  image: node:18
  only:
    - develop
  script:
    - npm install -g wrangler
    - wrangler deploy --env staging
    - npm run test:staging

deploy-production:
  stage: deploy-production
  image: node:18
  only:
    - main
  when: manual
  script:
    - npm install -g wrangler
    - wrangler deploy --env production
    - npm run test:production
```

## 📦 Deployment Strategies

### Blue-Green Deployment

```bash
# Deploy blue version
wrangler deploy --env production-blue

# Verify blue version
curl https://api-blue.yourdomain.com/health

# Switch traffic to blue
wrangler routes replace \
  --zone yourdomain.com \
  --pattern "https://api.yourdomain.com/*" \
  --destination "https://api-blue.yourdomain.com"

# Keep green version as fallback
```

### Canary Deployment

```bash
# Deploy canary version
wrangler deploy --env production-canary

# Route 10% traffic to canary
wrangler routes replace \
  --zone yourdomain.com \
  --pattern "https://api.yourdomain.com/*" \
  --destination "https://api-canary.yourdomain.com" \
  --weight 10

# Monitor canary performance
wrangler tail --env production-canary

# Gradually increase traffic (25%, 50%, 100%)
```

### Rolling Deployment

```bash
# Deploy to multiple workers sequentially
for i in {1..10}; do
  wrangler deploy --env production-worker-$i
  sleep 30
done
```

## 🔍 Pre-Deployment Checklist

### Code Review
- [ ] All tests passing
- [ ] Code reviewed and approved
- [ ] No console.log statements
- [ ] Error handling complete
- [ ] Documentation updated

### Configuration
- [ ] Environment variables set
- [ ] Secrets configured
- [ ] Routes configured
- [ ] DNS records updated
- [ ] SSL certificates valid

### Testing
- [ ] Unit tests passing
- [ ] Integration tests passing
- [ ] Smoke tests passing
- [ ] Performance tests within limits
- [ ] Security tests passing

### Monitoring
- [ ] Analytics enabled
- [ ] Error tracking configured
- [ ] Health checks set up
- [ ] Alerting configured
- [ ] Logging enabled

## 🚨 Post-Deployment Validation

### Automated Checks

```bash
# Run automated validation script
./validate-deployment.sh

# Check health endpoint
curl https://api.yourdomain.com/health

# Run smoke tests
npm run test:smoke

# Verify metrics
curl https://api.yourdomain.com/metrics
```

### Manual Checks

```bash
# Monitor real-time logs
wrangler tail --env production

# Check cache performance
curl https://api.yourdomain.com/analytics/cache

# Verify rate limiting
curl -H "X-User-ID: test-user" https://api.yourdomain.com/rate-limit

# Test provider failover
curl https://api.yourdomain.com/providers/health
```

## 🔄 Rollback Procedures

### Immediate Rollback

```bash
# Rollback to previous version
wrangler rollback --env production

# Verify rollback
curl https://api.yourdomain.com/health

# Monitor logs
wrangler tail --env production
```

### Manual Rollback

```bash
# Revert to specific version
wrangler deploy --env production --version 42

# Or deploy from backup
wrangler deploy --env production ai-proxy/src/index.backup.js
```

### Emergency Rollback

```bash
# Deploy maintenance page
wrangler deploy --env maintenance

# Switch all traffic to maintenance
wrangler routes replace \
  --zone yourdomain.com \
  --pattern "https://api.yourdomain.com/*" \
  --destination "https://maintenance.yourdomain.com"
```

## 📊 Monitoring Deployments

### Deployment Metrics

```javascript
// Track deployment metrics in your code
export async function trackDeployment(deploymentData) {
  return {
    deploymentId: deploymentData.id,
    version: deploymentData.version,
    environment: deploymentData.env,
    timestamp: new Date().toISOString(),
    duration: deploymentData.duration,
    success: deploymentData.success,
    errors: deploymentData.errors,
    performance: {
      deployTime: deploymentData.deployTime,
      warmupTime: deploymentData.warmupTime,
      firstResponseTime: deploymentData.firstResponseTime
    }
  };
}
```

### Deployment Dashboard

Create a deployment dashboard to track:

- **Deployment History**: All deployments with timestamps
- **Success Rate**: Percentage of successful deployments
- **Rollback Rate**: Frequency of rollbacks
- **Deployment Time**: Average deployment duration
- **Error Rate**: Errors during and after deployment

## 🛡️ Best Practices

### Security
- Never commit secrets to git
- Use environment variables for sensitive data
- Rotate API keys regularly
- Enable IP restrictions for deployments
- Use two-factor authentication for Cloudflare

### Reliability
- Always test in staging first
- Use blue-green deployments for critical updates
- Implement canary deployments for experimental features
- Keep previous versions available for rollback
- Monitor deployments actively after release

### Performance
- Optimize bundle size before deployment
- Enable compression
- Configure appropriate cache TTLs
- Monitor cold start times
- Use edge caching for static assets

### Monitoring
- Set up deployment alerts
- Track deployment metrics
- Monitor error rates post-deployment
- Analyze performance trends
- Create deployment dashboards

## 🔗 Related Documentation

- [Getting Started Guide](../user-guides/getting-started.md)
- [Configuration Guide](../user-guides/configuration.md)
- [Monitoring Setup](./monitoring.md)
- [Incident Response](./incident-response.md)
- [Performance Tuning](./performance-tuning.md)
