#!/bin/bash
# CI/CD Doppler Integration Script
# This script sets up Doppler for automated testing in CI/CD pipelines

set -e

# Configuration
DOPPLER_PROJECT="${DOPPLER_PROJECT:-bot}"
DOPPLER_CONFIG="${DOPPLER_CONFIG:-e2e}"
DOPPLER_TOKEN="${DOPPLER_TOKEN}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo -e "${GREEN}Setting up Doppler for CI/CD testing${NC}"

# Validate environment
if [ -z "$DOPPLER_TOKEN" ]; then
    echo -e "${RED}Error: DOPPLER_TOKEN environment variable is required${NC}"
    exit 1
fi

# Verify Doppler CLI
if ! command -v doppler &> /dev/null; then
    echo -e "${YELLOW}Installing Doppler CLI...${NC}"

    # Install Doppler (Linux CI/CD)
    curl -Ls --tlsv1.2 --proto "=https" --retry 3 https://cli.doppler.com/install.sh | sudo sh

    if ! command -v doppler &> /dev/null; then
        echo -e "${RED}Failed to install Doppler CLI${NC}"
        exit 1
    fi
fi

echo -e "${GREEN}Doppler CLI version: $(doppler --version)${NC}"

# Authenticate with service token
echo -e "${YELLOW}Authenticating with Doppler...${NC}"
export DOPPLER_TOKEN="$DOPPLER_TOKEN"

# Verify authentication
if ! doppler me --project "$DOPPLER_PROJECT" &> /dev/null; then
    echo -e "${RED}Doppler authentication failed${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Doppler authentication successful${NC}"

# Verify project and config access
if ! doppler secrets --project "$DOPPLER_PROJECT" --config "$DOPPLER_CONFIG" &> /dev/null; then
    echo -e "${RED}Cannot access Doppler project/config: $DOPPLER_PROJECT/$DOPPLER_CONFIG${NC}"
    exit 1
fi

echo -e "${GREEN}✓ Doppler project/config access verified${NC}"

# Set environment variables for tests
echo -e "${YELLOW}Setting up test environment...${NC}"

# Export Doppler environment variables
eval "$(doppler secrets download --project "$DOPPLER_PROJECT" --config "$DOPPLER_CONFIG" --format env)"

# Set additional test variables
export GO_ENV="test"
export TEST_WITH_DOPPLER="true"
export DOPPLER_PROJECT="$DOPPLER_PROJECT"
export DOPPLER_CONFIG="$DOPPLER_CONFIG"

echo -e "${GREEN}✓ Test environment configured${NC}"

# Optional: Run pre-test validation
if [ "$RUN_PREFLIGHT" = "true" ]; then
    echo -e "${YELLOW}Running pre-flight checks...${NC}"

    # Check required environment variables
    required_vars=("TELEGRAM_BOT_TOKEN" "TEST_DATABASE_URL")
    for var in "${required_vars[@]}"; do
        if [ -z "${!var}" ]; then
            echo -e "${RED}Missing required environment variable: $var${NC}"
            exit 1
        fi
    done

    echo -e "${GREEN}✓ Pre-flight checks passed${NC}"
fi

echo -e "${GREEN}Doppler CI/CD setup complete!${NC}"
echo "Environment variables are now available for testing."