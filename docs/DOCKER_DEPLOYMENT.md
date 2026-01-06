# Docker Deployment Guide

This guide provides comprehensive instructions for deploying the Obsidian Bot using Docker containers.

## 🚀 Quick Start

### Prerequisites
- Docker Engine 20.10+
- Docker Compose v2.0+
- Git

### Basic Setup

1. **Clone and Setup**
   ```bash
   git clone https://github.com/Ablay19/obsidian-vault.git
   cd obsidian-vault
   ```

2. **Configure Environment**
   ```bash
   cp .env.docker .env
   # Edit .env with your configuration
   nano .env
   ```

3. **Deploy**
   ```bash
   ./docker-deploy.sh production
   ```

## 📋 Configuration

### Environment Variables

Create `.env` file with the following variables:

```bash
# Application Configuration
APP_VERSION=v1.0.0
BOT_PORT=8080

# Database (Required)
TURSO_DATABASE_URL=libsql://your-db.turso.io
TURSO_AUTH_TOKEN=your-turso-token

# Telegram (Required)
TELEGRAM_BOT_TOKEN=your-telegram-bot-token

# Security (Required)
SESSION_SECRET=your-secret-key

# AI Providers (Optional)
CLOUDFLARE_WORKER_URL=https://your-worker.workers.dev
GEMINI_API_KEY=your-gemini-api-key

# Google Cloud Logging (Optional)
GOOGLE_CLOUD_PROJECT=your-gcp-project
GOOGLE_APPLICATION_CREDENTIALS_PATH=/path/to/credentials.json
ENABLE_GOOGLE_LOGGING=true
```

### Service Configuration

#### Main Services (Always Running)
- **obsidian-bot**: Main application (port 8080)
- **obsidian-redis**: Redis cache (port 6379)

#### Optional Services (Profiles)
- **ssh-server**: SSH access (port 2222, API port 8081)
- **vault**: HashiCorp Vault (port 8200)
- **nginx**: Reverse proxy (ports 80, 443)

Enable optional services:
```bash
export COMPOSE_PROFILES=ssh,vault
./docker-deploy.sh start
```

## 🛠️ Deployment Modes

### Production Deployment
```bash
./docker-deploy.sh production
```

**Features:**
- Optimized Docker images
- Health checks enabled
- Resource limits configured
- Google Cloud logging integration
- Persistent volumes
- Automatic restarts

### Development Setup
```bash
./docker-deploy.sh development
```

**Features:**
- Development configuration
- Debug logging enabled
- All services started
- Status display

### Manual Service Control

```bash
# Start services
./docker-deploy.sh start

# Stop services
./docker-deploy.sh stop

# Check status
./docker-deploy.sh status

# Health check
./docker-deploy.sh health

# Rebuild images
./docker-deploy.sh build --no-cache

# Cleanup resources
./docker-deploy.sh cleanup
```

## 📊 Service Management

### Monitoring

```bash
# View all services
docker-compose ps

# View logs
docker-compose logs -f bot

# View resource usage
docker stats

# Health check
curl http://localhost:8080/api/services/status
```

### Scaling

```bash
# Scale bot service
docker-compose up -d --scale bot=2

# With resource limits
docker-compose up -d --scale bot=2 --scale redis=1
```

### Updates

```bash
# Pull latest code
git pull origin main

# Rebuild and restart
./docker-deploy.sh build --no-cache
./docker-deploy.sh start
```

## 🔧 Advanced Configuration

### Custom Dockerfile

For custom builds, use `Dockerfile.production`:

```bash
docker build \
  -f Dockerfile.production \
  --build-arg APP_VERSION=v1.0.0 \
  --build-arg GO_VERSION=1.25.4 \
  -t obsidian-bot:custom \
  .
```

### Multi-Architecture Builds

```bash
# Build for multiple platforms
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -f Dockerfile.production \
  -t obsidian-bot:multiarch \
  --push .
```

### Custom Networks

```yaml
networks:
  obsidian-network:
    driver: bridge
    ipam:
      config:
        - subnet: 172.20.0.0/16
```

## 🔒 Security Configuration

### Production Security

1. **Non-root User**: Containers run as `appuser` (UID 1000)
2. **Read-only Filesystem**: Optional for extra security
3. **Resource Limits**: Memory and CPU limits configured
4. **Health Checks**: Automated health monitoring

### SSL/TLS with Nginx

```bash
# Enable with profile
export COMPOSE_PROFILES=nginx
./docker-deploy.sh start

# Configure SSL certificates
mkdir -p nginx/ssl
# Place certificates in nginx/ssl/
```

### Secrets Management

Using HashiCorp Vault:

```bash
# Enable Vault profile
export COMPOSE_PROFILES=vault
./docker-deploy.sh start

# Access Vault UI
http://localhost:8200
# Token: root (development)
```

## 📈 Performance Optimization

### Resource Tuning

```bash
# Resource limits in .env
BOT_MEMORY_LIMIT=512M
BOT_CPU_LIMIT=0.5
REDIS_MEMORY_LIMIT=256M
REDIS_MAX_MEMORY=256mb
```

### Caching Configuration

```bash
# Redis optimization
docker-compose exec redis redis-cli CONFIG SET maxmemory 512mb
docker-compose exec redis redis-cli CONFIG SET maxmemory-policy allkeys-lru
```

### Database Optimization

```bash
# Connection pooling in .env
DATABASE_POOL_SIZE=10
DATABASE_MAX_CONNECTIONS=20
```

## 🔍 Troubleshooting

### Common Issues

#### Container Won't Start
```bash
# Check logs
docker-compose logs bot

# Check configuration
docker-compose config

# Rebuild without cache
./docker-deploy.sh build --no-cache
```

#### Health Check Failures
```bash
# Manual health check
curl http://localhost:8080/api/services/status

# Container health
docker inspect obsidian-bot | grep Health -A 10
```

#### Performance Issues
```bash
# Resource usage
docker stats

# Container inspection
docker inspect obsidian-bot
```

### Debug Mode

```bash
# Enable debug logging
echo "LOG_LEVEL=DEBUG" >> .env
./docker-deploy.sh restart

# Access container shell
docker-compose exec bot /bin/sh
```

### Logs Collection

```bash
# Collect all logs
docker-compose logs > obsidian-bot-logs-$(date +%Y%m%d).log

# Specific service logs
docker-compose logs --tail=100 bot > bot-logs.log
docker-compose logs --tail=100 redis > redis-logs.log
```

## 🔄 Backup and Recovery

### Data Backup

```bash
# Backup volumes
docker run --rm -v obsidian_data:/data -v $(pwd):/backup alpine tar czf /backup/obsidian-data-$(date +%Y%m%d).tar.gz -C /data .

# Database backup
docker-compose exec bot ./scripts/backup-database.sh
```

### Recovery

```bash
# Restore volumes
docker run --rm -v obsidian_data:/data -v $(pwd):/backup alpine tar xzf /backup/obsidian-data-YYYYMMDD.tar.gz -C /data

# Restart services
./docker-deploy.sh restart
```

## 📚 Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Compose Documentation](https://docs.docker.com/compose/)
- [Obsidian Bot README](README-SETUP.md)
- [Troubleshooting Guide](docs/TROUBLESHOOTING.md)

## 🆘 Support

For Docker deployment issues:

1. Check the troubleshooting section above
2. Review container logs: `docker-compose logs`
3. Check system resources: `docker stats`
4. Verify environment configuration
5. Create an issue in the GitHub repository

---

**Note**: This Docker setup is production-ready and includes security best practices, health monitoring, and resource management for reliable deployment.