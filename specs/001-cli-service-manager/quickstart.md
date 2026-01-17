# Quickstart: CLI Service Manager for Termux

**Feature**: CLI Service Manager for Termux
**Date**: January 17, 2025

## Overview

The CLI Service Manager provides a unified command-line interface to manage all services in the AI Manim Video Generator system. It works on desktop Linux/macOS/Windows and Android Termux.

## Prerequisites

### System Requirements
- **Node.js**: 18.0.0 or higher
- **Docker**: For containerized services (optional but recommended)
- **kubectl**: For Kubernetes deployments (optional)
- **Termux**: For Android support (optional)

### Installation

#### Via npm (Recommended)
```bash
npm install -g @ai-manim/cli-manager
```

#### Via apt (Termux)
```bash
apt update && apt install cli-service-manager
```

#### Via pip (Fallback)
```bash
pip install cli-service-manager
```

#### Manual Installation
```bash
git clone <repository>
cd cli-service-manager
npm install
npm run build
npm link
```

## Getting Started

### 1. Initialize Configuration
```bash
# Create default configuration
cli init

# Or specify custom config location
cli init --config ~/.config/cli-manager/config.json
```

### 2. Add Services
```bash
# Add the AI Manim Worker
cli service add ai-manim-worker \
  --type docker \
  --image ai-manim-worker:latest \
  --ports 8787:8787 \
  --dependencies manim-renderer

# Add the Manim Renderer
cli service add manim-renderer \
  --type docker \
  --image manim-renderer:latest \
  --ports 8000:8000
```

### 3. Start Services
```bash
# Start all services
cli start all

# Or start specific service
cli start ai-manim-worker

# Check status
cli status
```

## Basic Usage

### Service Management
```bash
# List all services
cli service list

# View service details
cli service info ai-manim-worker

# Edit service configuration
cli service edit ai-manim-worker

# Remove service
cli service remove ai-manim-worker
```

### Environment Management
```bash
# Switch to development environment
cli env use dev

# List available environments
cli env list

# Show current environment
cli env current
```

### Monitoring & Debugging
```bash
# View service health
cli health

# View service logs
cli logs ai-manim-worker

# Follow logs in real-time
cli logs ai-manim-worker --follow

# View command history
cli history

# View system information
cli info
```

## Development Workflow

### Setting Up Development Environment
```bash
# Setup development environment
cli dev setup

# Start development services with hot reload
cli dev start

# Run tests
cli dev test

# View development logs
cli dev logs
```

### Working with Different Environments
```bash
# Development
cli env use dev
cli start all

# Staging
cli env use staging
cli deploy all

# Production
cli env use prod
cli deploy all --confirm
```

## Termux-Specific Usage

### Mobile-Optimized Commands
```bash
# Start with resource limits for mobile
cli start all --mobile

# Monitor battery and network status
cli status --mobile

# Use offline mode when network is poor
cli --offline start ai-manim-worker
```

### Termux Installation
```bash
# Update packages
pkg update && pkg upgrade

# Install Node.js
pkg install nodejs

# Install CLI manager
npm install -g @ai-manim/cli-manager

# Configure for Termux
cli config set mobile true
cli config set resource-limits "cpu=0.5,memory=512Mi"
```

## Configuration

### Configuration File
```json
{
  "version": "1.0",
  "environments": {
    "dev": {
      "services": ["ai-manim-worker", "manim-renderer"],
      "baseUrl": "http://localhost:3000"
    },
    "prod": {
      "services": ["ai-manim-worker", "manim-renderer"],
      "baseUrl": "https://api.example.com"
    }
  },
  "services": {
    "ai-manim-worker": {
      "type": "docker",
      "image": "ai-manim-worker:latest",
      "ports": [{"container": 8787, "host": 8787}],
      "dependencies": ["manim-renderer"],
      "environment": {
        "NODE_ENV": "production",
        "LOG_LEVEL": "info"
      }
    }
  },
  "mobile": {
    "enabled": false,
    "resourceLimits": "cpu=1.0,memory=1Gi",
    "networkTimeout": 30
  }
}
```

### Environment Variables
```bash
# Set default environment
export CLI_MANAGER_ENV=dev

# Set config location
export CLI_MANAGER_CONFIG=~/.config/cli-manager/config.json

# Enable debug logging
export CLI_MANAGER_DEBUG=true

# Set Docker host for remote Docker
export DOCKER_HOST=tcp://localhost:2376
```

## Troubleshooting

### Common Issues

#### Services Won't Start
```bash
# Check service dependencies
cli service deps ai-manim-worker

# View service logs
cli logs ai-manim-worker --lines 50

# Check Docker/Kubernetes status
cli info
```

#### Network Issues
```bash
# Test connectivity
cli network test

# Use offline mode
cli --offline status

# Configure proxy
cli config set proxy "http://proxy.company.com:8080"
```

#### Permission Issues
```bash
# Check current user permissions
cli permissions check

# Run with sudo (not recommended)
sudo cli start all

# Fix Docker permissions
sudo usermod -aG docker $USER
```

### Debug Mode
```bash
# Enable verbose logging
cli --verbose status

# View debug logs
cli logs --debug

# Run diagnostic checks
cli doctor
```

## Advanced Usage

### Custom Scripts
```bash
# Create custom startup script
cli script create pre-start "echo 'Starting services...'"

# List available scripts
cli script list

# Run custom script
cli script run pre-start
```

### Bulk Operations
```bash
# Export service configurations
cli service export > services.json

# Import service configurations
cli service import services.json

# Backup current state
cli backup create

# Restore from backup
cli backup restore latest
```

### Integration with CI/CD
```bash
# CI/CD deployment
cli env use prod
cli deploy all --wait --timeout 300

# Health checks for CI
cli health --format json | jq '.services[].health.status'

# Rollback on failure
cli rollback --service ai-manim-worker --version previous
```

## API Reference

The CLI includes a REST API for programmatic access:

```bash
# Start HTTP server
cli server start --port 3000

# API endpoints are available at http://localhost:3000
# See contracts/api.yaml for full API specification
```

## Contributing

### Development Setup
```bash
# Clone repository
git clone <repository>
cd cli-service-manager

# Install dependencies
npm install

# Run tests
npm test

# Build for distribution
npm run build
```

### Code Structure
```
src/
├── commands/     # CLI command implementations
├── services/     # Service management logic
├── utils/        # Utility functions
├── types/        # TypeScript type definitions
└── index.ts      # Main CLI entry point
```

## Support

### Getting Help
```bash
# Show help for any command
cli --help
cli <command> --help

# Report issues
cli bug-report

# Check for updates
cli update check
```

### Community Resources
- **Documentation**: https://docs.cli-manager.dev
- **GitHub Issues**: Report bugs and request features
- **Discord**: Join our community for support

---

**Happy managing! 🎉**