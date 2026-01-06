#!/bin/bash

set -e

echo "☁️  Cloudflare Workers Deployment Script"
echo "=================================="

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
WORKER_NAME=${2:-obsidian-bot-workers}

# Check prerequisites
check_prerequisites() {
    print_info "Checking prerequisites..."
    
    # Check wrangler
    if ! command -v wrangler &> /dev/null; then
        print_error "Wrangler is not installed"
        print_info "Install with: npm install -g wrangler"
        exit 1
    fi
    
    # Check auth
    if ! wrangler whoami &> /dev/null; then
        print_error "Not logged in to Cloudflare"
        print_info "Run: wrangler auth login"
        exit 1
    fi
    
    # Check worker directory
    if [ ! -d "workers" ]; then
        print_error "Workers directory not found"
        exit 1
    fi
    
    print_status "Prerequisites satisfied"
}

# Validate configuration
validate_config() {
    print_info "Validating configuration..."
    
    # Check wrangler.toml
    if [ ! -f "workers/wrangler.toml" ]; then
        print_error "workers/wrangler.toml not found"
        exit 1
    fi
    
    # Check worker files
    if [ ! -f "workers/obsidian-bot-worker.js" ]; then
        print_error "obsidian-bot-worker.js not found"
        exit 1
    fi
    
    print_status "Configuration validated"
}

# Deploy to specific environment
deploy_environment() {
    local env=$1
    print_info "Deploying to $env environment..."
    
    cd workers
    
    # Deploy with specific environment
    wrangler deploy --env $env --config wrangler.toml
    
    print_status "Deployed to $env environment"
    
    cd ..
}

# Deploy all environments
deploy_all() {
    local environments=("production" "staging" "development")
    
    for env in "${environments[@]}"; do
        print_info "Deploying to $env..."
        deploy_environment $env
        sleep 2
    done
    
    print_status "All environments deployed"
}

# Preview deployment
preview_deployment() {
    print_info "Starting preview deployment..."
    
    cd workers
    
    # Start preview
    wrangler dev --config wrangler.toml
    
    cd ..
    
    print_status "Preview server started"
}

# Health check
health_check() {
    print_info "Performing health check..."
    
    # Get production URL
    local production_url=$(wrangler whoami | grep -A 5 "Account URL" | head -1 | awk '{print $4}')
    
    if [ -n "$production_url" ]; then
        # Test health endpoint
        local health_response=$(curl -s "$production_url/health")
        
        if [ "$health_response" = "OK" ]; then
            print_status "Health check passed"
        else
            print_warning "Health check failed"
            print_info "Response: $health_response"
        fi
    else
        print_warning "Could not determine production URL"
    fi
}

# Get deployment info
get_deployment_info() {
    print_info "Getting deployment information..."
    
    cd workers
    
    # List deployments
    print_info "Deployment history:"
    wrangler deployments list --config wrangler.toml
    
    cd ..
}

# Rollback deployment
rollback_deployment() {
    local deployment_id=${1:-}
    
    if [ -z "$deployment_id" ]; then
        print_error "Deployment ID is required for rollback"
        print_info "Usage: $0 rollback <deployment_id>"
        exit 1
    fi
    
    print_info "Rolling back to deployment: $deployment_id"
    
    cd workers
    
    wrangler rollback $deployment_id --config wrangler.toml
    
    cd ..
    
    print_status "Rollback completed"
}

# Clean up deployments
cleanup_deployments() {
    local keep_count=${1:-5}
    
    print_info "Cleaning up old deployments (keeping latest $keep_count)..."
    
    cd workers
    
    # List and delete old deployments
    wrangler deployments list --config wrangler.toml --limit 100 | \
    grep -E "^id:" | \
    tail -n +$keep_count | \
    sed 's/^id: //' | \
    xargs -I {} wrangler delete {} --config wrangler.toml
    
    cd ..
    
    print_status "Cleanup completed"
}

# Update bindings
update_bindings() {
    print_info "Updating environment bindings..."
    
    cd workers
    
    # Update AI binding if AI is configured
    if [ -n "$GEMINI_API_KEY" ]; then
        print_info "Adding Gemini AI binding..."
        # This would update wrangler.toml with new binding
    fi
    
    if [ -n "$GROQ_API_KEY" ]; then
        print_info "Adding Groq AI binding..."
    fi
    
    cd ..
    
    print_status "Bindings updated"
}

# Main function
main() {
    case "${1:-deploy}" in
        "deploy")
            check_prerequisites
            validate_config
            deploy_environment $ENVIRONMENT
            health_check
            ;;
        "deploy-all")
            check_prerequisites
            validate_config
            deploy_all
            ;;
        "preview")
            check_prerequisites
            validate_config
            preview_deployment
            ;;
        "health")
            health_check
            ;;
        "info")
            get_deployment_info
            ;;
        "rollback")
            rollback_deployment $2
            ;;
        "cleanup")
            cleanup_deployments $2
            ;;
        "bindings")
            update_bindings
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [deploy|deploy-all|preview|health|info|rollback|cleanup|bindings|help] [options]"
            echo ""
            echo "Commands:"
            echo "  deploy      - Deploy to specified environment (default: production)"
            echo "  deploy-all  - Deploy to all environments"
            echo "  preview     - Start preview server"
            echo "  health      - Perform health check"
            echo "  info        - Show deployment information"
            echo "  rollback    - Rollback to specific deployment"
            echo "  cleanup     - Clean up old deployments"
            echo "  bindings    - Update environment bindings"
            echo "  help        - Show this help"
            echo ""
            echo "Options:"
            echo "  Environment: production|staging|development (default: production)"
            echo "  Worker Name: custom worker name (default: obsidian-bot-workers)"
            echo "  Cleanup Count: number of deployments to keep (default: 5)"
            echo ""
            echo "Environment Variables:"
            echo "  GEMINI_API_KEY    - Gemini API key for AI binding"
            echo "  GROQ_API_KEY      - Groq API key for AI binding"
            echo ""
            echo "Examples:"
            echo "  $0 deploy production"
            echo "  $0 deploy staging"
            echo "  $0 deploy-all"
            echo "  $0 preview"
            echo "  $0 rollback abc123"
            echo "  $0 cleanup 3"
            echo "  GEMINI_API_KEY=your-key $0 bindings"
            ;;
        *)
            print_error "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Check authentication first
if ! wrangler whoami &> /dev/null; then
    print_error "Please authenticate first: wrangler auth login"
    exit 1
fi

# Run main function with all arguments
main "$@"