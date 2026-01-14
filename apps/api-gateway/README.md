# API Gateway

The API Gateway is the main entry point for all HTTP requests to the Go backend services.

## Quick Start

```bash
# Run locally
cd apps/api-gateway && go run cmd/main.go

# Build
cd apps/api-gateway && go build -o bin/api-gateway ./cmd/main.go

# Run tests
cd apps/api-gateway && go test ./...
```

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/workers` | List all workers |
| GET | `/api/v1/workers/:id` | Get worker by ID |
| GET | `/api/v1/go-applications` | List all Go applications |
| GET | `/api/v1/shared-packages` | List all shared packages |

## Configuration

Environment variables:
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `PORT`: HTTP server port (default: 8080)

## Project Structure

```
apps/api-gateway/
├── cmd/
│   └── main.go           # Application entry point
├── internal/
│   ├── handlers/         # HTTP request handlers
│   ├── models/           # Data models
│   └── services/         # Business logic
├── deploy/
│   ├── docker/           # Docker configuration
│   └── k8s/              # Kubernetes manifests
└── tests/
    └── unit/             # Unit tests
```
