# Opencode Logs

This directory contains system logs, activity tracking, and debugging information.

## Log Structure

### Log Files
- `opencode.log` - Main application log
- `sessions.log` - Session activity and errors
- `tools.log` - Tool execution logs
- `agents.log` - Agent performance and decisions
- `api.log` - API requests and responses
- `security.log` - Security events and access controls
- `performance.log` - Performance metrics and timing
- `errors.log` - Error tracking and stack traces

### Log Organization
```
logs/
├── current/           # Current day's logs
├── archive/          # Historical logs (rotated)
├── analysis/         # Log analysis and reports
└── debug/           # Debug-specific logs
```

## Log Formats

### Standard Log Entry
```json
{
  "timestamp": "2026-01-13T10:30:45.123Z",
  "level": "info",
  "component": "bash",
  "session_id": "ses_abc123",
  "operation": "execute_command",
  "message": "Command completed successfully",
  "metadata": {
    "command": "npm install",
    "duration": 15234,
    "exit_code": 0,
    "user_id": "user_456"
  }
}
```

### Error Log Entry
```json
{
  "timestamp": "2026-01-13T10:30:45.123Z",
  "level": "error",
  "component": "read",
  "session_id": "ses_abc123",
  "operation": "read_file",
  "message": "File not found",
  "error": {
    "code": "ENOENT",
    "message": "ENOENT: no such file or directory",
    "stack": "Error: ENOENT: no such file or directory...",
    "file_path": "/path/to/nonexistent.js"
  },
  "metadata": {
    "file_path": "src/components.js",
    "user_id": "user_456"
  }
}
```

### Performance Log Entry
```json
{
  "timestamp": "2026-01-13T10:30:45.123Z",
  "level": "debug",
  "component": "performance",
  "operation": "tool_execution",
  "message": "Tool performance metrics",
  "metrics": {
    "tool_name": "task",
    "execution_time": 15432,
    "memory_usage": "45MB",
    "tokens_used": 2500,
    "cache_hit": false,
    "parallel_operations": 3
  }
}
```

## Log Levels

### **DEBUG** (0)
- Detailed diagnostic information
- Development and troubleshooting
- Internal state changes
- Cache operations

### **INFO** (1)
- General operational information
- Successful operations
- Configuration changes
- Session start/end

### **WARN** (2)
- Warning conditions
- Performance degradation
- Resource usage alerts
- Fallback operations

### **ERROR** (3)
- Error conditions
- Failed operations
- Exception handling
- Security violations

### **FATAL** (4)
- Critical system failures
- Service unavailability
- Data corruption
- Security breaches

## Log Configuration

### Settings (`config/logging.yaml`)
```yaml
level: "info"                    # Minimum log level
format: "json"                  # json or text
rotation: "daily"               # daily, weekly, monthly
retention_days: 7              # How long to keep logs
max_file_size: "100MB"         # Max size per log file
compression: true               # Compress rotated logs

outputs:
  file:
    enabled: true
    directory: "./.opencode/logs"
    pattern: "{component}.log"
  
  console:
    enabled: false
    colors: true
    level: "info"
  
  remote:
    enabled: false
    endpoint: "https://logs.opencode.ai"
    api_key: "${LOG_API_KEY}"
    batch_size: 100
    flush_interval: 60

components:
  bash:
    level: "info"
    include_stderr: true
  
  agents:
    level: "debug"
    include_thinking: true
  
  security:
    level: "warn"
    include_sensitive: false
  
  performance:
    level: "info"
    include_metrics: true
```

## Log Analysis

### Built-in Analysis Tools
```bash
# Generate daily summary
opencode logs summary --today

# Analyze errors
opencode logs errors --last 24h

# Performance metrics
opencode logs performance --component task

# Security events
opencode logs security --level warn
```

### Custom Queries
```bash
# Search logs by component
opencode logs search component=bash level=error

# Filter by time range
opencode logs search --from "2026-01-13" --to "2026-01-14"

# User activity
opencode logs search user_id=user_456 level=info

# Performance analysis
opencode logs search operation=execute_command duration>5000
```

### Log Reports
Generate comprehensive reports:
```bash
# Daily operations report
opencode logs report --type daily --date today

# Performance analysis
opencode logs report --type performance --last 7d

# Security audit
opencode logs report --type security --last 30d

# Error analysis
opencode logs report --type errors --last 24h
```

## Log Monitoring

### Real-time Monitoring
```bash
# Follow live logs
opencode logs follow

# Component-specific monitoring
opencode logs follow --component agents

# Error monitoring
opencode logs follow --level error
```

### Alerts and Notifications
Configure alerts for specific events:
```yaml
# .opencode/config/alerts.yaml
alerts:
  high_error_rate:
    condition: "error_rate > 0.1"
    window: "5m"
    action: "notify"
    channels: ["email", "slack"]
  
  slow_operations:
    condition: "duration > 30s"
    action: "log"
    level: "warn"
  
  security_violations:
    condition: "level == 'fatal' AND component == 'security'"
    action: "immediate_notify"
    channels: ["email", "pagerduty"]
```

## Log Privacy and Security

### Data Sanitization
- Remove sensitive information (passwords, tokens, keys)
- Sanitize file paths and user data
- Mask personal information
- Encrypt sensitive log entries

### Access Control
- Role-based log access
- Audit trail for log access
- Secure log storage
- Compliance with data protection regulations

### Retention Policies
- Automatic log rotation
- Secure deletion of old logs
- Archive important logs
- Comply with legal requirements

## Troubleshooting with Logs

### Common Issues

#### **Tool Execution Failures**
```bash
# Find failed tool calls
opencode logs search level=error component=bash

# Check recent tool history
opencode logs search --last 1h operation=execute_command
```

#### **Performance Issues**
```bash
# Slow operations
opencode logs search duration>10000

# Memory usage
opencode logs search memory_usage>100MB
```

#### **Session Problems**
```bash
# Session lifecycle
opencode logs search session_id=ses_abc123

# Session errors
opencode logs search component=sessions level=error
```

### Debug Mode
Enable debug logging for detailed troubleshooting:
```bash
opencode config set log_level debug
opencode config set debug.tools true
opencode config set debug.agents true
```

## Log Analytics

### Metrics Tracked
- Tool execution frequency and duration
- Agent performance and success rates
- Error rates by component
- User activity patterns
- Resource usage trends

### Dashboard Integration
Integrate with monitoring dashboards:
- Grafana dashboards
- Prometheus metrics
- Kibana log analysis
- Custom analytics platforms

### Performance Optimization
- Identify bottlenecks
- Optimize slow operations
- Resource usage analysis
- Capacity planning insights