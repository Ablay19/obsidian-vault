#!/bin/bash

set -e

echo "🐳 Obsidian Bot Docker Deployment System"
echo "======================================="

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

# Check if Docker is installed
check_docker() {
    if ! command -v docker &> /dev/null; then
        print_error "Docker is not installed. Please install Docker first."
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        print_error "Docker Compose is not installed. Please install Docker Compose first."
        exit 1
    fi
    
    print_status "Docker and Docker Compose are installed"
}

# Check if environment is configured
check_environment() {
    if [ ! -f ".env" ]; then
        print_warning ".env file not found. Creating from template..."
        if [ -f ".env.docker" ]; then
            cp .env.docker .env
            print_info "Created .env from .env.docker template"
            print_warning "Please edit .env file with your configuration before running deployment"
        else
            print_error "No .env.docker template found. Please create .env file manually."
            exit 1
        fi
    fi
    
    # Check critical environment variables
    source .env
    
    local missing_vars=()
    
    [ -z "$TELEGRAM_BOT_TOKEN" ] && missing_vars+=("TELEGRAM_BOT_TOKEN")
    [ -z "$TURSO_DATABASE_URL" ] && missing_vars+=("TURSO_DATABASE_URL")
    [ -z "$SESSION_SECRET" ] && missing_vars+=("SESSION_SECRET")
    
    if [ ${#missing_vars[@]} -gt 0 ]; then
        print_warning "Missing critical environment variables: ${missing_vars[*]}"
        print_warning "Please update .env file with these values"
        return 1
    fi
    
    print_status "Environment configuration validated"
    return 0
}

# Build Docker images
build_images() {
    print_info "Building Docker images..."
    
    if [ "$1" = "--no-cache" ]; then
        docker-compose build --no-cache
        print_status "Docker images built (no cache)"
    else
        docker-compose build
        print_status "Docker images built"
    fi
}

# Start services
start_services() {
    print_info "Starting Obsidian Bot services..."
    
    if [ -n "$COMPOSE_PROFILES" ]; then
        print_info "Starting with profiles: $COMPOSE_PROFILES"
        docker-compose --profile "${COMPOSE_PROFILES}" up -d
    else
        docker-compose up -d
    fi
    
    print_status "Services started"
}

# Stop services
stop_services() {
    print_info "Stopping Obsidian Bot services..."
    docker-compose down
    print_status "Services stopped"
}

# Show service status
show_status() {
    print_info "Service Status:"
    docker-compose ps
    
    print_info "Service Logs (last 20 lines):"
    docker-compose logs --tail=20 bot
}

# Health check
health_check() {
    print_info "Performing health check..."
    
    # Wait for services to be ready
    sleep 10
    
    # Check if main service is running
    if docker-compose ps | grep -q "Up.*obsidian-bot"; then
        print_status "Obsidian Bot container is running"
    else
        print_error "Obsidian Bot container is not running"
        return 1
    fi
    
    # Check if health endpoint responds
    if curl -f http://localhost:${BOT_PORT:-8080}/api/services/status > /dev/null 2>&1; then
        print_status "Health endpoint is responding"
    else
        print_warning "Health endpoint is not responding yet (still starting up)"
    fi
}

# Cleanup function
cleanup() {
    print_info "Cleaning up Docker resources..."
    docker-compose down -v --remove-orphans
    docker system prune -f
    print_status "Cleanup completed"
}

# Production deployment
deploy_production() {
    print_info "Starting production deployment..."
    
    # Validate environment
    if ! check_environment; then
        print_error "Environment validation failed"
        exit 1
    fi
    
    # Build images
    build_images
    
    # Start services
    start_services
    
    # Health check
    health_check
    
    print_status "Production deployment completed!"
    print_info "Dashboard: http://localhost:${BOT_PORT:-8080}/dashboard/overview"
}

# Development setup
setup_development() {
    print_info "Setting up development environment..."
    
    # Check environment
    check_environment
    
    # Build and start
    build_images
    start_services
    
    # Show status
    show_status
    
    print_status "Development environment ready!"
}

# Main menu
main_menu() {
    echo ""
    print_info "Docker Deployment Options:"
    echo "1. Production Deployment"
    echo "2. Development Setup"
    echo "3. Start Services"
    echo "4. Stop Services"
    echo "5. Show Status"
    echo "6. Health Check"
    echo "7. Rebuild Images (--no-cache)"
    echo "8. Cleanup Resources"
    echo "9. Exit"
    echo ""
    
    read -p "Choose an option (1-9): " choice
    
    case $choice in
        1)
            deploy_production
            ;;
        2)
            setup_development
            ;;
        3)
            start_services
            ;;
        4)
            stop_services
            ;;
        5)
            show_status
            ;;
        6)
            health_check
            ;;
        7)
            read -p "Use --no-cache? (y/N): " nocache
            if [[ $nocache =~ ^[Yy]$ ]]; then
                build_images --no-cache
            else
                build_images
            fi
            ;;
        8)
            cleanup
            ;;
        9)
            print_info "Exiting..."
            exit 0
            ;;
        *)
            print_error "Invalid option"
            main_menu
            ;;
    esac
}

# Main script logic
if [ $# -eq 0 ]; then
    check_docker
    main_menu
else
    case $1 in
        "production"|"prod")
            deploy_production
            ;;
        "development"|"dev")
            setup_development
            ;;
        "start")
            start_services
            ;;
        "stop")
            stop_services
            ;;
        "status")
            show_status
            ;;
        "health")
            health_check
            ;;
        "build")
            build_images "$2"
            ;;
        "cleanup")
            cleanup
            ;;
        "help"|"-h"|"--help")
            echo "Usage: $0 [command]"
            echo ""
            echo "Commands:"
            echo "  production, prod    - Production deployment"
            echo "  development, dev    - Development setup"
            echo "  start              - Start services"
            echo "  stop               - Stop services"
            echo "  status             - Show service status"
            echo "  health             - Health check"
            echo "  build [--no-cache] - Build Docker images"
            echo "  cleanup            - Cleanup Docker resources"
            echo "  help               - Show this help"
            echo ""
            echo "Environment variables:"
            echo "  COMPOSE_PROFILES   - Additional profiles (ssh, vault)"
            ;;
        *)
            print_error "Unknown command: $1"
            echo "Use '$0 help' for usage information"
            exit 1
            ;;
    esac
fi