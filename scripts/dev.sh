#!/bin/bash
# dev.sh - Development environment with hot reload for obsidian-vault
# Part of US2: Hot Reload Configuration

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

print_status() { echo -e "${BLUE}[•]${NC} $1"; }
print_success() { echo -e "${GREEN}[✓]${NC} $1"; }
print_warning() { echo -e "${YELLOW}[⚠]${NC} $1"; }

show_menu() {
    echo ""
    echo -e "${BLUE}Obsidian Vault Development${NC}"
    echo ""
    echo "1. 🚀 Start API Gateway (with hot reload)"
    echo "2. ⚡ Start AI Worker (wrangler dev)"
    echo "3. 🏗️  Build all components"
    echo "4. 🧪 Run all tests"
    echo "5. 📦 Full build & test"
    echo "0. 🚪 Exit"
    echo ""
}

start_api_gateway() {
    print_status "Starting API Gateway with hot reload..."
    if command -v air &> /dev/null; then
        cd apps/api-gateway && air
    else
        print_warning "air not found, using standard go run"
        print_info "Install air: go install github.com/cosmtrek/air@latest"
        cd apps/api-gateway && go run ./cmd/main.go
    fi
}

start_ai_worker() {
    print_status "Starting AI Worker with hot reload..."
    if [ -d "workers/ai-worker" ]; then
        cd workers/ai-worker && npx wrangler dev
    else
        print_error "AI Worker not found"
    fi
}

build_all() {
    print_status "Building all components..."
    ./scripts/build-all.sh
}

run_tests() {
    print_status "Running all tests..."
    ./scripts/test-all.sh
}

full_build_test() {
    build_all
    run_tests
}

main() {
    while true; do
        show_menu
        read -p "Select an option: " choice
        
        case $choice in
            1) start_api_gateway ;;
            2) start_ai_worker ;;
            3) build_all ;;
            4) run_tests ;;
            5) full_build_test ;;
            0) print_success "Goodbye!"; exit 0 ;;
            *) print_error "Invalid option" ;;
        esac
    done
}

main
