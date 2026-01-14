#!/bin/bash
# build-all.sh - Build all Go applications and workers
# Part of US2: Build Process Independence

set -e

echo "=============================================="
echo "Building All Components"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BUILD_DIR="bin"
mkdir -p "$BUILD_DIR"

# Track build status
FAILED=0
BUILT=0

# Build Go applications
echo ""
echo -e "${YELLOW}Building Go applications...${NC}"
echo "----------------------------------------------"

for app in apps/*/; do
    if [ -f "${app}go.mod" ]; then
        app_name=$(basename "$app")
        echo "Building $app_name..."

        mkdir -p "${app}${BUILD_DIR}"

        if cd "$app" && go build -o "${BUILD_DIR}/${app_name}" ./cmd/main.go 2>/dev/null; then
            echo -e "${GREEN}✓${NC} $app_name built successfully"
            BUILT=$((BUILT + 1))
        else
            echo -e "${RED}✗${NC} $app_name build failed"
            FAILED=$((FAILED + 1))
        fi
    fi
done

# Build workers
echo ""
echo -e "${YELLOW}Building Cloudflare Workers...${NC}"
echo "----------------------------------------------"

for worker in workers/*/; do
    if [ -f "${worker}package.json" ]; then
        worker_name=$(basename "$worker")
        echo "Building $worker_name..."

        if cd "$worker" && npm run build 2>/dev/null; then
            echo -e "${GREEN}✓${NC} $worker_name built successfully"
            BUILT=$((BUILT + 1))
        else
            echo -e "${YELLOW}!${NC} $worker_name has no build step (Cloudflare Workers don't require build)"
            BUILT=$((BUILT + 1))
        fi
    fi
done

# Summary
echo ""
echo "=============================================="
echo "Build Summary"
echo "=============================================="
echo -e "Built: ${GREEN}$BUILT${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "Failed: ${RED}$FAILED${NC}"
    exit 1
else
    echo -e "Failed: ${RED}0${NC}"
    echo ""
    echo -e "${GREEN}All builds completed successfully!${NC}"
fi
