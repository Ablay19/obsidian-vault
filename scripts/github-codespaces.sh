#!/bin/bash

echo "🐙 GitHub Codespaces SSH Setup"
echo "========================="

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

# Check if gh CLI is installed
check_gh_cli() {
    if ! command -v gh &> /dev/null; then
        print_error "GitHub CLI (gh) is not installed"
        print_info "Install with: curl -fsSL https://cli.github.com/packages/github-cli.sh | bash"
        exit 1
    fi
    
    print_status "GitHub CLI found: $(gh version --cli)"
}

# Check authentication
check_authentication() {
    if ! gh auth status &> /dev/null; then
        print_error "Not logged in to GitHub"
        print_info "Run: gh auth login"
        exit 1
    fi
    
    local auth_status=$(gh auth status)
    echo "$auth_status"
}

# Create Codespace
create_codespace() {
    local codespace_name=${1:-obsidian-bot-codespace}
    local repository=${2:-Ablay19/obsidian-vault}
    local branch=${3:-main}
    
    print_info "Creating Codespace: $codespace_name"
    print_info "Repository: $repository"
    print_info "Branch: $branch"
    
    # Create Codespace with your bot
    gh codespace create \
        --repo "$repository" \
        --branch "$branch" \
        --machine "standardLinux_x64" \
        "$codespace_name" \
        --displayName "Obsidian Bot Development" \
        --idle-timeout 120 \
        --retain-period 1d \
        --cpu 2 \
        --memory 4gb \
        --disk 10gb
    
    print_status "Codespace created"
    
    # Wait for creation
    sleep 5
    
    # Get Codespace information
    local codespace_info=$(gh codespace view "$codespace_name" --json)
    local codespace_id=$(echo "$codespace_info" | jq -r '.id')
    local container_id=$(echo "$codespace_info" | jq -r '.container.id')
    
    print_status "Codespace ID: $codespace_id"
    print_status "Container ID: $container_id"
    
    # SSH into the Codespace
    print_info "SSHing into Codespace..."
    gh codespace ssh --container "$container_id" --machine "$codespace_id" --repo "$repository"
}

# Setup environment in Codespace
setup_environment() {
    local codespace_name=${1:-obsidian-bot-codespace}
    
    print_info "Setting up environment in Codespace..."
    
    # SSH into Codespace
    gh codespace ssh --container "$(gh codespace view "$codespace_name" --json | jq -r '.container.id')" --machine "$codespace_name"
    
    print_status "Connected to Codespace terminal"
    
    # Once inside, run these commands:
    print_info "Once connected, run these commands:"
    echo ""
    echo "  # Clone your repository (if not done automatically)"
    echo "  git clone https://github.com/Ablay19/obsidian-vault.git"
    echo ""
    echo "  # Set up your environment"
    echo "  export TELEGRAM_BOT_TOKEN='your-token'"
    echo "  export TURSO_DATABASE_URL='your-db-url'"
    echo "  export TURSO_AUTH_TOKEN='your-auth-token'"
    echo ""
    echo "  # Run your bot"
    echo "  ./docker-deploy.sh development"
    echo ""
    echo "  # Or deploy to Cloudflare Workers"
    echo "  cd workers && ./deploy.sh production"
}

# List existing Codespaces
list_codespaces() {
    print_info "Listing existing Codespaces:"
    gh codespace list --json | jq -r '.[].name'
}

# Delete Codespace
delete_codespace() {
    local codespace_name=${1:-}
    
    if [ -z "$codespace_name" ]; then
        print_error "Codespace name is required"
        print_info "Usage: $0 delete <codespace_name>"
        exit 1
    fi
    
    print_info "Deleting Codespace: $codespace_name"
    gh codespace delete --codespace "$codespace_name" --force
    
    print_status "Codespace deleted"
}

# Start Codespace
start_codespace() {
    local codespace_name=${1:-}
    
    if [ -z "$codespace_name" ]; then
        print_error "Codespace name is required"
        print_info "Usage: $0 start <codespace_name>"
        exit 1
    fi
    
    print_info "Starting Codespace: $codespace_name"
    gh codespace start --codespace "$codespace_name"
    
    print_status "Codespace started"
}

# Stop Codespace
stop_codespace() {
    local codespace_name=${1:-}
    
    if [ -z "$codespace_name" ]; then
        print_error "Codespace name is required"
        print_info "Usage: $0 stop <codespace_name>"
        exit 1
    fi
    
    print_info "Stopping Codespace: $codespace_name"
    gh codespace stop --codespace "$codespace_name"
    
    print_status "Codespace stopped"
}

# Deploy bot in Codespace
deploy_bot_in_codespace() {
    local codespace_name=${1:-obsidian-bot-codespace}
    
    print_info "Setting up bot deployment in Codespace..."
    
    # SSH into Codespace and run deployment
    gh codespace ssh --container "$(gh codespace view "$codespace_name" --json | jq -r '.container.id')" --machine "$codespace_name" --repo "Ablay19/obsidian-vault" -- \
        "cd obsidian-vault && echo 'Setting up environment...' && \
        export TELEGRAM_BOT_TOKEN='\$TELEGRAM_BOT_TOKEN' && \
        export TURSO_DATABASE_URL='\$TURSO_DATABASE_URL' && \
        export TURSO_AUTH_TOKEN='\$TURSO_AUTH_TOKEN' && \
        echo 'Environment set up' && \
        echo 'Starting deployment...' && \
        ./docker-deploy.sh development"
    
    print_status "Deployment command sent to Codespace"
}

# Main function
main() {
    case "${1:-create}" in
        "create")
            check_gh_cli
            check_authentication
            create_codespace "$2" "$3" "$4"
            ;;
        "ssh")
            setup_environment "$2"
            ;;
        "list")
            list_codespaces
            ;;
        "delete")
            delete_codespace "$2"
            ;;
        "start")
            start_codespace "$2"
            ;;
        "stop")
            stop_codespace "$2"
            ;;
        "deploy")
            deploy_bot_in_codespace "$2"
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [create|ssh|list|delete|start|stop|deploy|help] [options]"
            echo ""
            echo "Commands:"
            echo "  create  - Create new Codespace (default)"
            echo "  ssh     - SSH into existing Codespace"
            echo "  list    - List existing Codespaces"
            echo "  delete  - Delete Codespace"
            echo "  start   - Start stopped Codespace"
            echo "  stop    - Stop Codespace"
            echo "  deploy  - Deploy bot in Codespace"
            echo "  help    - Show this help"
            echo ""
            echo "Examples:"
            echo "  $0 create                    # Create with defaults"
            echo "  $0 create my-dev-space        # Custom name"
            echo "  $0 create my-dev-space Ablay19/obsidian-vault develop  # Custom repo and branch"
            echo "  $0 ssh my-dev-space        # SSH into existing space"
            echo "  $0 deploy my-dev-space      # Deploy bot in Codespace"
            echo ""
            echo "Options:"
            echo "  Repository name (default: Ablay19/obsidian-vault)"
            echo "  Branch name (default: main)"
            echo "  Codespace name (default: obsidian-bot-codespace)"
            echo ""
            echo "Features:"
            echo "  • Free GitHub Codespaces (60 hours/month)"
            echo "  • 2 CPU cores, 4GB RAM, 10GB storage"
            echo "  • Pre-configured with common development tools"
            echo "  • Integrated with GitHub Actions"
            ;;
        *)
            print_error "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
}

# Run main function
if [ $# -gt 0 ]; then
    main "$@"
else
    main "create"
fi