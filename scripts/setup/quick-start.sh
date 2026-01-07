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
echo -e "${CYAN}║              Quick Start & Run Scripts                        ║${NC}"
echo -e "${CYAN}║                  One-Click Operations                        ║${NC}"
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

# Quick start menu
show_quick_menu() {
    echo -e "${CYAN}Quick Start Options:${NC}"
    echo ""
    echo "1. 🚀 Quick Start (Build & Run)"
    echo "2. 🧪 Development Mode"
    echo "3. 🏗️  Production Build"
    echo "4. 🐳 Docker Quick Start"
    echo "5. 📊 Start with Dashboard"
    echo "6. 🔧 Setup Environment"
    echo "7. 🧹 Clean & Restart"
    echo "8. 📋 Check Status"
    echo "9. 🛑 Stop Services"
    echo "0. 🚪 Exit"
    echo ""
}

# Quick start operations
quick_operations() {
    case $1 in
        1) quick_start ;;
        2) development_mode ;;
        3) production_build ;;
        4) docker_quick_start ;;
        5) start_with_dashboard ;;
        6) setup_environment ;;
        7) clean_restart ;;
        8) check_status ;;
        9) stop_services ;;
        0) exit 0 ;;
        *) print_error "Invalid option" ;;
    esac
}

quick_start() {
    print_status "Quick starting application..."
    
    # Build if needed
    if [ ! -f "obsidian-automation" ]; then
        print_status "Building application..."
        go build -o obsidian-automation ./cmd/bot/main.go
    fi
    
    # Create necessary directories
    mkdir -p logs attachments vault
    
    # Start the application
    print_success "Starting application..."
    print_info "Dashboard: http://localhost:8080"
    print_info "Press Ctrl+C to stop"
    
    ./obsidian-automation
}

development_mode() {
    print_status "Starting in development mode..."
    
    # Set development environment
    export ENVIRONMENT_MODE=dev
    export ENABLE_COLORFUL_LOGS=true
    export ENABLE_DEBUG_LOGS=true
    
    print_info "Environment: Development"
    print_info "Debug logs: Enabled"
    print_info "Colorful logs: Enabled"
    
    # Build with development flags
    print_status "Building for development..."
    go build -tags=dev -o obsidian-automation-dev ./cmd/bot/main.go
    
    # Start development server
    print_success "Starting development server..."
    ./obsidian-automation-dev
}

production_build() {
    print_status "Building for production..."
    
    # Clean previous builds
    rm -f obsidian-automation
    
    # Build with optimization flags
    print_status "Compiling with optimizations..."
    go build \
        -ldflags "-s -w" \
        -tags=prod \
        -o obsidian-automation \
        ./cmd/bot/main.go
    
    # Verify build
    if [ -f "obsidian-automation" ]; then
        print_success "Production build completed"
        print_info "Binary size: $(du -h obsidian-automation | cut -f1)"
        print_info "Run with: ./obsidian-automation"
    else
        print_error "Production build failed"
        exit 1
    fi
}

docker_quick_start() {
    print_status "Quick starting with Docker..."
    
    # Check if Docker is available
    if ! command -v docker &> /dev/null; then
        print_error "Docker not found. Please install Docker."
        exit 1
    fi
    
    # Build Docker image
    print_status "Building Docker image..."
    docker build -t obsidian-automation:latest .
    
    # Stop existing container
    docker stop obsidian-automation 2>/dev/null || true
    docker rm obsidian-automation 2>/dev/null || true
    
    # Start container
    print_status "Starting Docker container..."
    docker run -d \
        --name obsidian-automation \
        -p 8080:8080 \
        -v "$(pwd)/attachments:/app/attachments" \
        -v "$(pwd)/logs:/app/logs" \
        -v "$(pwd)/vault:/app/vault" \
        obsidian-automation:latest
    
    print_success "Docker container started"
    print_info "Container: obsidian-automation"
    print_info "Dashboard: http://localhost:8080"
    print_info "Logs: docker logs -f obsidian-automation"
}

start_with_dashboard() {
    print_status "Starting application with dashboard focus..."
    
    # Build if needed
    if [ ! -f "obsidian-automation" ]; then
        go build -o obsidian-automation ./cmd/bot/main.go
    fi
    
    # Start in background
    print_status "Starting application in background..."
    nohup ./obsidian-automation > logs/app.log 2>&1 &
    APP_PID=$!
    
    # Wait for startup
    print_status "Waiting for application to start..."
    sleep 3
    
    # Check if it's running
    if curl -s http://localhost:8080/health > /dev/null; then
        print_success "Application started successfully"
        print_info "PID: $APP_PID"
        print_info "Dashboard: http://localhost:8080"
        print_info "Logs: tail -f logs/app.log"
        
        # Open dashboard in browser if possible
        if command -v xdg-open &> /dev/null; then
            xdg-open http://localhost:8080
        elif command -v open &> /dev/null; then
            open http://localhost:8080
        fi
    else
        print_error "Application failed to start"
        kill $APP_PID 2>/dev/null || true
        exit 1
    fi
}

setup_environment() {
    print_status "Setting up environment..."
    
    # Check if .env exists
    if [ ! -f ".env" ]; then
        print_status "Creating .env file from template..."
        if [ -f ".env.example" ]; then
            cp .env.example .env
            print_success ".env file created"
        else
            print_error ".env.example not found"
            exit 1
        fi
    fi
    
    # Create necessary directories
    print_status "Creating directories..."
    mkdir -p logs attachments vault/{Inbox,Attachments,physics,math,chemistry,biology,admin,document}
    
    # Set permissions
    chmod -R 755 logs attachments vault
    
    # Check Go dependencies
    print_status "Checking Go dependencies..."
    go mod download
    go mod tidy
    
    # Validate environment
    print_status "Validating environment..."
    if command -v go &> /dev/null; then
        print_success "Go is installed: $(go version)"
    else
        print_error "Go is not installed"
        exit 1
    fi
    
    print_success "Environment setup completed"
    print_info "Next steps:"
    print_info "1. Edit .env file with your API keys"
    print_info "2. Run: ./scripts/quick-start.sh"
    print_info "3. Open: http://localhost:8080"
}

clean_restart() {
    print_status "Cleaning and restarting..."
    
    # Stop existing processes
    print_status "Stopping existing processes..."
    pkill -f obsidian-automation 2>/dev/null || true
    docker stop obsidian-automation 2>/dev/null || true
    docker rm obsidian-automation 2>/dev/null || true
    
    # Clean build artifacts
    print_status "Cleaning build artifacts..."
    rm -f obsidian-automation obsidian-automation-dev
    
    # Clean logs (optional)
    read -p "Clean logs? (y/N): " clean_logs
    if [[ $clean_logs =~ ^[Yy]$ ]]; then
        rm -f logs/*.log
        print_status "Logs cleaned"
    fi
    
    # Restart
    print_status "Restarting..."
    quick_start
}

check_status() {
    print_status "Checking application status..."
    
    echo -e "${BLUE}Application Status:${NC}"
    
    # Check if running
    if pgrep -f "obsidian-automation" > /dev/null; then
        PID=$(pgrep -f "obsidian-automation")
        print_success "Application: RUNNING (PID: $PID)"
    else
        print_error "Application: STOPPED"
    fi
    
    # Check ports
    if netstat -tuln 2>/dev/null | grep -q ":8080"; then
        print_success "Port 8080: OPEN"
    else
        print_warning "Port 8080: CLOSED"
    fi
    
    # Check health endpoint
    if curl -s http://localhost:8080/api/services/status > /dev/null; then
        print_success "Health Check: PASSING"
    else
        print_error "Health Check: FAILING"
    fi
    
    # Check Docker
    if docker ps --format "table {{.Names}}" | grep -q "obsidian-automation"; then
        print_success "Docker Container: RUNNING"
    else
        print_info "Docker Container: NOT RUNNING"
    fi
    
    # Check environment
    echo -e "\n${BLUE}Environment:${NC}"
    if [ -f ".env" ]; then
        print_success ".env file: EXISTS"
        
        # Check key variables
        if grep -q "TELEGRAM_BOT_TOKEN=" .env && ! grep -q "TELEGRAM_BOT_TOKEN=$" .env; then
            print_success "Telegram Bot: CONFIGURED"
        else
            print_warning "Telegram Bot: NOT CONFIGURED"
        fi
        
        if grep -q "GEMINI_API_KEY=" .env && ! grep -q "GEMINI_API_KEY=$" .env; then
            print_success "AI Service: CONFIGURED"
        else
            print_warning "AI Service: NOT CONFIGURED"
        fi
    else
        print_error ".env file: NOT FOUND"
    fi
    
    # Check logs
    if [ -f "logs/app.log" ]; then
        LOG_SIZE=$(du -h logs/app.log | cut -f1)
        print_info "Log file: $LOG_SIZE"
    fi
}

stop_services() {
    print_status "Stopping all services..."
    
    # Stop application
    if pgrep -f "obsidian-automation" > /dev/null; then
        print_status "Stopping application..."
        pkill -f obsidian-automation
        print_success "Application stopped"
    else
        print_info "Application not running"
    fi
    
    # Stop Docker container
    if docker ps --format "table {{.Names}}" | grep -q "obsidian-automation"; then
        print_status "Stopping Docker container..."
        docker stop obsidian-automation
        docker rm obsidian-automation
        print_success "Docker container stopped"
    else
        print_info "Docker container not running"
    fi
    
    # Clean up
    print_status "Cleaning up..."
    rm -f obsidian-automation obsidian-automation-dev
    
    print_success "All services stopped"
}

# Main execution
main() {
    while true; do
        show_quick_menu
        read -p "Select an option: " choice
        echo ""
        
        quick_operations $choice
        
        echo ""
        read -p "Press Enter to continue..."
        clear
    done
}

# Handle command line arguments
if [ $# -eq 1 ]; then
    case $1 in
        "start") quick_start ;;
        "dev") development_mode ;;
        "build") production_build ;;
        "docker") docker_quick_start ;;
        "status") check_status ;;
        "stop") stop_services ;;
        "setup") setup_environment ;;
        *) echo "Usage: $0 [start|dev|build|docker|status|stop|setup]" ;;
    esac
    exit 0
fi

# Start interactive mode
main