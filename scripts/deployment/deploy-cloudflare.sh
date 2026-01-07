#!/bin/bash

# Cloudflare Deployment Script for Obsidian Bot
set -e

echo "🚀 Starting Cloudflare deployment for Obsidian Bot..."

# Check if Wrangler is installed
if ! command -v wrangler &> /dev/null; then
    echo "❌ Wrangler CLI not found. Installing..."
    npm install -g wrangler
fi

# Check environment variables
if [ -z "$CLOUDFLARE_API_TOKEN" ]; then
    echo "❌ CLOUDFLARE_API_TOKEN environment variable is required"
    exit 1
fi

if [ -z "$CLOUDFLARE_ACCOUNT_ID" ]; then
    echo "❌ CLOUDFLARE_ACCOUNT_ID environment variable is required"
    exit 1
fi

echo "✅ Environment variables verified"

# Deploy AI Proxy Worker
echo "📦 Deploying AI Proxy Worker..."
cd workers/ai-proxy
if [ -f "package.json" ]; then
    npm install
fi

wrangler deploy --env production

# Deploy Analytics Worker
echo "📊 Deploying Analytics Worker..."
cd ../analytics
if [ -f "package.json" ]; then
    npm install
fi

wrangler deploy --env production

# Deploy Cache Worker
echo "⚡ Deploying Cache Worker..."
cd ../cache
if [ -f "package.json" ]; then
    npm install
fi

wrangler deploy --env production

cd ../..

# Create R2 bucket if it doesn't exist
echo "📦 Creating R2 storage bucket..."
if ! wrangler r2 bucket list | grep -q "obsidian-bot-media"; then
    wrangler r2 bucket create obsidian-bot-media
    echo "✅ Created R2 bucket: obsidian-bot-media"
else
    echo "✅ R2 bucket already exists"
fi

# Create D1 database if it doesn't exist
echo "🗄️ Creating D1 database..."
if ! wrangler d1 list | grep -q "obsidian-bot-db"; then
    wrangler d1 create obsidian-bot-db
    echo "✅ Created D1 database: obsidian-bot-db"
else
    echo "✅ D1 database already exists"
fi

# Deploy database schema
echo "🔧 Deploying database schema to D1..."
wrangler d1 execute obsidian-bot-db --file=./database/d1_schema.sql

# Create KV namespaces
echo "🗂️ Creating KV namespaces..."
if ! wrangler kv:namespace list | grep -q "AI_CACHE"; then
    wrangler kv:namespace create "AI_CACHE"
    echo "✅ Created KV namespace: AI_CACHE"
else
    echo "✅ KV namespace already exists"
fi

if ! wrangler kv:namespace list | grep -q "ANALYTICS"; then
    wrangler kv:namespace create "ANALYTICS"
    echo "✅ Created KV namespace: ANALYTICS"
else
    echo "✅ KV namespace already exists"
fi

# Update wrangler.toml with generated IDs
echo "🔧 Updating configuration with generated IDs..."
KV_CACHE_ID=$(wrangler kv:namespace list | grep "AI_CACHE" | jq -r '.id')
KV_ANALYTICS_ID=$(wrangler kv:namespace list | grep "ANALYTICS" | jq -r '.id')
D1_DATABASE_ID=$(wrangler d1 list | grep "obsidian-bot-db" | jq -r '.id')

# Update wrangler.toml with actual IDs
sed -i "s/your-kv-namespace-id/$KV_CACHE_ID/g" workers/wrangler.toml
sed -i "s/your-preview-kv-namespace-id/$KV_CACHE_ID/g" workers/wrangler.toml

echo "🎉 Cloudflare deployment completed!"
echo ""
echo "📋 Next steps:"
echo "1. Update your DNS to point to Cloudflare"
echo "2. Update application configuration to use AI proxy endpoints"
echo "3. Set up environment variables for R2 and D1"
echo "4. Test the integration"
echo ""
echo "🔗 Worker URLs:"
echo "AI Proxy: https://obsidian-bot-workers-prod.your-subdomain.workers.dev/ai/proxy"
echo "Analytics: https://obsidian-bot-workers-prod.your-subdomain.workers.dev/analytics"
echo ""
echo "📊 KV Namespace IDs:"
echo "AI Cache: $KV_CACHE_ID"
echo "Analytics: $KV_ANALYTICS_ID"
echo "D1 Database: $D1_DATABASE_ID"