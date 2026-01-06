#!/bin/bash

echo "🚀 Deploying Simple Cloudflare AI Worker..."

# Check if wrangler is installed
if ! command -v wrangler &> /dev/null; then
    echo "❌ Wrangler CLI not found. Installing..."
    npm install -g wrangler
fi

# Authenticate if needed
if ! wrangler whoami &> /dev/null; then
    echo "🔐 Please authenticate with Cloudflare:"
    wrangler auth login
fi

# Deploy simple worker
echo "📦 Deploying simple AI worker..."
cp workers/simple-cloudflare-ai.js workers/index.js
cd workers

# Create minimal wrangler.toml for AI
cat > wrangler-simple.toml << 'EOF'
name = "obsidian-bot-ai"
main = "index.js"
compatibility_date = "2024-01-01"

[ai]
binding = "AI"

[vars]
ENVIRONMENT = "production"
EOF

# Deploy
wrangler deploy --config wrangler-simple.toml

echo "✅ Simple AI worker deployed!"
echo ""
echo "🔗 Worker URL: https://obsidian-bot-ai.your-subdomain.workers.dev"
echo ""
echo "📋 Update your .env file:"
echo "CLOUDFLARE_WORKER_URL=https://obsidian-bot-ai.your-subdomain.workers.dev"
echo ""
echo "🧪 Test with:"
echo "curl https://obsidian-bot-ai.your-subdomain.workers.dev/health"