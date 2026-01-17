# Architecture Overview

## System Architecture

The Mauritania CLI is designed as a modular, transport-agnostic system that enables remote development through various communication channels. The architecture follows clean separation of concerns with clear boundaries between components.

## Core Components

### 1. CLI Layer (`cmd/mauritania-cli/`)
The command-line interface that users interact with directly.

**Responsibilities:**
- Command parsing and validation
- User interaction and feedback
- Configuration management
- Queue management and status display

**Key Files:**
- `main.go` - Application entry point
- `cmd/*.go` - Individual CLI commands
- Configuration handling and persistence

### 2. Transport Layer (`internal/transports/`)
Handles communication with various external services.

**Supported Transports:**
- **WhatsApp** - WhatsMeow-based integration
- **Telegram** - Bot API integration
- **Facebook** - Messenger API integration
- **SM APOS Shipper** - Secure network provider

**Interface Contract:**
```go
type TransportClient interface {
    SendMessage(recipient, message string) (*MessageResponse, error)
    ReceiveMessages() ([]*IncomingMessage, error)
    GetStatus() (*TransportStatus, error)
    ValidateCredentials() error
    GetRateLimit() (*RateLimit, error)
}
```

### 3. Service Layer (`internal/services/`)
Business logic and orchestration.

**Key Services:**
- **CommandAuthService** - Authentication and authorization
- **ShipperSessionManager** - Session management for secure transport
- **ShipperCommandExecutor** - Command execution orchestration
- **TransportSelector** - Intelligent transport selection

### 4. Model Layer (`internal/models/`)
Data structures and type definitions.

**Key Models:**
- `Command` - Command representation
- `CommandResult` - Execution results
- `ShipperSession` - Session management
- `TransportStatus` - Transport health status

### 5. Utility Layer (`internal/utils/`)
Shared utilities and helpers.

**Key Utilities:**
- **Config** - Configuration management
- **Logger** - Structured logging
- **Network** - Connectivity monitoring
- **Queue** - Offline command queuing
- **RateLimiter** - Rate limiting implementation
- **CommandEncryption** - Secure command transport

## Data Flow Architecture

```
┌─────────────┐    ┌──────────────┐    ┌─────────────┐
│   User CLI  │───▶│  Transport   │───▶│   Remote    │
│             │    │  Selection   │    │   Service   │
└─────────────┘    └──────────────┘    └─────────────┘
        ▲                    │               │
        │                    ▼               │
┌─────────────┐    ┌──────────────┐         │
│  Command    │◀───│   Queue &    │         │
│   Results   │    │   Retry      │         │
└─────────────┘    └──────────────┘         ▼
                                               │
                                               ▼
                                    ┌─────────────┐
                                    │   Results   │
                                    │  Display    │
                                    └─────────────┘
```

## Security Architecture

### Transport Security
- **End-to-end encryption** for sensitive commands
- **Transport layer security** (TLS/HTTPS)
- **API key protection** and rotation
- **Webhook signature validation**

### Command Security
- **Input validation** and sanitization
- **Command allowlisting** for authorized commands
- **Length limits** and complexity checks
- **Injection prevention** measures

### Session Security
- **Session token encryption** at rest
- **Automatic session expiration**
- **Secure credential storage**
- **Audit logging** for all operations

## Network Resilience

### Offline-First Design
- **Command queuing** when offline
- **Automatic retry** when connectivity returns
- **Intelligent backoff** strategies
- **Partial result handling**

### Multi-Transport Failover
- **Automatic transport selection** based on availability
- **Health monitoring** and status tracking
- **Rate limit awareness** across transports
- **Cost optimization** between transport options

## Performance Characteristics

### Mobile Optimization
- **Low memory footprint** (< 15MB binary)
- **Battery-efficient** background processing
- **Minimal network usage** with compression
- **Fast startup times** (< 100ms)

### Scalability Considerations
- **Horizontal scaling** through multiple transport instances
- **Load balancing** across available transports
- **Resource pooling** for connection reuse
- **Caching layers** for frequently accessed data

## Deployment Architecture

### Termux Deployment
```bash
# Single binary deployment
mauritania-cli-termux
├── Configuration: ~/.mauritania-cli/
├── Logs: ~/.mauritania-cli/logs/
├── Cache: ~/.mauritania-cli/cache/
└── Database: ~/.mauritania-cli/commands.db
```

### Directory Structure
```
/usr/local/bin/
└── mauritania-cli          # Main binary

~/.mauritania-cli/
├── config.toml            # Configuration file
├── commands.db            # SQLite database
├── logs/                  # Log files
│   ├── mauritania-cli.log
│   └── transports.log
└── cache/                 # Cached data
    ├── whatsapp/
    ├── telegram/
    └── shipper/
```

## Monitoring and Observability

### Logging Architecture
- **Structured logging** with JSON format
- **Multiple log levels** (DEBUG, INFO, WARN, ERROR)
- **Transport-specific logs** for debugging
- **Performance metrics** logging

### Metrics Collection
- **Command execution times**
- **Transport success/failure rates**
- **Network latency measurements**
- **Queue depth and processing rates**

### Health Checks
- **Transport connectivity** monitoring
- **Database health** verification
- **Memory usage** tracking
- **Goroutine monitoring**

## Error Handling Strategy

### Error Classification
- **Network Errors** - Temporary connectivity issues
- **Authentication Errors** - Credential or permission issues
- **Rate Limit Errors** - API quota exceeded
- **Validation Errors** - Input validation failures
- **Execution Errors** - Command execution failures

### Recovery Strategies
- **Retry Logic** - Exponential backoff for transient errors
- **Fallback Transports** - Automatic transport switching
- **Queue Persistence** - Command preservation across restarts
- **Graceful Degradation** - Reduced functionality when services unavailable

## Future Extensibility

### Plugin Architecture
- **Transport plugins** for new communication channels
- **Command plugins** for specialized command types
- **Storage plugins** for different database backends
- **UI plugins** for alternative interfaces

### API Extensions
- **REST API** for programmatic access
- **Webhook integrations** for external services
- **GraphQL API** for complex queries
- **WebSocket support** for real-time updates

This architecture provides a solid foundation for remote development in low-connectivity environments while maintaining security, reliability, and extensibility.