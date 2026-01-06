# 🆓 Free SSH Access Alternatives Guide

Complete guide for accessing your bot services without paying for SSH hosting.

## 📋 Table of Contents

1. [Render.com Alternatives](#rendercom-alternatives)
2. [Railway.app Setup](#railway-setup)
3. [Fly.io Deployment](#flyio-deployment)
4. [Vercel Integration](#vercel-integration)
5. [GitHub Codespaces](#github-codespaces)
6. [Local Port Forwarding](#local-port-forwarding)
7. [Cloud Tunnel Solutions](#cloud-tunnel-solutions)
8. [Comparison Table](#comparison-table)

## 🎯 **Quick Decision Matrix**

| Need | Best Free Option | Setup Time | Features |
|------|-----------------|-------------|---------|
| **SSH Access** | GitHub Codespaces | 5 minutes | Full dev environment |
| **Static Deploy** | Vercel | 10 minutes | Serverless, global |
| **Serverless** | Fly.io | 15 minutes | Edge computing, built-in DB |
| **Container** | Railway | 10 minutes | Managed containers |
| **Full Control** | Local Development | 0 minutes | Maximum flexibility |

## 🌐 Render.com Alternatives

### 🚀 **Option 1: Use Render.com Free Tier (Best for SSH)**

#### Setup Steps:
```bash
# 1. Get your Render.com account at https://render.com
# 2. Deploy your bot using the Railway approach (see below)
# 3. You'll get SSH access via Render dashboard
```

#### Railway Deployment Method:
```bash
# Install Railway CLI
npm install -g @railway/cli

# Deploy to Railway (free tier)
railway deploy \
  --service web \
  --name obsidian-bot \
  --buildpack static \
  --cwd . \
  --env="TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN" \
  --env="TURSO_DATABASE_URL=$TURSO_DATABASE_URL" \
  --env="TURSO_AUTH_TOKEN=$TURSO_AUTH_TOKEN"
```

#### SSH Access to Railway:
```bash
# Find your deployed service URL
railway status

# Access SSH (enabled in dashboard or use Railway CLI)
railway ssh obsidian-bot
```

### 🚀 **Option 2: Use Render Web Service (No SSH)**
```bash
# Deploy static version to Render
railway deploy \
  --service static \
  --name obsidian-bot-static \
  --dist ./build
```

## 🚂 Railway.app Setup

### **Free Tier Features:**
- ✅ **500 hours/month** of runtime
- ✅ **3 GB RAM** per instance
- ✅ **Unlimited builds**  
- ✅ **Environment variables**
- ✅ **Database storage** (Redis)
- ✅ **Built-in CI/CD**

### **Deploy Command:**
```bash
# Install Railway CLI
npm install -g @railway/cli

# Login to Railway
railway login

# Deploy your bot
railway deploy \
  --service web \
  --name obsidian-bot \
  --buildpack static \
  --env="TELEGRAM_BOT_TOKEN=$TELEGRAM_BOT_TOKEN" \
  --env="TURSO_DATABASE_URL=$TURSO_DATABASE_URL" \
  --env="TURSO_AUTH_TOKEN=$TURSO_AUTH_TOKEN"
```

### **Access Methods:**
```bash
# SSH into Railway instance
railway ssh obsidian-bot

# Check logs
railway logs obsidian-bot

# Restart service
railway restart obsidian-bot

# Check status
railway status
```

## 🚁 Fly.io Deployment

### **Free Tier Features:**
- ✅ **160 hours** of CPU time/month
- ✅ **3 shared CPUs** (equivalent to ~1.5 full CPU)
- ✅ **256 MB RAM** per app
- ✅ **Up to 10 apps**
- ✅ **Free SSL certificates**
- ✅ **Built-in Postgres**
- ✅ **Global CDN**

### **Deploy Command:**
```bash
# Install Fly CLI
curl -L https://fly.io/install.sh | sh

# Login to Fly
fly auth login

# Create and deploy app
fly launch obsidian-bot \
  --image obsidian-bot \
  --region ewr \
  --vm-size shared-cpu-1x \
  --env TELEGRAM_BOT_TOKEN \
  --env TURSO_DATABASE_URL \
  --env TURSO_AUTH_TOKEN
```

### **SSH Access to Fly.io:**
```bash
# Get app instance
fly apps list

# SSH into your app instance
ssh root@obsidian-bots.internal
```

### **Database Access (Free Postgres):**
```bash
# Connect to built-in Postgres
fly postgres connect obsidian-bot

# Access via external tool
psql postgresql://user:password@obsidian-bots.internal:5432/obsidian_bot
```

## 🌍 Vercel Integration

### **Serverless Deployment:**
```bash
# Install Vercel CLI
npm install -g vercel

# Deploy your bot (serverless)
vercel --prod \
  --name obsidian-bot \
  --env TELEGRAM_BOT_TOKEN \
  --env TURSO_DATABASE_URL \
  --env TURSO_AUTH_TOKEN
```

### **HTTP API Endpoint:**
```bash
# Your bot will be accessible at:
# https://obsidian-bot.vercel.app

# Test the deployment
curl -X POST https://obsidian-bot.vercel.app/api \
  -H "Content-Type: application/json" \
  -d '{"message": {"text": "Hello from Vercel!"}}'
```

## 🐙 GitHub Codespaces (Best for SSH)

### **Free Tier Features:**
- ✅ **60 hours/month** of compute time
- ✅ **2 CPU cores** per Codespace
- ✅ **4 GB RAM** per Codespace
- ✅ **10 GB storage** per Codespace
- ✅ **Pre-installed tools** (git, docker, etc.)
- ✅ **Full SSH access** via web terminal or VS Code
- ✅ **Persistent storage** between sessions

### **Quick Setup:**
```bash
# Use our automated script
./scripts/github-codespaces.sh create

# Or manual setup
gh codespace create \
  --repo Ablay19/obsidian-vault \
  --machine standardLinux_x64 \
  --idle-timeout 120 \
  --cpu 2 \
  --memory 4gb

# SSH into Codespace
gh codespace ssh <codespace-name>
```

### **Development Workflow in Codespace:**
```bash
# Once SSH'd into Codespace:
git clone https://github.com/Ablay19/obsidian-vault.git
cd obsidian-vault

# Set up environment
export TELEGRAM_BOT_TOKEN="your-token"
export TURSO_DATABASE_URL="your-db-url"
export TURSO_AUTH_TOKEN="your-auth-token"

# Deploy locally for testing
./docker-deploy.sh development

# Test locally
curl http://localhost:8080/api/services/status
```

## 🔧 Local Port Forwarding

### **If you have local deployment:**
```bash
# Access your local bot from anywhere
ssh -R 8080:localhost:8080 your-user@your-server

# Or use SSH tunnel for easier access
sshuttle -r user@your-server 8080
```

### **Cloudflare Tunnel (Free):**
```bash
# Install cloudflared
npm install -g cloudflared

# Create tunnel to your local bot
cloudflared tunnel --url http://localhost:8080

# Or expose to public URL
cloudflared tunnel --url http://localhost:8080 --hostname obsidian-bot
```

## 🔍 Cloud Tunnel Solutions

### **Ngrok (Free Tier)**
```bash
# Download and run ngrok
# 1. Download: https://ngrok.com/download
# 2. Run: ./ngrok http 8080
# 3. Your bot will be accessible at: https://random.ngrok.io
```

### **Cloudflare Tunnel (Free Tier)**
```bash
# Use built-in cloudflared from Workers SDK
# More reliable than ngrok and integrated with Cloudflare

cloudflared tunnel --url http://localhost:8080 --hostname obsidian-bot
```

### **LocalTunnel (Free)**
```bash
# Install Localtunnel
npm install -g localtunnel

# Create tunnel
localtunnel --port 8080 --subdomain obsidian-bot
```

## 📊 Comparison Table

| Service | Free Hours | CPU | RAM | Storage | SSH Access | Setup Time | Best For |
|----------|-------------|-----|-----|----------|-------------|-----------|
| **GitHub Codespaces** | 60/mo | 2 cores | 4 GB | ✅ Yes | 5 min | SSH Access |
| **Fly.io** | 160/mo | 1.5 cores | 256 MB | ❌ No | 15 min | Edge Computing |
| **Railway** | 500/mo | 1 core | 3 GB | ✅ Yes | 10 min | App Deployment |
| **Render.com** | 750/mo | ? | ? | ✅ Yes | 10 min | Web Services |
| **Vercel** | ∞ Serverless | ∞ Compute | ? | ❌ No | 10 min | Static Sites |
| **Ngrok** | ∞ Unlimited | ∞ Compute | 0 | ❌ No | 5 min | Local Tunnels |
| **Cloudflare Tunnel** | ∞ Unlimited | ∞ Compute | 0 | ❌ No | 5 min | Local Tunnels |

## 🎯 **Recommended Solutions**

### **For SSH Access (Best Choice):**
```bash
# Option 1: GitHub Codespaces (Recommended for development)
./scripts/github-codespaces.sh create

# Option 2: Railway (Good alternative)
railway deploy --service web --name obsidian-bot
```

### **For Serverless Deployment:**
```bash
# Option 1: Vercel (Easiest)
vercel --prod

# Option 2: Fly.io (More features)
fly launch obsidian-bot
```

### **For Quick Testing:**
```bash
# Option 1: Ngrok (Simplest)
ngrok http 8080

# Option 2: Cloudflare Tunnel (Most reliable)
cloudflared tunnel --url http://localhost:8080
```

## 🛠️ **Pro Tips for Free Services**

### **GitHub Codespaces Pro Tips:**
```bash
# Maximize your free hours
gh codespace config set idle-timeout 30  # Auto-stop after 30 min idle

# Persistent storage
gh codespace config set retain-period 7d  # Keep files for 7 days

# Better performance
gh codespace config set machine standardLinux_x64  # More CPU/RAM
```

### **Railway Money Saving Tips:**
```bash
# Use environment variables to avoid rebuilds
railway variables set TELEGRAM_BOT_TOKEN "your-token"

# Automatic deploys
git push origin main  # Auto-deploys via GitHub integration
```

### **Fly.io Optimization Tips:**
```bash
# Use volume for persistence
fly volumes create obsidian-data

# Scale efficiently
fly scale count 2  # Use both free CPU hours

# Monitor usage
fly hours --org personal
```

## 🎉 **Bottom Line**

**You have multiple excellent free options for SSH access:**

### 🥇 **Winner: GitHub Codespaces**
- **Best SSH access** with web terminal
- **Generous free tier** with 60 hours/month
- **Pre-installed development tools**
- **Integration with GitHub Actions**
- **Persistent storage between sessions**

### 🥈 **Runner-Up: Railway**
- **Container-based deployment** with SSH
- **Generous free tier** with 500 hours/month
- **Built-in CI/CD**
- **Database and storage included**

### 🌍 **Serverless Champion: Vercel**
- **Completely free** serverless hosting
- **Global CDN** distribution
- **Custom domains** support
- **Instant rollbacks**

---

## 🚀 **Getting Started Immediately**

### **For SSH Access Now:**
```bash
# Copy this script and run:
./scripts/github-codespaces.sh create

# Then set your environment:
gh codespace ssh obsidian-bot-codespace
cd obsidian-vault
nano .env  # Add your tokens
```

### **For Serverless Now:**
```bash
# Deploy to Vercel (easiest):
npm install -g vercel
vercel --prod

# Test deployment:
curl -X POST https://obsidian-bot-<random>.vercel.app/api \
  -H "Content-Type: application/json" \
  -d '{"message": {"text": "Hello from free deployment!"}}'
```

---

**No more excuses! You have everything you need to access your bot for FREE! 🎉**