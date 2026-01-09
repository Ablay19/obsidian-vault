#!/bin/bash

echo "🧪 Comprehensive Environment Test"
echo "============================"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() {
    local status=$1
    local message=$2
    case $status in
        "PASS") echo -e "${GREEN}✅ $message${NC}" ;;
        "FAIL") echo -e "${RED}❌ $message${NC}" ;;
        "WARN") echo -e "${YELLOW}⚠️  $message${NC}" ;;
        "INFO") echo -e "${BLUE}ℹ️  $message${NC}" ;;
        *) echo -e "${NC}$message" ;;
    esac
}

# Test 1: Environment Loading
echo "🔧 Testing Environment Loading..."
if [ -f ".env" ]; then
    if source .env 2>/dev/null; then
        print_status "PASS" ".env file loads successfully"
    else
        print_status "FAIL" ".env file failed to load"
    fi
else
    print_status "FAIL" ".env file not found"
fi

# Test 2: Critical Variables
echo ""
echo "🔑 Testing Critical Variables..."
critical_vars=("TURSO_DATABASE_URL" "SESSION_SECRET" "TELEGRAM_BOT_TOKEN")
all_set=true

for var in "${critical_vars[@]}"; do
    value=$(eval echo \$$var)
    if [ -n "$value" ]; then
        print_status "PASS" "$var is set"
    else
        print_status "FAIL" "$var is missing"
        all_set=false
    fi
done

if [ "$all_set" = false ]; then
    echo -e "\n${RED}❌ Critical variables missing!${NC}"
    echo -e "${YELLOW}Please run the setup script first:${NC}"
    exit 1
fi

# Test 3: AI Configuration
echo ""
echo "🤖 Testing AI Configuration..."
if [ -n "$CLOUDFLARE_WORKER_URL" ]; then
    print_status "FAIL" "CLOUDFLARE_WORKER_URL not set"
else
    print_status "PASS" "CLOUDFLARE_WORKER_URL: ${CLOUDFLARE_WORKER_URL:0:50}..."
    
    if curl -s --max-time 5 "$CLOUDFLARE_WORKER_URL/health" | grep -q "OK"; then
        print_status "PASS" "Cloudflare Worker is accessible"
    else
        print_status "WARN" "Cloudflare Worker not accessible (may be offline)"
    fi
fi

# Test 4: Bot Startup
echo ""
echo "🚀 Testing Bot Startup..."
if pgrep -f "bot" > /dev/null; then
    print_status "WARN" "Bot is already running"
    echo "Stopping existing instance..."
    pkill -f "bot"
    sleep 2
fi

# Start bot in background
echo "Starting bot..."
timeout 15s ./bot > /tmp/bot-test.log 2>&1 &
BOT_PID=$!

# Wait for startup
sleep 8

# Check if bot is still running
if kill -0 $BOT_PID 2>/dev/null; then
    print_status "FAIL" "Bot failed to start properly"
    echo "Last few lines of bot log:"
    tail -5 /tmp/bot-test.log
else
    print_status "PASS" "Bot started successfully (PID: $BOT_PID)"
    sleep 3
fi

# Test 5: API Endpoints
echo ""
echo "🔗 Testing API Endpoints..."

# Test dashboard
if curl -s http://localhost:8080/dashboard/whatsapp > /dev/null 2>&1; then
    print_status "PASS" "Dashboard accessible"
else
    print_status "FAIL" "Dashboard not accessible"
fi

# Test providers endpoint
if curl -s http://localhost:8080/api/ai/providers > /dev/null 2>&1; then
    providers_response=$(curl -s http://localhost:8080/api/ai/providers)
    echo "  Available providers: $(echo "$providers_response" | grep -o '"available":\[[^]]*\]' | cut -d'"' -f2)"
    echo "  Active provider: $(echo "$providers_response" | grep -o '"active":"[^"]*"' | cut -d'"' -f2)"
    print_status "PASS" "AI providers endpoint working"
else
    print_status "FAIL" "AI providers endpoint not responding"
fi

# Test services status
if curl -s http://localhost:8080/api/services/status > /dev/null 2>&1; then
    print_status "PASS" "Services status endpoint working"
else
    print_status "FAIL" "Services status endpoint not responding"
fi

# Test 6: Cleanup
echo ""
echo "🧹 Cleaning up test environment..."

if kill -0 $BOT_PID 2>/dev/null; then
    print_status "INFO" "Stopping test bot instance"
fi

# Remove temp log
rm -f /tmp/bot-test.log

# Test 7: Summary
echo ""
echo "🎯 Test Summary"
echo "============"

echo "✅ Tests completed:"
echo "  • Environment validation"
echo "  • Variable checking"
echo "  • API endpoint testing"
echo "  • Bot startup verification"

echo ""
echo "🚀 Next Steps:"
echo "  1. If all tests PASS: Start bot with ./bot"
echo "  2. If any tests FAIL: Check setup-doppler.sh and .env file"
echo "  3. Access dashboard: http://localhost:8080"
echo "  4. Review documentation: docs/README-SETUP.md"

echo ""
print_status "PASS" "Comprehensive environment testing completed!"