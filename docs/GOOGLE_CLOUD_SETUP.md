# Google Cloud Setup Guide

This comprehensive guide covers setting up Google Cloud services for the Obsidian Bot, including Cloud Logging, authentication, and production deployment.

## 📋 Table of Contents

1. [Prerequisites](#prerequisites)
2. [Google Cloud Project Setup](#google-cloud-project-setup)
3. [Service Account Configuration](#service-account-configuration)
4. [Cloud Logging Setup](#cloud-logging-setup)
5. [Environment Configuration](#environment-configuration)
6. [Docker Deployment with Google Cloud](#docker-deployment-with-google-cloud)
7. [Monitoring and Troubleshooting](#monitoring-and-troubleshooting)
8. [Security Best Practices](#security-best-practices)
9. [Cost Optimization](#cost-optimization)

## 🚀 Prerequisites

### Required Tools
- **Google Cloud CLI** (`gcloud`)
- **Docker** (for container deployment)
- **Git** (for source code management)

### Install Google Cloud CLI

#### Ubuntu/Debian
```bash
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
echo "deb https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
sudo apt-get update && sudo apt-get install -y google-cloud-sdk
```

#### macOS
```bash
brew install google-cloud-sdk
```

#### Windows
Download from: https://cloud.google.com/sdk/docs/install

### Initial Authentication
```bash
# Authenticate with Google Cloud
gcloud auth login

# Set your application default credentials
gcloud auth application-default login

# Verify authentication
gcloud auth list
gcloud config list
```

## 🏗️ Google Cloud Project Setup

### 1. Create a New Project
```bash
# Create project
gcloud projects create obsidian-bot-prod \
    --name="Obsidian Bot Production" \
    --organization=YOUR_ORG_ID

# Set active project
gcloud config set project obsidian-bot-prod
```

Or create via [Google Cloud Console](https://console.cloud.google.com/projectcreate)

### 2. Enable Required APIs
```bash
# Enable Cloud Logging API
gcloud services enable logging.googleapis.com

# Enable Cloud Resource Manager API
gcloud services enable cloudresourcemanager.googleapis.com

# Enable Cloud Monitoring API (optional but recommended)
gcloud services enable monitoring.googleapis.com

# Enable Secret Manager API (for secrets management)
gcloud services enable secretmanager.googleapis.com

# Verify enabled APIs
gcloud services list --enabled --filter="name~logging"
```

### 3. Configure Logging
```bash
# Create log-based metric (optional)
gcloud logging metrics create obsidian_bot_errors \
    --description="Obsidian Bot Error Count" \
    --filter='resource.type="container" AND severity="ERROR"'

# Create log sink for export (optional)
gcloud logging sinks create obsidian-bot-logs \
    bigquery.googleapis.com/projects/obsidian-bot-prod/datasets/obsidian_logs \
    --log-filter='resource.type="container"' \
    --include-children
```

## 🔐 Service Account Configuration

### 1. Create Service Account
```bash
# Create service account
gcloud iam service-accounts create obsidian-bot-sa \
    --display-name="Obsidian Bot Service Account" \
    --description="Service account for Obsidian Bot application"

# Get service account email
SA_EMAIL=$(gcloud iam service-accounts list \
    --filter="displayName:Obsidian Bot Service Account" \
    --format="value(email)")

echo "Service Account Email: $SA_EMAIL"
```

### 2. Grant Required Permissions

#### Basic Logging Permissions
```bash
# Grant Logs Writer role
gcloud projects add-iam-policy-binding obsidian-bot-prod \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/logging.logWriter"

# Grant Monitoring Viewer role (optional)
gcloud projects add-iam-policy-binding obsidian-bot-prod \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/monitoring.viewer"
```

#### Enhanced Permissions (if using monitoring extensively)
```bash
# Grant Metric Writer role
gcloud projects add-iam-policy-binding obsidian-bot-prod \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/monitoring.metricWriter"

# Grant Secret Manager access (if using)
gcloud projects add-iam-policy-binding obsidian-bot-prod \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/secretmanager.secretAccessor"
```

### 3. Create and Download Service Account Key
```bash
# Create service account key
gcloud iam service-accounts keys create ~/obsidian-bot-key.json \
    --iam-account="$SA_EMAIL" \
    --key-format=json

# Set proper permissions
chmod 600 ~/obsidian-bot-key.json

echo "Service account key created: ~/obsidian-bot-key.json"
```

### 4. Set Up Application Default Credentials
```bash
# For local development
export GOOGLE_APPLICATION_CREDENTIALS="$HOME/obsidian-bot-key.json"
echo "export GOOGLE_APPLICATION_CREDENTIALS=\"$HOME/obsidian-bot-key.json\"" >> ~/.bashrc

# For Docker deployment
cp ~/obsidian-bot-key.json ./service-account-key.json
chmod 600 ./service-account-key.json
```

## 📊 Cloud Logging Setup

### 1. Log Structure Configuration
The Obsidian Bot automatically structures logs with the following fields:

```json
{
  "timestamp": "2026-01-06T18:50:03.498963815Z",
  "severity": "INFO",
  "project_id": "obsidian-bot-prod",
  "event": "system_event",
  "level": "INFO",
  "details": {
    "component": "google_logger_test",
    "test": "basic_logging_test"
  }
}
```

### 2. Configure Log Levels
```bash
# In your .env file
LOG_LEVEL=INFO          # DEBUG, INFO, WARN, ERROR
ENABLE_GOOGLE_LOGGING=true
GOOGLE_CLOUD_PROJECT=obsidian-bot-prod
```

### 3. View Logs in Console
```bash
# View all logs
gcloud logging read "resource.type=container" --limit=10 --format=json

# View error logs
gcloud logging read "resource.type=container AND severity=ERROR" --limit=10

# Stream logs in real-time
gcloud logging tail "resource.type=container" --format=json
```

### 4. Create Log-Based Metrics
```bash
# Create error rate metric
gcloud logging metrics create obsidian_bot_error_rate \
    --description="Obsidian Bot Error Rate" \
    --filter='resource.type="container" AND severity="ERROR"' \
    --aggregations-alignment-period=60s \
    --aggregations-per-series-aligner=ALIGN_RATE

# Create request count metric
gcloud logging metrics create obsidian_bot_requests \
    --description="Obsidian Bot HTTP Requests" \
    --filter='resource.type="container" AND jsonPayload.method="GET"' \
    --aggregations-alignment-period=60s
```

## ⚙️ Environment Configuration

### 1. Local Development Setup
```bash
# Create .env.google file
cat > .env.google << EOF
# Google Cloud Configuration
GOOGLE_CLOUD_PROJECT=obsidian-bot-prod
GOOGLE_APPLICATION_CREDENTIALS=$HOME/obsidian-bot-key.json
ENABLE_GOOGLE_LOGGING=true

# Log Configuration
LOG_LEVEL=INFO
GOOGLE_LOG_NAME=obsidian-bot-logs
EOF

# Source the configuration
source .env.google
```

### 2. Production Environment Setup
```bash
# Update main .env file for production
cat >> .env << EOF

# Google Cloud Production Settings
GOOGLE_CLOUD_PROJECT=obsidian-bot-prod
GOOGLE_APPLICATION_CREDENTIALS=/app/credentials.json
ENABLE_GOOGLE_LOGGING=true
LOG_LEVEL=INFO
EOF
```

### 3. Docker Configuration
```yaml
# docker-compose.yml updates
services:
  obsidian-bot:
    environment:
      - GOOGLE_CLOUD_PROJECT=${GOOGLE_CLOUD_PROJECT:-obsidian-bot-prod}
      - GOOGLE_APPLICATION_CREDENTIALS=/app/credentials.json
      - ENABLE_GOOGLE_LOGGING=${ENABLE_GOOGLE_LOGGING:-true}
      - LOG_LEVEL=INFO
    volumes:
      - ./service-account-key.json:/app/credentials.json:ro
```

## 🐳 Docker Deployment with Google Cloud

### 1. Production Docker Deployment
```bash
# Set Google Cloud environment variables
export GOOGLE_CLOUD_PROJECT=obsidian-bot-prod
export GOOGLE_APPLICATION_CREDENTIALS_PATH="./service-account-key.json"
export ENABLE_GOOGLE_LOGGING=true

# Deploy with Docker
./docker-deploy.sh production

# Or manually
docker-compose --profile production up -d
```

### 2. Multi-Environment Setup
```bash
# Development (local logging)
export ENABLE_GOOGLE_LOGGING=false
./docker-deploy.sh development

# Production (Google Cloud logging)
export ENABLE_GOOGLE_LOGGING=true
./docker-deploy.sh production

# Staging (mixed logging)
export LOG_LEVEL=DEBUG
export GOOGLE_CLOUD_PROJECT=obsidian-bot-staging
./docker-deploy.sh staging
```

### 3. Cloud Build Integration
```bash
# Create cloudbuild.yaml
cat > cloudbuild.yaml << EOF
steps:
- name: 'gcr.io/cloud-builders/docker'
  args: ['build', '-t', 'gcr.io/$PROJECT_ID/obsidian-bot:$BUILD_ID', '.']
- name: 'gcr.io/cloud-builders/docker'
  args: ['push', 'gcr.io/$PROJECT_ID/obsidian-bot:$BUILD_ID']
- name: 'gcr.io/cloud-builders/kubectl'
  args: ['set', 'image', 'deployment/obsidian-bot', 'obsidian-bot=gcr.io/$PROJECT_ID/obsidian-bot:$BUILD_ID']
EOF

# Submit build
gcloud builds submit --config cloudbuild.yaml .
```

## 📈 Monitoring and Troubleshooting

### 1. Cloud Monitoring Dashboard
```bash
# Create custom dashboard
gcloud monitoring dashboards create --config-from-file=dashboard.json

# Example dashboard.json
cat > dashboard.json << EOF
{
  "displayName": "Obsidian Bot Metrics",
  "widgets": [
    {
      "title": "Error Rate",
      "xyChart": {
        "dataSets": [{
          "timeSeriesQuery": {
            "prometheusQuery": {
              "query": "rate(obsidian_bot_requests_total[5m])"
            }
          }
        }]
      }
    }
  ]
}
EOF
```

### 2. Alert Policies
```bash
# Create alert for high error rate
gcloud alpha monitoring policies create --policy-from-file=alert-policy.yaml

# Example alert-policy.yaml
cat > alert-policy.yaml << EOF
{
  "displayName": "Obsidian Bot High Error Rate",
  "conditions": [{
    "displayName": "Error rate > 5%",
    "conditionThreshold": {
      "filter": 'metric.type="logging.googleapis.com/user/obsidian_bot_error_rate"',
      "comparison": "COMPARISON_GT",
      "thresholdValue": 0.05
    }
  }],
  "notificationChannels": ["projects/obsidian-bot-prod/notificationChannels/123456789"]
}
EOF
```

### 3. Common Issues and Solutions

#### Authentication Issues
```bash
# Check credentials
gcloud auth application-default print-access-token

# Reset credentials
gcloud auth application-default revoke
gcloud auth application-default login

# Verify service account
gcloud auth test-iam-permissions obsidian-bot-prod \
    --iam-account=obsidian-bot-sa@obsidian-bot-prod.iam.gserviceaccount.com
```

#### Logging Issues
```bash
# Check if logs are being written
gcloud logging read "resource.type=container" --limit=5 --format=json

# Check log permissions
gcloud projects get-iam-policy obsidian-bot-prod \
    --flatten="bindings[].members" \
    --format="table(bindings.role, bindings.members)"
```

#### Permission Issues
```bash
# Grant missing permissions
gcloud projects add-iam-policy-binding obsidian-bot-prod \
    --member="serviceAccount:obsidian-bot-sa@obsidian-bot-prod.iam.gserviceaccount.com" \
    --role="roles/logging.logWriter"

# Force propagate changes
gcloud auth application-default set-quota-project obsidian-bot-prod
```

## 🔒 Security Best Practices

### 1. Service Account Security
```bash
# Use least privilege principle
# Only grant necessary roles (avoid using Editor or Owner)
gcloud projects add-iam-policy-binding obsidian-bot-prod \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/logging.logWriter"  # Specific role only

# Rotate service account keys regularly
gcloud iam service-accounts keys create ~/obsidian-bot-key-$(date +%Y%m).json \
    --iam-account="$SA_EMAIL"

# Delete old keys
gcloud iam service-accounts keys delete KEY_ID \
    --iam-account="$SA_EMAIL"
```

### 2. Secret Management
```bash
# Use Secret Manager for sensitive data
gcloud secrets create telegram-bot-token
echo -n "your-telegram-token" | gcloud secrets versions add telegram-bot-token --data-file=-

# Access secret in application
export TELEGRAM_BOT_TOKEN=$(gcloud secrets versions access latest --secret=telegram-bot-token)
```

### 3. Network Security
```bash
# Enable VPC Service Controls if needed
gcloud access-context-manager policies create \
    --organization=YOUR_ORG_ID \
    --title="Obsidian Bot Policy"

# Configure private Google Cloud access
gcloud services enable private.googleapis.com
gcloud compute addresses create obsidian-bot-private \
    --region=us-central1 \
    --subnet=default \
    --addresses=10.128.0.5
```

## 💰 Cost Optimization

### 1. Logging Costs
```bash
# Use log routing to reduce costs
gcloud logging sinks create obsidian-bot-cost-optimized \
    bigquery.googleapis.com/projects/obsidian-bot-prod/datasets/obsidian_logs_optimized \
    --log-filter='resource.type="container" AND severity>=WARNING' \
    --include-children

# Set log exclusions
gcloud logging log-exclusions create debug-logs \
    --description="Exclude DEBUG level logs" \
    --filter='severity="DEBUG"'
```

### 2. Monitoring Costs
```bash
# Use cost-effective monitoring
gcloud monitoring dashboards create --config-from-file=cost-optimized-dashboard.json

# Create alert policies only for critical metrics
gcloud alpha monitoring policies create --policy-from-file=critical-alerts.yaml
```

### 3. Resource Optimization
```bash
# Use appropriate machine types
gcloud container clusters create obsidian-bot-cluster \
    --num-nodes=1 \
    --machine-type=e2-small \
    --region=us-central1

# Enable autoscaling
gcloud container clusters update obsidian-bot-cluster \
    --enable-autoscaling \
    --min-nodes=1 \
    --max-nodes=3
```

## 📚 Additional Resources

### Documentation Links
- [Google Cloud Logging Documentation](https://cloud.google.com/logging/docs)
- [Service Account Best Practices](https://cloud.google.com/iam/docs/best-practices-for-service-accounts)
- [Cloud Monitoring Guide](https://cloud.google.com/monitoring/docs)
- [Google Cloud Security Best Practices](https://cloud.google.com/security/best-practices)

### Quick Reference Commands
```bash
# Quick status check
gcloud config list project
gcloud auth list
gcloud services list --enabled

# Common logging commands
gcloud logging tail "resource.type=container"
gcloud logging read "severity=ERROR" --limit=10
gcloud logging metrics list

# Service account management
gcloud iam service-accounts list
gcloud projects get-iam-policy $(gcloud config get-value project)
```

## 🆘 Support and Troubleshooting

### Common Error Messages

#### `PERMISSION_DENIED`
```bash
# Solution: Grant proper IAM roles
gcloud projects add-iam-policy-binding $PROJECT_ID \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/logging.logWriter"
```

#### `INVALID_ARGUMENT`
```bash
# Solution: Verify project ID and credentials
gcloud config set project $PROJECT_ID
gcloud auth application-default login
```

#### `RESOURCE_EXHAUSTED`
```bash
# Solution: Check quotas and request increase
gcloud project describe $PROJECT_ID --format="table(quotas)"
# Request quota increase via: https://console.cloud.google.com/iam-admin/quotas
```

### Getting Help
- **Google Cloud Support**: https://cloud.google.com/support
- **Stack Overflow**: Use tags `google-cloud-platform` and `google-cloud-logging`
- **GitHub Issues**: Create issue in repository with detailed logs

---

**This guide provides comprehensive setup instructions for Google Cloud integration with the Obsidian Bot. Follow each section carefully for a successful production deployment.**