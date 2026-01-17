# Data Model: Mauritania Network Integration

**Date**: January 17, 2025
**Feature**: Mauritania Network Integration

## Overview

The Mauritania Network Integration manages remote command execution through multiple transport mechanisms. Data is stored locally for offline operation and security, with synchronization to remote services when connectivity allows.

## Core Entities

### SocialMediaCommand
Command sent via social media transport with execution metadata.

**Fields:**
- `id: string` - Unique command identifier (UUID)
- `senderId: string` - Social media user identifier
- `platform: "facebook" | "twitter" | "whatsapp" | "telegram"` - Social media platform
- `command: string` - Command text to execute
- `timestamp: Date` - When command was received
- `priority: "low" | "normal" | "high" | "urgent"` - Execution priority
- `status: CommandStatus` - Current execution status
- `transportId: string` - Social media message/post identifier
- `sessionId: string` - Associated execution session

**Validation Rules:**
- `command` length must not exceed 4000 characters (social media limits)
- `senderId` must be validated against whitelist for security
- `timestamp` must be within last 24 hours to prevent replay attacks

**State Transitions:**
- `received` → `queued` → `executing` → `completed` | `failed` → `responded`

### NetworkRoute
Available network transport path with performance characteristics.

**Fields:**
- `id: string` - Unique route identifier
- `type: "social_media" | "sm_apos" | "direct" | "nrt"` - Transport mechanism
- `provider: string` - Network provider name (e.g., "Mauritel", "Chinguitel")
- `costPerMB: number` - Cost in local currency per megabyte
- `bandwidth: number` - Available bandwidth in Mbps
- `latency: number` - Typical latency in milliseconds
- `reliability: number` - Uptime percentage (0-100)
- `lastTested: Date` - When route was last tested
- `isActive: boolean` - Whether route is currently available

**Validation Rules:**
- `costPerMB` must be >= 0
- `reliability` must be between 0 and 100
- `bandwidth` and `latency` must be measured values

### ShipperSession
Authenticated session with SM APOS Shipper service.

**Fields:**
- `id: string` - Session identifier
- `userId: string` - Authenticated user identifier
- `token: string` - Session authentication token (encrypted)
- `createdAt: Date` - Session creation timestamp
- `expiresAt: Date` - Session expiration timestamp
- `permissions: string[]` - Allowed command types and operations
- `rateLimit: RateLimit` - Current rate limiting status
- `lastActivity: Date` - Last command execution time

**Validation Rules:**
- `expiresAt` must be in the future for active sessions
- `token` must be encrypted at rest
- `permissions` must include at least basic command execution

### CommandResult
Result of command execution with output and metadata.

**Fields:**
- `id: string` - Result identifier (matches command ID)
- `commandId: string` - Reference to executed command
- `status: "success" | "failure" | "partial" | "timeout"` - Execution outcome
- `exitCode: number` - Process exit code (0 for success)
- `stdout: string` - Command standard output
- `stderr: string` - Command standard error
- `executionTime: number` - Execution time in milliseconds
- `transportUsed: string` - Which transport mechanism was used
- `cost: number` - Execution cost in local currency
- `completedAt: Date` - When execution finished

**Validation Rules:**
- `stdout` + `stderr` combined size must not exceed 10MB
- `executionTime` must be reasonable (max 300 seconds)
- `cost` must match transport pricing

### OfflineQueue
Queued commands for execution when connectivity is restored.

**Fields:**
- `id: string` - Queue entry identifier
- `command: SocialMediaCommand` - The queued command
- `queuedAt: Date` - When command was queued
- `priority: number` - Numeric priority for ordering
- `retryCount: number` - Number of retry attempts
- `lastRetry: Date` - When last retry was attempted
- `nextRetry: Date` - When next retry should occur
- `failureReason: string` - Why command is queued (if known)

**Validation Rules:**
- `retryCount` should not exceed 10 to prevent infinite loops
- `nextRetry` should use exponential backoff
- `priority` affects queue ordering

## Supporting Types

### CommandStatus
```typescript
type CommandStatus =
  | "received"      // Command received from social media
  | "validated"     // Command syntax and permissions validated
  | "queued"        // Command queued for execution
  | "executing"     // Command currently executing
  | "completed"     // Command executed successfully
  | "failed"        // Command execution failed
  | "responded"     // Response sent back via social media
  | "expired";      // Command expired without execution
```

### RateLimit
```typescript
interface RateLimit {
  requestsPerHour: number;    // Allowed requests per hour
  requestsRemaining: number;  // Remaining requests this hour
  resetTime: Date;           // When limit resets
  isThrottled: boolean;      // Whether currently throttled
}
```

### TransportMetrics
```typescript
interface TransportMetrics {
  transportType: string;
  successRate: number;        // Percentage of successful transmissions
  averageLatency: number;     // Average response time in ms
  totalCost: number;          // Total cost incurred
  messagesSent: number;       // Total messages sent
  lastUsed: Date;            // When transport was last used
}
```

## Relationships

### Command Execution Flow
- **SocialMediaCommand** → **OfflineQueue** (if offline)
- **SocialMediaCommand** → **NetworkRoute** (transport selection)
- **SocialMediaCommand** → **ShipperSession** (authentication)
- **SocialMediaCommand** → **CommandResult** (execution outcome)

### Network Optimization
- **NetworkRoute** tracks performance metrics over time
- **TransportMetrics** aggregated from multiple command executions
- Route selection algorithm uses historical performance data

### Session Management
- **ShipperSession** manages authentication state
- Sessions automatically refresh tokens before expiration
- Rate limiting tracked per session to prevent abuse

### Queue Management
- Commands automatically queued when transport unavailable
- Queue prioritizes urgent commands and retries failed executions
- Queue persistence survives application restarts

## Data Storage Strategy

### Local SQLite Database
- **Command History**: All executed commands with results
- **Session Data**: Authentication tokens and permissions
- **Queue Persistence**: Offline commands survive restarts
- **Metrics Storage**: Performance and cost tracking

### Configuration Files
- **Transport Configs**: API keys, endpoints, rate limits
- **Route Preferences**: Cost and performance preferences
- **Security Settings**: Whitelisted users, command restrictions

### In-Memory Caching
- **Active Sessions**: Fast lookup of authenticated sessions
- **Route Metrics**: Real-time performance data
- **Command Status**: Current execution state

### Backup and Recovery
- **Automatic Backups**: Configuration and critical data
- **Incremental Sync**: Changes synchronized when online
- **Data Integrity**: Checksums and validation on restore