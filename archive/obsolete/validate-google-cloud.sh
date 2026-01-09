#!/bin/bash

set -e

echo "☁️  Google Cloud Configuration Validation"
echo "====================================="

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

# Check if gcloud CLI is installed
check_gcloud_cli() {
    if ! command -v gcloud &> /dev/null; then
        print_error "Google Cloud CLI (gcloud) is not installed"
        print_info "Install from: https://cloud.google.com/sdk/docs/install"
        return 1
    fi
    
    local gcloud_version=$(gcloud version --format='value(bundledPython.version)' 2>/dev/null || gcloud version --format='value(version)')
    print_status "Google Cloud CLI found (version: $gcloud_version)"
    return 0
}

# Check authentication
check_authentication() {
    print_info "Checking Google Cloud authentication..."
    
    # Check if application default credentials are set
    if ! gcloud auth application-default print-access-token &> /dev/null; then
        print_error "No application default credentials found"
        print_info "Run: gcloud auth application-default login"
        return 1
    fi
    
    # Get current account
    local account=$(gcloud config get-value account 2>/dev/null || echo "Not configured")
    if [ "$account" = "Not configured" ]; then
        print_error "No account configured"
        print_info "Run: gcloud auth login"
        return 1
    fi
    
    print_status "Authenticated as: $account"
    return 0
}

# Check project configuration
check_project() {
    print_info "Checking project configuration..."
    
    local project_id=$(gcloud config get-value project 2>/dev/null || echo "Not configured")
    if [ "$project_id" = "Not configured" ]; then
        print_error "No project configured"
        print_info "Run: gcloud config set project YOUR_PROJECT_ID"
        return 1
    fi
    
    print_status "Project configured: $project_id"
    
    # Check if project exists and is accessible
    if ! gcloud projects describe "$project_id" &> /dev/null; then
        print_error "Project '$project_id' is not accessible or doesn't exist"
        return 1
    fi
    
    print_status "Project '$project_id' is accessible"
    return 0
}

# Check required APIs
check_apis() {
    print_info "Checking required Google Cloud APIs..."
    
    local required_apis=("logging.googleapis.com" "cloudresourcemanager.googleapis.com")
    local missing_apis=()
    
    for api in "${required_apis[@]}"; do
        if ! gcloud services list --enabled --filter="name=$api" --format="value(name)" | grep -q "$api"; then
            missing_apis+=("$api")
        else
            print_status "API enabled: $api"
        fi
    done
    
    if [ ${#missing_apis[@]} -gt 0 ]; then
        print_error "Missing required APIs:"
        for api in "${missing_apis[@]}"; do
            echo "  - $api"
        done
        print_info "Enable with: gcloud services enable ${missing_apis[*]}"
        return 1
    fi
    
    print_status "All required APIs are enabled"
    return 0
}

# Check service account
check_service_account() {
    print_info "Checking service account configuration..."
    
    local GOOGLE_APPLICATION_CREDENTIALS=${GOOGLE_APPLICATION_CREDENTIALS:-}
    local GOOGLE_CLOUD_PROJECT=${GOOGLE_CLOUD_PROJECT:-$(gcloud config get-value project)}
    
    if [ -z "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
        print_warning "GOOGLE_APPLICATION_CREDENTIALS not set"
        return 1
    fi
    
    if [ ! -f "$GOOGLE_APPLICATION_CREDENTIALS" ]; then
        print_error "Service account key file not found: $GOOGLE_APPLICATION_CREDENTIALS"
        return 1
    fi
    
    # Validate JSON file
    if ! jq empty "$GOOGLE_APPLICATION_CREDENTIALS" &> /dev/null; then
        print_error "Service account key file is not valid JSON: $GOOGLE_APPLICATION_CREDENTIALS"
        return 1
    fi
    
    # Extract service account email
    local sa_email=$(jq -r '.client_email' "$GOOGLE_APPLICATION_CREDENTIALS" 2>/dev/null || echo "Invalid format")
    if [ "$sa_email" = "Invalid format" ] || [ -z "$sa_email" ]; then
        print_error "Invalid service account key format"
        return 1
    fi
    
    print_status "Service account file found: $GOOGLE_APPLICATION_CREDENTIALS"
    print_status "Service account email: $sa_email"
    
    # Check permissions
    if ! gcloud projects get-iam-policy "$GOOGLE_CLOUD_PROJECT" --format=json | jq -r '.bindings[] | select(.members[] | contains("serviceAccount:'$sa_email'")) | .role' | grep -q "roles/logging.logWriter"; then
        print_warning "Service account may not have logging permissions"
        print_info "Grant with: gcloud projects add-iam-policy-binding $GOOGLE_CLOUD_PROJECT --member=\"serviceAccount:$sa_email\" --role=\"roles/logging.logWriter\""
    else
        print_status "Service account has logging permissions"
    fi
    
    return 0
}

# Check application configuration
check_app_config() {
    print_info "Checking application configuration..."
    
    local GOOGLE_CLOUD_PROJECT=${GOOGLE_CLOUD_PROJECT:-}
    local ENABLE_GOOGLE_LOGGING=${ENABLE_GOOGLE_LOGGING:-false}
    
    if [ -z "$GOOGLE_CLOUD_PROJECT" ]; then
        print_warning "GOOGLE_CLOUD_PROJECT not set"
        print_info "Set with: export GOOGLE_CLOUD_PROJECT=your-project-id"
        return 1
    fi
    
    print_status "Google Cloud Project: $GOOGLE_CLOUD_PROJECT"
    print_status "Google Cloud Logging: $ENABLE_GOOGLE_LOGGING"
    
    if [ "$ENABLE_GOOGLE_LOGGING" = "true" ]; then
        print_status "Google Cloud logging is enabled"
    else
        print_warning "Google Cloud logging is disabled"
        print_info "Enable with: export ENABLE_GOOGLE_LOGGING=true"
    fi
    
    return 0
}

# Test logging functionality
test_logging() {
    print_info "Testing Google Cloud logging..."
    
    if [ "$ENABLE_GOOGLE_LOGGING" != "true" ]; then
        print_warning "Google Cloud logging is disabled - skipping test"
        return 0
    fi
    
    # Run test if Google logger test exists
    if [ -f "test-google-cloud-logging.sh" ]; then
        print_info "Running Google Cloud logging test..."
        if ./test-google-cloud-logging.sh; then
            print_status "Google Cloud logging test passed"
            return 0
        else
            print_error "Google Cloud logging test failed"
            return 1
        fi
    else
        print_warning "Google Cloud logging test script not found"
        return 1
    fi
}

# Generate configuration summary
generate_summary() {
    echo ""
    print_info "Configuration Summary:"
    echo "========================"
    
    local project_id=$(gcloud config get-value project 2>/dev/null || echo "Not configured")
    local account=$(gcloud config get-value account 2>/dev/null || echo "Not configured")
    local GOOGLE_CLOUD_PROJECT=${GOOGLE_CLOUD_PROJECT:-Not set}
    local ENABLE_GOOGLE_LOGGING=${ENABLE_GOOGLE_LOGGING:-false}
    local GOOGLE_APPLICATION_CREDENTIALS=${GOOGLE_APPLICATION_CREDENTIALS:-Not set}
    
    echo "Project ID: $project_id"
    echo "Account: $account"
    echo "Environment GOOGLE_CLOUD_PROJECT: $GOOGLE_CLOUD_PROJECT"
    echo "Environment ENABLE_GOOGLE_LOGGING: $ENABLE_GOOGLE_LOGGING"
    echo "Credentials File: $GOOGLE_APPLICATION_CREDENTIALS"
    
    if [ -f ".env" ]; then
        echo ""
        print_info "Environment file (.env) found"
        grep -E "^GOOGLE_" .env || print_warning "No Google Cloud variables found in .env"
    else
        print_warning "No .env file found"
    fi
}

# Main validation function
main() {
    echo ""
    print_info "Starting Google Cloud configuration validation..."
    echo ""
    
    local failed_checks=0
    
    # Run all checks
    if ! check_gcloud_cli; then ((failed_checks++)); fi
    if ! check_authentication; then ((failed_checks++)); fi
    if ! check_project; then ((failed_checks++)); fi
    if ! check_apis; then ((failed_checks++)); fi
    if ! check_service_account; then ((failed_checks++)); fi
    if ! check_app_config; then ((failed_checks++)); fi
    
    # Test logging if enabled
    if [ "$ENABLE_GOOGLE_LOGGING" = "true" ]; then
        if ! test_logging; then ((failed_checks++)); fi
    fi
    
    echo ""
    generate_summary
    
    if [ $failed_checks -eq 0 ]; then
        echo ""
        print_status "All checks passed! Google Cloud is properly configured."
        print_info "You can now run the bot with Google Cloud logging enabled."
        echo ""
        print_info "Next steps:"
        echo "1. Deploy: ./docker-deploy.sh production"
        echo "2. Monitor: gcloud logging tail \"resource.type=container\""
        echo "3. Dashboard: https://console.cloud.google.com/logging"
    else
        echo ""
        print_error "$failed_checks check(s) failed. Please fix the issues above."
        echo ""
        print_info "For help, see: docs/GOOGLE_CLOUD_SETUP.md"
        return 1
    fi
}

# Command line interface
case "${1:-full}" in
    "cli")
        check_gcloud_cli
        ;;
    "auth")
        check_authentication
        ;;
    "project")
        check_project
        ;;
    "apis")
        check_apis
        ;;
    "service-account"|"sa")
        check_service_account
        ;;
    "config")
        check_app_config
        ;;
    "test")
        test_logging
        ;;
    "summary")
        generate_summary
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [check|summary|help]"
        echo ""
        echo "Checks:"
        echo "  cli            Check gcloud CLI installation"
        echo "  auth           Check authentication"
        echo "  project        Check project configuration"
        echo "  apis           Check required APIs"
        echo "  service-account Check service account configuration"
        echo "  config         Check application configuration"
        echo "  test           Test logging functionality"
        echo "  summary        Show configuration summary"
        echo ""
        echo "Default: Run all checks"
        ;;
    *)
        main
        ;;
esac