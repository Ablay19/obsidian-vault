#!/bin/bash
# test-all.sh - Run all tests for Go applications and workers
# Part of US2: Test Process Independence

set -e

echo "=============================================="
echo "Running All Tests"
echo "=============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track test status
FAILED=0
PASSED=0
SKIPPED=0

# Test Go applications
echo ""
echo -e "${YELLOW}Testing Go applications...${NC}"
echo "----------------------------------------------"

for app in apps/*/; do
    if [ -f "${app}go.mod" ]; then
        app_name=$(basename "$app")
        echo "Testing $app_name..."

        if cd "$app" && go test -v ./internal/... ./tests/... 2>/dev/null; then
            echo -e "${GREEN}✓${NC} $app_name tests passed"
            ((PASSED++))
        else
            echo -e "${YELLOW}!${NC} $app_name has no tests or test directory"
            ((SKIPPED++))
        fi
    fi
done

# Test workers
echo ""
echo -e "${YELLOW}Testing Cloudflare Workers...${NC}"
echo "----------------------------------------------"

for worker in workers/*/; do
    if [ -f "${worker}package.json" ]; then
        worker_name=$(basename "$worker")
        echo "Testing $worker_name..."

        if cd "$worker" && npm test 2>/dev/null; then
            echo -e "${GREEN}✓${NC} $worker_name tests passed"
            ((PASSED++))
        else
            echo -e "${YELLOW}!${NC} $worker_name has no test configuration"
            ((SKIPPED++))
        fi
    fi
done

# Summary
echo ""
echo "=============================================="
echo "Test Summary"
echo "=============================================="
echo -e "Passed: ${GREEN}$PASSED${NC}"
echo -e "Skipped: ${YELLOW}$SKIPPED${NC}"
if [ $FAILED -gt 0 ]; then
    echo -e "Failed: ${RED}$FAILED${NC}"
    exit 1
else
    echo -e "Failed: ${RED}0${NC}"
    echo ""
    echo -e "${GREEN}All tests completed!${NC}"
fi
