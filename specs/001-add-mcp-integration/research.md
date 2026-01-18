# Research Findings: MCP Integration for Problem Handling

**Date**: January 18, 2026
**Researcher**: opencode
**Feature**: 001-add-mcp-integration

## Decision: MCP SDK Selection

**Chosen**: github.com/modelcontextprotocol/go-sdk

**Rationale**: Official Anthropic MCP SDK with comprehensive Go support, active development, and alignment with MCP specification. Provides both stdio and HTTP transports out-of-the-box.

**Alternatives Considered**:
- mark3labs/mcp-go: Mature library with good documentation but less official
- mcp4go: Comprehensive but potentially over-engineered for CLI use case

## Decision: Transport Implementation

**Chosen**: Support both stdio and HTTP transports

**Rationale**: Stdio for local AI client integration (e.g., Claude Desktop), HTTP for remote/web-based AI access. Follows MCP best practices from specification.

**Implementation Approach**:
- Command-line flags for transport selection
- Default to stdio for security (local only)
- HTTP with configurable host/port for flexibility

## Decision: Tool Architecture

**Chosen**: Individual tool files with shared utilities

**Rationale**: Clean separation of concerns, easier testing, and maintainability. Each diagnostic tool (status, logs, etc.) in separate file.

**Pattern**:
- tools/status.go: Transport connectivity checks
- tools/logs.go: Log retrieval with filtering
- tools/diagnostics.go: Health checks and validation
- tools/config.go: Sanitized configuration exposure
- tools/test_conn.go: Connection testing utilities
- tools/metrics.go: Error aggregation and trends

## Decision: Data Sanitization Strategy

**Chosen**: Multi-layer sanitization with allowlists

**Rationale**: Critical for privacy compliance - prevent accidental exposure of credentials, keys, or sensitive paths.

**Implementation**:
- Remove API keys, passwords, tokens from all outputs
- Redact file paths outside project directory
- Limit log exposure to last 24 hours by default
- Structured allowlist for configuration fields

## Decision: Error Handling and Resilience

**Chosen**: Graceful degradation with informative errors

**Rationale**: MCP server should not crash CLI, provide meaningful feedback to AIs.

**Pattern**:
- Tool-specific error types with context
- Timeout handling for long-running operations
- Fallback responses for partial failures
- Structured error logging for debugging

## Decision: Performance Optimization

**Chosen**: Caching and async processing

**Rationale**: Meet <2s response requirement for AI usability.

**Techniques**:
- Cache transport status for 30 seconds
- Async log parsing with pagination
- Connection pooling for diagnostics
- Rate limiting at 100 requests/hour per session

## Decision: Integration with Existing CLI

**Chosen**: Separate process with shared utilities

**Rationale**: Avoid interference with main CLI, enable independent scaling.

**Architecture**:
- New `mcp-server` command in cmd/
- Shared internal packages for data access
- Process isolation for stability
- Configuration inheritance from main CLI

## Decision: Testing Strategy

**Chosen**: TDD with integration tests

**Rationale**: Constitution requirement for test-first, critical for AI tool reliability.

**Coverage**:
- Unit tests for each tool handler
- Integration tests for MCP protocol compliance
- Mock transports for offline testing
- Performance benchmarks for response times

## Key Security Considerations

Based on MCP security research:

1. **Input Validation**: All tool parameters validated against schemas, prevent injection attacks
2. **Access Control**: Rate limiting (100/hour), session timeouts, no persistent auth
3. **Data Protection**: Sanitization pipeline, no sensitive data exposure, path traversal prevention
4. **Transport Security**: HTTPS for HTTP transport, no plaintext credentials
5. **Audit Logging**: All tool calls logged with timestamps, but sanitized
6. **Fail-Safe Defaults**: Conservative permissions, minimal data exposure
7. **Session Isolation**: No cross-session data leakage, clean session teardown

## Performance Benchmarks

Target metrics from research:
- Tool response: <2 seconds (95th percentile)
- Concurrent sessions: 10 simultaneous
- Memory usage: <100MB per server instance
- Startup time: <5 seconds

## Integration Patterns

From Go MCP server examples:
- Command-line transport selection
- Structured tool registration
- Error handling with proper MCP error types
- Resource cleanup on shutdown
- Configuration via environment variables

## Risks Identified

1. **Data Leakage**: Mitigated by sanitization layers
2. **DoS via Tools**: Mitigated by rate limiting and timeouts
3. **Session Hijacking**: Mitigated by proper session management
4. **Path Traversal**: Mitigated by path validation and allowlists
5. **Performance Degradation**: Mitigated by caching and async processing

## Open Questions

- Specific AI client compatibility requirements?
- Detailed performance requirements for concurrent usage?
- Integration with existing monitoring/alerting systems?