#!/bin/bash

set -e

echo "🔐 Kubernetes Secrets Management"
echo "============================="

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
ENVIRONMENT=${1:-production}
NAMESPACE=${2:-obsidian-system}

# Encode function
base64_encode() {
    echo -n "$1" | base64 -w 0
}

# Create secrets from environment variables
create_secrets_from_env() {
    print_info "Creating secrets from environment variables..."
    
    # Check required environment variables
    local required_vars=(
        "TURSO_DATABASE_URL"
        "TURSO_AUTH_TOKEN"
        "TELEGRAM_BOT_TOKEN"
        "SESSION_SECRET"
    )
    
    local missing_vars=()
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            missing_vars+=("$var")
        fi
    done
    
    if [ ${#missing_vars[@]} -gt 0 ]; then
        print_error "Missing required environment variables:"
        for var in "${missing_vars[@]}"; do
            echo "  - $var"
        done
        print_info "Set environment variables or use interactive mode"
        exit 1
    fi
    
    # Create secrets.yaml
    cat > k8s/base/secrets.yaml << EOF
apiVersion: v1
kind: Secret
metadata:
  name: obsidian-bot-secrets
  namespace: $NAMESPACE
  labels:
    app: obsidian-bot
    component: secrets
  annotations:
    secret-management: "kubectl"
type: Opaque
data:
  turso-database-url: $(base64_encode "$TURSO_DATABASE_URL")
  turso-auth-token: $(base64_encode "$TURSO_AUTH_TOKEN")
  telegram-bot-token: $(base64_encode "$TELEGRAM_BOT_TOKEN")
  session-secret: $(base64_encode "$SESSION_SECRET")
EOF
    
    # Add optional secrets if available
    if [ -n "$GEMINI_API_KEY" ]; then
        echo "  gemini-api-key: $(base64_encode "$GEMINI_API_KEY")" >> k8s/base/secrets.yaml
    fi
    
    if [ -n "$GROQ_API_KEY" ]; then
        echo "  groq-api-key: $(base64_encode "$GROQ_API_KEY")" >> k8s/base/secrets.yaml
    fi
    
    if [ -n "$CLOUDFLARE_API_KEY" ]; then
        echo "  cloudflare-api-key: $(base64_encode "$CLOUDFLARE_API_KEY")" >> k8s/base/secrets.yaml
    fi
    
    print_status "Secrets created from environment variables"
}

# Create Google Cloud credentials secret
create_gcp_credentials() {
    print_info "Creating Google Cloud credentials secret..."
    
    local gcp_credentials_file=${GOOGLE_APPLICATION_CREDENTIALS:-}
    
    if [ -z "$gcp_credentials_file" ]; then
        print_error "GOOGLE_APPLICATION_CREDENTIALS environment variable not set"
        print_info "Set with: export GOOGLE_APPLICATION_CREDENTIALS=/path/to/credentials.json"
        exit 1
    fi
    
    if [ ! -f "$gcp_credentials_file" ]; then
        print_error "Google Cloud credentials file not found: $gcp_credentials_file"
        exit 1
    fi
    
    # Create GCP credentials secret
    local gcp_credentials_base64=$(cat "$gcp_credentials_file" | base64 -w 0)
    
    cat >> k8s/base/secrets.yaml << EOF

---
apiVersion: v1
kind: Secret
metadata:
  name: obsidian-bot-gcp-credentials
  namespace: $NAMESPACE
  labels:
    app: obsidian-bot
    component: gcp-credentials
  annotations:
    secret-management: "google-cloud"
type: Opaque
data:
  credentials.json: $gcp_credentials_base64
EOF
    
    print_status "Google Cloud credentials secret created"
}

# Create TLS certificate secret
create_tls_secret() {
    print_info "Creating TLS certificate secret..."
    
    local tls_cert=${TLS_CERT_PATH:-}
    local tls_key=${TLS_KEY_PATH:-}
    
    if [ -z "$tls_cert" ] || [ -z "$tls_key" ]; then
        print_warning "TLS certificate paths not provided, skipping TLS secret"
        return
    fi
    
    if [ ! -f "$tls_cert" ] || [ ! -f "$tls_key" ]; then
        print_error "TLS certificate files not found"
        print_info "Set TLS_CERT_PATH and TLS_KEY_PATH environment variables"
        return
    fi
    
    local tls_cert_base64=$(cat "$tls_cert" | base64 -w 0)
    local tls_key_base64=$(cat "$tls_key" | base64 -w 0)
    
    cat >> k8s/base/secrets.yaml << EOF

---
apiVersion: v1
kind: Secret
metadata:
  name: obsidian-bot-tls
  namespace: $NAMESPACE
  labels:
    app: obsidian-bot
    component: tls
type: kubernetes.io/tls
data:
  tls.crt: $tls_cert_base64
  tls.key: $tls_key_base64
EOF
    
    print_status "TLS certificate secret created"
}

# Interactive secrets creation
create_secrets_interactive() {
    print_info "Interactive secrets creation..."
    
    # Backup existing secrets
    if [ -f "k8s/base/secrets.yaml" ]; then
        cp k8s/base/secrets.yaml k8s/base/secrets.yaml.backup
        print_warning "Existing secrets backed up"
    fi
    
    echo ""
    print_info "Please enter your configuration values:"
    echo "Press Enter to use default or leave empty to skip optional values"
    echo ""
    
    # Read required values
    read -p "Turso Database URL: " TURSO_DATABASE_URL
    read -p "Turso Auth Token: " TURSO_AUTH_TOKEN
    read -p "Telegram Bot Token: " TELEGRAM_BOT_TOKEN
    read -p "Session Secret: " SESSION_SECRET
    
    # Read optional values
    read -p "Gemini API Key (optional): " GEMINI_API_KEY
    read -p "Groq API Key (optional): " GROQ_API_KEY
    read -p "Cloudflare API Key (optional): " CLOUDFLARE_API_KEY
    
    # Read Google Cloud credentials
    read -p "Google Cloud Credentials JSON path (optional): " gcp_credentials_file
    if [ -n "$gcp_credentials_file" ] && [ -f "$gcp_credentials_file" ]; then
        export GOOGLE_APPLICATION_CREDENTIALS="$gcp_credentials_file"
    fi
    
    # Create secrets
    create_secrets_from_env
    if [ -n "$gcp_credentials_file" ]; then
        create_gcp_credentials
    fi
    
    echo ""
    print_status "Interactive secrets creation completed"
}

# Validate secrets
validate_secrets() {
    print_info "Validating secrets..."
    
    if [ ! -f "k8s/base/secrets.yaml" ]; then
        print_error "Secrets file not found"
        return 1
    fi
    
    # Validate YAML syntax
    if ! kubectl apply --dry-run=client -f k8s/base/secrets.yaml &> /dev/null; then
        print_error "Invalid YAML syntax in secrets file"
        return 1
    fi
    
    print_status "Secrets validation passed"
}

# Show secrets summary
show_secrets_summary() {
    print_info "Secrets summary:"
    
    if [ -f "k8s/base/secrets.yaml" ]; then
        local secret_count=$(kubectl apply --dry-run=client -f k8s/base/secrets.yaml -o json | jq '.items | length')
        echo "  Total secrets: $secret_count"
        
        # List secrets
        echo "  Secret names:"
        kubectl apply --dry-run=client -f k8s/base/secrets.yaml -o json | jq -r '.items[].metadata.name' | sed 's/^/    /  - /'
    else
        echo "  No secrets file found"
    fi
}

# Apply secrets
apply_secrets() {
    print_info "Applying secrets to cluster..."
    
    if [ ! -f "k8s/base/secrets.yaml" ]; then
        print_error "Secrets file not found"
        print_info "Create secrets first with: $0 create"
        return 1
    fi
    
    # Create namespace if it doesn't exist
    if ! kubectl get namespace $NAMESPACE &> /dev/null; then
        kubectl create namespace $NAMESPACE
        print_status "Namespace $NAMESPACE created"
    fi
    
    # Apply secrets
    kubectl apply -f k8s/base/secrets.yaml -n $NAMESPACE
    print_status "Secrets applied to namespace $NAMESPACE"
}

# Main function
main() {
    case "${1:-help}" in
        "create")
            if [ "${2:-env}" = "interactive" ]; then
                create_secrets_interactive
            else
                create_secrets_from_env
                create_gcp_credentials
                create_tls_secret
            fi
            ;;
        "validate")
            validate_secrets
            ;;
        "apply")
            apply_secrets
            ;;
        "summary")
            show_secrets_summary
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [create|validate|apply|summary|help] [options]"
            echo ""
            echo "Commands:"
            echo "  create        - Create secrets (default: from environment)"
            echo "  create interactive - Create secrets interactively"
            echo "  validate      - Validate secrets file"
            echo "  apply         - Apply secrets to cluster"
            echo "  summary       - Show secrets summary"
            echo "  help          - Show this help"
            echo ""
            echo "Environment Variables:"
            echo "  NAMESPACE      - Kubernetes namespace (default: obsidian-system)"
            echo "  ENVIRONMENT   - Environment (production|staging|development)"
            echo ""
            echo "Required Environment Variables for 'create':"
            echo "  TURSO_DATABASE_URL"
            echo "  TURSO_AUTH_TOKEN"
            echo "  TELEGRAM_BOT_TOKEN"
            echo "  SESSION_SECRET"
            echo ""
            echo "Optional Environment Variables:"
            echo "  GEMINI_API_KEY"
            echo "  GROQ_API_KEY"
            echo "  CLOUDFLARE_API_KEY"
            echo "  GOOGLE_APPLICATION_CREDENTIALS"
            echo "  TLS_CERT_PATH"
            echo "  TLS_KEY_PATH"
            echo ""
            echo "Examples:"
            echo "  $0 create"
            echo "  $0 create interactive"
            echo "  $0 validate"
            echo "  $0 apply"
            ;;
        *)
            print_error "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"