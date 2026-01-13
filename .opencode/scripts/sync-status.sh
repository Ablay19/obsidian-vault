#!/bin/bash
# Opencode Multi-Instance Synchronization Script

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"
SHARED_DIR="$PROJECT_ROOT/.opencode/shared"

echo "🔄 Opencode Multi-Instance Synchronization"
echo "========================================="

# Check if shared directory exists
if [ ! -d "$SHARED_DIR" ]; then
    echo "❌ Shared directory not found: $SHARED_DIR"
    echo "Creating shared directory..."
    mkdir -p "$SHARED_DIR"
fi

# Load coordination configuration
COORDINATION_FILE="$SHARED_DIR/coordination.yaml"

if [ ! -f "$COORDINATION_FILE" ]; then
    echo "❌ Coordination file not found: $COORDINATION_FILE"
    echo "Please ensure coordination.yaml exists in shared directory"
    exit 1
fi

echo "📋 Current Coordination Status:"
echo "----------------------------"

# Read current instance (from environment or config)
CURRENT_INSTANCE="${OPENCODE_INSTANCE:-$(hostname)}"
echo "Current Instance: $CURRENT_INSTANCE"

# Check for pending sync messages
echo ""
echo "📬 Pending Sync Messages:"
echo "------------------------"

if [ -d "$SHARED_DIR/messages" ]; then
    for message_file in "$SHARED_DIR/messages"/*.md; do
        if [ -f "$message_file" ]; then
            echo "📄 $(basename "$message_file")"
            head -10 "$message_file" | grep "^##"
        fi
    done
else
    echo "No pending messages"
fi

# Check coordination status
echo ""
echo "🤝 Coordination Status:"
echo "----------------------"

if command -v yq >/dev/null 2>&1; then
    # Use yq if available for YAML parsing
    echo "📊 Task Distribution:"
    yq '.projects | to_entries | .[] | "\(.key): \(.value.status) (owner: \(.value.owner_instance))"' "$COORDINATION_FILE"
    
    echo ""
    echo "⚡ Active Conflicts:"
    yq '.active_conflicts | length' "$COORDINATION_FILE" "active conflicts"
    
    echo ""
    echo "📞 Pending Messages:"
    yq '.messages.pending | length' "$COORDINATION_FILE" "pending messages"
else
    echo "⚠️  yq not found. Install yq for better YAML parsing:"
    echo "   brew install yq  # macOS"
    echo "   pip install yq   # Python"
fi

# Sync actions
echo ""
echo "🚀 Sync Actions:"
echo "---------------"

echo "1. Status check: $SCRIPT_DIR/sync-status.sh"
echo "2. Send message: $SCRIPT_DIR/send-message.sh"
echo "3. Update tasks: $SCRIPT_DIR/update-tasks.sh"
echo "4. Resolve conflicts: $SCRIPT_DIR/resolve-conflicts.sh"

# Auto-sync if requested
if [ "$1" = "--auto" ]; then
    echo ""
    echo "🔄 Auto-sync enabled..."
    
    # Update coordination file with current status
    TIMESTAMP=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
    echo "Last sync: $TIMESTAMP"
    
    # Update last_sync in coordination file
    if command -v yq >/dev/null 2>&1; then
        yq eval ".sync_state.last_sync = \"$TIMESTAMP\"" -i "$COORDINATION_FILE"
        echo "✅ Updated last sync timestamp"
    fi
fi

# Check for conflicts
echo ""
echo "⚠️  Conflict Detection:"
echo "----------------------"

# Look for task overlaps
if [ -f "$PROJECT_ROOT/specs/002-telegram-ai-bot/tasks.md" ]; then
    echo "📋 Checking task completion status..."
    
    # Count completed vs pending tasks
    COMPLETED=$(grep -c "\[x\] T0" "$PROJECT_ROOT/specs/002-telegram-ai-bot/tasks.md" || echo "0")
    PENDING=$(grep -c "\[ \] T0" "$PROJECT_ROOT/specs/002-telegram-ai-bot/tasks.md" || echo "0")
    
    echo "Completed tasks: $COMPLETED"
    echo "Pending tasks: $PENDING"
    
    if [ "$PENDING" -gt 0 ]; then
        echo "📝 Pending tasks:"
        grep "\[ \] T0" "$PROJECT_ROOT/specs/002-telegram-ai-bot/tasks.md" | head -5
    fi
fi

# Recommendations
echo ""
echo "💡 Recommendations:"
echo "------------------"

echo "1. 📞 Send updates to other instances about your progress"
echo "2. 🔄 Sync task completion status regularly"
echo "3. 🤝 Coordinate on shared resources (AI models, files)"
echo "4. 📋 Update coordination file when tasks are completed"
echo "5. 🚀 Plan next phase completion together"

echo ""
echo "✅ Sync check completed!"
echo "Run '$0 --auto' to update timestamp automatically"