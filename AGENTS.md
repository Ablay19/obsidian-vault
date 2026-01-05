# AGENTS.md

## Overview

This file contains build/lint/test commands and code style guidelines for agentic coding in this repository.

## Core Architecture

Our agents are built on these principles:
- **Service-Based Architecture**: All major components use dependency injection
- **Interface-First Design**: Clear separation between public interfaces and implementations
- **Error Handling**: Comprehensive error handling with proper logging
- **Configuration Management**: Centralized configuration with environment variable validation
- **Testing-Driven Development**: Comprehensive unit and integration test coverage

## Agent Categories

### 1. **WhatsApp Service Agent**
**File**: `internal/whatsapp/service.go`
**Purpose**: Handle WhatsApp webhook processing and message handling
**Key Components**:
- Service interface for dependency injection
- Message type support (text, image, document, audio, video)
- Media download and storage capabilities
- Comprehensive validation and error handling

### 2. **Authentication Agent**
**Files**: `internal/auth/oauth.go`, `internal/auth/session.go`
**Purpose**: Secure user authentication and session management
**Key Components**:
- JWT-based session management with automatic refresh
- OAuth flow separation from session handling
- Secure repository pattern with prepared statements
- Multiple AI provider support

### 3. **CLI Agent**
**Files**: `cmd/cli/tui/views/*` (refactored)
**Purpose**: Command-line interface with consistent user experience
**Key Components**:
- Unified color palette and styling system
- Router-based navigation with back stack management
- Component-based architecture for reusability
- Consistent error handling and user feedback

### 4. **Visual Problem-Solving Agent**
**Files**: `internal/visualizer/*`
**Purpose**: AI-powered problem analysis and visual solution generation
**Key Components**:
- AI service for pattern recognition and root cause analysis
- Multi-modal input support (text, images, code, visual content)
- Dynamic diagram generation (Mermaid, PlantUML, GraphViz)
- Learning system based on user interactions

### 5. **Configuration Validation Agent**
**Files**: `internal/config/env_validation.go`
**Purpose**: Comprehensive environment variable validation and security checking
**Key Components**:
- Validation for all required and optional variables
- Production security warnings
- Comprehensive testing infrastructure
- Benchmarking capabilities

### 6. **Testing Agent**
**Files**: `tests/integration/*`
**Purpose**: End-to-end testing and validation
**Key Components**:
- Component-level unit tests with mocking
- Integration tests for complete workflows
- Performance testing and benchmarking
- Environment validation testing

## Build Commands

### Development Commands
```bash
# Build all components
make build

# Build specific services
make build-whatsapp
make build-auth
make build-cli

# Run tests
make test
make test-integration
make test-performance

# Environment validation
make validate-env
make test-security
```

### Code Quality Commands
```bash
# Format and lint
make fmt
make lint
make security-scan

# Static analysis
make analyze
```

## Lint and Quality Gates

### Required Checks
- `make lint` must pass without errors
- `go fmt` must format all files
- `go vet` must pass without warnings
- Security scan must find no critical vulnerabilities

### Security Standards
- No hardcoded credentials
- SQL injection prevention (prepared statements)
- Proper input validation and sanitization
- Secure session management with proper expiration

## Testing Standards

### Unit Tests
```bash
# Run all unit tests
make test-unit

# Run specific package tests
make test-auth
make test-whatsapp
make test-cli
```

### Integration Tests
```bash
# Run end-to-end tests
make test-integration

# Run performance tests
make test-performance

# Environment validation tests
make test-env-validation
```

## Deployment Commands

### Production Deployment
```bash
# Build for production
make build-prod

# Run comprehensive tests
make test-all

# Deploy with validation
make deploy-with-validation
```

## Agent Behavior Guidelines

### 1. **Error Handling**
- Always handle errors explicitly and return meaningful error messages
- Use structured logging with appropriate log levels
- Implement graceful degradation when possible
- Never expose internal errors to clients

### 2. **Logging Standards**
- Use structured logging (zap)
- Include context information in all log entries
- Use appropriate log levels (Debug, Info, Warn, Error)
- Avoid logging sensitive information

### 3. **Configuration Management**
- Validate all configuration on startup
- Use environment variables with proper defaults
- Implement configuration hot-reloading where appropriate
- Document all configuration options

### 4. **Security Standards**
- Never log or expose sensitive data
- Validate all inputs and sanitize outputs
- Use secure defaults for all configurations
- Implement proper authentication and authorization

### 5. **Performance Standards**
- Implement connection pooling and resource management
- Use appropriate timeouts for all external calls
- Monitor and log performance metrics
- Implement caching for frequently accessed data

## Agent Communication

### Inter-Agent Communication
- Use well-defined interfaces for agent-to-agent communication
- Implement message passing protocols with proper error handling
- Support asynchronous communication where appropriate
- Implement agent discovery and registration

### Agent Lifecycle Management**
- Proper initialization and cleanup procedures
- Graceful shutdown handling
- Resource usage monitoring and limits
- Health check capabilities for all agents

## Documentation Requirements

Each agent module must include:
- Purpose and scope definition
- Interface documentation
- Usage examples and integration guides
- Testing procedures and coverage reports
- Security considerations and best practices

## Compliance and Standards

### Code Standards
- Follow Go best practices and idioms
- Maintain consistent naming conventions
- Use interfaces for dependency injection
- Implement proper error handling
- Write comprehensive tests

### Security Standards
- Follow OWASP security guidelines
- Implement proper input validation
- Use secure authentication and session management
- Follow principle of least privilege
- Implement proper logging without sensitive data exposure

### Performance Standards
- Monitor resource usage and response times
- Implement efficient algorithms and data structures
- Use connection pooling and caching appropriately
- Regular performance testing and optimization

## Integration Testing

### Test Scenarios
- Complete user workflows from end-to-end
- Multi-agent communication and coordination
- Error handling and recovery scenarios
- Performance under load testing
- Configuration management and validation

### Test Coverage
- Minimum 80% line coverage for all critical paths
- 100% test coverage for all agent interfaces
- Integration tests for all major workflows
- Performance benchmarks for all critical operations

---

## Agent Development Process

1. **Planning**: Define agent scope and requirements
2. **Interface Design**: Create clear interface definitions
3. **Implementation**: Build core functionality with tests
4. **Integration**: Test agent in system context
5. **Documentation**: Complete documentation and examples
6. **Review**: Code review and security assessment

## Agent Deployment

### Development Environment
- Use Docker containers for isolation
- Implement health checks and monitoring
- Support environment-specific configurations
- Use proper secret management and rotation
- Implement logging and observability

### Production Environment
- High availability and scalability requirements
- Load balancing and failover capabilities
- Comprehensive monitoring and alerting
- Security hardening and compliance checks

---

## Agent Maintenance

### Regular Updates
- Bug fixes and security patches
- Performance optimization and feature enhancements
- Documentation updates and improvements
- Compatibility updates for new dependencies

### Support Procedures
- Troubleshooting guides and common issues
- Performance monitoring and optimization
- User feedback collection and analysis

---

**All agents must follow these guidelines to ensure consistent, secure, and maintainable contributions to the system.**