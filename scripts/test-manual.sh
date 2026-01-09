#!/bin/bash

set -e

# Manual Testing Checklist for Obsidian Bot
# Guides through testing key functionality after deployment

source "$(dirname "$0")/setup/common.sh"

echo "🧪 Obsidian Bot Manual Testing Checklist"
echo "======================================="

# Check if services are running
print_header "Service Status Check"
if docker-compose ps | grep -q "Up"; then
    print_success "Docker services are running"
else
    print_error "Some services are not running. Check with: docker-compose ps"
    exit 1
fi

# Check dashboard
print_status "Testing dashboard..."
if curl -f http://localhost:8080/api/services/status > /dev/null 2>&1; then
    print_success "Dashboard is responding"
else
    print_error "Dashboard is not responding"
    exit 1
fi

echo ""
print_header "Manual Testing Checklist"
echo "Follow these steps to verify functionality:"
echo ""

echo "📱 Telegram Bot Testing:"
echo "  1. Send /start to your bot"
echo "  2. Send /help to see available commands"
echo "  3. Send /stats to check statistics"
echo "  4. Send a PDF or image file"
echo "  5. Send /process to process the file"
echo "  6. Check if note was created in vault"
echo ""

echo "🌐 Dashboard Testing:"
echo "  1. Open http://localhost:8080"
echo "  2. Check service status page"
echo "  3. Upload a file via dashboard"
echo "  4. Verify real-time updates work"
echo "  5. Test AI provider switching"
echo ""

echo "🤖 AI Features Testing:"
echo "  1. Test different AI providers (/setprovider)"
echo "  2. Send complex prompts"
echo "  3. Test file analysis with images/PDFs"
echo "  4. Check response quality and speed"
echo ""

echo "🔄 Integration Testing:"
echo "  1. Test WebSocket connections"
echo "  2. Verify database persistence"
echo "  3. Check Redis caching"
echo "  4. Test concurrent users"
echo ""

echo "📊 Performance Testing:"
echo "  1. Monitor memory usage"
echo "  2. Check response times"
echo "  3. Test with large files"
echo "  4. Verify error handling"
echo ""

print_info "Testing Commands:"
echo "  • View logs: docker-compose logs -f"
echo "  • Check resource usage: docker stats"
echo "  • Restart services: docker-compose restart"
echo "  • Stop all: docker-compose down"
echo ""

print_success "🎯 Ready for manual testing!"
print_info "Mark each item as you complete it."
print_info "Report any issues found during testing."