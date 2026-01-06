#!/bin/bash

set -e

echo "🤖 Telegram Webhook Setup for Cloudflare Workers"
echo "==========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

# Configuration
TELEGRAM_BOT_TOKEN=${TELEGRAM_BOT_TOKEN:-}
WEBHOOK_URL=${WEBHOOK_URL:-https://api.obsidian-bot.com/webhook}
WEBHOOK_SECRET=${WEBHOOK_SECRET:-}

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check wrangler
    if ! command -v wrangler &> /dev/null; then
        print_error "Wrangler is not installed"
        print_info "Install with: npm install -g wrangler"
        exit 1
    fi
    
    # Check telegram token
    if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
        print_error "TELEGRAM_BOT_TOKEN environment variable is required"
        print_info "Set with: export TELEGRAM_BOT_TOKEN=your_token"
        exit 1
    fi
    
    print_status "Prerequisites satisfied"
}

# Set Telegram webhook
setup_webhook() {
    print_info "Setting up Telegram webhook..."
    
    # Get bot info
    local bot_info=$(curl -s "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getMe")
    if [ -z "$bot_info" ]; then
        print_error "Failed to get bot info. Check your token."
        exit 1
    fi
    
    local bot_username=$(echo "$bot_info" | jq -r '.result.username')
    print_status "Bot: @$bot_username"
    
    # Set webhook
    local webhook_data=$(jq -n \
        --arg url "$WEBHOOK_URL" \
        --arg secret "$WEBHOOK_SECRET" \
        '{
            "url": $url,
            "max_connections": 40,
            "allowed_updates": ["message", "callback_query"],
            "secret_token": $secret
        }')
    
    local webhook_response=$(curl -s \
        -X POST \
        -H "Content-Type: application/json" \
        -d "$webhook_data" \
        "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/setWebhook")
    
    local success=$(echo "$webhook_response" | jq -r '.ok')
    if [ "$success" = "true" ]; then
        print_status "Webhook set successfully"
        print_info "Webhook URL: $WEBHOOK_URL"
    else
        local error_desc=$(echo "$webhook_response" | jq -r '.description // "Unknown error"')
        print_error "Failed to set webhook: $error_desc"
        exit 1
    fi
}

# Verify webhook
verify_webhook() {
    print_info "Verifying webhook..."
    
    # Get current webhook info
    local webhook_info=$(curl -s "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getWebhookInfo")
    
    local webhook_url=$(echo "$webhook_info" | jq -r '.result.url // "Not set"')
    local pending_update_count=$(echo "$webhook_info" | jq -r '.result.pending_update_count // 0')
    
    if [ "$webhook_url" = "$WEBHOOK_URL" ]; then
        print_status "Webhook URL verified"
    else
        print_warning "Webhook URL mismatch"
        print_info "Current: $webhook_url"
        print_info "Expected: $WEBHOOK_URL"
    fi
    
    if [ "$pending_update_count" = "0" ]; then
        print_status "No pending updates"
    else
        print_warning "$pending_update_count pending updates"
    fi
}

# Test webhook connection
test_webhook() {
    print_info "Testing webhook connection..."
    
    # Get webhook info (includes connection info)
    local webhook_info=$(curl -s "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getWebhookInfo")
    
    local last_error_date=$(echo "$webhook_info" | jq -r '.result.last_error_date // null')
    local last_error_message=$(echo "$webhook_info" | jq -r '.result.last_error_message // null')
    
    if [ -n "$last_error_date" ]; then
        print_warning "Last webhook error:"
        print_info "Date: $last_error_date"
        print_info "Message: $last_error_message"
        return 1
    fi
    
    print_status "Webhook connection looks good"
    return 0
}

# Create webhook secret
create_webhook_secret() {
    print_info "Generating webhook secret..."
    
    # Generate random secret
    local webhook_secret=$(openssl rand -hex 32)
    
    # Save to .env file
    if [ -f ".env" ]; then
        if ! grep -q "WEBHOOK_SECRET" .env; then
            echo "WEBHOOK_SECRET=$webhook_secret" >> .env
        else
            sed -i "s/WEBHOOK_SECRET=.*/WEBHOOK_SECRET=$webhook_secret/" .env
        fi
    else
        echo "WEBHOOK_SECRET=$webhook_secret" > .env
    fi
    
    print_status "Webhook secret generated and saved to .env"
    echo "Secret: $webhook_secret"
}

# Update wrangler.toml with webhook
update_wrangler_config() {
    print_info "Updating wrangler.toml..."
    
    # Check if wrangler.toml exists
    if [ ! -f "workers/wrangler.toml" ]; then
        print_error "wrangler.toml not found"
        exit 1
    fi
    
    # Update webhook secret in wrangler.toml
    if grep -q "WEBHOOK_SECRET" workers/wrangler.toml; then
        sed -i "s/WEBHOOK_SECRET = .*/WEBHOOK_SECRET = \"$WEBHOOK_SECRET\"/" workers/wrangler.toml
    else
        echo "WEBHOOK_SECRET = \"$WEBHOOK_SECRET\"" >> workers/wrangler.toml
    fi
    
    print_status "wrangler.toml updated"
}

# Deploy workers
deploy_workers() {
    print_info "Deploying Cloudflare Workers..."
    
    # Change to workers directory
    cd workers
    
    # Deploy to production
    wrangler deploy --env production
    
    print_status "Workers deployed to production"
    
    # Get worker URL
    local worker_url=$(wrangler whoami | grep -A 10 "Account URL" | head -1 | awk '{print $4}')
    
    if [ -n "$worker_url" ]; then
        print_status "Worker URL: $worker_url"
        print_info "Webhook endpoint: $worker_url/webhook"
    fi
    
    cd ..
}

# Get webhook info
show_webhook_info() {
    print_info "Current webhook configuration:"
    
    local webhook_info=$(curl -s "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/getWebhookInfo")
    
    local url=$(echo "$webhook_info" | jq -r '.result.url // "Not set"')
    local has_custom_cert=$(echo "$webhook_info" | jq -r '.result.has_custom_certificate // false')
    local pending_updates=$(echo "$webhook_info" | jq -r '.result.pending_update_count // 0')
    local ip_address=$(echo "$webhook_info" | jq -r '.result.ip_address // "Not set"')
    local max_connections=$(echo "$webhook_info" | jq -r '.result.max_connections // 0')
    local allowed_updates=$(echo "$webhook_info" | jq -r '.result.allowed_updates[]' | tr '\n' ',' | sed 's/,$//')
    
    echo ""
    echo "URL: $url"
    echo "Custom Certificate: $has_custom_cert"
    echo "Pending Updates: $pending_updates"
    echo "IP Address: $ip_address"
    echo "Max Connections: $max_connections"
    echo "Allowed Updates: $allowed_updates"
}

# Remove webhook
remove_webhook() {
    print_info "Removing webhook..."
    
    local response=$(curl -s -X POST "https://api.telegram.org/bot$TELEGRAM_BOT_TOKEN/deleteWebhook")
    local success=$(echo "$response" | jq -r '.ok')
    
    if [ "$success" = "true" ]; then
        print_status "Webhook removed successfully"
    else
        local error_desc=$(echo "$response" | jq -r '.description // "Unknown error"')
        print_error "Failed to remove webhook: $error_desc"
        exit 1
    fi
}

# Main function
main() {
    case "${1:-setup}" in
        "setup")
            check_prerequisites
            create_webhook_secret
            update_wrangler_config
            deploy_workers
            sleep 5  # Wait for deployment
            setup_webhook
            verify_webhook
            ;;
        "verify")
            verify_webhook
            ;;
        "test")
            test_webhook
            ;;
        "info")
            show_webhook_info
            ;;
        "remove")
            remove_webhook
            ;;
        "deploy")
            deploy_workers
            ;;
        "secret")
            create_webhook_secret
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [setup|verify|test|info|remove|deploy|secret|help]"
            echo ""
            echo "Commands:"
            echo "  setup   - Complete webhook setup (default)"
            echo "  verify  - Verify webhook configuration"
            echo "  test    - Test webhook connection"
            echo "  info     - Show current webhook info"
            echo "  remove  - Remove webhook"
            echo "  deploy   - Deploy workers only"
            echo "  secret   - Generate webhook secret only"
            echo "  help     - Show this help"
            echo ""
            echo "Environment Variables:"
            echo "  TELEGRAM_BOT_TOKEN  - Telegram bot token (required)"
            echo "  WEBHOOK_URL           - Webhook URL (default: https://api.obsidian-bot.com/webhook)"
            echo "  WEBHOOK_SECRET        - Webhook secret (auto-generated if not provided)"
            echo ""
            echo "Examples:"
            echo "  $0 setup"
            echo "  $0 verify"
            echo "  WEBHOOK_SECRET=custom123 $0 setup"
            ;;
        *)
            print_error "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
if [ $# -gt 0 ]; then
    main "$@"
else
    main "setup"
fi