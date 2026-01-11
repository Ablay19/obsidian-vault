#!/bin/bash

# Deployment Readiness Check for Obsidian Bot
# This script verifies that all required environment variables and configurations are set

set -e

echo "🚀 Obsidian Bot Deployment Readiness Check"
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check environment variable
check_env_var() {
    local var_name="$1"
    local description="$2"
    local required="${3:-true}"

    if [ -n "${!var_name}" ]; then
        echo -e "✅ ${var_name}: ${GREEN}Set${NC} (${description})"
        return 0
    else
        if [ "$required" = "true" ]; then
            echo -e "❌ ${var_name}: ${RED}Missing${NC} (${description}) - REQUIRED"
            return 1
        else
            echo -e "⚠️  ${var_name}: ${YELLOW}Not set${NC} (${description}) - Optional"
            return 0
        fi
    fi
}

# Function to check file existence
check_file() {
    local file_path="$1"
    local description="$2"
    local required="${3:-true}"

    if [ -f "$file_path" ]; then
        echo -e "✅ ${file_path}: ${GREEN}Exists${NC} (${description})"
        return 0
    else
        if [ "$required" = "true" ]; then
            echo -e "❌ ${file_path}: ${RED}Missing${NC} (${description}) - REQUIRED"
            return 1
        else
            echo -e "⚠️  ${file_path}: ${YELLOW}Not found${NC} (${description}) - Optional"
            return 0
        fi
    fi
}

# Function to check directory
check_directory() {
    local dir_path="$1"
    local description="$2"
    local required="${3:-true}"

    if [ -d "$dir_path" ]; then
        echo -e "✅ ${dir_path}: ${GREEN}Exists${NC} (${description})"
        return 0
    else
        if [ "$required" = "true" ]; then
            echo -e "❌ ${dir_path}: ${RED}Missing${NC} (${description}) - REQUIRED"
            return 1
        else
            echo -e "⚠️  ${dir_path}: ${YELLOW}Not found${NC} (${description}) - Optional"
            return 0
        fi
    fi
}

echo ""
echo "🔧 Required Environment Variables:"
echo "-----------------------------------"

required_vars_passed=0
total_required=0

# Required variables
check_env_var "TELEGRAM_BOT_TOKEN" "Telegram bot API token" && ((required_vars_passed++))
((total_required++))
check_env_var "TURSO_DATABASE_URL" "Turso database connection URL" && ((required_vars_passed++))
((total_required++))

echo ""
echo "🤖 AI Provider API Keys (at least one required):"
echo "------------------------------------------------"

ai_providers_configured=0

# AI Provider keys
check_env_var "GEMINI_API_KEY" "Google Gemini API key" false && ((ai_providers_configured++))
check_env_var "DEEPSEEK_API_KEY" "DeepSeek API key" false && ((ai_providers_configured++))
check_env_var "GROQ_API_KEY" "Groq API key" false && ((ai_providers_configured++))
check_env_var "OPENROUTER_API_KEY" "OpenRouter API key" false && ((ai_providers_configured++))
check_env_var "HUGGINGFACE_API_KEY" "HuggingFace API key" false && ((ai_providers_configured++))
check_env_var "CLOUDFLARE_API_TOKEN" "Cloudflare API token" false && ((ai_providers_configured++))

if [ $ai_providers_configured -eq 0 ]; then
    echo -e "${RED}❌ No AI providers configured - at least one is required${NC}"
    ((required_vars_passed--))
    ((total_required++))
else
    echo -e "✅ AI Providers: ${GREEN}$ai_providers_configured configured${NC}"
fi

echo ""
echo "🔐 Optional Security & Authentication:"
echo "--------------------------------------"

# Optional variables
check_env_var "SESSION_SECRET" "Session encryption secret" false
check_env_var "TURSO_AUTH_TOKEN" "Turso authentication token" false
check_env_var "REDIS_ADDR" "Redis server address" false

echo ""
echo "📁 Required Files & Directories:"
echo "---------------------------------"

# Required files
check_file ".env" "Environment configuration file"
check_file "config.yaml" "Application configuration file"

# Required directories
check_directory "vault" "Obsidian vault directory" false
check_directory "attachments" "File attachments directory" false
check_directory ".data" "Application data directory" false

echo ""
echo "🛠️  System Dependencies:"
echo "-----------------------"

# Check system dependencies
dependencies=("tesseract" "pdftotext" "convert" "pandoc")
missing_deps=0

for dep in "${dependencies[@]}"; do
    if command -v "$dep" &> /dev/null; then
        echo -e "✅ ${dep}: ${GREEN}Available${NC}"
    else
        echo -e "❌ ${dep}: ${RED}Not found${NC}"
        ((missing_deps++))
    fi
done

echo ""
echo "📊 Deployment Readiness Summary:"
echo "================================="

if [ $required_vars_passed -eq $total_required ] && [ $missing_deps -eq 0 ]; then
    echo -e "${GREEN}🎉 DEPLOYMENT READY!${NC}"
    echo "All required components are configured and available."
    echo ""
    echo "🚀 To start the bot:"
    echo "   GIN_MODE=release ./bot"
    echo ""
    echo "📋 Next steps:"
    echo "   1. Ensure database migrations are run"
    echo "   2. Test bot commands (/help, /setprovider)"
    echo "   3. Send test images for OCR processing"
    exit 0
else
    echo -e "${RED}❌ DEPLOYMENT NOT READY${NC}"
    echo "Issues found that need to be resolved:"
    echo "  - Required variables: $required_vars_passed/$total_required configured"
    echo "  - Missing dependencies: $missing_deps"
    echo ""
    echo "🔧 Fix the issues above and re-run this check."
    exit 1
fi