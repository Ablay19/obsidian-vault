# Deployment Guide for Operations Team

This guide provides comprehensive instructions for deploying and managing the AI Platform components in production environments.

## 📋 Overview

The AI Platform consists of multiple independent components that can be deployed separately:

- **Go Applications**: API Gateway and backend services (Kubernetes)
- **Cloudflare Workers**: AI Worker and related edge functions
- **Monitoring Stack**: Prometheus, Grafana, Alertmanager
- **Shared Packages**: Reusable libraries

## 🏗️ Infrastructure Requirements

### Kubernetes Cluster
- **Version**: 1.24+
- **Nodes**: Minimum 3 nodes for high availability
- **Resources**: 4 CPU cores, 8GB RAM per node minimum
- **Storage**: Persistent volumes for databases and logs

### Cloudflare Account
- **Workers**: Paid plan for production deployments
- **KV Storage**: For session management
- **R2 Storage**: For video file storage
- **Analytics**: For performance monitoring

### Networking
- **Load Balancer**: Nginx Ingress Controller
- **SSL/TLS**: Let's Encrypt certificates
- **DNS**: Domain configuration for API endpoints

## 🚀 Deployment Procedures

### 1. Initial Setup

#### Kubernetes Cluster Setup

```bash
# Create namespaces
kubectl create namespace arch-separation
kubectl create namespace monitoring

# Install NGINX Ingress Controller
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm install nginx-ingress ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace

# Install cert-manager for SSL
helm repo add jetstack https://charts.jetstack.io
helm install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set installCRDs=true
```

#### Cloudflare Workers Setup

```bash
# Install Wrangler CLI
npm install -g wrangler

# Login to Cloudflare
wrangler auth login

# Create KV namespace
wrangler kv:namespace create "SESSION_STORE"

# Create R2 bucket
wrangler r2 bucket create ai-platform-videos
```

### 2. Go Applications Deployment

#### API Gateway Deployment

```bash
# Build and push Docker image
cd apps/api-gateway
docker build -t your-registry/api-gateway:latest .
docker push your-registry/api-gateway:latest

# Deploy to Kubernetes
kubectl apply -f deploy/k8s/go-services/api-gateway.yaml

# Verify deployment
kubectl get pods -n arch-separation
kubectl logs -f deployment/api-gateway -n arch-separation
```

#### Environment Configuration

```yaml
# Create configmap for environment variables
apiVersion: v1
kind: ConfigMap
metadata:
  name: api-gateway-config
  namespace: arch-separation
data:
  LOG_LEVEL: "info"
  PORT: "8080"
  API_GATEWAY_URL: "https://api.your-domain.com"
  DATABASE_URL: "postgres://user:pass@db-host:5432/dbname"
```

#### Secrets Management

```yaml
# Create secrets for sensitive data
apiVersion: v1
kind: Secret
metadata:
  name: api-gateway-secrets
  namespace: arch-separation
type: Opaque
data:
  DATABASE_PASSWORD: <base64-encoded-password>
  JWT_SECRET: <base64-encoded-secret>
  CLOUDFLARE_API_TOKEN: <base64-encoded-token>
```

### 3. Cloudflare Workers Deployment

#### AI Worker Deployment

```bash
cd workers/ai-worker

# Deploy to staging
wrangler deploy --env staging

# Verify deployment
curl https://ai-worker.your-domain.workers.dev/health

# Promote to production
wrangler deploy --env production
```

#### Environment Variables

```bash
# Set environment variables
wrangler secret put LOG_LEVEL --env production
wrangler secret put API_GATEWAY_URL --env production
wrangler secret put CLOUDFLARE_API_TOKEN --env production
```

### 4. Monitoring Stack Deployment

#### Prometheus and Grafana Setup

```bash
cd monitoring

# Deploy monitoring stack
docker-compose up -d

# Access Grafana
# URL: http://localhost:3000
# Default credentials: admin/admin
```

#### Configure Data Sources

1. **Prometheus Data Source**:
   - URL: http://prometheus:9090
   - Access: Server (default)

2. **Import Dashboards**:
   - Upload `monitoring/grafana/dashboards/ai-platform-dashboard.json`

## 🔄 Update Procedures

### Rolling Updates

#### Go Applications

```bash
# Update Docker image
kubectl set image deployment/api-gateway api-gateway=your-registry/api-gateway:v1.1.0 -n arch-separation

# Monitor rollout
kubectl rollout status deployment/api-gateway -n arch-separation

# Rollback if needed
kubectl rollout undo deployment/api-gateway -n arch-separation
```

#### Cloudflare Workers

```bash
# Deploy new version
wrangler deploy --env production

# Monitor traffic
wrangler tail --format=pretty
```

### Blue-Green Deployments

```bash
# Create new deployment with different labels
kubectl apply -f deploy/k8s/go-services/api-gateway-green.yaml

# Switch ingress to green deployment
kubectl patch ingress api-gateway -p '{"spec":{"rules":[{"host":"api.your-domain.com","http":{"paths":[{"path":"/","pathType":"Prefix","backend":{"service":{"name":"api-gateway-green","port":{"number":8080}}}}]}}]}}'

# Verify and remove blue deployment
kubectl delete deployment api-gateway-blue
```

## 📊 Monitoring and Alerting

### Key Metrics to Monitor

- **API Gateway**:
  - Request rate and latency
  - Error rates (4xx, 5xx)
  - Database connection pool usage

- **Cloudflare Workers**:
  - Response times
  - Error rates
  - KV storage usage

- **System Resources**:
  - CPU and memory usage
  - Disk space
  - Network I/O

### Alert Configuration

Critical alerts are configured for:
- Service downtime (>1 minute)
- High error rates (>5%)
- Resource exhaustion (>90% usage)
- Security incidents

## 🔒 Security Procedures

### SSL/TLS Configuration

```yaml
# cert-manager ClusterIssuer
apiVersion: cert-manager.io/v1
kind: ClusterIssuer
metadata:
  name: letsencrypt-prod
spec:
  acme:
    server: https://acme-v02.api.letsencrypt.org/directory
    email: admin@your-domain.com
    privateKeySecretRef:
      name: letsencrypt-prod
    solvers:
    - http01:
        ingress:
          class: nginx
```

### Access Control

- **RBAC**: Configure Kubernetes RBAC for team access
- **Cloudflare Access**: Protect admin endpoints
- **API Keys**: Rotate API keys regularly

### Backup Procedures

```bash
# Database backups
kubectl exec -n arch-separation deployment/postgres -- pg_dump dbname > backup.sql

# Configuration backups
kubectl get configmaps,secrets -n arch-separation -o yaml > config-backup.yaml

# PV backups
kubectl cp -n arch-separation postgres-pod:/var/lib/postgresql/data ./postgres-data-backup
```

## 🚨 Incident Response

### Service Outage

1. **Assess Impact**: Check monitoring dashboards
2. **Identify Root Cause**: Review logs and metrics
3. **Implement Fix**: Deploy hotfix or rollback
4. **Communicate**: Update stakeholders
5. **Post-Mortem**: Document lessons learned

### Rollback Procedures

```bash
# Go applications
kubectl rollout undo deployment/api-gateway -n arch-separation

# Cloudflare Workers
wrangler deploy --env production@1  # Deploy previous version
```

## 📈 Scaling Procedures

### Horizontal Scaling

```bash
# Scale Go deployments
kubectl scale deployment api-gateway --replicas=5 -n arch-separation

# Cloudflare Workers auto-scale based on traffic
```

### Vertical Scaling

```bash
# Update resource requests/limits
kubectl patch deployment api-gateway -n arch-separation --type='json' \
  -p='[{"op": "replace", "path": "/spec/template/spec/containers/0/resources/requests/cpu", "value":"1000m"}]'
```

## 🔧 Maintenance Tasks

### Regular Maintenance

- **Security Updates**: Apply security patches weekly
- **Dependency Updates**: Update dependencies monthly
- **Certificate Renewal**: Automatic with cert-manager
- **Log Rotation**: Configure log aggregation and retention

### Health Checks

```bash
# Kubernetes health
kubectl get pods -n arch-separation
kubectl get ingress -n arch-separation

# Application health
curl https://api.your-domain.com/health

# Worker health
curl https://ai-worker.your-domain.workers.dev/health
```

## 📞 Support and Contacts

- **Emergency**: +1-XXX-XXX-XXXX (24/7)
- **Email**: ops@your-domain.com
- **Slack**: #platform-ops
- **Documentation**: https://docs.your-domain.com/ops

## 📋 Checklist

### Pre-Deployment
- [ ] Infrastructure ready
- [ ] Secrets configured
- [ ] DNS configured
- [ ] SSL certificates ready
- [ ] Monitoring configured

### Post-Deployment
- [ ] Health checks pass
- [ ] Monitoring alerts configured
- [ ] Logs accessible
- [ ] Backups configured
- [ ] Team notified

### Maintenance
- [ ] Security patches applied
- [ ] Performance optimized
- [ ] Documentation updated
- [ ] Incident reviews completed