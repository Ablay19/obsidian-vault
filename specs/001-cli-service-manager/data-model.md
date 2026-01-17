# Data Model: CLI Service Manager for Termux

**Date**: January 17, 2025
**Feature**: CLI Service Manager for Termux

## Overview

The CLI Service Manager manages service definitions, runtime state, and operational history. All data is stored locally in JSON configuration files with optional SQLite persistence for command history.

## Core Entities

### ServiceConfig
Service definition and configuration data.

**Fields:**
- `id: string` - Unique service identifier (e.g., "ai-manim-worker")
- `name: string` - Human-readable service name
- `type: "docker" | "kubernetes" | "process"` - Service orchestration type
- `image?: string` - Docker image name (for Docker/K8s services)
- `command?: string[]` - Process command array (for process services)
- `ports: PortMapping[]` - Port mappings for the service
- `dependencies: string[]` - Array of service IDs this service depends on
- `environment: Record<string, string>` - Environment variables
- `resources: ResourceLimits` - CPU/memory limits and requests
- `healthCheck: HealthCheckConfig` - Health monitoring configuration
- `volumes?: VolumeMapping[]` - Volume mounts (optional)

**Validation Rules:**
- `id` must be lowercase, alphanumeric with hyphens
- `dependencies` cannot contain circular references
- `ports` must not conflict with other services
- `resources` must respect platform limitations (especially Termux)

**State Transitions:**
- `defined` → `starting` → `healthy` | `unhealthy` → `stopping` → `stopped`

### ServiceStatus
Real-time service operational status.

**Fields:**
- `serviceId: string` - Reference to ServiceConfig.id
- `state: ServiceState` - Current operational state
- `health: HealthStatus` - Health check results
- `lastSeen: Date` - Last successful communication
- `uptime: number` - Seconds since last start
- `resourceUsage: ResourceUsage` - Current CPU/memory usage
- `errorMessage?: string` - Last error message (if any)
- `restartCount: number` - Number of automatic restarts

**Validation Rules:**
- `health` must be updated every 30 seconds when service is running
- `resourceUsage` must not exceed configured limits
- `restartCount` should trigger alerts if > 5 in 5 minutes

### CommandHistory
Audit trail of CLI operations.

**Fields:**
- `id: string` - Unique command execution ID
- `timestamp: Date` - When command was executed
- `command: string` - CLI command that was run
- `args: string[]` - Command arguments
- `user: string` - User who executed command
- `environment: string` - Target environment (dev/staging/prod)
- `status: "success" | "failure" | "partial"` - Command execution result
- `duration: number` - Execution time in milliseconds
- `affectedServices: string[]` - Services impacted by command
- `errorMessage?: string` - Error details if failed

**Validation Rules:**
- `duration` should be tracked for performance monitoring
- `affectedServices` must match actual service operations
- Historical data should be retained for 90 days minimum

### EnvironmentProfile
Environment-specific configuration.

**Fields:**
- `name: string` - Environment identifier (dev/staging/prod)
- `baseUrl: string` - Base URL for the environment
- `services: Record<string, ServiceOverride>` - Service-specific overrides
- `credentials: CredentialConfig` - Authentication credentials
- `network: NetworkConfig` - Network-specific settings
- `features: FeatureFlags` - Feature toggles for the environment

**Validation Rules:**
- `baseUrl` must be valid URL format
- `services` overrides must match existing service configurations
- `credentials` should use secure storage mechanisms

## Supporting Types

### PortMapping
```typescript
interface PortMapping {
  container: number;    // Internal container port
  host: number;         // External host port
  protocol: "tcp" | "udp";
}
```

### ResourceLimits
```typescript
interface ResourceLimits {
  cpu: {
    request: string;    // e.g., "100m", "0.1"
    limit: string;      // e.g., "500m", "0.5"
  };
  memory: {
    request: string;    // e.g., "64Mi", "128Mi"
    limit: string;      // e.g., "256Mi", "512Mi"
  };
}
```

### HealthCheckConfig
```typescript
interface HealthCheckConfig {
  type: "http" | "tcp" | "command";
  endpoint?: string;     // For HTTP checks
  port?: number;         // For TCP checks
  command?: string[];    // For command checks
  interval: number;      // Check interval in seconds
  timeout: number;       // Timeout in seconds
  retries: number;       // Number of retries before failure
}
```

### ServiceState
```typescript
type ServiceState =
  | "stopped"      // Service is not running
  | "starting"     // Service is being started
  | "healthy"      // Service is running and healthy
  | "unhealthy"    // Service is running but failing health checks
  | "stopping"     // Service is being stopped
  | "failed"       // Service failed to start or crashed
  | "unknown";     // Status cannot be determined
```

## Relationships

### Service Dependencies
- Services form a directed acyclic graph (DAG)
- Dependency resolution uses topological sorting
- Circular dependencies are detected and prevented at configuration time

### Environment Overrides
- Base service configurations can be overridden per environment
- Overrides are merged with base configuration
- Environment-specific credentials are isolated

### Command Auditing
- All CLI operations are logged for troubleshooting
- Command history enables rollback and debugging
- Performance metrics are tracked for optimization

## Data Storage Strategy

### Configuration Files
- Service definitions stored in JSON files
- Environment profiles in separate JSON files
- Version controlled with Git for change tracking

### Runtime State
- Service status cached in memory with periodic persistence
- Command history stored in SQLite database
- Automatic cleanup of old history records

### Backup and Recovery
- Configuration files automatically backed up before changes
- Command history enables replay of operations
- Service state can be exported for migration