# Quickstart: MCP Integration for Problem Handling

**Feature**: 001-add-mcp-integration
**Date**: January 18, 2026

## Overview

The MCP server enables AI assistants to access diagnostic tools for troubleshooting Mauritania CLI issues. Get started in 3 steps.

## Prerequisites

- Mauritania CLI installed and configured
- Go 1.21+ for development
- AI client supporting MCP (e.g., Claude Desktop, custom integrations)

## Step 1: Start the MCP Server

### Local Development (Stdio Transport)
```bash
# Start MCP server for local AI client integration
mauritania-cli mcp-server --transport stdio
```

### HTTP Transport (Remote Access)
```bash
# Start HTTP server on port 8080
mauritania-cli mcp-server --transport http --port 8080

# With custom host
mauritania-cli mcp-server --transport http --host 0.0.0.0 --port 8080
```

## Step 2: Configure AI Client

### Claude Desktop Configuration
Add to `claude_desktop_config.json`:
```json
{
  "mcpServers": {
    "mauritania-cli": {
      "command": "mauritania-cli",
      "args": ["mcp-server", "--transport", "stdio"]
    }
  }
}
```

### Custom AI Integration
For HTTP transport:
```json
{
  "mcpServers": {
    "mauritania-cli": {
      "url": "http://localhost:8080/mcp"
    }
  }
}
```

## Step 3: Test Diagnostic Tools

### Check Transport Status
```
AI: What's the current status of the WhatsApp transport?
MCP Server: WhatsApp transport is connected. Last message sent 5 minutes ago.
```

### View Recent Logs
```
AI: Show me error logs from the last hour
MCP Server: Found 3 errors:
- [14:30] WhatsApp connection timeout
- [14:45] Telegram API rate limit exceeded
- [15:00] Configuration validation failed
```

### Run Diagnostics
```
AI: Run full system diagnostics
MCP Server: Diagnostics completed:
✓ All transports configured correctly
✓ Database connections healthy
✓ No security vulnerabilities detected
⚠ Rate limiting active (normal)
```

### Test Connections
```
AI: Test connection to Telegram transport
MCP Server: Telegram connection test successful. Response time: 250ms.
```

### View Error Metrics
```
AI: Show error trends for the last 24 hours
MCP Server: Error metrics:
- Total errors: 12
- Error rate: 0.5 per hour
- Top error: Connection timeout (40%)
```

## Available Tools

| Tool | Purpose | Parameters |
|------|---------|------------|
| `status` | Check transport connectivity | None |
| `logs` | Retrieve filtered logs | `level`, `hours`, `component` |
| `diagnostics` | Run system health checks | None |
| `config` | View sanitized configuration | `transport` |
| `test_connection` | Test specific transport | `transport` |
| `error_metrics` | Get error statistics | `window` |

## Troubleshooting

### Server Won't Start
- Check Go version: `go version` (need 1.21+)
- Verify CLI installation: `mauritania-cli --version`
- Check permissions for log/config access

### AI Client Can't Connect
- Stdio: Ensure AI client supports MCP stdio transport
- HTTP: Verify server is running on correct host/port
- Check firewall settings for HTTP transport

### Tools Return Errors
- Verify CLI configuration is valid
- Check log files for detailed error messages
- Ensure transports are properly configured

## Development

### Building from Source
```bash
git clone <repository>
cd mauritania-cli
go build -o mauritania-cli ./cmd/mauritania-cli
```

### Running Tests
```bash
go test ./internal/mcp/...
```

### Adding New Tools
1. Create new tool file in `internal/mcp/tools/`
2. Implement handler function with MCP interfaces
3. Register tool in server initialization
4. Add tests and update documentation

## Security Notes

- All sensitive data is automatically sanitized
- Rate limiting prevents abuse (100 requests/hour)
- No persistent authentication required
- Local transport (stdio) recommended for security
- HTTP transport should use HTTPS in production

## Next Steps

- Integrate with your preferred AI client
- Customize tool parameters for your use case
- Monitor MCP server performance and usage
- Contribute improvements back to the project