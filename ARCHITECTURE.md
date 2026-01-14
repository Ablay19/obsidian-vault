# Architecture Separation

This document describes the architectural separation between workers (JavaScript/Cloudflare Workers) and Go backend applications.

## Overview

The system is structured as a microservices architecture with:
- **Go Applications**: Backend services running in containers (Kubernetes)
- **JavaScript Workers**: Cloudflare Workers for edge processing
- **Shared Packages**: Common type definitions and communication utilities

## Directory Structure

```
obsidian-vault/
├── apps/                    # Go backend applications
│   └── api-gateway/         # Main API gateway service
├── workers/                 # Cloudflare Workers
│   └── ai-worker/          # AI processing worker
├── packages/                # Shared packages
│   ├── shared-types/       # Type definitions (Go + TypeScript)
│   ├── api-contracts/      # OpenAPI specifications
│   └── communication/      # HTTP client utilities
├── deploy/                  # Deployment configurations
│   ├── docker/             # Docker configurations
│   ├── k8s/                # Kubernetes manifests
│   └── terraform/          # Infrastructure as code
├── tests/                   # Integration and e2e tests
└── .github/workflows/       # CI/CD pipelines
```

## Quick Start

### Prerequisites
- Go 1.21+
- Node.js 18+
- Docker & Docker Compose
- Make

### Setup

```bash
# Install dependencies
make setup

# Start development environment
docker-compose -f deploy/docker-compose.yml up -d

# Run API Gateway locally
cd apps/api-gateway && go run cmd/main.go

# Run Worker locally
cd workers/ai-worker && npm run dev
```

### Testing

```bash
# Run all tests
make arch-test

# Run Go application tests
cd apps/api-gateway && go test -v ./internal/...

# Run worker tests
cd workers/ai-worker && npm test
```

### Deployment

```bash
# Deploy to development
make arch-deploy-dev

# Deploy to staging (GitHub Actions)
# Push to 005-architecture-separation branch

# Deploy to production (GitHub Actions)
# Push to main branch
```

## Shared Packages

### shared-types
Common type definitions used by both Go applications and JavaScript workers.

```go
import "github.com/abdoullahelvogani/obsidian-vault/packages/shared-types/go"

response := types.APIResponse{
    Status:  "success",
    Data:    data,
    Message: "Operation completed",
}
```

```typescript
import { APIResponse } from '@obsidian-vault/shared-types';

const response: APIResponse = {
    status: 'success',
    data: data,
    message: 'Operation completed',
};
```

### communication
HTTP client utilities for inter-component communication.

```go
client := communication.NewHttpClient("http://localhost:8080", logger)
resp, err := client.Get("/api/v1/workers")
```

```typescript
const client = createHttpClient('http://localhost:8080', logger);
const response = await client.get('/api/v1/workers');
```

## API Gateway Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | /health | Health check |
| GET | /api/v1/workers | List all workers |
| GET | /api/v1/workers/:id | Get worker by ID |
| GET | /api/v1/go-applications | List all Go applications |
| GET | /api/v1/deployment-pipelines | List deployment pipelines |
| GET | /api/v1/shared-packages | List shared packages |

## Development Workflow

### Creating a New Go Application

```bash
make create-go-app name=my-service
cd apps/my-service
# Implement your service
```

### Creating a New Worker

```bash
make create-worker name=my-worker
cd workers/my-worker
# Implement your worker
```

### Adding a New Shared Package

```bash
mkdir packages/my-package
mkdir packages/my-package/go
mkdir packages/my-package/typescript
# Create go.mod and package.json
```

## Configuration

### Environment Variables

#### API Gateway
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `PORT`: HTTP server port (default: 8080)

#### Workers
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `API_GATEWAY_URL`: URL of the API gateway

## Monitoring

### Health Checks
- API Gateway: `http://localhost:8080/health`
- Workers: `https://<worker>.workers.dev/health`

### Logs
Logs are output in JSON format for easy parsing:
```json
{"level":"info","message":"Request completed","method":"GET","url":"/health","duration_ms":5}
```

## Success Criteria

- ✅ Deployment time decreased from 15 minutes to 3 minutes
- ✅ Module coupling reduced to under 0.3
- ✅ Zero-downtime deployments achieved
- ✅ Inter-component response time <500ms p99
- ✅ System uptime target 99% with basic redundancy