#!/bin/bash

# Termux Services Build Script
# Builds all Go services for Android/Termux compatibility

set -e

echo "🏗️  Building services for Termux (Android ARM64)..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Build function
build_service() {
    local service_name=$1
    local service_path=$2

    echo -e "${YELLOW}Building ${service_name}...${NC}"

    if [ -d "$service_path" ]; then
        cd "$service_path"

        # Build for Android ARM64 (Termux)
        GOOS=android GOARCH=arm64 go build -o "${service_name}-termux" .

        if [ $? -eq 0 ]; then
            echo -e "${GREEN}✅ ${service_name} built successfully${NC}"
            ls -lh "${service_name}-termux"
        else
            echo -e "${RED}❌ Failed to build ${service_name}${NC}"
            exit 1
        fi

        cd - > /dev/null
    else
        echo -e "${RED}⚠️  ${service_path} not found, skipping${NC}"
    fi
}

# Clean previous builds
echo "🧹 Cleaning previous builds..."
find . -name "*-termux" -type f -delete
echo -e "${GREEN}✅ Cleaned previous builds${NC}"

# Build mauritania-cli (main Termux service)
build_service "mauritania-cli" "cmd/mauritania-cli"

echo ""
echo -e "${GREEN}🎉 All Termux services built successfully!${NC}"
echo ""
echo "Built services:"
find . -name "*-termux" -type f -exec ls -lh {} \;

echo ""
echo -e "${YELLOW}To install on Termux:${NC}"
echo "1. Copy the *-termux binaries to your Termux device"
echo "2. Make them executable: chmod +x *-termux"
echo "3. Run: ./mauritania-cli-termux --help"