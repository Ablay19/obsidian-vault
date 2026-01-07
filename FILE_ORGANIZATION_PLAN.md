# Obsidian Vault - File Organization Guide

## 📁 Current State

The project has been partially organized but needs better structure. Key issues identified:
- Missing README in root
- Configuration files scattered
- Build artifacts not organized
- Source code mixed with generated files
- Archives and deprecated files in root

## 🎯 Target Organization

```
obsidian-vault/
├── README.md                   # Main project README
├── AGENTS.md                  # Agent development guidelines  
├── Makefile                   # Build and automation
├── go.mod & go.sum           # Go module definition
├── .git/                      # Git configuration
├── .config/                    # Configuration files
│   ├── local/               # Local development config
│   ├── staging/             # Staging environment config
│   └── production/          # Production config
├── .data/                      # Application data
│   ├── local/               # Local development data
│   └── production/          # Production data
├── .cache/                     # Build and dependency cache
│   ├── logs/                # Log files
│   └── build/               # Build artifacts
├── build/                      # Build output directory
│   ├── local/               # Local builds
│   └── production/          # Production builds
├── .backups/                    # Backup directory
├── tmp/                        # Temporary files
├── docs/                       # Documentation
│   ├── guides/              # User guides
│   ├── api/                 # API documentation
│   ├── deployment/          # Deployment guides
│   ├── development/          # Development guides
│   └── architecture/        # System architecture
├── scripts/                     # Automation scripts
│   ├── deploy/              # Deployment scripts
│   ├── setup/               # Setup scripts
│   ├── maintenance/          # Maintenance scripts
│   └── utils/               # Utility scripts
├── k8s/                        # Kubernetes manifests
│   ├── base/                # Base manifests
│   ├── overlays/             # Environment overlays
│   └── environments/         # Environment configs
├── tests/                       # Test files
│   ├── unit/                # Unit tests
│   ├── integration/         # Integration tests
│   └── e2e/                # End-to-end tests
├── cmd/                         # Application entry points
│   ├── bot/                 # Main bot application
│   ├── api/                 # API server
│   ├── cli/                 # CLI tool
│   └── workers/              # Worker processes
├── internal/                    # Internal packages
│   ├── bot/                 # Bot logic
│   ├── api/                 # API handlers
│   ├── auth/                # Authentication
│   ├── config/              # Configuration
│   ├── dashboard/           # Dashboard
│   └── [other packages]   # Other services
├── pkg/                         # Public packages
├── deployments/                 # Deployment configurations
├── configs/                     # Configuration templates
└── archive/                     # Archived files
    ├── old/                 # Previous versions
    └── deprecated/          # Deprecated components
```

## 🚀 Organization Steps

### Step 1: Create Directory Structure
```bash
mkdir -p .config/{local,staging,production} .data/{local,production} .cache/{logs,build} build/{local,production} .backups tmp pkg deployments configs
```

### Step 2: Move Configuration Files
```bash
# Move sensitive config files
mv .env .config/local/
mv config.yml .config/local/
mv cli.yml .config/local/

# Move project config files
mv sqlc.yaml .config/local/
mv k8s/ configs/
```

### Step 3: Organize Source Code
```bash
# Keep current internal/ structure
# Move cmd/ to proper structure if needed
# Keep go.mod/go.sum in root
```

### Step 4: Organize Documentation
```bash
# Organize docs/ by type
mv docs/*.md docs/guides/
# Create api/ subdirectory for API docs
mkdir -p docs/api/
```

### Step 5: Organize Scripts
```bash
# scripts/ already well organized
# Ensure proper permissions
chmod +x scripts/**/*.sh
```

### Step 6: Organize Tests
```bash
# Create test structure
mkdir -p tests/{unit,integration,e2e}
# Move test files to appropriate categories
```

### Step 7: Clean Up Root
```bash
# Remove loose files from root
# Keep only essential files in root
```

## 📋 File Categories

### Configuration Files (.config/)
- Environment-specific configurations
- Secrets and API keys
- Database connection strings
- Feature flags

### Data Files (.data/)
- Local development data
- Production data snapshots
- Database files (SQLite, etc.)

### Build Artifacts (build/, .cache/)
- Compiled binaries
- Build logs
- Dependency cache
- Package files

### Documentation (docs/)
- User guides and tutorials
- API reference documentation
- Deployment instructions
- Architecture documentation

### Scripts (scripts/)
- Deployment automation
- Development tools
- Maintenance scripts
- Utility functions

### Kubernetes (k8s/)
- Base manifests
- Environment-specific overlays
- Configuration templates
- Deployment scripts

### Tests (tests/)
- Unit test suites
- Integration test scenarios
- End-to-end test cases
- Performance benchmarks

### Archive (archive/)
- Previous project versions
- Deprecated components
- Historical documentation
- Backup configurations

## 🎯 Benefits

1. **Clarity**: Clear separation of concerns
2. **Security**: Sensitive files in .config/
3. **Scalability**: Organized growth structure
4. **Maintainability**: Predictable file locations
5. **Collaboration**: Standard structure for team members
6. **Deployment**: Environment-specific configurations
7. **Testing**: Organized test structure

## 🔧 Maintenance Guidelines

### Regular Tasks
- Weekly: Clean up tmp/ directory
- Monthly: Review and archive old files
- Quarterly: Update documentation structure

### File Naming Conventions
- Use kebab-case for directories
- Use descriptive names for files
- Include version numbers for releases
- Use .md for documentation files

### Git Management
- Add .gitignore for sensitive files
- Use branches for development/staging/production
- Tag releases appropriately
- Use submodules for external dependencies

---

**Status**: Draft - Ready for implementation  
**Next**: Execute organization steps and validate structure