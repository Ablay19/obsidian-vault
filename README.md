# Obsidian Vault - README

## 🎯 Welcome to Obsidian Vault

A comprehensive automation platform that combines AI, communication, and productivity tools to create an intelligent vault of knowledge and automation.

## 🚀 Quick Start

### Environment Setup
```bash
# Validate environment variables
./scripts/validate-env.sh

# Set up environment interactively
./scripts/setup/env-setup.sh
```

### Development
```bash
# Quick start with Docker
docker-compose up -d

# Manual testing checklist
./scripts/test-manual.sh

# View logs
docker-compose logs -f
```

### Production Deployment
```bash
# Deploy to staging
./scripts/deploy.sh staging deploy

# Deploy to production
./scripts/deploy.sh production deploy

# Rollback if needed
./scripts/deploy.sh production rollback
```

## 📁 Project Overview

Obsidian Vault is an advanced automation platform that provides:
- **AI-Powered Knowledge Management**: Intelligent document analysis and organization
- **Multi-Channel Communication**: WhatsApp, web dashboard, API integration
- **Secure Automation**: Scripted workflows with proper error handling
- **Real-Time Monitoring**: Comprehensive system health tracking
- **Scalable Architecture**: Modular design supporting growth and extension

## 🏗️ Key Features

### 🤖 AI Integration
- **Document Analysis**: Automatic categorization and tagging
- **Smart Search**: Natural language processing for knowledge discovery
- **Content Generation**: AI-assisted document creation
- **Insight Engine**: Pattern recognition across data sources

### 📱 Communication Hub
- **WhatsApp Integration**: Direct messaging and automation
- **Web Dashboard**: Modern UI with real-time updates
- **API Management**: RESTful API for external integrations
- **Notification System**: Multi-channel alerting and updates

### 🛠️ Automation Engine
- **Script Management**: Organized automation scripts
- **Task Scheduling**: Time-based and event-driven automation
- **Workflows**: Complex multi-step processes
- **Security Scanning**: Automated vulnerability assessment

### 📊 Analytics & Monitoring
- **Performance Metrics**: System and application performance tracking
- **User Behavior Analytics**: Interaction patterns and usage insights
- **Business Intelligence**: Data-driven decision support
- **Real-Time Alerts**: Proactive system notifications

## 🔧 Technology Stack

### Backend
- **Language**: Go (v1.25.4)
- **Framework**: Gin Web Framework
- **Database**: SQLite with SQLC migrations
- **Caching**: Redis for performance
- **Authentication**: JWT-based with OAuth support
- **AI Integration**: Multiple providers (OpenAI, Gemini, etc.)

### Frontend
- **Dashboard**: HTML/CSS/JavaScript with Alpine.js
- **UI Framework**: Custom with responsive design
- **Real-Time Updates**: WebSocket connections
- **Component Library**: Modular and reusable components

### Infrastructure
- **Containerization**: Docker support
- **Orchestration**: Kubernetes manifests
- **CI/CD**: GitHub Actions workflows
- **Monitoring**: Comprehensive health checks
- **Security**: RBAC and secrets management

## 🚀 Quick Commands

```bash
# Environment validation
./scripts/validate-env.sh

# Start all services with Docker
docker-compose up -d

# View service status
docker-compose ps

# View logs
docker-compose logs -f

# Stop all services
docker-compose down

# Deploy to production
./scripts/deploy.sh production deploy
```

## 📚 Documentation

### For Users
- [**Setup Guide**](docs/guides/quick-start.md) - Get started quickly
- [**Development Guide**](docs/development/) - Development workflows
- [**Deployment Guide**](docs/deployment/) - Production deployment
- [**API Reference**](docs/api/) - API documentation
- [**Architecture**](docs/architecture/) - System design

### For Developers
- [**Agent Development**](AGENTS.md) - Contribute to AI agents
- [**Code Standards**](conductor/code_styleguides/go.md) - Go coding standards
- [**File Organization**](docs/guides/FILE_ORGANIZATION.md) - Project structure guide

## 🔧 Configuration

### Environment Variables
See `.env.example` for required configuration
- Database configuration in `config/sqlc.yaml`
- AI provider settings in config files
- Build and deployment options

## 🤝 Getting Help

### Commands
```bash
# Validate environment
./scripts/validate-env.sh

# Deploy application
./scripts/deploy.sh [staging|production] [build|deploy|rollback]

# Manual testing guide
./scripts/test-manual.sh

# Environment setup
./scripts/setup/env-setup.sh
```

### Support

- **Documentation**: Comprehensive guides available
- **Script Runner**: Universal access to all automation
- **Error Handling**: Clear error messages and troubleshooting
- **Community**: Active development and support channels

---

**Version**: 2.1 (Refactored)
**Last Updated**: January 2026
**Maintainer**: Obsidian Vault Team

## 🎯 Ready to Start?

Choose your journey:
- 🔧 **Environment Setup**: `./scripts/setup/env-setup.sh`
- 🚀 **Quick Development**: `docker-compose up -d`
- 📦 **Production Deployment**: `./scripts/deploy.sh production deploy`
- 🧪 **Manual Testing**: `./scripts/test-manual.sh`
- 🤖 **AI Agent Development**: See [AGENTS.md](AGENTS.md)

**Obsidian Vault** - Where knowledge meets automation. 🏦