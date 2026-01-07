# Obsidian Bot - Project File Organization

This document explains the organized structure of the Obsidian Bot project files and provides guidelines for maintaining this organization.

## 📁 Project Structure Overview

```
obsidian-vault/
├── README.md                     # Main project documentation
├── Makefile                      # Build and automation targets
├── go.mod & go.sum              # Go module definition
├── AGENTS.md                     # Agent development guidelines
├── scripts/                       # Organized automation scripts
├── cmd/                          # Application entry points
│   ├── ssh-server/               # SSH management server
│   ├── bot/                      # Main bot application
│   ├── render-tui/               # TUI rendering
│   └── test-*/                   # Test applications
├── internal/                      # Internal Go packages
│   ├── dashboard/                # Dashboard components
│   ├── ssh/                      # SSH server internals
│   └── [other packages]        # Other internal modules
├── docs/                          # Documentation
│   ├── guides/                   # User guides and tutorials
│   ├── deployment/               # Deployment documentation
│   ├── development/               # Development guides
│   └── architecture/             # System architecture
├── config/                        # Configuration files
│   ├── docker/                   # Docker configurations
│   ├── k8s/                     # Kubernetes manifests
│   ├── database/                 # Database configurations
│   └── [other configs]           # Other configuration files
├── k8s/                          # Kubernetes deployment
│   ├── base/                     # Base manifests
│   ├── overlays/                  # Environment-specific overlays
│   └── [k8s files]               # Additional K8s resources
├── tests/                          # Test files and configs
│   ├── integration/               # Integration tests
│   ├── e2e/                     # End-to-end tests
│   └── performance/              # Performance tests
└── archive/                        # Archived/deprecated files
    ├── old/                     # Old project files
    └── deprecated/              # Deprecated components
```

## 📚 Documentation Organization (`docs/`)

### 📖 Guides (`docs/guides/`)
- **WhatsApp Setup**: WhatsApp integration setup guide
- **Performance Optimization**: System optimization techniques
- **SSH Access Alternatives**: Remote access methods
- **Google Cloud Setup**: GCP configuration
- **Cloudflare Setup**: Cloudflare Workers deployment
- **Architecture**: System design and architecture documents
- **Refactoring Documentation**: Code refactoring guidelines

### 🚀 Deployment (`docs/deployment/`)
- **Docker Deployment**: Container-based deployment
- **Google Cloud Quick Start**: GCP quick deployment
- **Cloudflare Monitoring**: Cloudflare deployment monitoring
- **Cloudflare Workers**: Cloudflare Workers deployment
- **Kubernetes**: K8s deployment guide

### 🔨 Development (`docs/development/`)
- Development guides and best practices

### 🏗️ Architecture (`docs/architecture/`)
- **Conductor Tracks**: AI provider expansion plans
- **Product Guidelines**: Product development guidelines
- **Workflow**: Development workflow documentation
- **Tech Stack**: Technology stack overview
- **Code Style**: Go coding standards

## ⚙️ Configuration Organization (`config/`)

### 🐳 Docker (`config/docker/`)
- `docker-compose.yml`: Multi-service orchestration

### ☸️ Kubernetes (`config/k8s/`)
- Base manifests for deployments
- Configuration maps and secrets
- Service and deployment configurations

### 🗄️ Database (`config/database/`)
- `sqlc.yaml`: SQL code generation config

### 📋 Other Configs
- `config.yml`: Application configuration
- `cli.yml`: CLI tool configuration
- `render.yaml`: Rendering configuration

## 🐳 Kubernetes Organization (`k8s/`)

### 📦 Base (`k8s/base/`)
- Core Kubernetes manifests
- Service definitions
- RBAC and network policies

### 🔧 Overlays (`k8s/overlays/`)
- Environment-specific customizations
- `development/`: Development environment
- `staging/`: Staging environment
- `production/`: Production environment

### 🌍 Environments
- Separate overlays for different deployment environments
- Kustomize configurations for each environment

## 🧪 Scripts Organization (`scripts/`)

### 📦 Scripts Runner
- `run.sh`: Universal script runner with auto-discovery

### 🚀 Deployment (`scripts/deployment/`)
- Cloudflare deployment scripts
- Docker deployment automation
- Simple deployment options

### 🔧 Services (`scripts/services/`)
- Service management and control
- Start/stop/restart operations

### ⚙️ Setup (`scripts/setup/`)
- Environment setup and configuration
- Quick start and initialization

### 📊 Monitoring (`scripts/monitoring/`)
- Health checks and monitoring
- Integration testing scripts

### 🛠️ Utilities (`scripts/utilities/`)
- System checks and maintenance
- Migration and utility scripts

### ☸️ Kubernetes (`scripts/k8s/`)
- K8s deployment automation
- Secrets management

### 🔨 Development (`scripts/dev/`)
- Development environment setup
- Debugging and testing tools

## 🗃️ Archive Organization (`archive/`)

### 📂 Old Files (`archive/old/`)
- Previous versions of setup files
- Deprecated configuration files
- Old documentation

### 🚫 Deprecated (`archive/deprecated/`)
- Components no longer in use
- Outdated deployment methods
- Legacy code organization

## 🎯 File Organization Guidelines

### 📝 Naming Conventions

#### Files
- Use **kebab-case** for file names (e.g., `docker-compose.yml`)
- Use **descriptive names** that clearly indicate purpose
- Include **version numbers** for compatibility files when needed

#### Directories
- Use **lowercase** for directory names
- Use **singular form** (e.g., `config/` not `configs/`)
- Use **short, meaningful names** (e.g., `docs/` not `documentation/`)

### 🏗️ Structural Principles

#### 1. Separation of Concerns
- **Configuration**: Separate from code
- **Documentation**: Organized by type and purpose
- **Deployment**: Environment-specific configurations
- **Tests**: Separate test types and environments

#### 2. Logical Grouping
- **By Function**: Related files grouped together
- **By Lifecycle**: Development vs. production files
- **By Technology**: Docker, K8s, database configs

#### 3. Accessibility
- **Common Locations**: Frequently used files in accessible locations
- **Clear Hierarchy**: Obvious where to find specific file types
- **Consistent Patterns**: Predictable structure across components

### 📋 File Type Guidelines

#### Configuration Files
- **YAML**: For structured configuration
- **JSON**: For data interchange and simple config
- **ENV**: For environment variables
- **Markdown**: For documentation and comments

#### Documentation Files
- **README**: Project overview and quick start
- **GUIDE**: Step-by-step instructions
- **SPEC**: Technical specifications
- **FAQ**: Common questions and answers

#### Deployment Files
- **Compose Files**: Multi-container orchestration
- **Dockerfiles**: Container build definitions
- **K8s Manifests**: Kubernetes deployment
- **CI/CD**: Pipeline configurations

## 🔄 Maintenance Guidelines

### 📅 Regular Tasks
- **Weekly**: Review and archive old files
- **Monthly**: Update documentation and reorganize if needed
- **Quarterly**: Review structure and optimize
- **Annually**: Major restructuring and cleanup

### 🧹 Cleanup Process
1. **Identify** unused or duplicate files
2. **Review** file purposes and relevance
3. **Archive** old but potentially needed files
4. **Delete** truly obsolete files
5. **Update** references and documentation

### 📝 Documentation Updates
- **README.md**: Update when structure changes
- **Internal docs**: Keep team documentation current
- **Comments**: Add file purpose comments to new files
- **Diagrams**: Update architecture diagrams

## 🚀 Getting Started with This Structure

### For New Developers
1. **Clone repository**: `git clone <repo-url>`
2. **Read documentation**: Start with `README.md`
3. **Explore structure**: Use `tree` or `find` to understand layout
4. **Set up environment**: Follow `docs/development/` guides
5. **Run tests**: Execute from `tests/` directory

### For Operations
1. **Deployment**: Use `scripts/deployment/` automation
2. **Configuration**: Modify files in `config/` directories
3. **Monitoring**: Use scripts in `scripts/monitoring/`
4. **Troubleshooting**: Check `docs/guides/` for common issues

### For Contributors
1. **Code changes**: Update appropriate `internal/` packages
2. **Documentation**: Add to relevant `docs/` subdirectory
3. **Tests**: Add to appropriate `tests/` subdirectory
4. **Scripts**: Add automation to `scripts/` categories

## 📊 Benefits of This Organization

### 🎯 Efficiency
- **Fast Discovery**: Find files quickly with logical structure
- **Reduced Cognitive Load**: Predictable file locations
- **Team Collaboration**: Shared understanding of structure
- **Onboarding**: New team members adapt quickly

### 🔧 Maintainability
- **Scalable Growth**: Clear places to add new components
- **Dependency Management**: Related files grouped together
- **Version Control**: Meaningful commit scopes
- **Refactoring**: Easier to reorganize related code

### 📚 Documentation
- **Targeted Docs**: Documentation with relevant code
- **Version Alignment**: Docs match code structure
- **User Experience**: Better-organized help and guides
- **Search Optimization**: Easier to find relevant information

### 🚀 Deployment
- **Environment Management**: Separate configs per environment
- **Pipeline Efficiency**: Clear targets for CI/CD
- **Configuration Validation**: Easier to validate setups
- **Rollback Support**: Clear configuration history

---

**Last Updated**: January 2026  
**Maintainer**: Obsidian Bot Team  
**Version**: 1.0.0  

For questions or suggestions about this organization, please open an issue or contact the development team.