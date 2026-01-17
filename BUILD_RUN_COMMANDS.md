# Build & Run Commands Guide

This document provides comprehensive commands for building and running all components.

## Quick Reference

| Command | Description |
|---------|-------------|
| `make build-all` | Build everything (Go apps + Workers + Docker) |
| `make run-all` | Run all services in background |
| `make run-all-stop` | Stop all running services |

---

## Build Commands

### Build Everything
```bash
make build-all
```
**Builds:**
- ✅ All Go applications (`apps/*/`)
- ✅ All Go binaries (`cmd/*/`)
- ✅ All Cloudflare Workers (`workers/*/`)
- ✅ Docker image for bot (`.config/local/docker/Dockerfile`)

### Build Only (No Docker)
```bash
make build-all-no-docker
```
**Builds:**
- ✅ All Go applications
- ✅ All Workers
- ❌ No Docker images

### Build Separately
```bash
make build-go-apps    # Build all Go applications
make build-workers    # Build all workers
make build            # Build Go binary only
make build-prod       # Build production binary
```

### Build Docker Images Only
```bash
make docker-build     # Build bot Docker image
make build-ssh        # Build SSH server image
make docker-build-all # Build all Docker images
```

---

## Run Commands

### Run Everything
```bash
make run-all
```
**Starts:**
- 🌐 Go Bot API on `http://localhost:8080`
- ⚡ AI Worker on `http://localhost:8787`
- 🎬 AI Manim Worker on `http://localhost:8788`
- 🖼️ Image Worker on `http://localhost:8789`

### Run in Separate Terminals

**Terminal 1 - Go Bot:**
```bash
make dev
# or
go run ./cmd/bot/main.go
```

**Terminal 2 - AI Worker:**
```bash
cd workers/ai-worker && npm run dev
```

**Terminal 3 - AI Manim Worker:**
```bash
cd workers/ai-manim-worker && npm run dev
```

**Terminal 4 - Image Worker:**
```bash
cd workers/image-worker && npm run dev
```

### Run with Docker
```bash
make docker-up       # Start all Docker services
make docker-down     # Stop Docker services
make docker-logs     # View logs
make docker-status   # Show status
```

---

## Stop Commands

```bash
make run-all-stop    # Stop all services (Go + Workers)
make down            # Stop bot container
make docker-down     # Stop all Docker services
```

---

## Status Commands

```bash
make status-all      # Show all service statuses
make health-all      # Perform all health checks
make docker-status   # Show Docker status
```

---

## Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                        make build-all                            │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────┐    ┌──────────────────┐                   │
│  │   Go Apps        │    │   Workers        │                   │
│  │   (apps/*/)      │    │   (workers/*/)   │                   │
│  └────────┬─────────┘    └────────┬─────────┘                   │
│           │                       │                              │
│           ▼                       ▼                              │
│  ┌──────────────────┐    ┌──────────────────┐                   │
│  │   Go Binaries    │    │   Wrangler       │                   │
│  │   (build/)       │    │   Deploy         │                   │
│  └────────┬─────────┘    └────────┬─────────┘                   │
│           │                       │                              │
│           └───────────┬───────────┘                              │
│                       ▼                                          │
│            ┌─────────────────────┐                               │
│            │   Docker Images     │                               │
│            │   (obsidian-bot)    │                               │
│            └─────────────────────┘                               │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                        make run-all                              │
├─────────────────────────────────────────────────────────────────┤
│                                                                 │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │  Go Bot (cmd/bot/main.go)                                │   │
│  │  • HTTP API: http://localhost:8080                       │   │
│  │  • Telegram Bot                                          │   │
│  │  • Web Dashboard                                         │   │
│  └──────────────────────────────────────────────────────────┘   │
│                              │                                   │
│         ┌────────────────────┼────────────────────┐             │
│         │                    │                    │             │
│         ▼                    ▼                    ▼             │
│  ┌─────────────┐      ┌─────────────┐      ┌─────────────┐    │
│  │ AI Worker   │      │Manim Worker │      │Image Worker │    │
│  │ :8787       │      │ :8788       │      │ :8789       │    │
│  └─────────────┘      └─────────────┘      └─────────────┘    │
│                                                                 │
└─────────────────────────────────────────────────────────────────┘
```

---

## Service Ports

| Service | Port | URL | Command |
|---------|------|-----|---------|
| **Go Bot API** | 8080 | http://localhost:8080 | `make dev` or `go run ./cmd/bot/main.go` |
| **AI Worker** | 8787 | http://localhost:8787 | `cd workers/ai-worker && npm run dev` |
| **AI Manim Worker** | 8788 | http://localhost:8788 | `cd workers/ai-manim-worker && npm run dev` |
| **Image Worker** | 8789 | http://localhost:8789 | `cd workers/image-worker && npm run dev` |
| **Redis** | 6379 | localhost:6379 | `docker-compose up -d redis` |
| **Vault** | 8200 | http://localhost:8200 | `docker-compose --profile vault up -d` |
| **RabbitMQ** | 5672 | localhost:5672 | `docker-compose up -d rabbitmq` |

---

## Docker Compose Services

```bash
# Start all services
docker-compose up -d

# Start specific services
docker-compose up -d bot redis

# Start with profiles
docker-compose --profile ssh up -d bot ssh
docker-compose --profile vault up -d bot vault

# View logs
docker-compose logs -f bot
docker-compose logs -f redis

# Stop all
docker-compose down
```

---

## Health Endpoints

| Service | Health Check |
|---------|-------------|
| Go Bot | `curl http://localhost:8080/health` |
| AI Worker | `curl http://localhost:8787/health` |
| Redis | `docker exec obsidian-redis redis-cli ping` |

---

## Examples

### Full Development Setup
```bash
# 1. Build everything
make build-all

# 2. Start all services
make run-all

# 3. Check status
make status-all

# 4. Check health
make health-all

# 5. When done, stop everything
make run-all-stop
```

### Docker-Only Setup
```bash
# 1. Build and start Docker services
make docker-up

# 2. Check status
make docker-status

# 3. View logs
make docker-logs

# 4. Stop when done
make docker-down
```

### Mixed Setup (Go + Workers + Docker)
```bash
# 1. Start Docker services (Redis, etc.)
docker-compose up -d redis

# 2. Run Go bot
make dev

# 3. Run workers in separate terminals
cd workers/ai-worker && npm run dev
cd workers/ai-manim-worker && npm run dev
```
