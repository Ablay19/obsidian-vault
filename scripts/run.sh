#!/bin/bash

# Obsidian Bot Scripts Directory
# This file provides easy access to all organized scripts

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Help function
show_help() {
    echo -e "${CYAN}Obsidian Bot Scripts - Usage Guide${NC}"
    echo "=================================="
    echo ""
    echo -e "${YELLOW}Deployment Scripts:${NC}"
    echo "  deploy-cloudflare     - Deploy to Cloudflare Workers"
    echo "  deploy-simple        - Simple deployment script"
    echo "  docker-deploy        - Docker deployment"
    echo ""
    echo -e "${YELLOW}Service Management:${NC}"
    echo "  start-services       - Start all services (web + SSH)"
    echo "  control              - Control running services"
    echo "  maintenance          - Maintenance operations"
    echo ""
    echo -e "${YELLOW}Setup Scripts:${NC}"
    echo "  env-setup            - Environment setup"
    echo "  quick-start          - Quick project setup"
    echo "  complete-setup       - Complete project setup"
    echo "  setup-doppler        - Doppler secrets setup"
    echo "  setup-google-cloud   - Google Cloud setup"
    echo ""
    echo -e "${YELLOW}Kubernetes Scripts:${NC}"
    echo "  k8s/deploy           - Deploy to Kubernetes"
    echo "  k8s/secrets          - Manage Kubernetes secrets"
    echo ""
    echo -e "${YELLOW}Monitoring & Testing:${NC}"
    echo "  monitor-cloudflare   - Monitor Cloudflare deployment"
    echo "  test-cloudflare      - Test Cloudflare setup"
    echo "  test-doppler         - Test Doppler configuration"
    echo "  test-dashboard       - Test dashboard functionality"
    echo "  validate-whatsapp    - Validate WhatsApp setup"
    echo "  run-migrations       - Run database migrations"
    echo ""
    echo -e "${YELLOW}Development:${NC}"
    echo "  dev                  - Development mode"
    echo "  dashboard            - Dashboard development"
    echo "  debug-cloudflare    - Debug Cloudflare issues"
    echo "  final-test           - Final integration tests"
    echo ""
    echo -e "${YELLOW}Utilities:${NC}"
    echo "  system-check         - System requirements check"
    echo "  github-codespaces    - GitHub Codespaces setup"
    echo ""
    echo -e "${GREEN}Examples:${NC}"
    echo "  ./scripts/deploy-cloudflare.sh"
    echo "  make deploy"
    echo "  ./scripts/start-services.sh"
    echo ""
}

# Check if help is requested
if [[ "$1" == "help" || "$1" == "-h" || "$1" == "--help" ]]; then
    show_help
    exit 0
fi

# Auto-discover and run script
SCRIPT_NAME="$1"
SCRIPT_DIR="$(dirname "$0")"

if [[ -z "$SCRIPT_NAME" ]]; then
    echo -e "${RED}Error: No script specified${NC}"
    echo ""
    show_help
    exit 1
fi

# Check if script exists in organized directories
POSSIBLE_PATHS=(
    "$SCRIPT_DIR/deployment/$SCRIPT_NAME.sh"
    "$SCRIPT_DIR/services/$SCRIPT_NAME.sh"
    "$SCRIPT_DIR/monitoring/$SCRIPT_NAME.sh"
    "$SCRIPT_DIR/setup/$SCRIPT_NAME.sh"
    "$SCRIPT_DIR/utilities/$SCRIPT_NAME.sh"
    "$SCRIPT_DIR/k8s/$SCRIPT_NAME.sh"
    "$SCRIPT_DIR/dev/$SCRIPT_NAME.sh"
    "$SCRIPT_DIR/$SCRIPT_NAME.sh"  # Direct fallback
)

SCRIPT_FOUND=""
for path in "${POSSIBLE_PATHS[@]}"; do
    if [[ -f "$path" ]]; then
        SCRIPT_FOUND="$path"
        break
    fi
done

if [[ -n "$SCRIPT_FOUND" ]]; then
    echo -e "${GREEN}✓ Running script: $SCRIPT_NAME${NC}"
    echo -e "${BLUE}Path: $SCRIPT_FOUND${NC}"
    echo ""
    bash "$SCRIPT_FOUND" "${@:2}"
else
    echo -e "${RED}❌ Script not found: $SCRIPT_NAME${NC}"
    echo ""
    echo -e "${YELLOW}Available scripts:${NC}"
    ls -1 "$SCRIPT_DIR"/*.sh 2>/dev/null | sed 's/.sh$//' | head -20
    echo ""
    echo "Use '$0 help' for detailed usage guide."
    exit 1
fi