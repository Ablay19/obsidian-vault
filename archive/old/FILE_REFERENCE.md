# 📋 Quick File Reference Guide

## 🎯 Quick Access to Common File Types

### 🚀 Scripts
```bash
# Universal script runner
./scripts/run.sh help
./scripts/run.sh quick-start
./scripts/run.sh deploy-cloudflare
./scripts/run.sh start-services
```

### 📚 Documentation
```bash
# Main guides and tutorials
ls docs/guides/
ls docs/deployment/
ls docs/development/
```

### ⚙️ Configuration
```bash
# Application and deployment configs
ls config/
cat config/config.yml
cat config/docker/docker-compose.yml
```

### ☸️ Kubernetes
```bash
# K8s manifests and environments
ls k8s/
ls k8s/overlays/production/
```

### 🧪 Tests
```bash
# Test files and suites
ls tests/integration/
ls tests/performance/
```

## 📁 Key Directory Mappings

| Purpose | Old Location | New Location | Example |
|---------|-------------|--------------|---------|
| Service Scripts | `scripts/` | `scripts/services/` | `start-services.sh` |
| Deployment | `docs/*.md` | `docs/deployment/` | `cloudflare-setup.md` |
| Docker Config | `*.yml` | `config/docker/` | `docker-compose.yml` |
| K8s Config | `k8s/*.yaml` | `config/k8s/` | `deployment.yaml` |
| Setup Guides | `docs/*setup*` | `docs/guides/` | `google-cloud-setup.md` |
| Old Files | `*old*` | `archive/old/` | `deprecated-config.md` |

## 🔧 Common Tasks

### 1. Start Development Environment
```bash
# Old way
./scripts/dev.sh

# New way (multiple options)
./scripts/run.sh dev
make dev
./scripts/dev/dev.sh
```

### 2. Deploy to Production
```bash
# Old way
./scripts/deploy-cloudflare.sh

# New way (organized)
./scripts/run.sh deploy-cloudflare
./scripts/deployment/deploy-cloudflare.sh
make deploy-cloudflare
```

### 3. Quick Setup
```bash
# Old way
./scripts/quick-start.sh

# New way (organized)
./scripts/run.sh quick-start
./scripts/setup/quick-start.sh
make quick-start
```

### 4. System Health Check
```bash
# Old way
./scripts/system-check.sh

# New way (organized)
./scripts/run.sh system-check
./scripts/utilities/system-check.sh
make system-check
```

## 📖 What's Where?

### Missing Something? Try These Locations:
```bash
# Configuration files
find config/ -name "*.yml" -o -name "*.yaml" -o -name "*.json"

# Documentation
find docs/ -name "*.md" | head -10

# Scripts by category
find scripts/ -name "*.sh" | grep -E "(deploy|setup|service)"

# Kubernetes manifests
find k8s/ -name "*.yaml"

# Test files
find tests/ -name "*.yml" -o -name "*.go"
```

### Recent Changes:
- **Scripts**: Now organized by function (deployment, services, setup, etc.)
- **Documentation**: Categorized into guides, deployment, development
- **Configuration**: Separated by technology (docker, k8s, database)
- **K8s**: Organized with base manifests and environment overlays
- **Archive**: Old and deprecated files moved to `archive/`

## 🚨 Migration Tips

### Updating References:
If you find old references in scripts or documentation:
1. **Search for old paths**: `grep -r "old-script-name.sh" .`
2. **Update to new paths**: Replace with organized structure
3. **Test the changes**: Ensure scripts still work after updates

### Adding New Files:
1. **Follow the structure**: Use appropriate directory category
2. **Update docs**: Add to relevant documentation
3. **Update scripts**: Include new files in automation

### For Contributors:
- **Check structure first**: Use `find` to locate correct file type
- **Follow naming**: Use existing patterns for consistency  
- **Document**: Add comments and update relevant docs

---

**Need help?** Check the full guide: `docs/guides/FILE_ORGANIZATION.md`