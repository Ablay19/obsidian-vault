# Opencode Session Management

This directory manages persistent session data, conversation history, and user preferences.

## Session Structure

Each session contains:
- `metadata.json` - Session information and settings
- `conversation.json` - Full conversation history
- `context.json` - Current working context and state
- `preferences.json` - User preferences and customizations
- `cache.json` - Cached data and optimizations

## Session Files

### `metadata.json`
```json
{
  "session_id": "ses_abc123",
  "created_at": "2026-01-13T10:30:00Z",
  "last_activity": "2026-01-13T11:45:00Z",
  "user_agent": "opencode/1.0.0",
  "environment": "development",
  "status": "active"
}
```

### `conversation.json`
```json
{
  "messages": [
    {
      "id": "msg_001",
      "timestamp": "2026-01-13T10:30:00Z",
      "role": "user",
      "content": "Help me analyze this codebase",
      "tool_calls": [],
      "metadata": {}
    }
  ],
  "total_messages": 1,
  "context_window_used": 1000
}
```

### `context.json`
```json
{
  "working_directory": "/home/user/project",
  "active_files": ["src/main.js", "README.md"],
  "recent_commands": ["read", "grep", "bash"],
  "state": {
    "current_task": "codebase_analysis",
    "phase": "exploration",
    "findings": []
  }
}
```

## Session Lifecycle

### 1. **Session Creation**
- Generate unique session ID
- Initialize metadata and conversation
- Load user preferences
- Set working environment

### 2. **Session Updates**
- Append new messages
- Update activity timestamps
- Maintain context window
- Cache relevant data

### 3. **Session Persistence**
- Save conversation history
- Store context and preferences
- Maintain working state
- Compress old data

### 4. **Session Cleanup**
- Remove inactive sessions
- Compress historical data
- Clear temporary caches
- Archive old conversations

## Session Configuration

### Settings (`config/sessions.yaml`)
```yaml
persistence:
  enabled: true
  backend: "file"
  directory: "./.opencode/sessions"
  format: "json"
  
retention:
  max_sessions: 50
  max_age_days: 30
  cleanup_interval_hours: 24
  
performance:
  context_window_size: 10000
  compression_enabled: true
  cache_ttl_seconds: 3600
  
security:
  encrypt_sessions: false
  access_control: true
  audit_logging: true
```

## Session API

### Create Session
```javascript
// Auto-generated on first interaction
const session = await SessionManager.create({
  user_id: "user123",
  preferences: { theme: "dark", language: "en" }
});
```

### Load Session
```javascript
const session = await SessionManager.load("session_id");
```

### Update Session
```javascript
await session.addMessage({
  role: "assistant",
  content: "Here's the analysis...",
  tool_calls: ["grep", "read"]
});
```

### Save Session
```javascript
await session.save();
```

## Session Features

### **Context Persistence**
- Maintain working directory state
- Remember active files and projects
- Preserve conversation context
- Track user preferences

### **Intelligent Caching**
- Cache file reads and searches
- Optimize repeated operations
- Predictive data loading
- Memory-efficient storage

### **Multi-Instance Support**
- Share sessions across instances
- Synchronize conversation history
- Coordinate context changes
- Prevent conflicts

### **Privacy & Security**
- Optional session encryption
- Access control and permissions
- Audit trail of all activities
- Secure data deletion

## Session Analytics

Track and analyze:
- Session duration and frequency
- Common usage patterns
- Tool usage statistics
- Error rates and recovery
- User satisfaction metrics

## Session Troubleshooting

### Common Issues
- **Corrupted session files** - Automatic backup recovery
- **Memory leaks** - Automatic context pruning
- **Synchronization conflicts** - Merge resolution strategies
- **Performance degradation** - Cache optimization

### Debug Information
Enable debug mode for detailed session logging:
```bash
/opencode config set debug.sessions true
```