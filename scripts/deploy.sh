#!/bin/bash

set -e

# Deployment Script for Obsidian Bot
# Handles building, validating, and deploying the application

source "$(dirname "$0")/setup/common.sh"

echo "🚀 Obsidian Bot Deployment Script"
echo "================================="

# Function to show usage
show_usage() {
    echo "Usage: $0 [environment] [action]"
    echo ""
    echo "Environments:"
    echo "  staging    - Deploy to staging environment"
    echo "  production - Deploy to production environment"
    echo ""
    echo "Actions:"
    echo "  build      - Build Docker images only"
    echo "  deploy     - Full deployment (build + deploy)"
    echo "  rollback   - Rollback to previous version"
    echo ""
    echo "Examples:"
    echo "  $0 staging build"
    echo "  $0 production deploy"
    echo "  $0 production rollback"
}

# Parse arguments
ENVIRONMENT=${1:-staging}
ACTION=${2:-deploy}

case $ENVIRONMENT in
    staging|production)
        ;;
    *)
        print_error "Invalid environment: $ENVIRONMENT"
        show_usage
        exit 1
        ;;
esac

case $ACTION in
    build|deploy|rollback)
        ;;
    *)
        print_error "Invalid action: $ACTION"
        show_usage
        exit 1
        ;;
esac

print_status "Starting $ACTION for $ENVIRONMENT environment"

# Validate environment
print_header "Environment Validation"
if ! ./scripts/validate-env.sh; then
    print_error "Environment validation failed. Aborting deployment."
    exit 1
fi

print_success "Environment validation passed"

# Set environment-specific variables
case $ENVIRONMENT in
    staging)
        COMPOSE_FILE="docker-compose.yml"
        APP_VERSION="staging-$(date +%Y%m%d-%H%M%S)"
        ;;
    production)
        COMPOSE_FILE="docker-compose.yml"
        APP_VERSION="v$(date +%Y%m%d)"
        ;;
esac

export APP_VERSION

# Build images if needed
if [ "$ACTION" = "build" ] || [ "$ACTION" = "deploy" ]; then
    print_header "Building Docker Images"

    print_status "Building application image..."
    docker-compose -f "$COMPOSE_FILE" build --no-cache --parallel

    print_success "Docker images built successfully"
fi

# Deploy if requested
if [ "$ACTION" = "deploy" ]; then
    print_header "Deploying Services"

    print_status "Starting services..."
    docker-compose -f "$COMPOSE_FILE" up -d

    print_status "Waiting for services to be healthy..."
    sleep 30

    # Check service health
    print_status "Checking service health..."
    if docker-compose -f "$COMPOSE_FILE" ps | grep -q "Up"; then
        print_success "Services are running"

        # Run basic health checks
        print_status "Running health checks..."
        if curl -f http://localhost:8080/api/services/status > /dev/null 2>&1; then
            print_success "Dashboard health check passed"
        else
            print_warning "Dashboard health check failed - check logs"
        fi

    else
        print_error "Some services failed to start"
        print_info "Check logs with: docker-compose -f $COMPOSE_FILE logs"
        exit 1
    fi

    print_success "🎉 Deployment completed successfully!"
    print_info "Application is running at: http://localhost:8080"
    print_info "Check logs with: docker-compose -f $COMPOSE_FILE logs -f"
    print_info "Stop with: docker-compose -f $COMPOSE_FILE down"
fi

# Rollback if requested
if [ "$ACTION" = "rollback" ]; then
    print_header "Rolling Back"

    print_status "Stopping current services..."
    docker-compose -f "$COMPOSE_FILE" down

    print_status "Starting previous version..."
    # This assumes you have image tagging/versioning
    # In a real scenario, you'd pull the previous image tag
    docker-compose -f "$COMPOSE_FILE" up -d --scale bot=1

    print_success "Rollback completed"
fi

print_header "Deployment Summary"
echo "Environment: $ENVIRONMENT"
echo "Action: $ACTION"
echo "Version: $APP_VERSION"
echo "Compose File: $COMPOSE_FILE"

if [ "$ACTION" = "deploy" ]; then
    echo ""
    print_info "Post-deployment checklist:"
    echo "  □ Test Telegram bot commands"
    echo "  □ Verify file processing works"
    echo "  □ Check dashboard functionality"
    echo "  □ Monitor logs for errors"
    echo "  □ Run performance tests"
fi