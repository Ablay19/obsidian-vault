# Google Cloud Quick Start Guide

Get your Obsidian Bot running with Google Cloud logging in minutes!

## 🚀 5-Minute Quick Setup

### 1. Install Google Cloud CLI
```bash
# Ubuntu/Debian
curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | sudo apt-key add -
echo "deb https://packages.cloud.google.com/apt cloud-sdk main" | sudo tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
sudo apt-get update && sudo apt-get install -y google-cloud-sdk

# macOS
brew install google-cloud-sdk
```

### 2. Authenticate and Create Project
```bash
# Login to Google Cloud
gcloud auth login
gcloud auth application-default login

# Create project (replace with your name)
gcloud projects create obsidian-bot-prod --name="Obsidian Bot Production"
gcloud config set project obsidian-bot-prod

# Enable required APIs
gcloud services enable logging.googleapis.com cloudresourcemanager.googleapis.com
```

### 3. Create Service Account
```bash
# Create service account
gcloud iam service-accounts create obsidian-bot-sa \
    --display-name="Obsidian Bot Service Account"

# Get service account email
SA_EMAIL=$(gcloud iam service-accounts list \
    --filter="displayName:Obsidian Bot Service Account" \
    --format="value(email)")

# Grant logging permissions
gcloud projects add-iam-policy-binding obsidian-bot-prod \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/logging.logWriter"

# Create and download key
gcloud iam service-accounts keys create ~/obsidian-bot-key.json \
    --iam-account="$SA_EMAIL"

echo "✅ Service account key created: ~/obsidian-bot-key.json"
```

### 4. Configure Environment
```bash
# Set up environment variables
export GOOGLE_CLOUD_PROJECT="obsidian-bot-prod"
export GOOGLE_APPLICATION_CREDENTIALS="$HOME/obsidian-bot-key.json"
export ENABLE_GOOGLE_LOGGING="true"

# Add to shell profile
echo "export GOOGLE_CLOUD_PROJECT=\"obsidian-bot-prod\"" >> ~/.bashrc
echo "export GOOGLE_APPLICATION_CREDENTIALS=\"$HOME/obsidian-bot-key.json\"" >> ~/.bashrc
echo "export ENABLE_GOOGLE_LOGGING=\"true\"" >> ~/.bashrc

# Configure for Obsidian Bot
echo "GOOGLE_CLOUD_PROJECT=obsidian-bot-prod" >> .env
echo "GOOGLE_APPLICATION_CREDENTIALS=$HOME/obsidian-bot-key.json" >> .env
echo "ENABLE_GOOGLE_LOGGING=true" >> .env
```

### 5. Deploy and Test
```bash
# Test Google Cloud logging
./test-google-cloud-logging.sh

# Deploy with Docker
./docker-deploy.sh production

# Or run locally
./bot
```

## ✅ Verification Steps

### Check Google Cloud Setup
```bash
# Verify project
gcloud config list project

# Verify service account
gcloud auth test-iam-permissions obsidian-bot-prod \
    --iam-account=obsidian-bot-sa@obsidian-bot-prod.iam.gserviceaccount.com

# Check logs
gcloud logging tail "resource.type=container" --format=json
```

### Check Application
```bash
# Test health endpoint
curl http://localhost:8080/api/services/status

# Check logs locally
tail -f bot.log | grep Google
```

## 📊 Monitor Logs in Cloud Console

### View Logs
1. Go to [Google Cloud Console](https://console.cloud.google.com)
2. Navigate to "Logging" → "Log Explorer"
3. Filter by: `resource.type="container"`
4. Look for `obsidian-bot-prod` logs

### Create Log-Based Metrics
```bash
# Error count metric
gcloud logging metrics create obsidian_errors \
    --description="Obsidian Bot Errors" \
    --filter='severity="ERROR"'

# Request count metric
gcloud logging metrics create obsidian_requests \
    --description="Obsidian Bot Requests" \
    --filter='jsonPayload.method'
```

## 🔧 Common Issues and Solutions

### Issue: `PERMISSION_DENIED`
```bash
# Solution: Grant proper permissions
gcloud projects add-iam-policy-binding obsidian-bot-prod \
    --member="serviceAccount:$SA_EMAIL" \
    --role="roles/logging.logWriter"
```

### Issue: `INVALID_ARGUMENT`
```bash
# Solution: Check project ID and credentials
gcloud config set project obsidian-bot-prod
export GOOGLE_APPLICATION_CREDENTIALS="$HOME/obsidian-bot-key.json"
```

### Issue: No logs appearing
```bash
# Check application is using Google Cloud logging
grep -i google bot.log

# Check service account permissions
gcloud projects get-iam-policy obsidian-bot-prod | grep obsidian-bot-sa
```

## 🎯 Production Deployment

### Docker with Google Cloud
```bash
# Set production variables
export GOOGLE_CLOUD_PROJECT="obsidian-bot-prod"
export GOOGLE_APPLICATION_CREDENTIALS_PATH="./obsidian-bot-key.json"
export ENABLE_GOOGLE_LOGGING="true"

# Deploy
./docker-deploy.sh production
```

### Environment Variables for Production
```bash
# Add to .env file
GOOGLE_CLOUD_PROJECT=obsidian-bot-prod
GOOGLE_APPLICATION_CREDENTIALS=/app/credentials.json
ENABLE_GOOGLE_LOGGING=true
LOG_LEVEL=INFO
```

## 💡 Pro Tips

### 1. Use Specific Log Levels
```bash
# Production: Only errors and warnings
LOG_LEVEL=WARN

# Development: Include debug logs
LOG_LEVEL=DEBUG
```

### 2. Create Custom Log Names
```bash
# In your application
export GOOGLE_LOG_NAME="obsidian-bot-production"

# Filter in Cloud Console
logName="projects/obsidian-bot-prod/logs/obsidian-bot-production"
```

### 3. Use Log-Based Alerts
```bash
# Create alert for errors
gcloud alpha monitoring policies create \
    --notification-channels=projects/obsidian-bot-prod/notificationChannels/123 \
    --condition-filter='severity="ERROR"' \
    --condition-threshold-value=1 \
    --condition-threshold-duration=60s \
    --display-name="Obsidian Bot Errors"
```

## 📚 Next Steps

1. **Read Complete Guide**: See [GOOGLE_CLOUD_SETUP.md](GOOGLE_CLOUD_SETUP.md) for detailed setup
2. **Monitor**: Set up dashboards and alerts
3. **Optimize**: Configure log routing and cost management
4. **Secure**: Implement secret management and security best practices

## 🆘 Need Help?

- **Google Cloud Documentation**: https://cloud.google.com/logging/docs
- **Obsidian Bot Issues**: Create GitHub issue with logs
- **Quick Commands**: Use the `setup-google-cloud.sh` script for guided setup

---

**Your Obsidian Bot is now ready for production with Google Cloud logging! 🎉**