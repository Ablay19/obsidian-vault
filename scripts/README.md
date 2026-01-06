# Obsidian Automation Scripts

This directory contains comprehensive scripts for managing, deploying, and maintaining the Obsidian Automation platform.

## 📋 Script Overview

### 🎯 Main Control Scripts

#### [`control.sh`](./control.sh) - Master Control Center
**All-in-one operations center for complete application management**

```bash
# Interactive mode
./scripts/control.sh

# Quick commands
./scripts/control.sh start      # Start application
./scripts/control.sh stop       # Stop application
./scripts/control.sh restart    # Restart application
./scripts/control.sh status     # Check status
./scripts/control.sh dev        # Development mode
./scripts/control.sh build      # Build application
./scripts/control.sh test       # Run tests
./scripts/control.sh deploy     # Deploy services
./scripts/control.sh clean      # Clean build artifacts
./scripts/control.sh health     # Health check
./scripts/control.sh setup      # Initial setup
```

**Features:**
- 🚀 Quick operations (start/stop/restart)
- 🔧 Development tools (build/test/lint)
- 📊 Build & deploy (Docker/Cloudflare/Render)
- 🛠️ Environment management
- 🧹 Maintenance & cleanup
- 📈 Monitoring & debugging
- 🐳 Docker operations
- ☁️ Cloud services
- 🔐 Security & authentication
- 📋 System information
- 🎯 One-click actions

---

#### [`quick-start.sh`](./quick-start.sh) - Quick Start & Run
**One-click operations for rapid development and deployment**

```bash
# Interactive mode
./scripts/quick-start.sh

# Quick commands
./scripts/quick-start.sh start      # Quick start
./scripts/quick-start.sh dev        # Development mode
./scripts/quick-start.sh build      # Production build
./scripts/quick-start.sh docker     # Docker quick start
./scripts/quick-start.sh status     # Check status
./scripts/quick-start.sh stop       # Stop services
./scripts/quick-start.sh setup      # Environment setup
```

**Features:**
- 🚀 Quick start (build & run)
- 🧪 Development mode with debug
- 🏗️ Production optimized build
- 🐳 Docker quick start
- 📊 Start with dashboard focus
- 🔧 Environment setup
- 🧹 Clean & restart
- 📋 Status checking
- 🛑 Service management

---

#### [`dev.sh`](./dev.sh) - Development Tools
**Comprehensive development environment management**

```bash
# Interactive mode
./scripts/dev.sh
```

**Features:**
- 🔨 Build & test operations
- 🚀 Deploy services (Render/Cloudflare/Docker)
- 📊 Monitor & debug (logs/health/performance)
- 🔧 Development tools (coming soon)
- 🧹 Maintenance tools (coming soon)
- 📋 Environment setup (coming soon)
- 🗄️ Database operations (coming soon)
- 📦 Package management (coming soon)
- 🔐 Security & auth (coming soon)

---

#### [`env-setup.sh`](./env-setup.sh) - Environment Setup
**Complete environment configuration and validation**

```bash
# Interactive mode
./scripts/env-setup.sh

# Quick commands
./scripts/env-setup.sh init       # Initial setup
./scripts/env-setup.sh config     # Configure .env
./scripts/env-setup.sh validate   # Validate configuration
./scripts/env-setup.sh test       # Test environment
```

**Features:**
- 🔧 Initial setup (system requirements/dependencies)
- 📝 Configure .env (interactive configuration)
- 🗄️ Database setup (connection/migrations)
- 🤖 AI services setup (testing/validation)
- 📱 Bot services setup (Telegram/WhatsApp)
- 🔐 Authentication setup (OAuth/session)
- 🐳 Docker environment (Dockerfile/docker-compose)
- ☁️ Cloud services (Cloudflare/Render)
- 🧪 Test environment (validation)
- 📊 Validate configuration (comprehensive checks)

---

#### [`maintenance.sh`](./maintenance.sh) - Maintenance & Cleanup
**System health, optimization, and cleanup operations**

```bash
# Interactive mode
./scripts/maintenance.sh

# Quick commands
./scripts/maintenance.sh clean      # Clean build artifacts
./scripts/maintenance.sh logs       # Clean logs
./scripts/maintenance.sh health     # Health check
./scripts/maintenance.sh deep       # Deep clean
./scripts/maintenance.sh optimize   # Optimize system
./scripts/maintenance.sh report     # Generate report
```

**Features:**
- 🧹 Clean build artifacts (binaries/cache/temp)
- 📊 Clean logs (archive/remove)
- 🗄️ Database maintenance (optimization/cleanup)
- 📦 Clean dependencies (modules/vendor)
- 🐳 Docker cleanup (containers/images)
- 🔄 Reset configuration (backup/reset)
- 📈 System health check (comprehensive)
- 🧹 Deep clean (complete cleanup)
- 📋 Generate report (system status)
- 🔧 Optimize system (performance)

---

### 🌐 Cloud Services Scripts

#### [`deploy-cloudflare.sh`](./deploy-cloudflare.sh) - Cloudflare Deployment
**Deploy and manage Cloudflare Workers and services**

```bash
./scripts/deploy-cloudflare.sh
```

**Features:**
- 🚀 Deploy AI Proxy Worker
- 📊 Deploy Analytics Worker
- ⚡ Deploy Cache Worker
- 📦 Create R2 storage bucket
- 🗄️ Create D1 database
- 🗂️ Create KV namespaces
- 🔧 Update configuration with IDs
- 📋 Deployment summary

#### [`monitor-cloudflare.sh`](./monitor-cloudflare.sh) - Cloudflare Monitoring
**Monitor Cloudflare Workers performance and health**

```bash
./scripts/monitor-cloudflare.sh
```

**Features:**
- 🏥 Health status check
- ⚡ Response time testing
- 📈 Worker status monitoring
- 📊 Log metrics collection
- 📋 Quick actions reference

---

### 📋 Legacy Scripts

#### [`complete-setup.sh`](./complete-setup.sh) - Complete Setup (Legacy)
**Original setup script - use env-setup.sh for new installations**

#### [`dashboard.sh`](./dashboard.sh) - Dashboard (Legacy)
**Dashboard management - integrated into control.sh**

#### [`system-check.sh`](./system-check.sh) - System Check (Legacy)
**System validation - integrated into maintenance.sh**

#### [`install_doppler`](./install_doppler) - Doppler Installation
**Install Doppler for secrets management**

#### [`requirements.txt`](./requirements.txt) - Python Requirements
**Python dependencies for auxiliary tools**

---

## 🚀 Quick Start Guide

### 1. Initial Setup
```bash
# Clone and navigate to project
git clone <repository-url>
cd obsidian-vault

# Run initial setup
./scripts/env-setup.sh init

# Configure environment
./scripts/env-setup.sh config
```

### 2. Development Mode
```bash
# Quick start in development
./scripts/quick-start.sh dev

# Or use master control
./scripts/control.sh dev
```

### 3. Production Build
```bash
# Build for production
./scripts/quick-start.sh build

# Start production
./scripts/quick-start.sh start
```

### 4. Docker Deployment
```bash
# Quick Docker start
./scripts/quick-start.sh docker

# Or use Docker menu
./scripts/control.sh → 7 (Docker Operations)
```

### 5. Cloud Deployment
```bash
# Deploy to Cloudflare
./scripts/control.sh → 8 → 1

# Deploy to Render
./scripts/control.sh → 3 → 4
```

---

## 📖 Usage Examples

### Development Workflow
```bash
# 1. Start development mode
./scripts/control.sh dev

# 2. Make changes and test
./scripts/control.sh → 2 → 2 (Run tests)

# 3. Build and check
./scripts/control.sh → 2 → 1 (Build)

# 4. View logs
./scripts/control.sh → 6 → 1 (View logs)
```

### Production Deployment
```bash
# 1. Build and test
./scripts/control.sh → 3 → 1 (Build)
./scripts/control.sh → 2 → 2 (Test)

# 2. Deploy to cloud
./scripts/control.sh → 8 → 1 (Cloudflare)

# 3. Monitor deployment
./scripts/control.sh → 6 → 5 (Cloudflare monitoring)
```

### Maintenance Operations
```bash
# 1. Health check
./scripts/control.sh → 5 → 7 (Health check)

# 2. Clean if needed
./scripts/control.sh → 5 → 1 (Clean artifacts)

# 3. Optimize system
./scripts/control.sh → 5 → 10 (Optimize)
```

---

## 🔧 Script Architecture

### Modular Design
- **Master Control**: `control.sh` orchestrates all operations
- **Specialized Scripts**: Each script handles specific domains
- **Unified Interface**: Consistent menus and commands
- **Error Handling**: Comprehensive error checking and recovery

### Color-Coded Output
- 🟢 **Green**: Success operations
- 🔴 **Red**: Errors and failures
- 🟡 **Yellow**: Warnings and cautions
- 🔵 **Blue**: Status and progress
- 🟣 **Purple**: Information and help

### Interactive & CLI Modes
All scripts support both:
- **Interactive Mode**: Menu-driven navigation
- **CLI Mode**: Direct command execution

---

## 🛠️ Customization

### Adding New Scripts
1. Create script in `scripts/` directory
2. Add executable permissions: `chmod +x scripts/new-script.sh`
3. Integrate with `control.sh` by adding menu options
4. Update documentation

### Environment Variables
Key environment variables used by scripts:
- `ENVIRONMENT_MODE`: dev/prod
- `TELEGRAM_BOT_TOKEN`: Bot authentication
- `GEMINI_API_KEY`: AI service
- `TURSO_DATABASE_URL`: Database connection
- `SESSION_SECRET`: Security

### Configuration Files
- `.env`: Environment configuration
- `.env.example`: Configuration template
- `Dockerfile`: Docker build configuration
- `docker-compose.yml`: Docker compose setup

---

## 📚 Best Practices

### Development
1. Always use `env-setup.sh` for initial configuration
2. Test in development mode before production
3. Use `maintenance.sh health` regularly
4. Keep dependencies updated with `env-setup.sh test`

### Production
1. Use production builds with optimization
2. Deploy with proper health checks
3. Monitor with cloud service tools
4. Regular maintenance with `maintenance.sh`

### Security
1. Never commit `.env` files
2. Use strong session secrets
3. Regular security audits
4. Keep dependencies updated

---

## 🆘 Troubleshooting

### Common Issues
1. **Permission Denied**: Run `chmod +x scripts/*.sh`
2. **Go Not Found**: Install Go 1.21+
3. **Docker Not Found**: Install Docker
4. **Environment Missing**: Run `./scripts/env-setup.sh init`

### Getting Help
- Use `control.sh → 10` for system information
- Run `maintenance.sh health` for diagnostics
- Check logs with `control.sh → 6 → 1`
- Generate report with `maintenance.sh report`

---

## 📄 License

These scripts are part of the Obsidian Automation project and follow the same license terms.

---

**🎉 Happy automating!** Use these scripts to streamline your development and deployment workflows.