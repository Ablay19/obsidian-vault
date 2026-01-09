#!/bin/bash

echo "☁️  Google Cloud Logging Setup for Production"
echo "============================================="

# Check if .env file exists
if [ ! -f ".env" ]; then
    echo "❌ Error: .env file not found. Please run setup-doppler.sh first."
    exit 1
fi

echo ""
echo "📋 Current Google Cloud Configuration:"
echo "===================================="

# Show current Google Cloud environment (if any)
if [ -f ".env.google" ]; then
    echo "📁 Found .env.google file:"
    grep -v '^#' .env.google | grep -v '^$' || echo "  (empty)"
else
    echo "❌ No .env.google file found"
fi

echo ""
echo "🔧 Google Cloud Environment Variables:"
echo "===================================="
echo "GOOGLE_CLOUD_PROJECT: ${GOOGLE_CLOUD_PROJECT:-"not set"}"
echo "GOOGLE_APPLICATION_CREDENTIALS: ${GOOGLE_APPLICATION_CREDENTIALS:-"not set"}"
echo "ENABLE_GOOGLE_LOGGING: ${ENABLE_GOOGLE_LOGGING:-"not set"}"

echo ""
echo "📝 Setup Instructions:"
echo "====================="
echo ""
echo "1. Create a Google Cloud Project:"
echo "   - Go to: https://console.cloud.google.com/"
echo "   - Create a new project or use existing one"
echo ""
echo "2. Enable Cloud Logging API:"
echo "   - In your project, enable 'Cloud Logging API'"
echo ""
echo "3. Create Service Account:"
echo "   - Go to IAM & Admin > Service Accounts"
echo "   - Create service account with 'Logs Writer' role"
echo "   - Download JSON key file"
echo ""
echo "4. Configure Environment:"
echo "   export GOOGLE_CLOUD_PROJECT=\"your-project-id\""
echo "   export GOOGLE_APPLICATION_CREDENTIALS=\"/path/to/key.json\""
echo "   export ENABLE_GOOGLE_LOGGING=\"true\""
echo ""
echo "5. Test Configuration:"
echo "   ./test-google-cloud-logging.sh"
echo ""

# Check if required variables are set
if [ -n "$GOOGLE_CLOUD_PROJECT" ] && [ -n "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
    echo "✅ Google Cloud environment variables are configured!"
    
    # Check if credentials file exists
    if [ -f "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
        echo "✅ Service account credentials file found"
        
        # Test Google Cloud connectivity
        echo ""
        echo "🧪 Testing Google Cloud Connectivity..."
        if command -v gcloud >/dev/null 2>&1; then
            echo "✅ gcloud CLI found"
            gcloud auth list --filter=status:ACTIVE --format="value(account)" | head -1 | xargs echo "  Active account:"
        else
            echo "⚠️  gcloud CLI not found (optional for testing)"
        fi
        
        echo ""
        echo "🚀 To enable Google Cloud logging:"
        echo "   export ENABLE_GOOGLE_LOGGING=\"true\""
        echo "   ./bot"
        
    else
        echo "❌ Service account credentials file not found: $GOOGLE_APPLICATION_CREDENTIALS"
    fi
else
    echo "⚠️  Google Cloud environment variables not configured"
    echo ""
    echo "📦 Use the production template:"
    echo "   cp .env.google.production .env.google"
    echo "   # Edit .env.google with your real credentials"
fi

echo ""
echo "💡 Development Mode:"
echo "==================="
echo "For development without real Google Cloud credentials:"
echo "  - The system will use local logging automatically"
echo "  - All functionality remains intact"
echo "  - Google Cloud logging will be gracefully disabled"

echo ""
echo "🔗 Next Steps:"
echo "=============="
echo "1. Set up Google Cloud project (if not done)"
echo "2. Configure environment variables"
echo "3. Run: ./test-google-cloud-logging.sh"
echo "4. Enable with: export ENABLE_GOOGLE_LOGGING=\"true\""
echo "5. Restart the bot to apply changes"

echo ""
echo "✅ Google Cloud logging setup guide completed!"