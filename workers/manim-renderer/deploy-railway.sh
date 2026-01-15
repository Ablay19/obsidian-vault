#!/bin/bash
# Manim Renderer Deployment Script for Railway

set -e

echo "🚀 Deploying Manim Renderer to Railway..."

# Check if railway CLI is installed
if ! command -v railway &> /dev/null; then
    echo "📦 Installing Railway CLI..."
    npm install -g @railway/cli
fi

# Login to Railway (requires interactive login)
echo "🔐 Please login to Railway (if not already logged in):"
 railway login

# Initialize Railway project (or link to existing)
if [ ! -f "railway.json" ]; then
    echo "📋 Creating Railway project..."
    railway init
fi

# Deploy to Railway
echo "🚀 Deploying to Railway..."
railway up

# Get the deployed URL
echo "🔗 Getting deployment URL..."
RAILWAY_URL=$(railway domain 2>/dev/null || echo "Deployment in progress")

echo "✅ Manim Renderer deployed successfully!"
echo "📝 Update your worker configuration with:"
echo "   MANIM_RENDERER_URL=$RAILWAY_URL"
