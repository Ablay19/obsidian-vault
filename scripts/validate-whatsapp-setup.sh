#!/bin/bash

echo "🧪 WhatsApp Setup Validation Tool"
echo "================================="

# Check environment variables
echo "📋 Checking WhatsApp Configuration Variables..."

vars=("WHATSAPP_ACCESS_TOKEN" "WHATSAPP_VERIFY_TOKEN" "WHATSAPP_APP_SECRET" "WHATSAPP_WEBHOOK_URL")
missing_vars=()

for var in "${vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Missing: $var"
        missing_vars+=("$var")
    else
        echo "✅ Found: $var"
    fi
done

if [ ${#missing_vars[@]} -gt 0 ]; then
    echo ""
    echo "🚨 Missing environment variables:"
    printf ' - %s\n' "${missing_vars[@]}"
    echo ""
    echo "📖 Please set these in your .env file:"
    echo ""
    echo "# WhatsApp Configuration"
    echo "WHATSAPP_ACCESS_TOKEN=\"EAADZ...your-access-token\""
    echo "WHATSAPP_VERIFY_TOKEN=\"your-custom-verify-token\""
    echo "WHATSAPP_APP_SECRET=\"your-app-secret-from-meta\""
    echo "WHATSAPP_WEBHOOK_URL=\"https://your-domain.com/api/v1/auth/whatsapp/webhook\""
    echo ""
    exit 1
fi

echo ""
echo "🌐 Testing Webhook URL Accessibility..."
if [ -n "$WHATSAPP_WEBHOOK_URL" ]; then
    # Test GET endpoint (verification)
    if curl -s --max-time 10 "$WHATSAPP_WEBHOOK_URL" > /dev/null 2>&1; then
        echo "✅ Webhook URL is accessible"
    else
        echo "❌ Webhook URL is not accessible"
        echo "   Please check:"
        echo "   - URL is correct and accessible from internet"
        echo "   - SSL certificate is valid"
        echo "   - Firewall allows incoming connections"
    fi
    
    # Test POST endpoint with sample payload
    echo ""
    echo "🧪 Testing webhook POST endpoint..."
    test_payload='{"object":"whatsapp_business_account","entry":[]}'
    
    if response=$(curl -s --max-time 10 -X POST "$WHATSAPP_WEBHOOK_URL" \
        -H "Content-Type: application/json" \
        -H "X-Hub-Signature-256: sha256=test" \
        -d "$test_payload" 2>&1); then
        
        if echo "$response" | grep -q "processed"; then
            echo "✅ Webhook POST endpoint working"
        else
            echo "⚠️  Webhook POST endpoint may have issues"
            echo "   Response: $response"
        fi
    else
        echo "❌ Cannot test webhook POST endpoint"
    fi
else
    echo "❌ WHATSAPP_WEBHOOK_URL not set"
fi

echo ""
echo "🔍 Checking Local Dependencies..."

# Check required binaries
required_bins=("curl" "openssl")
for bin in "${required_bins[@]}"; do
    if command -v "$bin" > /dev/null 2>&1; then
        echo "✅ $bin is available"
    else
        echo "❌ $bin is missing - please install it"
    fi
done

# Check optional binaries
optional_bins=("tesseract" "pdftotext")
for bin in "${optional_bins[@]}"; do
    if command -v "$bin" > /dev/null 2>&1; then
        echo "✅ $bin is available (optional)"
    else
        echo "⚠️  $bin is not available (optional - some features limited)"
    fi
done

echo ""
echo "📂 Checking Directory Structure..."

# Create media storage directory
media_dir="attachments/whatsapp"
if [ ! -d "$media_dir" ]; then
    echo "📁 Creating media storage directory: $media_dir"
    mkdir -p "$media_dir"
    echo "✅ Created: $media_dir"
else
    echo "✅ Media directory exists: $media_dir"
fi

# Check database
if [ -f "./test.db" ]; then
    echo "✅ Database exists: ./test.db"
elif [ -f "./dev.db" ]; then
    echo "✅ Database exists: ./dev.db"
else
    echo "⚠️  No database file found - will be created on startup"
fi

echo ""
echo "🔧 Configuration Summary:"
echo "================================="
echo "Webhook URL: ${WHATSAPP_WEBHOOK_URL:-Not Set}"
echo "Access Token: ${WHATSAPP_ACCESS_TOKEN:0:10}...${WHATSAPP_ACCESS_TOKEN: -4}"
echo "Verify Token: ${WHATSAPP_VERIFY_TOKEN:-Not Set}"
echo "App Secret: ${WHATSAPP_APP_SECRET:0:10}...${WHATSAPP_APP_SECRET: -4}"
echo "Media Storage: $media_dir"

echo ""
echo "🚀 Next Steps:"
echo "1. Start the bot: ./bot"
echo "2. Check dashboard: http://localhost:8080/dashboard/whatsapp"
echo "3. Send a test message to your WhatsApp number"
echo "4. Monitor logs: tail -f bot.log"

echo ""
echo "📊 Testing Checklist:"
echo "□ Webhook verification with curl"
echo "□ Message reception test"
echo "□ Media download test"
echo "□ Dashboard panel functionality"
echo "□ Error logging verification"

echo ""
echo "🔗 Useful URLs:"
echo "• Meta Developers: https://developers.facebook.com"
echo "• WhatsApp Business API: https://developers.facebook.com/docs/whatsapp/"
echo "• Webhook Tester: https://webhook.site"
echo "• HMAC Generator: https://www.freeformatter.com/hmac-generator.html"

echo ""
echo "✅ WhatsApp setup validation completed!"