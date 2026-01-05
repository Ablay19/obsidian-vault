# AGENTS.md

This file contains build/lint/test commands and code style guidelines for agentic coding agents working in this repository.

## Build Commands

### Core Build Commands
- `make build` - Build the main bot Docker image
- `make build-ssh` - Build the SSH server Docker image
- `make docker-build-all` - Build all Docker images
- `make run-local` - Run the bot locally (clears CGO flags to avoid onnxruntime issues)

### Database Commands
- `make sqlc-generate` - Generate SQLC code from queries

### Docker Commands
- `make up` - Build and start the main bot container
- `make ssh-up` - Build and start the SSH server container
- `make down` - Stop and remove the main bot container
- `make ssh-down` - Stop and remove the SSH server container
- `make restart` - Restart the main bot container
- `make ssh-restart` - Restart the SSH server container

### Kubernetes Commands
- `make k8s-apply` - Apply Kubernetes manifests
- `make k8s-delete` - Delete Kubernetes manifests

## Test Commands

### Running Tests
- `go test ./...` - Run all tests
- `go test -v ./...` - Run all tests with verbose output
- `go test ./internal/bot` - Run tests for a specific package
- `go test -run TestProcessFileWithAI_Success ./internal/bot` - Run a single test
- `go test -run TestProcessFileWithAI ./internal/bot` - Run tests matching a pattern

### Test Coverage
- `go test -cover ./...` - Run tests with coverage
- `go test -coverprofile=coverage.out ./...` - Generate coverage profile
- `go tool cover -html=coverage.out` - View coverage in HTML

## Lint Commands

### Go Formatting
- `go fmt ./...` - Format all Go code (REQUIRED before commits)
- `gofmt -w .` - Format and write changes to files

### Go Linting
- `go vet ./...` - Run Go vet for potential issues
- `golangci-lint run` - Run golangci-lint (if installed)

### Build Verification
- `go build ./...` - Build all packages to verify compilation

## Code Style Guidelines

### General Principles
- Follow Go conventions and idiomatic patterns
- Use `gofmt` for all code formatting (non-negotiable)
- Keep functions small and focused
- Prefer explicit error handling over panics
- Use interfaces for abstraction and testability

### Import Organization
- Group imports in three blocks: standard library, third-party, internal packages
- Use blank lines between import groups
- Sort imports alphabetically within each group
- Use `goimports` if available for automatic import management

Example:
```go
import (
    "context"
    "fmt"
    "os"

    "github.com/gin-gonic/gin"
    "github.com/spf13/cobra"

    "obsidian-automation/internal/bot"
    "obsidian-automation/internal/dashboard"
)
```

### Naming Conventions
- Use `MixedCaps` for exported names, `mixedCaps` for unexported
- Package names: short, single-word, lowercase
- Interface names: method name + `-er` suffix for single-method interfaces
- Avoid `Get` prefix for getters (e.g., `Owner()` not `GetOwner()`)

### Function and Variable Naming
- Use descriptive names that convey purpose
- Prefer clarity over brevity
- Use camelCase for local variables
- Use meaningful variable names, not single letters (except in loops)

### Error Handling
- Always handle errors explicitly, never ignore with `_`
- Return errors as the last return value: `(result, error)`
- Use `fmt.Errorf` for wrapping errors with context
- Create custom error types for domain-specific errors

Example:
```go
func processFile(filename string) error {
    data, err := os.ReadFile(filename)
    if err != nil {
        return fmt.Errorf("failed to read file %s: %w", filename, err)
    }
    // process data
    return nil
}
```

### Struct Design
- Use field tags for serialization (JSON, YAML, etc.)
- Group related fields together
- Use pointer types for optional fields or large structs
- Provide constructor functions for complex initialization

Example:
```go
type Config struct {
    Port     int    `json:"port"`
    Host     string `json:"host"`
    Database *DatabaseConfig `json:"database,omitempty"`
}

func NewConfig(port int, host string) *Config {
    return &Config{
        Port: port,
        Host: host,
    }
}
```

### Concurrency
- Use goroutines for concurrent operations
- Use channels for communication between goroutines
- Prefer buffered channels for known workloads
- Use `context.Context` for cancellation and timeouts
- Use `sync.WaitGroup` for coordinating multiple goroutines

### Testing Guidelines
- Use table-driven tests for multiple test cases
- Name tests descriptively: `TestFunctionName_Scenario_ExpectedResult`
- Use `t.Helper()` for helper functions in tests
- Mock external dependencies using interfaces
- Use testify/assert for assertions if available

Example:
```go
func TestProcessFile_WithValidData_ReturnsSuccess(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
        wantErr  bool
    }{
        {
            name:     "valid input",
            input:    "test data",
            expected: "processed",
            wantErr:  false,
        },
        // more test cases
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := processFile(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("processFile() error = %v, wantErr %v", err, tt.wantErr)
                return
            }
            if result != tt.expected {
                t.Errorf("processFile() = %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Package Structure
- `cmd/` - Main applications and CLI commands
- `internal/` - Private application code
- `pkg/` - Public library code (if any)
- `api/` - API definitions (OpenAPI, protobuf, etc.)
- `web/` - Web UI assets and templates
- `configs/` - Configuration files
- `scripts/` - Build and deployment scripts

### Documentation
- Add package comments explaining the purpose
- Document exported functions, types, and constants
- Use examples in documentation for complex APIs
- Keep comments up-to-date with code changes

### Performance Considerations
- Use `strings.Builder` for string concatenation in loops
- Pre-allocate slices with known capacity using `make([]T, 0, capacity)`
- Use `sync.Pool` for object reuse in hot paths
- Profile with `pprof` when optimizing performance

### Security Best Practices
- Never log sensitive information (passwords, tokens, keys)
- Use environment variables for configuration secrets
- Validate all external inputs
- Use parameterized queries for database operations
- Implement proper authentication and authorization

## Environment Setup

### Required Environment Variables
- `TELEGRAM_BOT_TOKEN` - Telegram bot token
- `TURSO_DATABASE_URL` - Turso database URL
- `TURSO_AUTH_TOKEN` - Turso database auth token

### Optional Environment Variables
- `GEMINI_API_KEYS` - Comma-separated list of Gemini API keys
- `GROQ_API_KEY` - Groq API key
- `DASHBOARD_PORT` - Web dashboard port (default: 8080)
- `SSH_PORT` - SSH server port (default: 2222)
- `SSH_API_PORT` - SSH server API port (default: 8081)

## Development Workflow

1. Before coding: `make run-local` to verify the application starts
2. During development: Use `go test ./...` frequently to verify changes
3. Before committing: `go fmt ./...` and `go vet ./...`
4. For Docker changes: `make build` to verify Docker build
5. For database changes: `make sqlc-generate` after modifying SQL queries

## Testing Strategy

- Unit tests for business logic and utilities
- Integration tests for external service interactions
- Table-driven tests for multiple scenarios
- Mock implementations for external dependencies
- End-to-end tests for critical user flows

## Common Patterns

### Dependency Injection
- Use constructor functions to inject dependencies
- Define interfaces for external services
- Use struct composition for related functionality

### Configuration Management
- Use Viper for configuration loading
- Support both environment variables and config files
- Provide default values for optional settings

### Logging
- Use structured logging with logrus or zap
- Include context information in log messages
- Use appropriate log levels (Debug, Info, Warn, Error)

### HTTP Handlers
- Use Gin framework for HTTP routing
- Implement middleware for common concerns
- Return consistent JSON response formats
- Handle errors gracefully with proper HTTP status codes