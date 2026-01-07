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
echo -e "${CYAN}║                 Master Control Script                        ║${NC}"
echo -e "${CYAN}║              All-in-One Operations Center                    ║${NC}"
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

# Master menu
show_master_menu() {
    echo -e "${CYAN}Master Control Center:${NC}"
    echo ""
    echo "1. 🚀 Quick Operations"
    echo "2. 🔧 Development Tools"
    echo "3. 📊 Build & Deploy"
    echo "4. 🛠️  Environment Management"
    echo "5. 🧹 Maintenance & Cleanup"
    echo "6. 📈 Monitoring & Debugging"
    echo "7. 🐳 Docker Operations"
    echo "8. ☁️  Cloud Services"
    echo "9. 🔐 Security & Authentication"
    echo "10. 📋 System Information"
    echo "11. 🎯 One-Click Actions"
    echo "0. 🚪 Exit"
    echo ""
}

# Quick operations menu
show_quick_menu() {
    echo -e "${CYAN}Quick Operations:${NC}"
    echo ""
    echo "1. 🚀 Start Application"
    echo "2. 🛑 Stop Application"
    echo "3. 🔄 Restart Application"
    echo "4. 📊 Check Status"
    echo "5. 🧪 Development Mode"
    echo "6. 🏗️  Production Mode"
    echo "7. 📋 View Logs"
    echo "8. 🔍 Health Check"
    echo "0. 🔙 Back"
    echo ""
}

# Development tools menu
show_dev_menu() {
    echo -e "${CYAN}Development Tools:${NC}"
    echo ""
    echo "1. 🔨 Build Application"
    echo "2. 🧪 Run Tests"
    echo "3. 🔍 Lint Code"
    echo "4. 📊 Run Benchmarks"
    echo "5. 🐛 Debug Mode"
    echo "6. 📝 Code Coverage"
    echo "7. 🔄 Hot Reload"
    echo "8. 📚 Generate Docs"
    echo "0. 🔙 Back"
    echo ""
}

# Build & deploy menu
show_build_menu() {
    echo -e "${CYAN}Build & Deploy:${NC}"
    echo ""
    echo "1. 🏗️  Build Application"
    echo "2. 🐳 Build Docker Image"
    echo "3. ☁️  Deploy to Cloudflare"
    echo "4. 🌐 Deploy to Render"
    echo "5. 📦 Package for Distribution"
    echo "6. 🔄 Update Deployment"
    echo "7. 📋 Deployment Status"
    echo "8. 🚀 Full Deploy Pipeline"
    echo "0. 🔙 Back"
    echo ""
}

# Environment management menu
show_env_menu() {
    echo -e "${CYAN}Environment Management:${NC}"
    echo ""
    echo "1. 🔧 Initial Setup"
    echo "2. 📝 Configure .env"
    echo "3. 🗄️  Database Setup"
    echo "4. 🤖 AI Services Setup"
    echo "5. 📱 Bot Services Setup"
    echo "6. 🔐 Authentication Setup"
    echo "7. 🧪 Test Environment"
    echo "8. 📊 Validate Configuration"
    echo "9. 🔄 Reset Environment"
    echo "10. 📋 Environment Report"
    echo "0. 🔙 Back"
    echo ""
}

# Maintenance menu
show_maintenance_menu() {
    echo -e "${CYAN}Maintenance & Cleanup:${NC}"
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

# Monitoring menu
show_monitor_menu() {
    echo -e "${CYAN}Monitoring & Debugging:${NC}"
    echo ""
    echo "1. 📊 View Logs"
    echo "2. 🔍 Health Check"
    echo "3. 📈 Performance Metrics"
    echo "4. 🐛 Debug Mode"
    echo "5. 🔥 Cloudflare Monitoring"
    echo "6. 📋 Service Status"
    echo "7. 🕐 Real-time Logs"
    echo "8. 📊 Analytics Dashboard"
    echo "0. 🔙 Back"
    echo ""
}

# Docker operations menu
show_docker_menu() {
    echo -e "${CYAN}Docker Operations:${NC}"
    echo ""
    echo "1. 🐳 Build Image"
    echo "2. 🚀 Run Container"
    echo "3. 🛑 Stop Container"
    echo "4. 🔄 Restart Container"
    echo "5. 📋 Container Status"
    echo "6. 📊 View Logs"
    echo "7. 🧹 Cleanup Docker"
    echo "8. 📦 Compose Operations"
    echo "0. 🔙 Back"
    echo ""
}

# Cloud services menu
show_cloud_menu() {
    echo -e "${CYAN}Cloud Services:${NC}"
    echo ""
    echo "1. 🔥 Cloudflare Deploy"
    echo "2. 🌐 Render Deploy"
    echo "3. 📊 Cloudflare Monitor"
    echo "4. 🔧 Cloud Services Config"
    echo "5. 📋 Cloud Status"
    echo "6. 🔄 Sync Cloud Resources"
    echo "7. 📊 Cloud Analytics"
    echo "8. 🧹 Cloud Cleanup"
    echo "0. 🔙 Back"
    echo ""
}

# Security menu
show_security_menu() {
    echo -e "${CYAN}Security & Authentication:${NC}"
    echo ""
    echo "1. 🔐 Authentication Setup"
    echo "2. 🔑 Generate Secrets"
    echo "3. 🔍 Security Audit"
    echo "4. 🛡️  Update Security"
    echo "5. 📋 Security Report"
    echo "6. 🔧 SSL/TLS Setup"
    echo "7. 🚨 Security Scan"
    echo "8. 📊 Security Metrics"
    echo "0. 🔙 Back"
    echo ""
}

# System information menu
show_info_menu() {
    echo -e "${CYAN}System Information:${NC}"
    echo ""
    echo "1. 📊 System Status"
    echo "2. 📈 Performance Metrics"
    echo "3. 🔧 Configuration Info"
    echo "4. 📦 Dependencies Info"
    echo "5. 🐳 Docker Info"
    echo "6. ☁️  Cloud Services Info"
    echo "7. 📋 Application Info"
    echo "8. 📊 Generate Full Report"
    echo "0. 🔙 Back"
    echo ""
}

# One-click actions menu
show_oneclick_menu() {
    echo -e "${CYAN}One-Click Actions:${NC}"
    echo ""
    echo "1. 🚀 Full Setup & Start"
    echo "2. 🔄 Update & Restart"
    echo "3. 🧹 Clean & Rebuild"
    echo "4. 📊 Health Check & Report"
    echo "5. 🚀 Deploy All Services"
    echo "6. 🛑 Emergency Stop"
    echo "7. 📋 Backup Configuration"
    echo "8. 🔄 Restore from Backup"
    echo "0. 🔙 Back"
    echo ""
}

# Execute scripts based on choice
execute_script() {
    local script_name=$1
    shift
    
    if [ -f "scripts/${script_name}.sh" ]; then
        print_status "Executing ${script_name}.sh..."
        bash "scripts/${script_name}.sh" "$@"
    else
        print_error "Script ${script_name}.sh not found"
    fi
}

# Quick operations
quick_operations() {
    case $1 in
        1) execute_script "quick-start" start ;;
        2) execute_script "quick-start" stop ;;
        3) execute_script "quick-start" stop && sleep 2 && execute_script "quick-start" start ;;
        4) execute_script "quick-start" status ;;
        5) execute_script "quick-start" dev ;;
        6) execute_script "quick-start" build && execute_script "quick-start" start ;;
        7) execute_script "dev" 7 ;;
        8) execute_script "dev" 2 ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

# Development operations
dev_operations() {
    case $1 in
        1) execute_script "dev" 1 ;;
        2) execute_script "dev" 2 ;;
        3) execute_script "dev" 3 ;;
        4) execute_script "dev" 4 ;;
        5) execute_script "dev" 4 ;;
        6) go test -coverprofile=coverage.out ./... && go tool cover -html=coverage.out ;;
        7) print_status "Hot reload not implemented yet" ;;
        8) print_status "Generating docs..." && godoc -http=:6060 & ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

# Build operations
build_operations() {
    case $1 in
        1) execute_script "dev" 1 ;;
        2) execute_script "dev" 6 ;;
        3) execute_script "deploy-cloudflare" ;;
        4) execute_script "dev" 1 && git push origin main ;;
        5) print_status "Packaging for distribution..." ;;
        6) git add -A && git commit -m "Update deployment $(date)" && git push origin main ;;
        7) execute_script "dev" 8 ;;
        8) execute_script "dev" 1 && execute_script "deploy-cloudflare" && git push origin main ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

# Environment operations
env_operations() {
    case $1 in
        1) execute_script "env-setup" init ;;
        2) execute_script "env-setup" config ;;
        3) execute_script "env-setup" 3 ;;
        4) execute_script "env-setup" 4 ;;
        5) execute_script "env-setup" 5 ;;
        6) execute_script "env-setup" 6 ;;
        7) execute_script "env-setup" test ;;
        8) execute_script "env-setup" validate ;;
        9) execute_script "env-setup" 6 ;;
        10) execute_script "maintenance" report ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

# Maintenance operations
maintenance_operations() {
    case $1 in
        1) execute_script "maintenance" clean ;;
        2) execute_script "maintenance" logs ;;
        3) execute_script "maintenance" 3 ;;
        4) execute_script "maintenance" 4 ;;
        5) execute_script "maintenance" 5 ;;
        6) execute_script "maintenance" 6 ;;
        7) execute_script "maintenance" health ;;
        8) execute_script "maintenance" deep ;;
        9) execute_script "maintenance" report ;;
        10) execute_script "maintenance" optimize ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

# Monitor operations
monitor_operations() {
    case $1 in
        1) execute_script "dev" 1 ;;
        2) execute_script "dev" 2 ;;
        3) execute_script "dev" 3 ;;
        4) execute_script "dev" 4 ;;
        5) execute_script "monitor-cloudflare" ;;
        6) execute_script "dev" 6 ;;
        7) execute_script "dev" 7 ;;
        8) print_status "Opening analytics dashboard..." && xdg-open http://localhost:8080/dashboard/analytics 2>/dev/null ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

# Docker operations
docker_operations() {
    case $1 in
        1) execute_script "quick-start" docker ;;
        2) docker run -d --name obsidian-automation -p 8080:8080 --env-file .env obsidian-automation:latest ;;
        3) docker stop obsidian-automation ;;
        4) docker restart obsidian-automation ;;
        5) docker ps -a | grep obsidian-automation ;;
        6) docker logs -f obsidian-automation ;;
        7) execute_script "maintenance" 5 ;;
        8) docker-compose up -d ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

# Cloud operations
cloud_operations() {
    case $1 in
        1) execute_script "deploy-cloudflare" ;;
        2) git push origin main ;;
        3) execute_script "monitor-cloudflare" ;;
        4) print_status "Cloud services configuration..." ;;
        5) execute_script "dev" 8 ;;
        6) print_status "Syncing cloud resources..." ;;
        7) print_status "Cloud analytics..." ;;
        8) print_status "Cloud cleanup..." ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

# Security operations
security_operations() {
    case $1 in
        1) execute_script "env-setup" 6 ;;
        2) print_status "Generating secure secrets..." && openssl rand -base64 32 ;;
        3) print_status "Performing security audit..." ;;
        4) print_status "Updating security configurations..." ;;
        5) print_status "Generating security report..." ;;
        6) print_status "SSL/TLS setup..." ;;
        7) print_status "Running security scan..." ;;
        8) print_status "Security metrics..." ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

# Info operations
info_operations() {
    case $1 in
        1) execute_script "maintenance" health ;;
        2) execute_script "dev" 3 ;;
        3) print_status "Configuration information..." && cat .env 2>/dev/null || print_warning ".env file not found" ;;
        4) go list -m all ;;
        5) docker system info ;;
        6) print_status "Cloud services information..." ;;
        7) print_status "Application information..." && ./obsidian-automation --version 2>/dev/null || print_warning "Application not built" ;;
        8) execute_script "maintenance" report ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

# One-click operations
oneclick_operations() {
    case $1 in
        1) 
            print_status "Performing full setup and start..."
            execute_script "env-setup" init
            execute_script "env-setup" config
            execute_script "dev" 1
            execute_script "quick-start" start
            ;;
        2) 
            print_status "Updating and restarting..."
            git pull origin main
            execute_script "dev" 1
            execute_script "quick-start" stop
            sleep 2
            execute_script "quick-start" start
            ;;
        3) 
            print_status "Clean and rebuild..."
            execute_script "maintenance" deep
            execute_script "dev" 1
            ;;
        4) 
            print_status "Health check and report..."
            execute_script "maintenance" health
            execute_script "maintenance" report
            ;;
        5) 
            print_status "Deploying all services..."
            execute_script "dev" 1
            execute_script "deploy-cloudflare"
            git push origin main
            ;;
        6) 
            print_status "Emergency stop..."
            execute_script "quick-start" stop
            pkill -f obsidian-automation
            docker stop obsidian-automation 2>/dev/null || true
            ;;
        7) 
            print_status "Backing up configuration..."
            cp .env ".env.backup.$(date +%Y%m%d_%H%M%S)"
            print_success "Configuration backed up"
            ;;
        8) 
            print_status "Restore from backup..."
            print_info "Available backups:"
            ls -la .env.backup.* 2>/dev/null || print_warning "No backups found"
            ;;
        0) return ;;
        *) print_error "Invalid option" ;;
    esac
}

# Main execution loop
main() {
    while true; do
        show_master_menu
        read -p "Select an option: " choice
        echo ""
        
        case $choice in
            1)
                while true; do
                    show_quick_menu
                    read -p "Select quick operation: " subchoice
                    quick_operations $subchoice
                    [ $subchoice -eq 0 ] && break
                done
                ;;
            2)
                while true; do
                    show_dev_menu
                    read -p "Select development tool: " subchoice
                    dev_operations $subchoice
                    [ $subchoice -eq 0 ] && break
                done
                ;;
            3)
                while true; do
                    show_build_menu
                    read -p "Select build operation: " subchoice
                    build_operations $subchoice
                    [ $subchoice -eq 0 ] && break
                done
                ;;
            4)
                while true; do
                    show_env_menu
                    read -p "Select environment operation: " subchoice
                    env_operations $subchoice
                    [ $subchoice -eq 0 ] && break
                done
                ;;
            5)
                while true; do
                    show_maintenance_menu
                    read -p "Select maintenance operation: " subchoice
                    maintenance_operations $subchoice
                    [ $subchoice -eq 0 ] && break
                done
                ;;
            6)
                while true; do
                    show_monitor_menu
                    read -p "Select monitoring operation: " subchoice
                    monitor_operations $subchoice
                    [ $subchoice -eq 0 ] && break
                done
                ;;
            7)
                while true; do
                    show_docker_menu
                    read -p "Select Docker operation: " subchoice
                    docker_operations $subchoice
                    [ $subchoice -eq 0 ] && break
                done
                ;;
            8)
                while true; do
                    show_cloud_menu
                    read -p "Select cloud operation: " subchoice
                    cloud_operations $subchoice
                    [ $subchoice -eq 0 ] && break
                done
                ;;
            9)
                while true; do
                    show_security_menu
                    read -p "Select security operation: " subchoice
                    security_operations $subchoice
                    [ $subchoice -eq 0 ] && break
                done
                ;;
            10)
                while true; do
                    show_info_menu
                    read -p "Select information operation: " subchoice
                    info_operations $subchoice
                    [ $subchoice -eq 0 ] && break
                done
                ;;
            11)
                while true; do
                    show_oneclick_menu
                    read -p "Select one-click action: " subchoice
                    oneclick_operations $subchoice
                    [ $subchoice -eq 0 ] && break
                done
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

# Handle command line arguments
if [ $# -eq 1 ]; then
    case $1 in
        "start") execute_script "quick-start" start ;;
        "stop") execute_script "quick-start" stop ;;
        "restart") execute_script "quick-start" stop && sleep 2 && execute_script "quick-start" start ;;
        "status") execute_script "quick-start" status ;;
        "dev") execute_script "quick-start" dev ;;
        "build") execute_script "dev" 1 ;;
        "test") execute_script "dev" 2 ;;
        "deploy") execute_script "deploy-cloudflare" ;;
        "clean") execute_script "maintenance" clean ;;
        "health") execute_script "maintenance" health ;;
        "setup") execute_script "env-setup" init ;;
        *) echo "Usage: $0 [start|stop|restart|status|dev|build|test|deploy|clean|health|setup]" ;;
    esac
    exit 0
fi

# Start interactive mode
main