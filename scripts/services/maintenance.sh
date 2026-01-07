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
echo -e "${CYAN}║                Maintenance & Cleanup Scripts                 ║${NC}"
echo -e "${CYAN}║                  System Health & Optimization                 ║${NC}"
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

# Maintenance menu
show_maintenance_menu() {
    echo -e "${CYAN}Maintenance Operations:${NC}"
    echo ""
    echo "1. 🧹 Clean Build Artifacts"
    echo "2. 📊 Clean Logs"
    echo "3. 🗄️  Database Maintenance"
    echo "4. 📦 Clean Dependencies"
    echo "5. 🐳 Docker Cleanup"
    echo "6. 🔄 Reset Configuration"
    echo "7. 📈 System Health Check"
    echo "8. 🧹 Deep Clean"
    echo "9. 📋 Generate Report"
    echo "10. 🔧 Optimize System"
    echo "0. 🔙 Back"
    echo ""
}

# Maintenance operations
maintenance_operations() {
    case $1 in
        1) clean_build_artifacts ;;
        2) clean_logs ;;
        3) database_maintenance ;;
        4) clean_dependencies ;;
        5) docker_cleanup ;;
        6) reset_configuration ;;
        7) system_health_check ;;
        8) deep_clean ;;
        9) generate_report ;;
        10) optimize_system ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

clean_build_artifacts() {
    print_status "Cleaning build artifacts..."
    
    # Remove binary files
    print_status "Removing binary files..."
    rm -f obsidian-automation obsidian-automation-* main
    print_success "Binary files removed"
    
    # Remove temporary files
    print_status "Removing temporary files..."
    rm -f *.tmp *.temp *.bak
    rm -f temp/* tmp/* 2>/dev/null || true
    print_success "Temporary files removed"
    
    # Remove test coverage files
    print_status "Removing test coverage files..."
    rm -f coverage.out coverage.html
    print_success "Coverage files removed"
    
    # Clean Go build cache
    print_status "Cleaning Go build cache..."
    go clean -cache
    go clean -testcache
    print_success "Go build cache cleaned"
    
    print_success "Build artifacts cleanup completed"
}

clean_logs() {
    print_status "Cleaning logs..."
    
    # Check if logs directory exists
    if [ ! -d "logs" ]; then
        print_warning "Logs directory not found"
        return
    fi
    
    # Show log sizes before cleanup
    print_status "Current log sizes:"
    du -sh logs/* 2>/dev/null || print_info "No log files found"
    
    # Ask for confirmation
    read -p "Remove all log files? (y/N): " confirm
    if [[ $confirm =~ ^[Yy]$ ]]; then
        # Remove log files
        print_status "Removing log files..."
        rm -f logs/*.log logs/*.out
        print_success "Log files removed"
        
        # Keep .gitkeep files
        touch logs/.gitkeep
    else
        # Archive old logs
        print_status "Archiving old logs..."
        ARCHIVE_NAME="logs-$(date +%Y%m%d_%H%M%S).tar.gz"
        tar -czf "$ARCHIVE_NAME" logs/
        print_success "Logs archived to: $ARCHIVE_NAME"
        
        # Remove old log files
        rm -f logs/*.log logs/*.out
        print_success "Old log files removed"
    fi
    
    print_success "Log cleanup completed"
}

database_maintenance() {
    print_status "Performing database maintenance..."
    
    # Check if database is configured
    if [ -z "$TURSO_DATABASE_URL" ]; then
        print_warning "Database not configured"
        return
    fi
    
    # Test database connection
    print_status "Testing database connection..."
    if go run ./cmd/bot/main.go --test-db 2>/dev/null; then
        print_success "Database connection: OK"
    else
        print_error "Database connection: FAILED"
        return
    fi
    
    # Run database optimizations
    print_status "Running database optimizations..."
    # Add database optimization commands here
    
    # Clean up old records
    print_status "Cleaning up old records..."
    # Add cleanup commands here
    
    print_success "Database maintenance completed"
}

clean_dependencies() {
    print_status "Cleaning dependencies..."
    
    # Clean Go modules
    print_status "Cleaning Go modules..."
    go mod tidy
    go clean -modcache 2>/dev/null || print_warning "Cannot clean modcache (may require sudo)"
    
    # Remove vendor directory
    if [ -d "vendor" ]; then
        print_status "Removing vendor directory..."
        rm -rf vendor
        print_success "Vendor directory removed"
    fi
    
    # Download fresh dependencies
    print_status "Downloading fresh dependencies..."
    go mod download
    go mod verify
    
    print_success "Dependency cleanup completed"
}

docker_cleanup() {
    print_status "Cleaning Docker resources..."
    
    # Check if Docker is available
    if ! command -v docker &> /dev/null; then
        print_warning "Docker not available"
        return
    fi
    
    # Stop containers
    print_status "Stopping containers..."
    docker stop obsidian-automation 2>/dev/null || true
    docker rm obsidian-automation 2>/dev/null || true
    
    # Remove images
    print_status "Removing Docker images..."
    docker rmi obsidian-automation:latest 2>/dev/null || true
    docker rmi obsidian-automation 2>/dev/null || true
    
    # Clean up unused resources
    print_status "Cleaning up unused Docker resources..."
    docker system prune -f
    
    print_success "Docker cleanup completed"
}

reset_configuration() {
    print_status "Resetting configuration..."
    
    # Warning and confirmation
    print_warning "This will reset your configuration to defaults"
    print_warning "All custom settings will be lost"
    read -p "Are you sure? (type 'yes' to confirm): " confirm
    
    if [ "$confirm" != "yes" ]; then
        print_info "Configuration reset cancelled"
        return
    fi
    
    # Backup current configuration
    if [ -f ".env" ]; then
        print_status "Backing up current configuration..."
        cp .env ".env.backup.$(date +%Y%m%d_%H%M%S)"
        print_success "Configuration backed up"
    fi
    
    # Reset to example
    if [ -f ".env.example" ]; then
        print_status "Resetting to example configuration..."
        cp .env.example .env
        print_success "Configuration reset to defaults"
    else
        print_error ".env.example not found"
        return
    fi
    
    print_success "Configuration reset completed"
    print_info "Please edit .env file with your settings"
}

system_health_check() {
    print_status "Performing system health check..."
    
    echo -e "${BLUE}System Health Report:${NC}"
    echo "========================"
    
    # Check disk space
    DISK_USAGE=$(df . | awk 'NR==2 {print $5}' | sed 's/%//')
    if [ $DISK_USAGE -lt 80 ]; then
        print_success "Disk usage: ${DISK_USAGE}% (OK)"
    elif [ $DISK_USAGE -lt 90 ]; then
        print_warning "Disk usage: ${DISK_USAGE}% (WARNING)"
    else
        print_error "Disk usage: ${DISK_USAGE}% (CRITICAL)"
    fi
    
    # Check memory usage
    if command -v free &> /dev/null; then
        MEM_USAGE=$(free | awk 'NR==2{printf "%.0f", $3*100/$2}')
        if [ $MEM_USAGE -lt 80 ]; then
            print_success "Memory usage: ${MEM_USAGE}% (OK)"
        elif [ $MEM_USAGE -lt 90 ]; then
            print_warning "Memory usage: ${MEM_USAGE}% (WARNING)"
        else
            print_error "Memory usage: ${MEM_USAGE}% (CRITICAL)"
        fi
    fi
    
    # Check Go installation
    if command -v go &> /dev/null; then
        GO_VERSION=$(go version)
        print_success "Go: $GO_VERSION"
    else
        print_error "Go: Not installed"
    fi
    
    # Check Git installation
    if command -v git &> /dev/null; then
        GIT_VERSION=$(git --version)
        print_success "Git: $GIT_VERSION"
    else
        print_error "Git: Not installed"
    fi
    
    # Check Docker installation
    if command -v docker &> /dev/null; then
        DOCKER_VERSION=$(docker --version)
        print_success "Docker: $DOCKER_VERSION"
    else
        print_warning "Docker: Not installed"
    fi
    
    # Check application status
    if pgrep -f "obsidian-automation" > /dev/null; then
        print_success "Application: Running"
    else
        print_warning "Application: Not running"
    fi
    
    # Check configuration
    if [ -f ".env" ]; then
        print_success "Configuration: .env exists"
        
        # Check required variables
        if [ -n "$TELEGRAM_BOT_TOKEN" ]; then
            print_success "Telegram token: Configured"
        else
            print_error "Telegram token: Missing"
        fi
    else
        print_error "Configuration: .env missing"
    fi
    
    echo "========================"
    print_success "System health check completed"
}

deep_clean() {
    print_status "Performing deep clean..."
    
    # Warning
    print_warning "This will perform a comprehensive cleanup"
    print_warning "All build artifacts, logs, cache, and temporary files will be removed"
    read -p "Continue? (y/N): " confirm
    
    if [[ ! $confirm =~ ^[Yy]$ ]]; then
        print_info "Deep clean cancelled"
        return
    fi
    
    # Stop all services
    print_status "Stopping all services..."
    pkill -f obsidian-automation 2>/dev/null || true
    docker stop obsidian-automation 2>/dev/null || true
    
    # Clean build artifacts
    clean_build_artifacts
    
    # Clean logs
    clean_logs
    
    # Clean dependencies
    clean_dependencies
    
    # Clean Docker
    docker_cleanup
    
    # Clean system temporary files
    print_status "Cleaning system temporary files..."
    rm -rf /tmp/obsidian-* 2>/dev/null || true
    
    # Reset file permissions
    print_status "Resetting file permissions..."
    find . -type f -name "*.sh" -exec chmod +x {} \;
    chmod -R 755 logs attachments vault 2>/dev/null || true
    
    print_success "Deep clean completed"
}

generate_report() {
    print_status "Generating system report..."
    
    REPORT_FILE="system-report-$(date +%Y%m%d_%H%M%S).txt"
    
    {
        echo "Obsidian Automation System Report"
        echo "=================================="
        echo "Generated: $(date)"
        echo ""
        
        echo "System Information:"
        echo "- OS: $(uname -s)"
        echo "- Kernel: $(uname -r)"
        echo "- Architecture: $(uname -m)"
        echo ""
        
        echo "Software Versions:"
        echo "- Go: $(go version 2>/dev/null || echo 'Not installed')"
        echo "- Git: $(git --version 2>/dev/null || echo 'Not installed')"
        echo "- Docker: $(docker --version 2>/dev/null || echo 'Not installed')"
        echo ""
        
        echo "Project Information:"
        echo "- Directory: $(pwd)"
        echo "- Git branch: $(git branch --show-current 2>/dev/null || echo 'Not a git repo')"
        echo "- Git commit: $(git rev-parse --short HEAD 2>/dev/null || echo 'N/A')"
        echo "- Go modules: $(go list -m all 2>/dev/null | wc -l) modules"
        echo ""
        
        echo "Resource Usage:"
        echo "- Disk usage: $(df -h . | awk 'NR==2 {print $3 "/" $2 " (" $5 ")"}')"
        if command -v free &> /dev/null; then
            echo "- Memory usage: $(free -h | awk 'NR==2{print $3 "/" $2}')"
        fi
        echo ""
        
        echo "Application Status:"
        if pgrep -f "obsidian-automation" > /dev/null; then
            echo "- Status: Running"
            echo "- PID: $(pgrep -f obsidian-automation)"
        else
            echo "- Status: Not running"
        fi
        echo ""
        
        echo "Configuration Status:"
        if [ -f ".env" ]; then
            echo "- .env file: Exists"
            echo "- Size: $(du -h .env | cut -f1)"
            echo "- Modified: $(stat -c %y .env 2>/dev/null || echo 'N/A')"
        else
            echo "- .env file: Missing"
        fi
        echo ""
        
        echo "Directory Structure:"
        echo "- Logs directory: $([ -d logs ] && echo 'Exists' || echo 'Missing')"
        echo "- Attachments directory: $([ -d attachments ] && echo 'Exists' || echo 'Missing')"
        echo "- Vault directory: $([ -d vault ] && echo 'Exists' || echo 'Missing')"
        echo "- Go modules: $([ -f go.mod ] && echo 'Exists' || echo 'Missing')"
        echo ""
        
        echo "Recent Activity:"
        echo "- Last build: $(ls -la obsidian-automation 2>/dev/null | awk '{print $6, $7, $8}' || echo 'No build found')"
        echo "- Last commit: $(git log -1 --format='%ci %s' 2>/dev/null || echo 'No git history')"
        
    } > "$REPORT_FILE"
    
    print_success "System report generated: $REPORT_FILE"
    print_info "View with: cat $REPORT_FILE"
}

optimize_system() {
    print_status "Optimizing system..."
    
    # Optimize Go modules
    print_status "Optimizing Go modules..."
    go mod tidy
    go mod download
    
    # Optimize file permissions
    print_status "Optimizing file permissions..."
    find . -type f -name "*.go" -exec chmod 644 {} \;
    find . -type f -name "*.sh" -exec chmod +x {} \;
    chmod -R 755 logs attachments vault 2>/dev/null || true
    
    # Optimize Git repository
    if [ -d ".git" ]; then
        print_status "Optimizing Git repository..."
        git gc --aggressive --prune=now
        print_success "Git repository optimized"
    fi
    
    # Optimize Docker images
    if command -v docker &> /dev/null; then
        print_status "Optimizing Docker resources..."
        docker system prune -f
        print_success "Docker resources optimized"
    fi
    
    # Create optimized build
    print_status "Creating optimized build..."
    go build \
        -ldflags "-s -w" \
        -tags=prod \
        -o obsidian-automation-optimized \
        ./cmd/bot/main.go
    
    if [ -f "obsidian-automation-optimized" ]; then
        ORIGINAL_SIZE=$(du -h obsidian-automation 2>/dev/null | cut -f1 || echo "N/A")
        OPTIMIZED_SIZE=$(du -h obsidian-automation-optimized | cut -f1)
        print_success "Optimized build created"
        print_info "Original size: $ORIGINAL_SIZE"
        print_info "Optimized size: $OPTIMIZED_SIZE"
    fi
    
    print_success "System optimization completed"
}

# Main execution
main() {
    while true; do
        show_maintenance_menu
        read -p "Select an option: " choice
        echo ""
        
        maintenance_operations $choice
        
        echo ""
        read -p "Press Enter to continue..."
        clear
    done
}

# Handle command line arguments
if [ $# -eq 1 ]; then
    case $1 in
        "clean") clean_build_artifacts ;;
        "logs") clean_logs ;;
        "health") system_health_check ;;
        "deep") deep_clean ;;
        "optimize") optimize_system ;;
        "report") generate_report ;;
        *) echo "Usage: $0 [clean|logs|health|deep|optimize|report]" ;;
    esac
    exit 0
fi

# Start interactive mode
main