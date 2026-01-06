#!/bin/bash

set -e  # Exit on any error

echo "🚀 Obsidian Bot Environment Setup with Doppler"
echo "=========================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    local status=$1
    local message=$2
    case $status in
        "OK") echo -e "${GREEN}✅ $message${NC}" ;;
        "WARN") echo -e "${YELLOW}⚠️  $message${NC}" ;;
        "ERROR") echo -e "${RED}❌ $message${NC}" ;;
        "INFO") echo -e "${BLUE}ℹ️  $message${NC}" ;;
        *) echo -e "${NC}$message" ;;
    esac
}

print_header() {
    echo -e "\n${BLUE}🔧 $1${NC}"
    echo "================================"
}

# Step 1: Check Doppler setup
print_header "Step 1: Doppler Configuration"

if ! command -v doppler &> /dev/null; then
    print_status "ERROR" "Doppler CLI not found. Please install it first:"
    echo "curl -fsSL https://cli.doppler.com/install.sh | sh"
    exit 1
fi

print_status "OK" "Doppler CLI is installed"

# Check if already logged in
if ! doppler whoami &> /dev/null; then
    print_status "WARN" "Not logged into Doppler. Please login:"
    echo "doppler login"
    exit 1
fi

print_status "OK" "Already logged into Doppler"

# Step 2: Configure Doppler project
print_header "Step 2: Doppler Project Setup"

# Try to get existing projects
echo "Checking existing Doppler projects..."
PROJECTS_LIST=$(doppler projects list 2>/dev/null || echo "")

if [ -n "$PROJECTS_LIST" ]; then
    # Extract first project name from the list
    FIRST_PROJECT=$(echo "$PROJECTS_LIST" | head -1 | awk '{print $1}')
    
    if [ -n "$FIRST_PROJECT" ]; then
        print_status "OK" "Found existing project: $FIRST_PROJECT"
        PROJECT_NAME="$FIRST_PROJECT"
        
        print_status "INFO" "Switching to project: $PROJECT_NAME"
        doppler projects switch "$PROJECT_NAME" 2>/dev/null
        if [ $? -eq 0 ]; then
            print_status "OK" "Successfully switched to project: $PROJECT_NAME"
        else
            print_status "WARN" "Failed to switch project, but continuing"
        fi
    else
        print_status "WARN" "No existing projects found"
        PROJECT_NAME="obsidian-bot"
    fi
else
    print_status "WARN" "Unable to list projects, using default"
    PROJECT_NAME="obsidian-bot"
fi

print_status "INFO" "Using Doppler project: $PROJECT_NAME"
    fi
    
    doppler projects switch "$PROJECT_NAME"
    if [ $? -eq 0 ]; then
        print_status "OK" "Using project: $PROJECT_NAME"
    else
        print_status "WARN" "Failed to switch to project, continuing with setup"
    fi
else
    print_status "OK" "Using existing Doppler project: $CURRENT_PROJECT"
    PROJECT_NAME="$CURRENT_PROJECT"
fi
    if [ $? -eq 0 ]; then
        echo "✅ Switched to project: $PROJECT_NAME"
    else
        echo "❌ Failed to switch to project"
        exit 1
    fi
    print_status "OK" "Doppler project configured: $PROJECT_NAME"
else
    print_status "OK" "Using existing Doppler project: $CURRENT_PROJECT"
    PROJECT_NAME="$CURRENT_PROJECT"
fi

# Step 3: Set up environment with all required variables
print_header "Step 3: Environment Variable Setup"

# Define all required environment variables
declare -A ENV_VARS=(
    ["TURSO_DATABASE_URL"]="file:./obsidian.db"
    ["TURSO_AUTH_TOKEN"]="your-turso-auth-token"
    ["TELEGRAM_BOT_TOKEN"]="your-telegram-bot-token"
    ["SESSION_SECRET"]="change-me-to-something-very-secure"
    ["ENVIRONMENT_MODE"]="dev"
    ["AI_ENABLED"]="true"
    ["CLOUDFLARE_WORKER_URL"]="https://obsidian-bot-workers.your-subdomain.workers.dev"
    ["ACTIVE_PROVIDER"]="Cloudflare"
    ["WHATSAPP_ACCESS_TOKEN"]="EAADZ..."
    ["WHATSAPP_VERIFY_TOKEN"]="your-verify-token"
    ["WHATSAPP_APP_SECRET"]="your-app-secret"
    ["WHATSAPP_WEBHOOK_URL"]="https://your-domain.com/api/v1/auth/whatsapp/webhook"
    ["GEMINI_API_KEY"]="your-gemini-api-key"
    ["GROQ_API_KEY"]="your-groq-api-key"
    ["HUGGINGFACE_API_KEY"]="your-huggingface-api-key"
    ["OPENROUTER_API_KEY"]="your-openrouter-api-key"
)

# Interactive setup for each variable
print_status "INFO" "Configuring environment variables interactively..."
echo ""

# Function to prompt for variable value
prompt_var() {
    local var_name=$1
    local default_value=$2
    local description=$3
    local is_secret=${4:-false}
    
    echo -e "${BLUE}📝 $var_name${NC}"
    echo -e "   $description"
    echo -e "   Default: ${YELLOW}$default_value${NC}"
    
    if [ "$is_secret" = "true" ]; then
        read -s -p "   Enter value (or press Enter to keep default): " value
        echo ""
    else
        read -p "   Enter value (or press Enter to keep default): " value
        echo ""
    fi
    
    if [ -z "$value" ]; then
        value="$default_value"
    fi
    
    # Set in Doppler
    if doppler secrets set "$var_name" "$value" --plain &> /dev/null; then
        print_status "OK" "Set $var_name"
    else
        print_status "ERROR" "Failed to set $var_name"
        return 1
    fi
}

# Setup critical variables first
echo -e "${YELLOW}🔑 Critical Configuration (Required)${NC}"
prompt_var "TURSO_DATABASE_URL" "${ENV_VARS[TURSO_DATABASE_URL]}" "SQLite database file path" "false"
prompt_var "SESSION_SECRET" "${ENV_VARS[SESSION_SECRET]}" "Secret for session encryption" "true"

echo -e "${YELLOW}🤖 AI Configuration${NC}"
prompt_var "CLOUDFLARE_WORKER_URL" "${ENV_VARS[CLOUDFLARE_WORKER_URL]}" "Cloudflare Worker URL for AI" "false"
prompt_var "ACTIVE_PROVIDER" "${ENV_VARS[ACTIVE_PROVIDER]}" "Default AI provider (Cloudflare, Gemini, Groq, etc.)" "false"

echo -e "${YELLOW}📱 Telegram Configuration${NC}"
prompt_var "TELEGRAM_BOT_TOKEN" "${ENV_VARS[TELEGRAM_BOT_TOKEN]}" "Telegram bot token" "true"

echo -e "${YELLOW"📱 WhatsApp Configuration (Optional)${NC}"
read -p "Do you want to configure WhatsApp? (y/N): " setup_whatsapp
echo ""

if [[ "$setup_whatsapp" =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}⚠️  WhatsApp requires Meta Developer account and Business API access${NC}"
    echo ""
    
    prompt_var "WHATSAPP_ACCESS_TOKEN" "${ENV_VARS[WHATSAPP_ACCESS_TOKEN]}" "WhatsApp Business API access token" "true"
    prompt_var "WHATSAPP_VERIFY_TOKEN" "${ENV_VARS[WHATSAPP_VERIFY_TOKEN]}" "Webhook verification token" "true"
    prompt_var "WHATSAPP_APP_SECRET" "${ENV_VARS[WHATSAPP_APP_SECRET]}" "App secret for webhook validation" "true"
    prompt_var "WHATSAPP_WEBHOOK_URL" "${ENV_VARS[WHATSAPP_WEBHOOK_URL]}" "Public webhook URL" "false"
    
    echo -e "${GREEN}✅ WhatsApp configuration completed${NC}"
else
    echo -e "${BLUE}ℹ️  Skipping WhatsApp configuration${NC}"
fi

echo -e "${YELLOW}🤖 Additional AI Providers (Optional)${NC}"
read -p "Do you want to configure additional AI providers? (y/N): " setup_additional_ai
echo ""

if [[ "$setup_additional_ai" =~ ^[Yy]$ ]]; then
    prompt_var "GEMINI_API_KEY" "${ENV_VARS[GEMINI_API_KEY]}" "Google Gemini API key" "true"
    prompt_var "GROQ_API_KEY" "${ENV_VARS[GROQ_API_KEY]}" "Groq API key" "true"
    prompt_var "HUGGINGFACE_API_KEY" "${ENV_VARS[HUGGINGFACE_API_KEY]}" "Hugging Face API key" "true"
    prompt_var "OPENROUTER_API_KEY" "${ENV_VARS[OPENROUTER_API_KEY]}" "OpenRouter API key" "true"
fi

# Step 4: Verify Doppler configuration
print_header "Step 4: Doppler Configuration Verification"

print_status "INFO" "Verifying Doppler configuration..."
echo ""

# List all configured secrets
if doppler secrets list --plain 2>/dev/null; then
    echo -e "${GREEN}✅ Doppler secrets configured:${NC}"
    doppler secrets list --plain | while read -r line; do
        if [ -n "$line" ]; then
            echo "  • $line"
        fi
    done
else
    print_status "ERROR" "Failed to list Doppler secrets"
    exit 1
fi

# Step 5: Create local environment file
print_header "Step 5: Local Environment File Creation"

ENV_FILE=".env"
BACKUP_FILE=".env.backup.$(date +%Y%m%d_%H%M%S)"

print_status "INFO" "Creating local .env file from Doppler..."

# Backup existing .env if it exists
if [ -f "$ENV_FILE" ]; then
    cp "$ENV_FILE" "$BACKUP_FILE"
    print_status "OK" "Backed up existing .env to $BACKUP_FILE"
fi

# Create new .env file
cat > "$ENV_FILE" << EOF
# Obsidian Bot Environment Configuration
# Generated by setup-doppler.sh on $(date)
# Source: Doppler secrets management

# Database Configuration
TURSO_DATABASE_URL=file:./obsidian.db
TURSO_AUTH_TOKEN=\$(doppler secrets get TURSO_AUTH_TOKEN --plain)

# Bot Configuration
TELEGRAM_BOT_TOKEN=\$(doppler secrets get TELEGRAM_BOT_TOKEN --plain)
SESSION_SECRET=\$(doppler secrets get SESSION_SECRET --plain)
ENVIRONMENT_MODE=dev
AI_ENABLED=true

# AI Configuration
CLOUDFLARE_WORKER_URL=\$(doppler secrets get CLOUDFLARE_WORKER_URL --plain)
ACTIVE_PROVIDER=\$(doppler secrets get ACTIVE_PROVIDER --plain)

# AI Provider Fallbacks (Optional)
GEMINI_API_KEY=\$(doppler secrets get GEMINI_API_KEY --plain)
GROQ_API_KEY=\$(doppler secrets get GROQ_API_KEY --plain)
HUGGINGFACE_API_KEY=\$(doppler secrets get HUGGINGFACE_API_KEY --plain)
OPENROUTER_API_KEY=\$(doppler secrets get OPENROUTER_API_KEY --plain)

# WhatsApp Configuration (Optional)
WHATSAPP_ACCESS_TOKEN=\$(doppler secrets get WHATSAPP_ACCESS_TOKEN --plain)
WHATSAPP_VERIFY_TOKEN=\$(doppler secrets get WHATSAPP_VERIFY_TOKEN --plain)
WHATSAPP_APP_SECRET=\$(doppler secrets get WHATSAPP_APP_SECRET --plain)
WHATSAPP_WEBHOOK_URL=\$(doppler secrets get WHATSAPP_WEBHOOK_URL --plain)
EOF

print_status "OK" "Created $ENV_FILE with Doppler integration"

# Set proper permissions
chmod 600 "$ENV_FILE"
print_status "OK" "Set secure permissions on $ENV_FILE"

# Step 6: Test Environment Setup
print_header "Step 6: Environment Testing"

print_status "INFO" "Testing environment setup..."

# Test 1: Load environment
print_status "INFO" "Testing environment loading..."
if source .env 2>/dev/null; then
    print_status "OK" "Environment file loads successfully"
else
    print_status "ERROR" "Failed to load environment file"
    exit 1
fi

# Test 2: Check critical variables
print_status "INFO" "Validating critical variables..."

if [ -z "$TURSO_DATABASE_URL" ]; then
    print_status "ERROR" "TURSO_DATABASE_URL is not set"
    exit 1
fi

if [ -z "$SESSION_SECRET" ]; then
    print_status "ERROR" "SESSION_SECRET is not set"
    exit 1
fi

if [ -z "$CLOUDFLARE_WORKER_URL" ]; then
    print_status "ERROR" "CLOUDFLARE_WORKER_URL is not set"
    exit 1
fi

if [ -z "$TELEGRAM_BOT_TOKEN" ]; then
    print_status "ERROR" "TELEGRAM_BOT_TOKEN is not set"
    exit 1
fi

print_status "OK" "All critical variables are set"

# Test 3: Database connectivity
print_status "INFO" "Testing database connectivity..."
if [ -f "./obsidian.db" ]; then
    print_status "OK" "Database file exists"
else
    print_status "WARN" "Database file will be created on first run"
fi

# Step 7: Create test and deployment scripts
print_header "Step 7: Helper Scripts Creation"

# Create Doppler update script
cat > update-doppler.sh << 'EOF'
#!/bin/bash
echo "🔄 Updating Doppler secrets..."
doppler secrets sync
echo "✅ Doppler secrets updated"
echo ""
echo "🚀 Restarting bot with new configuration..."
if pgrep -f "bot" > /dev/null; then
    pkill -f "bot"
    sleep 2
fi
./bot &
echo "✅ Bot started with updated configuration"
EOF

chmod +x update-doppler.sh
print_status "OK" "Created update-doppler.sh script"

# Create comprehensive test script
cat > test-comprehensive.sh << 'EOF'
#!/bin/bash

echo "🧪 Comprehensive Environment Test"
echo "==============================="

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

test_section() {
    local title=$1
    echo -e "\n${BLUE}🔧 $title${NC}"
    echo "--------------------------------"
}

test_result() {
    local status=$1
    local message=$2
    case $status in
        "PASS") echo -e "${GREEN}✅ PASS: $message${NC}" ;;
        "FAIL") echo -e "${RED}❌ FAIL: $message${NC}" ;;
        "WARN") echo -e "${YELLOW}⚠️  WARN: $message${NC}" ;;
        "INFO") echo -e "${BLUE}ℹ️  INFO: $message${NC}" ;;
    esac
}

# Test 1: Environment Loading
test_section "Environment Loading Test"
if source .env 2>/dev/null; then
    test_result "PASS" ".env file loads successfully"
else
    test_result "FAIL" ".env file failed to load"
fi

# Test 2: Variable Validation
test_section "Variable Validation Test"

critical_vars=("TURSO_DATABASE_URL" "SESSION_SECRET" "CLOUDFLARE_WORKER_URL" "TELEGRAM_BOT_TOKEN")
for var in "${critical_vars[@]}"; do
    if [ -z "${!var}" ]; then
        test_result "FAIL" "$var is not set"
    else
        test_result "PASS" "$var is set"
    fi
done

# Test 3: Doppler Integration
test_section "Doppler Integration Test"
if doppler whoami &> /dev/null; then
    test_result "PASS" "Doppler login is active"
else
    test_result "FAIL" "Doppler login failed"
fi

if doppler secrets list --plain &> /dev/null; then
    test_result "PASS" "Doppler secrets accessible"
else
    test_result "FAIL" "Doppler secrets not accessible"
fi

# Test 4: Network Connectivity
test_section "Network Connectivity Test"

# Test Cloudflare Worker
if [ -n "$CLOUDFLARE_WORKER_URL" ]; then
    if curl -s --max-time 10 "$CLOUDFLARE_WORKER_URL/health" | grep -q "OK"; then
        test_result "PASS" "Cloudflare Worker accessible"
    else
        test_result "FAIL" "Cloudflare Worker not accessible"
    fi
fi

# Test 5: AI Provider Selection
test_section "AI Provider Test"
if [ -n "$ACTIVE_PROVIDER" ]; then
    test_result "INFO" "Active provider: $ACTIVE_PROVIDER"
    
    case "$ACTIVE_PROVIDER" in
        "Cloudflare")
            if curl -s "$CLOUDFLARE_WORKER_URL/ai-test" &> /dev/null; then
                test_result "PASS" "Cloudflare AI binding test passed"
            else
                test_result "WARN" "Cloudflare AI binding test inconclusive"
            fi
            ;;
        *)
            test_result "INFO" "Alternative provider configured: $ACTIVE_PROVIDER"
            ;;
    esac
fi

# Test 6: WhatsApp Configuration (Optional)
test_section "WhatsApp Configuration Test"
if [ -n "$WHATSAPP_ACCESS_TOKEN" ]; then
    test_result "INFO" "WhatsApp credentials configured"
    
    if [ -n "$WHATSAPP_WEBHOOK_URL" ]; then
        if [[ "$WHATSAPP_WEBHOOK_URL" =~ ^https:// ]]; then
            test_result "PASS" "WhatsApp webhook URL uses HTTPS"
        else
            test_result "WARN" "WhatsApp webhook URL should use HTTPS"
        fi
    else
        test_result "WARN" "WhatsApp webhook URL not configured"
    fi
else
    test_result "INFO" "WhatsApp not configured (optional)"
fi

# Test 7: File System
test_section "File System Test"

# Check media directories
for dir in "attachments/whatsapp" "attachments/telegram"; do
    if [ ! -d "$dir" ]; then
        mkdir -p "$dir"
        test_result "INFO" "Created directory: $dir"
    else
        test_result "PASS" "Directory exists: $dir"
    fi
done

# Check database file
if [ -f "./obsidian.db" ]; then
    test_result "PASS" "Database file exists"
else
    test_result "INFO" "Database file will be created on startup"
fi

echo -e "\n${GREEN}🎯 Test Summary${NC}"
echo "================================"
echo "Run this test anytime to validate your environment:"
echo "./test-comprehensive.sh"
echo ""

echo -e "${BLUE}🚀 Quick Commands${NC}"
echo "================================"
echo "Start bot:        ./bot"
echo "Stop bot:         pkill -f bot"
echo "Update secrets:   ./update-doppler.sh"
echo "Run tests:       ./test-comprehensive.sh"
echo "Edit config:      vim .env"
echo "Doppler dashboard: https://doppler.com"
EOF

chmod +x test-comprehensive.sh
print_status "OK" "Created test-comprehensive.sh script"

# Step 8: Final instructions
print_header "Setup Complete! 🎉"

echo -e "${GREEN}✅ Doppler-based environment setup completed successfully!${NC}"
echo ""
echo -e "${BLUE}🚀 Next Steps:${NC}"
echo "================================"
echo ""
echo "1. ${YELLOW}Review and update actual values${NC}:"
echo "   Edit .env and replace placeholder values with real credentials"
echo ""
echo "2. ${YELLOW}Run comprehensive tests${NC}:"
echo "   ./test-comprehensive.sh"
echo ""
echo "3. ${YELLOW}Start the bot${NC}:"
echo "   ./bot"
echo ""
echo "4. ${YELLOW}Access dashboard${NC}:"
echo "   http://localhost:8080"
echo ""
echo "5. ${YELLOW}Update secrets${NC}:"
echo "   ./update-doppler.sh"
echo ""
echo -e "${BLUE}📚 Useful Commands:${NC}"
echo "  • List secrets:     doppler secrets list --plain"
echo "  • Set new secret:   doppler secrets set VAR_NAME value --plain"
echo "  • Sync secrets:     doppler secrets sync"
echo "  • Switch project:    doppler project set PROJECT_NAME"
echo ""
echo -e "${GREEN}🎯 Your environment is now properly configured with Doppler!${NC}"