# Deployment Guide

This guide covers deployment procedures for the architecture separation components.

## Quick Deployment

### API Gateway (Go)

```bash
# Build Docker image
docker build -f deploy/docker/api-gateway.Dockerfile -t api-gateway .

# Run with Docker Compose
docker-compose -f deploy/docker-compose.yml up -d api-gateway

# Deploy to Kubernetes
kubectl apply -f deploy/k8s/go-services/api-gateway.yaml
```

### AI Worker (Cloudflare Workers)

```bash
# Deploy to staging
cd workers/ai-worker && wrangler deploy --env staging

# Deploy to production
cd workers/ai-worker && wrangler deploy --env production
```

## CI/CD Pipeline

### GitHub Actions

The project uses GitHub Actions for CI/CD:

- **API Gateway**: `.github/workflows/api-gateway-ci.yml`
- **AI Worker**: `.github/workflows/ai-worker-ci.yml`

### Deployment Triggers

| Branch | Environment | Trigger |
|--------|-------------|---------|
| `005-architecture-separation` | Staging | Push |
| `main` | Production | Push |

## Environment Configuration

### API Gateway

| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_LEVEL` | Logging level | `info` |
| `PORT` | HTTP port | `8080` |

### AI Worker

| Variable | Description | Default |
|----------|-------------|---------|
| `LOG_LEVEL` | Logging level | `info` |
| `API_GATEWAY_URL` | API Gateway URL | `http://localhost:8080` |

## Health Checks

- **API Gateway**: `http://localhost:8080/health`
- **Worker**: `https://<worker>.workers.dev/health`
