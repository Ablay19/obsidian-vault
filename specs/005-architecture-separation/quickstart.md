# Quickstart Guide: Architectural Separation

**Feature**: Architectural Separation for Workers and Go Applications  
**Date**: January 15, 2025  
**Version**: 1.0.0

## Overview

This guide helps you set up the development environment for the new separated architecture with independent workers and Go applications.

## Prerequisites

### Required Tools
- **Go 1.21+** - For backend services
- **Node.js 18+** - For workers development
- **Docker** - For local testing and deployment
- **Git** - For version control
- **Make** - For build automation

### Optional Tools
- **Cloudflare Wrangler** - For workers deployment
- **kubectl** - For Kubernetes management
- **Terraform** - For infrastructure management

## Quick Setup (5 minutes)

### 1. Clone and Setup Repository

```bash
# Clone the repository
git clone <repository-url>
cd obsidian-vault

# Switch to the architecture separation branch
git checkout 005-architecture-separation

# Install development dependencies
make setup
```

### 2. Start Development Environment

```bash
# Start all services in development mode
make dev

# Verify everything is running
make health-check
```

### 3. Create Your First Components

```bash
# Create a new worker
make create-worker name=my-worker

# Create a new Go application
make create-go-app name=my-service

# Deploy to development environment
make deploy-dev
```

## Detailed Setup

### Go Applications Setup

#### 1. Create a New Go Application

```bash
# Navigate to apps directory
cd apps

# Create new application directory
mkdir my-new-service
cd my-new-service

# Initialize Go module
go mod init github.com/your-org/obsidian-vault/apps/my-new-service

# Create basic structure
mkdir -p cmd internal/{handlers,services,models} tests
```

#### 2. Application Template

Create `cmd/main.go`:
```go
package main

import (
    "log"
    "net/http"
    "github.com/your-org/obsidian-vault/apps/my-new-service/internal/handlers"
)

func main() {
    mux := http.NewServeMux()
    
    // Register handlers
    mux.HandleFunc("/health", handlers.Health)
    mux.HandleFunc("/api/v1/", handlers.API)
    
    log.Println("Starting my-new-service on :8080")
    if err := http.ListenAndServe(":8080", mux); err != nil {
        log.Fatal(err)
    }
}
```

#### 3. Build and Run

```bash
# Build the application
go build -o bin/my-new-service ./cmd

# Run locally
./bin/my-new-service

# Or use make
make build-go-app APP=my-new-service
make run-go-app APP=my-new-service
```

### Workers Setup

#### 1. Create a New Worker

```bash
# Navigate to workers directory
cd workers

# Create new worker directory
mkdir my-new-worker
cd my-new-worker

# Initialize npm package
npm init -y

# Install dependencies
npm install --save-dev wrangler
npm install @cloudflare/workers-types
```

#### 2. Worker Template

Create `src/index.ts`:
```typescript
import { WorkerRequest, WorkerResponse } from '../types';

export default {
    async fetch(request: Request, env: Env, ctx: ExecutionContext): Promise<Response> {
        const url = new URL(request.url);
        
        // Route handling
        if (url.pathname === '/health') {
            return new Response(JSON.stringify({ status: 'ok' }), {
                headers: { 'Content-Type': 'application/json' }
            });
        }
        
        if (url.pathname.startsWith('/api/')) {
            return handleAPI(request, env);
        }
        
        return new Response('Not Found', { status: 404 });
    }
};

async function handleAPI(request: Request, env: Env): Promise<Response> {
    // API logic here
    return new Response(JSON.stringify({ message: 'Hello from worker' }), {
        headers: { 'Content-Type': 'application/json' }
    });
}
```

#### 3. Wrangler Configuration

Create `wrangler.toml`:
```toml
name = "my-new-worker"
main = "src/index.ts"
compatibility_date = "2024-01-01"

[env.development]
name = "my-new-worker-dev"

[env.staging]
name = "my-new-worker-staging"

[env.production]
name = "my-new-worker"
```

#### 4. Deploy Worker

```bash
# Deploy to development
wrangler dev

# Deploy to staging
wrangler deploy --env staging

# Deploy to production
wrangler deploy --env production
```

### Shared Packages Setup

#### 1. Create Shared Types

```bash
# Navigate to packages directory
cd packages

# Create shared types package
mkdir shared-types
cd shared-types

# Go types
mkdir go
echo 'package types

type APIResponse struct {
    Status  string      `json:"status"`
    Data    interface{} `json:"data"`
    Message string      `json:"message"`
}' > go/types.go

# TypeScript types
mkdir typescript
npm init -y
echo 'export interface APIResponse {
    status: string;
    data: any;
    message: string;
}' > typescript/index.ts
```

#### 2. Use Shared Types

In Go application:
```go
import "github.com/your-org/obsidian-vault/packages/shared-types/go"

func handleAPI(w http.ResponseWriter, r *http.Request) {
    response := types.APIResponse{
        Status:  "success",
        Data:    nil,
        Message: "API working",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(response)
}
```

In Worker:
```typescript
import { APIResponse } from '@your-org/shared-types';

export default {
    async fetch(): Promise<Response> {
        const response: APIResponse = {
            status: 'success',
            data: null,
            message: 'Worker working'
        };
        
        return new Response(JSON.stringify(response), {
            headers: { 'Content-Type': 'application/json' }
        });
    }
};
```

## Development Workflow

### 1. Local Development

```bash
# Start all services
make dev

# View logs
make logs

# Run tests
make test

# Run integration tests
make test-integration
```

### 2. Making Changes

```bash
# Make changes to your code
# ...

# Run tests
make test-unit COMPONENT=my-worker
make test-unit COMPONENT=my-service

# Build and test locally
make build COMPONENT=my-worker
make build COMPONENT=my-service

# Deploy to development
make deploy-dev COMPONENT=my-worker
make deploy-dev COMPONENT=my-service
```

### 3. Testing

```bash
# Unit tests
make test-unit

# Integration tests
make test-integration

# End-to-end tests
make test-e2e

# Contract tests
make test-contract
```

### 4. Deployment

```bash
# Deploy to staging
make deploy-staging

# Deploy to production
make deploy-production

# Rollback
make rollback COMPONENT=my-worker VERSION=previous
```

## Common Tasks

### Adding a New Shared Package

```bash
# Create package directory
mkdir packages/my-new-package
cd packages/my-new-package

# Initialize Go module (if needed)
mkdir go && cd go
go mod init github.com/your-org/obsidian-vault/packages/my-new-package/go

# Initialize npm package (if needed)
cd ../
mkdir typescript && cd typescript
npm init -y

# Add to workspace configuration
# Update packages/package.json workspaces
```

### Setting Up API Communication

```bash
# Generate API client from contracts
make generate-api-client

# Update shared types
make update-types

# Validate API contracts
make validate-contracts
```

### Database Setup

```bash
# Start local database
make start-db

# Run migrations
make migrate DB=my-service

# Seed database
make seed DB=my-service
```

## Troubleshooting

### Common Issues

1. **Port Conflicts**
   ```bash
   # Check what's using ports
   make check-ports
   
   # Kill processes
   make cleanup
   ```

2. **Build Failures**
   ```bash
   # Clean build cache
   make clean
   
   # Rebuild
   make build-all
   ```

3. **Test Failures**
   ```bash
   # Run tests with verbose output
   make test-verbose
   
   # Run specific test
   make test COMPONENT=my-worker TEST=api
   ```

### Getting Help

- Check the logs: `make logs`
- Run health check: `make health-check`
- View documentation: `make docs`
- Open an issue: Create GitHub issue with logs and configuration

## Make Commands Reference

### Setup Commands
- `make setup` - Initialize development environment
- `make clean` - Clean build artifacts
- `make deps` - Install dependencies

### Development Commands
- `make dev` - Start development environment
- `make build` - Build all components
- `make test` - Run all tests
- `make logs` - View service logs

### Component Commands
- `make create-worker name=<name>` - Create new worker
- `make create-go-app name=<name>` - Create new Go application
- `make deploy-dev` - Deploy to development
- `make deploy-staging` - Deploy to staging
- `make deploy-production` - Deploy to production

### Testing Commands
- `make test-unit` - Run unit tests
- `make test-integration` - Run integration tests
- `make test-e2e` - Run end-to-end tests
- `make test-contract` - Run contract tests

### Utility Commands
- `make health-check` - Check system health
- `make check-ports` - Check port usage
- `make generate-api-client` - Generate API clients
- `make validate-contracts` - Validate API contracts

---

**Ready to start developing with the new separated architecture! 🚀**