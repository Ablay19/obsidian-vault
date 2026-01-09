#!/bin/bash

set -e

# Environment Variables Validation Script
# Validates required and optional environment variables for the Obsidian Bot

source "$(dirname "$0")/setup/common.sh"

echo "🔍 Obsidian Bot Environment Variables Validation"
echo "==============================================="

# Check if .env file exists
if [ ! -f ".env" ]; then
    print_error ".env file not found. Please run ./scripts/setup/env-setup.sh first."
    exit 1
fi

# Load environment variables
set -a
source .env
set +a

echo ""
print_status "Loaded environment variables from .env"

# Required variables validation
REQUIRED_VARS=(
    "TELEGRAM_BOT_TOKEN:Telegram Bot API token from @BotFather"
    "ENVIRONMENT_MODE:Environment mode (dev/prod)"
    "SESSION_SECRET:Session secret for authentication (min 32 chars)"
)

OPTIONAL_VARS=(
    "GEMINI_API_KEY:Google Gemini API key"
    "GROQ_API_KEY:Groq API key"
    "HUGGINGFACE_API_KEY:HuggingFace API key"
    "OPENROUTER_API_KEY:OpenRouter API key"
    "TURSO_DATABASE_URL:Turso database connection URL"
    "GOOGLE_CLIENT_ID:Google OAuth client ID"
    "GOOGLE_CLIENT_SECRET:Google OAuth client secret"
    "WHATSAPP_ACCESS_TOKEN:WhatsApp Business API token"
    "DASHBOARD_PORT:Dashboard port (default: 8080)"
    "BOT_MEMORY_LIMIT:Bot memory limit"
    "REDIS_MAX_MEMORY:Redis memory limit"
)

# Validate required variables
echo ""
print_header "Required Variables"
all_required_valid=true

for var_info in "${REQUIRED_VARS[@]}"; do
    var_name=$(echo "$var_info" | cut -d':' -f1)
    var_desc=$(echo "$var_info" | cut -d':' -f2)
    var_value="${!var_name}"

    if [ -z "$var_value" ]; then
        print_error "$var_name: MISSING - $var_desc"
        all_required_valid=false
    else
        print_success "$var_name: SET"

        # Additional validation for specific variables
        case $var_name in
            "ENVIRONMENT_MODE")
                if [[ ! "$var_value" =~ ^(dev|prod|staging)$ ]]; then
                    print_warning "$var_name: Invalid value '$var_value'. Must be dev, prod, or staging"
                fi
                ;;
            "SESSION_SECRET")
                if [ ${#var_value} -lt 32 ]; then
                    print_warning "$var_name: Too short (${#var_value} chars). Recommended: 32+ characters"
                fi
                ;;
            "TELEGRAM_BOT_TOKEN")
                if [[ ! "$var_value" =~ ^[0-9]+: ]]; then
                    print_warning "$var_name: Invalid format. Should start with bot ID"
                fi
                ;;
        esac
    fi
done

if [ "$all_required_valid" = false ]; then
    print_error "❌ Some required variables are missing. Please configure them in .env"
    exit 1
fi

print_success "✅ All required variables are configured"

# Validate optional variables
echo ""
print_header "Optional Variables"
configured_optional=0
total_optional=${#OPTIONAL_VARS[@]}

for var_info in "${OPTIONAL_VARS[@]}"; do
    var_name=$(echo "$var_info" | cut -d':' -f1)
    var_desc=$(echo "$var_info" | cut -d':' -f2)
    var_value="${!var_name}"

    if [ -n "$var_value" ]; then
        print_success "$var_name: CONFIGURED - $var_desc"
        configured_optional=$((configured_optional + 1))

        # Additional validation for specific variables
        case $var_name in
            "TURSO_DATABASE_URL")
                if [[ ! "$var_value" =~ ^libsql:// ]]; then
                    print_warning "$var_name: Should start with 'libsql://' for Turso"
                fi
                ;;
            "DASHBOARD_PORT")
                if ! [[ "$var_value" =~ ^[0-9]+$ ]] || [ "$var_value" -lt 1 ] || [ "$var_value" -gt 65535 ]; then
                    print_warning "$var_name: Invalid port number"
                fi
                ;;
        esac
    else
        print_info "$var_name: NOT SET - $var_desc"
    fi
done

# Summary
echo ""
print_header "Validation Summary"
print_info "Required variables: ${#REQUIRED_VARS[@]} configured"
print_info "Optional variables: $configured_optional/$total_optional configured"

if [ $configured_optional -eq 0 ]; then
    print_warning "⚠️  No AI services configured. The bot will have limited functionality."
    print_info "Configure at least one AI provider (GEMINI_API_KEY, GROQ_API_KEY, etc.)"
elif [ $configured_optional -lt 2 ]; then
    print_warning "⚠️  Only one AI service configured. Consider adding more for redundancy."
else
    print_success "✅ Multiple AI services configured for good redundancy"
fi

# Check for conflicting configurations
echo ""
print_header "Configuration Conflicts"
conflicts_found=false

# Check if both Turso and SQLite are configured
if [ -n "$TURSO_DATABASE_URL" ] && [ -n "$SQLITE_DATABASE_PATH" ]; then
    print_warning "Both TURSO_DATABASE_URL and SQLITE_DATABASE_PATH are set. Using Turso."
fi

# Check Google OAuth completeness
if [ -n "$GOOGLE_CLIENT_ID" ] && [ -z "$GOOGLE_CLIENT_SECRET" ]; then
    print_warning "GOOGLE_CLIENT_ID set but GOOGLE_CLIENT_SECRET missing"
    conflicts_found=true
elif [ -z "$GOOGLE_CLIENT_ID" ] && [ -n "$GOOGLE_CLIENT_SECRET" ]; then
    print_warning "GOOGLE_CLIENT_SECRET set but GOOGLE_CLIENT_ID missing"
    conflicts_found=true
fi

if [ "$conflicts_found" = false ]; then
    print_success "✅ No configuration conflicts detected"
fi

# Final result
echo ""
if [ "$all_required_valid" = true ]; then
    print_success "🎉 Environment validation completed successfully!"
    print_info "You can now proceed with: docker-compose up -d"
    exit 0
else
    print_error "❌ Environment validation failed. Please fix the issues above."
    exit 1
fi