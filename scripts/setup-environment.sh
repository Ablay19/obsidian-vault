#!/bin/bash

# Automatic Environment Setup for Obsidian Bot
# This script creates all necessary environment variables and configurations

set -e

echo "🔧 Automatic Environment Setup for Obsidian Bot"
echo "================================================"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Function to prompt for input with default
prompt_with_default() {
    local prompt="$1"
    local default="$2"
    local var_name="$3"

    echo -ne "${BLUE}${prompt}${NC} [${default}]: "
    read -r input

    if [ -z "$input" ]; then
        input="$default"
    fi

    eval "$var_name=\"$input\""
}

# Function to generate a secure random string
generate_secret() {
    openssl rand -hex 32
}

echo ""
echo "🤖 Telegram Bot Configuration:"
echo "-----------------------------"

# Telegram bot token
if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    echo "You'll need to get a bot token from @BotFather on Telegram"
    echo "1. Message @BotFather with /newbot"
    echo "2. Follow the instructions to create your bot"
    echo "3. Copy the API token here"
    echo ""
    prompt_with_default "Enter your Telegram Bot Token" "" TELEGRAM_BOT_TOKEN
else
    echo "✅ TELEGRAM_BOT_TOKEN: Already set"
fi

echo ""
echo "🗄️  Database Configuration:"
echo "-------------------------"

# Database configuration
if [ -z "$TURSO_DATABASE_URL" ]; then
    echo "For production, you should use Turso (https://turso.tech)"
    echo "For development/testing, we can use a local SQLite file"
    echo ""
    read -p "Do you want to use Turso database? (y/n) [n]: " use_turso
    case "${use_turso:-n}" in
        [Yy]*)
            echo "Create a database at https://turso.tech"
            echo "Then provide the database URL and auth token below:"
            echo ""
            prompt_with_default "Turso Database URL" "" TURSO_DATABASE_URL
            prompt_with_default "Turso Auth Token" "" TURSO_AUTH_TOKEN
            ;;
        *)
            TURSO_DATABASE_URL=".data/local/test.db"
            echo "✅ Using local SQLite database: $TURSO_DATABASE_URL"
            ;;
    esac
else
    echo "✅ TURSO_DATABASE_URL: Already set"
fi

echo ""
echo "🤖 AI Provider Configuration:"
echo "----------------------------"

# AI Providers
ai_configured=0

if [ -z "$GEMINI_API_KEY" ]; then
    echo "Google Gemini is FREE and works great for this bot!"
    echo "Get your API key from: https://makersuite.google.com/app/apikey"
    echo ""
    read -p "Do you want to configure Google Gemini? (y/n) [y]: " use_gemini
    case "${use_gemini:-y}" in
        [Yy]*)
            prompt_with_default "Google Gemini API Key" "" GEMINI_API_KEY
            ((ai_configured++))
            ;;
    esac
else
    echo "✅ GEMINI_API_KEY: Already configured"
    ((ai_configured++))
fi

if [ -z "$DEEPSEEK_API_KEY" ]; then
    echo "DeepSeek is FREE and excellent for image processing!"
    echo "Get your API key from: https://platform.deepseek.com/"
    echo ""
    read -p "Do you want to configure DeepSeek? (y/n) [y]: " use_deepseek
    case "${use_deepseek:-y}" in
        [Yy]*)
            prompt_with_default "DeepSeek API Key" "" DEEPSEEK_API_KEY
            ((ai_configured++))
            ;;
    esac
else
    echo "✅ DEEPSEEK_API_KEY: Already configured"
    ((ai_configured++))
fi

# Other AI providers (optional)
providers=("GROQ_API_KEY" "OPENROUTER_API_KEY" "HUGGINGFACE_API_KEY")
provider_names=("Groq" "OpenRouter" "HuggingFace")
provider_urls=("https://console.groq.com/keys" "https://openrouter.ai/keys" "https://huggingface.co/settings/tokens")

for i in "${!providers[@]}"; do
    var_name="${providers[$i]}"
    provider_name="${provider_names[$i]}"
    provider_url="${provider_urls[$i]}"

    if [ -z "${!var_name}" ]; then
        read -p "Configure $provider_name AI provider? (y/n) [n]: " configure_provider
        case "${configure_provider:-n}" in
            [Yy]*)
                echo "Get your API key from: $provider_url"
                prompt_with_default "$provider_name API Key" "" "$var_name"
                ((ai_configured++))
                ;;
        esac
    else
        echo "✅ $var_name: Already configured"
        ((ai_configured++))
    fi
done

if [ $ai_configured -eq 0 ]; then
    echo ""
    echo -e "${RED}❌ No AI providers configured!${NC}"
    echo "You need at least one AI provider. Google Gemini is recommended as it's free."
    echo "Please configure at least one AI provider above."
    exit 1
fi

echo ""
echo "🔐 Security Configuration:"
echo "-------------------------"

# Generate session secret if not set
if [ -z "$SESSION_SECRET" ]; then
    SESSION_SECRET=$(generate_secret)
    echo "✅ Generated SESSION_SECRET"
else
    echo "✅ SESSION_SECRET: Already set"
fi

echo ""
echo "⚙️  Optional Configuration:"
echo "-------------------------"

# Redis (optional)
if [ -z "$REDIS_ADDR" ]; then
    read -p "Configure Redis for caching? (y/n) [n]: " use_redis
    case "${use_redis:-n}" in
        [Yy]*)
            prompt_with_default "Redis Address" "localhost:6379" REDIS_ADDR
            ;;
    esac
fi

echo ""
echo "📝 Creating .env file..."
echo "========================"

# Create .env file
cat > .env << EOF
# Telegram Bot Configuration
TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN}

# Database Configuration
TURSO_DATABASE_URL=${TURSO_DATABASE_URL}
TURSO_AUTH_TOKEN=${TURSO_AUTH_TOKEN}

# AI Provider API Keys
GEMINI_API_KEY=${GEMINI_API_KEY}
DEEPSEEK_API_KEY=${DEEPSEEK_API_KEY}
GROQ_API_KEY=${GROQ_API_KEY}
OPENROUTER_API_KEY=${OPENROUTER_API_KEY}
HUGGINGFACE_API_KEY=${HUGGINGFACE_API_KEY}

# Security
SESSION_SECRET=${SESSION_SECRET}

# Optional Services
REDIS_ADDR=${REDIS_ADDR}

# Development Settings
GIN_MODE=release
ENABLE_COLORFUL_LOGS=true
EOF

echo "✅ .env file created successfully"

echo ""
echo "🔧 Creating required directories..."
echo "==================================="

# Create required directories
mkdir -p vault/Inbox
mkdir -p attachments
mkdir -p .data/local
mkdir -p pdfs

echo "✅ Directories created"

echo ""
echo "🧪 Creating debug templates..."
echo "=============================="

# Run the debug template setup
if [ -f "debug/setup_templates.sh" ]; then
    ./debug/setup_templates.sh > /dev/null 2>&1
    echo "✅ Debug templates created"
else
    echo "⚠️  Debug template setup script not found"
fi

echo ""
echo "📊 Configuration Summary:"
echo "========================="
echo "🤖 Telegram Bot: ${TELEGRAM_BOT_TOKEN:0:10}..."
echo "🗄️  Database: ${TURSO_DATABASE_URL}"
echo "🤖 AI Providers: $ai_configured configured"
echo "🔐 Session Secret: Generated"
echo "📁 Directories: Created"
echo "🧪 Debug Templates: Ready"

echo ""
echo "🎉 Setup Complete!"
echo "=================="
echo ""
echo "🚀 To start the bot:"
echo "   ./scripts/check-deployment.sh  # Verify everything is ready"
echo "   GIN_MODE=release ./bot        # Start the bot"
echo ""
echo "📋 Next steps:"
echo "   1. Run the deployment check: ./scripts/check-deployment.sh"
echo "   2. Start the bot: GIN_MODE=release ./bot"
echo "   3. Test commands: /help, /setprovider, /mode"
echo "   4. Send test images for processing"
echo ""
echo "💡 Pro tip: Use ./debug/test_bot.sh scenarios for testing ideas"