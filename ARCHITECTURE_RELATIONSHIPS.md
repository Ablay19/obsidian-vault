# Architecture Relationships & Docker Configuration

This document describes the relationships between `cmd/`, `apps/`, `workers/`, and Docker configuration.

## Directory Structure Overview

```
obsidian-vault/
├── cmd/                          # Go application entry points (compiled binaries)
│   ├── bot/main.go              # Main Obsidian Bot (HTTP API + Telegram bot)
│   ├── cli/root.go              # CLI with TUI, email, SSH management
│   ├── whatsapp-cli/            # WhatsApp integration
│   ├── ssh-server/              # SSH daemon
│   └── ... (utilities)
│
├── apps/                         # Go microservices (future state)
│   ├── api-gateway/             # API gateway service
│   ├── auth-service/            # Authentication microservice
│   └── user-service/            # User management microservice
│
├── workers/                      # Cloudflare Workers (edge functions)
│   ├── ai-worker/               # AI processing (TypeScript)
│   ├── ai-manim-worker/         # Manim video generation (TypeScript)
│   ├── image-worker/            # Image processing (TypeScript)
│   ├── ai-proxy/                # AI provider proxy (JavaScript)
│   └── manim-renderer/          # Manim rendering service
│
├── internal/                     # Shared Go packages (imported by cmd/ and apps/)
│   ├── ai/                      # AI service integrations
│   ├── bot/                     # Telegram bot
│   ├── auth/                    # Authentication
│   ├── config/                  # Configuration
│   ├── dashboard/               # Web dashboard
│   ├── database/                # Database operations
│   ├── middleware/              # HTTP middleware
│   ├── rag/                     # RAG implementation
│   ├── ssh/                     # SSH management
│   ├── state/                   # Runtime state
│   ├── telemetry/               # Metrics
│   ├── vectorstore/             # Vector database
│   ├── vision/                  # Computer vision (OCR)
│   └── ... (20+ packages)
│
├── packages/                     # Cross-language shared packages
│   ├── shared-types/            # Type definitions (Go + TypeScript)
│   ├── api-contracts/           # OpenAPI specs
│   └── communication/           # HTTP clients
│
├── .config/local/docker/
│   ├── Dockerfile               # Go application Dockerfile
│   ├── Dockerfile.ssh           # SSH server Dockerfile
│   └── Dockerfile.production    # Production variant
│
├── Dockerfile                   # ⚠️ Cloudflare Workers Dockerfile (Node.js)
│
├── docker-compose.yml           # Docker orchestration (Go services)
│
└── wrangler.toml                # Workers configuration
```

## Dockerfile Relationships

### Go Application Dockerfile
- **Location**: `.config/local/docker/Dockerfile`
- **Purpose**: Builds Go applications (`cmd/bot/main.go`)
- **Base Image**: `golang:1.25.4-alpine`
- **Build Target**: `./cmd/bot/main.go`
- **Used by**:
  - `docker-compose.yml` (bot service)
  - `Makefile` (build, up targets)
  - `Makefile` (docker-build-all target)

### Cloudflare Workers Dockerfile
- **Location**: `Dockerfile` (root)
- **Purpose**: Builds Cloudflare Workers deployment
- **Base Image**: `node:18-alpine`
- **Build Target**: `npm run deploy` (wrangler)
- **Used by**:
  - Workers deployment pipeline
  - NOT for Go services

### SSH Server Dockerfile
- **Location**: `.config/local/docker/Dockerfile.ssh`
- **Purpose**: Builds SSH server container
- **Used by**:
  - `docker-compose.yml` (ssh service)
  - `Makefile` (build-ssh target)

## Service Relationships

```
┌─────────────────────────────────────────────────────────────────────┐
│                     docker-compose.yml                               │
│  Orchestrates: bot, redis, vault, ssh, rabbitmq                      │
└─────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│  cmd/bot/main.go                                                     │
│  - HTTP API Server (port 8080)                                       │
│  - Telegram Bot Integration                                          │
│  - Imports: internal/* packages                                      │
└─────────────────────────────────────────────────────────────────────┘
                                   │
         ┌─────────────────────────┼─────────────────────────┐
         │                         │                         │
         ▼                         ▼                         ▼
┌─────────────────┐   ┌─────────────────────┐   ┌─────────────────────┐
│   internal/     │   │   apps/            │   │   workers/          │
│   Shared Go     │   │   Go Microservices │   │   Cloudflare Workers│
│   Packages      │   │   (future state)   │   │   (edge functions)  │
└─────────────────┘   └─────────────────────┘   └─────────────────────┘
         │                         │                         │
         └─────────────────────────┼─────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│  packages/shared-types                                               │
│  - Common Type Definitions (Go + TypeScript)                         │
│  - Enables type-safe communication                                   │
└─────────────────────────────────────────────────────────────────────┘
```

## Docker Build Commands

### Build Go Application
```bash
# Using docker-compose
docker-compose build bot

# Using Makefile
make build  # Docker image
make build-go-apps  # Binary
```

### Build Workers
```bash
# Using wrangler
cd workers/ai-worker && npm run deploy

# Using Makefile
make build-workers
```

### Build SSH Server
```bash
# Using docker-compose
docker-compose --profile ssh build ssh

# Using Makefile
make build-ssh
```

## Communication Flow

```
┌─────────────┐         ┌─────────────┐         ┌─────────────┐
│ Telegram    │────────▶│  cmd/bot    │────────▶│  Workers    │
│ User        │         │  (Go)       │         │  (Edge)     │
└─────────────┘         └──────┬──────┘         └─────────────┘
                               │
                               ▼
                        ┌─────────────┐
                        │  internal/  │
                        │  packages   │
                        └──────┬──────┘
                               │
                               ▼
                        ┌─────────────┐
                        │  packages/  │
                        │  shared-    │
                        │  types      │
                        └─────────────┘
```

## Key Configuration Files

| File | Purpose | Language |
|------|---------|----------|
| `.config/local/docker/Dockerfile` | Go application container | Go |
| `.config/local/docker/Dockerfile.ssh` | SSH server container | Go |
| `Dockerfile` | Workers deployment container | Node.js |
| `wrangler.toml` | Workers configuration | TOML |
| `docker-compose.yml` | Service orchestration | YAML |
| `go.mod` | Go dependencies | Go |
| `workers/package.json` | Workers dependencies | Node.js |

## Common Issues & Fixes

### Issue: Wrong Dockerfile Referenced
**Problem**: `docker-compose.yml` referenced root `Dockerfile` (for Workers)
**Fix**: Updated to use `.config/local/docker/Dockerfile`

```yaml
# Before (incorrect)
bot:
  build:
    dockerfile: Dockerfile  # Node.js - wrong!

# After (correct)
bot:
  build:
    dockerfile: .config/local/docker/Dockerfile  # Go - correct!
```

### Issue: Makefile Docker Variable
**Problem**: `DOCKERFILE` variable pointed to wrong file
**Fix**: Updated to correct path

```makefile
# Before
DOCKERFILE ?= Dockerfile

# After
DOCKERFILE ?= .config/local/docker/Dockerfile
```

## Deployment Pipeline

```
┌─────────────────────────────────────────────────────────────────────┐
│                         Development                                  │
│  • make dev          - Run Go bot locally                           │
│  • cd workers/ai-worker && npm run dev - Run worker locally         │
└─────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Docker Build                                 │
│  • docker-compose build bot      - Build Go container               │
│  • docker-compose build ssh      - Build SSH container              │
└─────────────────────────────────────────────────────────────────────┘
                                   │
                                   ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         Deployment                                   │
│  • docker-compose up -d          - Start Go services                │
│  • cd workers/ai-worker && npm run deploy - Deploy worker           │
└─────────────────────────────────────────────────────────────────────┘
```

## Summary

- **cmd/**: Go entry points (bot, CLI, utilities)
- **apps/**: Go microservices (future architecture)
- **workers/**: Cloudflare Workers (edge functions)
- **internal/**: Shared Go packages
- **packages/**: Cross-language type definitions
- **Docker**: Go apps use `.config/local/docker/Dockerfile`
