# Data Model: MCP Integration for Problem Handling

**Date**: January 18, 2026
**Feature**: 001-add-mcp-integration

## Overview

The data model supports MCP server functionality for exposing diagnostic tools to AI assistants. All entities are designed for read-only access with strong data sanitization to prevent sensitive information exposure.

## Core Entities

### MCPServer

**Purpose**: Central server entity managing MCP protocol implementation and tool orchestration.

**Fields**:
- `Name`: string (required) - Server identifier ("Mauritania CLI Diagnostics")
- `Version`: string (required) - Semantic version ("1.0.0")
- `Transport`: enum (stdio, http) - Active transport mechanism
- `Tools`: []Tool (required) - Registered diagnostic tools
- `SessionCount`: int - Active concurrent sessions (max 10)
- `Uptime`: duration - Server runtime since start

**Validation**:
- Name: alphanumeric + hyphens, max 50 chars
- Version: semantic version format
- SessionCount: 0-10 range

**Relationships**:
- 1:many with DiagnosticTool
- 1:many with TransportStatus (read-only access)

### DiagnosticTool

**Purpose**: Individual MCP tool providing specific diagnostic functionality.

**Fields**:
- `Name`: string (required) - Tool identifier ("status", "logs", "diagnostics")
- `Description`: string (required) - Human-readable purpose
- `Schema`: JSONSchema (required) - Input parameter validation schema
- `Enabled`: bool - Tool availability flag
- `RateLimit`: int - Requests per hour (default 100)
- `LastExecuted`: timestamp - Most recent tool call

**Validation**:
- Name: lowercase alphanumeric + underscores, max 32 chars
- Description: max 200 chars
- Schema: valid JSON Schema structure
- RateLimit: 1-1000 range

**Relationships**:
- belongs to MCPServer
- 1:many with LogEntry (for logs tool)
- 1:many with ErrorMetric (for metrics tool)

### TransportStatus

**Purpose**: Real-time connectivity status for messaging transports.

**Fields**:
- `Transport`: enum (whatsapp, telegram, facebook, shipper) (required)
- `Connected`: bool - Current connection state
- `LastConnected`: timestamp - Most recent successful connection
- `LastError`: string - Last connection error (sanitized)
- `MessageCount`: int - Messages sent in current session
- `Uptime`: duration - Transport uptime percentage

**Validation**:
- Transport: valid enum value
- LastError: max 500 chars, no sensitive data
- MessageCount: non-negative integer
- Uptime: 0.0-1.0 range

**Relationships**:
- accessed by MCPServer (read-only)
- referenced by DiagnosticTool (status tool)

### LogEntry

**Purpose**: Structured log data for troubleshooting and analysis.

**Fields**:
- `Timestamp`: timestamp (required) - Log entry time
- `Level`: enum (debug, info, warn, error) (required)
- `Message`: string (required) - Log content
- `Component`: string - Source component (e.g., "whatsapp-transport")
- `RequestID`: string - Associated request identifier
- `ErrorCode`: string - Standardized error code
- `Metadata`: map[string]string - Additional context (sanitized)

**Validation**:
- Message: max 1000 chars, no sensitive data patterns
- Component: max 50 chars, alphanumeric + hyphens
- RequestID: UUID format if present
- Metadata: max 10 key-value pairs, values max 100 chars

**Relationships**:
- queried by DiagnosticTool (logs tool)
- filtered by time range and level

### ErrorMetric

**Purpose**: Aggregated error statistics and trends for monitoring.

**Fields**:
- `TimeWindow`: enum (hour, day, week) (required) - Aggregation period
- `StartTime`: timestamp (required) - Window start
- `EndTime`: timestamp (required) - Window end
- `ErrorCount`: int - Total errors in window
- `ErrorRate`: float - Errors per minute
- `TopErrors`: []ErrorSummary - Most frequent errors (max 5)
- `TransportErrors`: map[string]int - Errors by transport

**Validation**:
- TimeWindow: valid enum
- EndTime > StartTime
- ErrorCount: non-negative
- ErrorRate: non-negative float
- TopErrors: max 5 items

**Relationships**:
- calculated by DiagnosticTool (metrics tool)
- used for trend analysis

**Embedded Type - ErrorSummary**:
- `Code`: string - Error code
- `Count`: int - Occurrence count
- `Description`: string - Human-readable description

### SanitizedConfig

**Purpose**: Configuration data with sensitive fields redacted for safe exposure.

**Fields**:
- `TransportConfigs`: map[string]TransportConfig - Transport settings (sanitized)
- `GeneralSettings`: GeneralConfig - Non-sensitive general settings
- `SecuritySettings`: SecurityConfig - Security configuration (redacted)
- `LastModified`: timestamp - Configuration last change

**Validation**:
- All sensitive fields (API keys, passwords, tokens) must be redacted
- TransportConfigs limited to connection settings only
- No file paths or system-specific details

**Embedded Types**:
- `TransportConfig`: Connection settings (host, port, timeout)
- `GeneralConfig`: Log level, rate limits (non-sensitive)
- `SecurityConfig`: Redacted security flags (enabled/disabled only)

## Data Flow

1. **AI Request** → MCPServer receives tool call
2. **Tool Execution** → DiagnosticTool accesses read-only data sources
3. **Data Sanitization** → Sensitive information removed/filtered
4. **Response Formation** → Structured MCP response with sanitized data
5. **AI Consumption** → AI receives safe diagnostic information

## Validation Rules

### Global Rules
- No API keys, passwords, or tokens in any entity
- File paths limited to relative project paths
- Timestamps in UTC with nanosecond precision
- String fields trimmed and validated for length
- Numeric fields within reasonable bounds

### Business Rules
- TransportStatus updated every 30 seconds maximum
- LogEntry retention limited to 24 hours by default
- ErrorMetric calculations real-time with caching
- SanitizedConfig reflects current active configuration

## Performance Considerations

- TransportStatus: Cached for 30 seconds, async updates
- LogEntry: Indexed by timestamp and level for fast queries
- ErrorMetric: Pre-calculated aggregations for quick access
- SanitizedConfig: Cached with invalidation on config changes