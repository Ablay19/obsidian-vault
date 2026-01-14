# Architecture Documentation

## Overview

This document describes the architecture of the obsidian-vault project, which separates workers (JavaScript/TypeScript) from Go backend applications.

## Architecture Diagram

```
                    ┌─────────────────────────────────────────┐
                    │         Internet / Users                │
                    └────────────────┬────────────────────────┘
                                     │
                    ┌────────────────▼────────────────────────┐
                    │           API Gateway                   │
                    │         (apps/api-gateway)              │
                    │    Port 8080 · Go 1.21+                 │
                    └────────┬────────────────┬───────────────┘
                             │                │
              ┌──────────────┴───────────────┴──────────────┐
              │                                              │
     ┌────────▼────────┐                          ┌────────▼────────┐
     │  Auth Service   │                          │  User Service   │
     │ (apps/auth-svc) │                          │ (apps/user-svc) │
     │  Port 8081      │                          │  Port 8082      │
     └─────────────────┘                          └─────────────────┘
              │                                              │
              └──────────────┬───────────────────────────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
  ┌──────▼──────┐    ┌──────▼──────┐    ┌──────▼──────┐
  │  AI Worker  │    │Image Worker │    │ Other Workers│
  │(workers/ai-)│    │ (workers/   │    │  (workers/)  │
  │   worker)   │    │ image-wrk)  │    │              │
  └─────────────┘    └─────────────┘    └─────────────┘
         │                   │                   │
         └───────────────────┼───────────────────┘
                             │
                    ┌────────▼────────┐
                    │  External APIs  │
                    │  (Internet)     │
                    └─────────────────┘
```

## Component Description

### Go Applications (`apps/`)

| Service | Port | Responsibility |
|---------|------|----------------|
| api-gateway | 8080 | Request routing, load balancing |
| auth-service | 8081 | Authentication, authorization |
| user-service | 8082 | User management, profiles |

### Cloudflare Workers (`workers/`)

| Worker | Type | Responsibility |
|--------|------|----------------|
| ai-worker | AI | AI processing, natural language |
| image-worker | Vision | Image processing, OCR |

### Shared Packages (`packages/`)

| Package | Purpose |
|---------|---------|
| shared-types | Common type definitions |
| api-contracts | OpenAPI specifications |
| communication | HTTP client utilities |

## Design Principles

1. **Independent Deployment** - Each component deploys independently
2. **Clear Boundaries** - Separate tech stacks, no cross-dependencies
3. **Eventual Consistency** - API-based synchronization between components
4. **Fail-Fast** - No retries on inter-component API failures
5. **Internal Network** - No external authentication between services

## Communication Patterns

- **REST/JSON** - Primary communication protocol
- **Eventual Consistency** - For cross-service data
- **No Retries** - Fail-fast error handling
- **Internal Network Only** - Network isolation

## Deployment

### Local Development
```bash
# Start all services
./scripts/dev.sh

# Or use Docker Compose
docker-compose -f .docker-compose.dev.yml up
```

### Build All
```bash
./scripts/build-all.sh
```

### Test All
```bash
./scripts/test-all.sh
```

## Performance Targets

- **Deployment Time**: < 3 minutes per component
- **Inter-component Response**: < 500ms (p99)
- **Build Time**: 50% improvement through modular compilation
- **Coupling Ratio**: < 0.3 (dependency analysis)

## Security

- Internal network communication only
- Network isolation via Kubernetes NetworkPolicies
- Fail-fast error handling
- No authentication between internal services
