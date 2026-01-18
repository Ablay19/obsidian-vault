#!/bin/bash
# Doppler Setup Script for E2E Testing
# This script installs and configures Doppler CLI for the Mauritania CLI project

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Setting up Doppler for Mauritania CLI E2E Testing${NC}"

# Check if Doppler is already installed
if command -v doppler &> /dev/null; then
    echo -e "${GREEN}Doppler CLI is already installed$(doppler --version)${NC}"
else
    echo -e "${YELLOW}Installing Doppler CLI...${NC}"

    # Install Doppler CLI based on OS
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        # Linux
        curl -Ls --tlsv1.2 --proto "=https" --retry 3 https://cli.doppler.com/install.sh | sudo sh
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        brew install dopplerhq/cli/doppler
    else
        echo -e "${RED}Unsupported OS: $OSTYPE${NC}"
        echo "Please install Doppler CLI manually from https://docs.doppler.com/docs/cli"
        exit 1
    fi
fi

# Verify installation
if ! command -v doppler &> /dev/null; then
    echo -e "${RED}Doppler CLI installation failed${NC}"
    exit 1
fi

echo -e "${GREEN}Doppler CLI installed successfully$(doppler --version)${NC}"

# Check if user is logged in
if ! doppler me &> /dev/null; then
    echo -e "${YELLOW}Please login to Doppler:${NC}"
    echo "doppler login"
    echo ""
    echo -e "${YELLOW}After logging in, run this script again or set up your project manually.${NC}"
    exit 0
fi

echo -e "${GREEN}Doppler authentication verified${NC}"

# Project setup instructions
echo ""
echo -e "${YELLOW}Project Setup Instructions:${NC}"
echo "1. Create a Doppler project: doppler projects create mauritania-cli"
echo "2. Create environments: doppler environments create dev staging prod"
echo "3. Add secrets to configs:"
echo "   - doppler secrets set API_KEY --config dev"
echo "   - doppler secrets set TELEGRAM_BOT_TOKEN --config dev"
echo "   - doppler secrets set WHATSAPP_API_KEY --config dev"
echo ""

# Service token setup for CI/CD
echo -e "${YELLOW}CI/CD Setup:${NC}"
echo "Generate service tokens for automated testing:"
echo "doppler service-tokens create e2e-testing --config dev"
echo ""
echo "Set DOPPLER_TOKEN in your CI/CD environment variables"

echo -e "${GREEN}Doppler setup complete!${NC}"
echo "Run 'doppler --help' for more commands"