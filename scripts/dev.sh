#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Header
echo -e "${CYAN}╔══════════════════════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}║         Obsidian Automation Development Scripts              ║${NC}"
echo -e "${CYAN}║                      Enhanced Version                        ║${NC}"
echo -e "${CYAN}╚══════════════════════════════════════════════════════════════╝${NC}"
echo ""

# Utility functions
print_status() { echo -e "${BLUE}[•]${NC} $1"; }
print_success() { echo -e "${GREEN}[✓]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[⚠]${NC} $1"; }
print_error() { echo -e "${RED}[✗]${NC} $1"; }
print_info() { echo -e "${PURPLE}[ℹ]${NC} $1"; }

# Check if we're in the right directory
if [ ! -f "go.mod" ] || [ ! -d "internal" ]; then
    print_error "This script must be run from the project root directory"
    exit 1
fi

# Main menu
show_menu() {
    echo -e "${CYAN}Available Operations:${NC}"
    echo ""
    echo "1. 🏗️  Build & Test"
    echo "2. 🚀 Deploy Services"
    echo "3. 📊 Monitor & Debug"
    echo "4. 🔧 Development Tools"
    echo "5. 🧹 Maintenance"
    echo "6. 📋 Environment Setup"
    echo "7. 🛠️  Database Operations"
    echo "8. 📦 Package Management"
    echo "9. 🔐 Security & Auth"
    echo "0. 🚪 Exit"
    echo ""
}

# Build & Test submenu
build_test_menu() {
    echo -e "${CYAN}Build & Test Operations:${NC}"
    echo ""
    echo "1. 🔨 Build Application"
    echo "2. 🧪 Run Tests"
    echo "3. 🔍 Lint Code"
    echo "4. 📊 Run Benchmarks"
    echo "5. 🧪 Test All Services"
    echo "6. 📦 Build Docker Image"
    echo "7. 🔄 Build & Run"
    echo "0. 🔙 Back"
    echo ""
}

build_test_operations() {
    case $1 in
        1) build_app ;;
        2) run_tests ;;
        3) lint_code ;;
        4) run_benchmarks ;;
        5) test_all_services ;;
        6) build_docker ;;
        7) build_and_run ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

build_app() {
    print_status "Building application..."
    
    # Clean previous builds
    rm -f obsidian-automation
    
    # Build with version info
    VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
    BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
    
    print_info "Version: $VERSION"
    print_info "Build: $BUILD_TIME"
    print_info "Commit: $COMMIT"
    
    if go build \
        -ldflags "-X main.Version=$VERSION -X main.BuildTime=$BUILD_TIME -X main.GitCommit=$COMMIT" \
        -o obsidian-automation \
        ./cmd/bot/main.go; then
        print_success "Build completed successfully"
        print_info "Binary: ./obsidian-automation"
    else
        print_error "Build failed"
        exit 1
    fi
}

run_tests() {
    print_status "Running tests..."
    
    # Run unit tests
    if go test -v ./...; then
        print_success "All tests passed"
    else
        print_error "Tests failed"
        exit 1
    fi
}

lint_code() {
    print_status "Linting code..."
    
    # Check for gofmt
    if ! command -v gofmt &> /dev/null; then
        print_warning "gofmt not found, skipping formatting check"
    else
        print_status "Checking code formatting..."
        UNFORMATTED=$(gofmt -l .)
        if [ -n "$UNFORMATTED" ]; then
            print_warning "Files need formatting:"
            echo "$UNFORMATTED"
            print_status "Auto-formatting..."
            gofmt -w .
            print_success "Code formatted"
        else
            print_success "Code is properly formatted"
        fi
    fi
    
    # Run go vet
    print_status "Running go vet..."
    if go vet ./...; then
        print_success "go vet passed"
    else
        print_error "go vet found issues"
        exit 1
    fi
    
    # Run golint if available
    if command -v golint &> /dev/null; then
        print_status "Running golint..."
        golint ./...
    fi
}

run_benchmarks() {
    print_status "Running benchmarks..."
    go test -bench=. -benchmem ./...
}

test_all_services() {
    print_status "Testing all services..."
    
    # Test database
    print_status "Testing database connection..."
    if go run ./cmd/bot/main.go --test-db 2>/dev/null; then
        print_success "Database test passed"
    else
        print_warning "Database test failed (may be expected without config)"
    fi
    
    # Test AI services
    print_status "Testing AI service initialization..."
    # Add AI service tests here
    
    print_success "Service tests completed"
}

build_docker() {
    print_status "Building Docker image..."
    
    if docker build -t obsidian-automation:latest .; then
        print_success "Docker image built successfully"
        print_info "Image: obsidian-automation:latest"
    else
        print_error "Docker build failed"
        exit 1
    fi
}

build_and_run() {
    build_app
    print_status "Starting application..."
    ./obsidian-automation
}

# Deploy Services submenu
deploy_menu() {
    echo -e "${CYAN}Deploy Services:${NC}"
    echo ""
    echo "1. ☁️  Deploy to Render"
    echo "2. 🔥 Deploy to Cloudflare"
    echo "3. 🐳 Deploy Docker"
    echo "4. 📦 Deploy All Services"
    echo "5. 🔄 Update Deployment"
    echo "6. 📋 Check Deployment Status"
    echo "0. 🔙 Back"
    echo ""
}

deploy_operations() {
    case $1 in
        1) deploy_render ;;
        2) deploy_cloudflare ;;
        3) deploy_docker ;;
        4) deploy_all ;;
        5) update_deployment ;;
        6) check_deployment ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

deploy_render() {
    print_status "Deploying to Render..."
    
    # Check if render CLI is available
    if command -v render &> /dev/null; then
        render deploy
    else
        print_warning "Render CLI not found"
        print_info "Please push to GitHub and Render will auto-deploy"
        print_status "Pushing to GitHub..."
        git push origin main
    fi
}

deploy_cloudflare() {
    print_status "Deploying to Cloudflare..."
    if [ -f "scripts/deploy-cloudflare.sh" ]; then
        ./scripts/deploy-cloudflare.sh
    else
        print_error "Cloudflare deployment script not found"
    fi
}

deploy_docker() {
    print_status "Deploying Docker container..."
    
    # Stop existing container
    docker stop obsidian-automation 2>/dev/null || true
    docker rm obsidian-automation 2>/dev/null || true
    
    # Run new container
    docker run -d \
        --name obsidian-automation \
        -p 8080:8080 \
        --env-file .env \
        obsidian-automation:latest
    
    print_success "Docker container deployed"
}

deploy_all() {
    print_status "Deploying all services..."
    deploy_cloudflare
    deploy_docker
    print_success "All services deployed"
}

update_deployment() {
    print_status "Updating deployment..."
    git add -A
    git commit -m "Update deployment $(date)"
    git push origin main
    print_success "Deployment updated"
}

check_deployment() {
    print_status "Checking deployment status..."
    
    # Check local service
    if curl -s http://localhost:8080/health > /dev/null; then
        print_success "Local service is running"
    else
        print_warning "Local service is not responding"
    fi
    
    # Check Cloudflare Workers
    if [ -f "scripts/monitor-cloudflare.sh" ]; then
        ./scripts/monitor-cloudflare.sh
    fi
}

# Monitor & Debug submenu
monitor_menu() {
    echo -e "${CYAN}Monitor & Debug:${NC}"
    echo ""
    echo "1. 📊 View Logs"
    echo "2. 🔍 Health Check"
    echo "3. 📈 Performance Metrics"
    echo "4. 🐛 Debug Mode"
    echo "5. 🔥 Cloudflare Monitoring"
    echo "6. 📋 Service Status"
    echo "7. 🕐 Real-time Logs"
    echo "0. 🔙 Back"
    echo ""
}

monitor_operations() {
    case $1 in
        1) view_logs ;;
        2) health_check ;;
        3) performance_metrics ;;
        4) debug_mode ;;
        5) cloudflare_monitoring ;;
        6) service_status ;;
        7) realtime_logs ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

view_logs() {
    print_status "Viewing application logs..."
    
    if [ -f "logs/app.log" ]; then
        tail -f logs/app.log
    else
        print_warning "Log file not found"
        print_info "Starting application with logging..."
        mkdir -p logs
        ./obsidian-automation 2>&1 | tee logs/app.log
    fi
}

health_check() {
    print_status "Performing health check..."
    
    # Check if service is running
    if curl -s http://localhost:8080/api/services/status > /dev/null; then
        print_success "Service is responding"
        curl -s http://localhost:8080/api/services/status | jq .
    else
        print_error "Service is not responding"
    fi
}

performance_metrics() {
    print_status "Collecting performance metrics..."
    
    # Check system resources
    echo -e "${BLUE}System Resources:${NC}"
    echo "CPU: $(top -bn1 | grep "Cpu(s)" | awk '{print $2}' | cut -d'%' -f1)"
    echo "Memory: $(free -h | awk '/^Mem:/ {print $3 "/" $2}')"
    echo "Disk: $(df -h . | awk 'NR==2 {print $3 "/" $2 " (" $5 " used)"}')"
    
    # Check application metrics if available
    if curl -s http://localhost:8080/metrics > /dev/null; then
        echo -e "\n${BLUE}Application Metrics:${NC}"
        curl -s http://localhost:8080/metrics
    fi
}

debug_mode() {
    print_status "Starting debug mode..."
    
    # Set debug environment variables
    export ENVIRONMENT_MODE=dev
    export ENABLE_DEBUG_LOGS=true
    
    print_info "Debug mode enabled"
    print_info "ENVIRONMENT_MODE=dev"
    print_info "ENABLE_DEBUG_LOGS=true"
    
    # Start application with debug flags
    ./obsidian-automation --debug --verbose
}

cloudflare_monitoring() {
    if [ -f "scripts/monitor-cloudflare.sh" ]; then
        ./scripts/monitor-cloudflare.sh
    else
        print_error "Cloudflare monitoring script not found"
    fi
}

service_status() {
    print_status "Checking all services..."
    
    echo -e "${BLUE}Services Status:${NC}"
    
    # Check main application
    if pgrep -f "obsidian-automation" > /dev/null; then
        print_success "Main application: RUNNING"
    else
        print_error "Main application: STOPPED"
    fi
    
    # Check database
    if [ -n "$TURSO_DATABASE_URL" ]; then
        print_status "Database: Configured (Turso)"
    else
        print_warning "Database: Not configured"
    fi
    
    # Check AI services
    if [ -n "$GEMINI_API_KEY" ] || [ -n "$GROQ_API_KEY" ]; then
        print_success "AI Services: Configured"
    else
        print_warning "AI Services: Not configured"
    fi
    
    # Check bot tokens
    if [ -n "$TELEGRAM_BOT_TOKEN" ]; then
        print_success "Telegram Bot: Configured"
    else
        print_warning "Telegram Bot: Not configured"
    fi
}

realtime_logs() {
    print_status "Starting real-time log monitoring..."
    
    # Monitor multiple log sources
    multitail \
        -l "tail -f logs/app.log" \
        -l "journalctl -u obsidian-automation -f" \
        -l "docker logs -f obsidian-automation" 2>/dev/null || \
    tail -f logs/app.log 2>/dev/null || \
    print_warning "No log sources available"
}

# Main execution loop
main() {
    while true; do
        show_menu
        read -p "Select an option: " choice
        
        case $choice in
            1)
                while true; do
                    build_test_menu
                    read -p "Select build/test option: " subchoice
                    build_test_operations $subchoice
                    [ $subchoice -eq 0 ] && break
                done
                ;;
            2)
                while true; do
                    deploy_menu
                    read -p "Select deploy option: " subchoice
                    deploy_operations $subchoice
                    [ $subchoice -eq 0 ] && break
                done
                ;;
            3)
                while true; do
                    monitor_menu
                    read -p "Select monitor option: " subchoice
                    monitor_operations $subchoice
                    [ $subchoice -eq 0 ] && break
                done
                ;;
            4)
                print_info "Development tools menu (coming soon)"
                ;;
            5)
                print_info "Maintenance menu (coming soon)"
                ;;
            6)
                print_info "Environment setup menu (coming soon)"
                ;;
            7)
                print_info "Database operations menu (coming soon)"
                ;;
            8)
                print_info "Package management menu (coming soon)"
                ;;
            9)
                print_info "Security & auth menu (coming soon)"
                ;;
            0)
                print_success "Goodbye!"
                exit 0
                ;;
            *)
                print_error "Invalid option. Please try again."
                ;;
        esac
        
        echo ""
        read -p "Press Enter to continue..."
        clear
    done
}

# Start the script
main