# Kubernetes Deployment Guide

This comprehensive guide covers deploying Obsidian Bot to Kubernetes with production-ready configurations, monitoring, and scaling.

## 📋 Table of Contents

1. [Prerequisites](#prerequisites)
2. [Quick Start](#quick-start)
3. [Configuration](#configuration)
4. [Deployment](#deployment)
5. [Environments](#environments)
6. [Monitoring](#monitoring)
7. [Scaling](#scaling)
8. [Security](#security)
9. [Troubleshooting](#troubleshooting)

## 🚀 Prerequisites

### Required Tools
- **Kubernetes Cluster** (v1.24+)
- **kubectl** (configured cluster access)
- **kustomize** (v4.0+)
- **Docker** (for building images)
- **Container Registry** access (GCR, Docker Hub, etc.)

### Quick Install Commands
```bash
# Install kubectl
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/

# Install kustomize
curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh" | bash

# Verify installation
kubectl version --client
kustomize version
```

### Cluster Requirements
- **Nodes**: Minimum 2 nodes for HA
- **Storage**: Persistent storage classes available
- **Networking**: Ingress controller for external access
- **Load Balancer**: For external service exposure

## ⚡ Quick Start

### 1. One-Command Deployment
```bash
# Deploy to production
./k8s/scripts/deploy.sh deploy production

# Deploy to staging
./k8s/scripts/deploy.sh deploy staging

# Deploy to development
./k8s/scripts/deploy.sh deploy development
```

### 2. Manual Deployment
```bash
# Create secrets
./k8s/scripts/secrets.sh create interactive

# Deploy
kubectl apply -k k8s/overlays/production/

# Check status
kubectl get pods -n obsidian-system
kubectl get services -n obsidian-system
```

## ⚙️ Configuration

### 1. Secrets Configuration

Create secrets using environment variables:
```bash
export TURSO_DATABASE_URL="libsql://your-db.turso.io"
export TURSO_AUTH_TOKEN="your-auth-token"
export TELEGRAM_BOT_TOKEN="your-telegram-token"
export SESSION_SECRET="your-session-secret"
export GEMINI_API_KEY="your-gemini-key"

./k8s/scripts/secrets.sh create
```

Or use interactive mode:
```bash
./k8s/scripts/secrets.sh create interactive
```

### 2. Environment Variables

| Variable | Required | Description |
|----------|-----------|-------------|
| `TURSO_DATABASE_URL` | Yes | Turso database connection URL |
| `TURSO_AUTH_TOKEN` | Yes | Turso authentication token |
| `TELEGRAM_BOT_TOKEN` | Yes | Telegram bot token |
| `SESSION_SECRET` | Yes | Session encryption secret |
| `GEMINI_API_KEY` | No | Google Gemini API key |
| `GROQ_API_KEY` | No | Groq API key |
| `GOOGLE_APPLICATION_CREDENTIALS` | No | GCP service account JSON file |
| `ENABLE_GOOGLE_LOGGING` | No | Enable Google Cloud logging |

### 3. Image Configuration

```yaml
# k8s/base/deployment.yaml
spec:
  template:
    spec:
      containers:
      - name: obsidian-bot
        image: obsidian-bot:latest
        imagePullPolicy: IfNotPresent
```

## 🚢 Deployment

### 1. Production Deployment

#### Deploy with Custom Values
```bash
# Deploy with specific registry and tag
./k8s/scripts/deploy.sh deploy production obsidian-system gcr.io/obsidian-bot-prod v1.0.0

# Deploy with custom namespace
./k8s/scripts/deploy.sh deploy production my-obsidian-namespace
```

#### Verify Deployment
```bash
# Check pod status
kubectl get pods -n obsidian-system -l app=obsidian-bot

# Check rollout status
kubectl rollout status deployment/obsidian-bot -n obsidian-system

# Get service URL
kubectl get ingress -n obsidian-system

# Get logs
kubectl logs -n obsidian-system -l app=obsidian-bot --tail=50
```

### 2. Using Kustomize

#### Build and Apply
```bash
# Build manifests
kustomize build k8s/overlays/production > manifests.yaml

# Apply manifests
kubectl apply -f manifests.yaml

# Delete deployment
kubectl delete -f manifests.yaml
```

#### Multiple Environments
```bash
# Deploy all environments
for env in development staging production; do
    echo "Deploying to $env..."
    kubectl apply -k k8s/overlays/$env
done
```

## 🌍 Environments

### 1. Production
- **Replicas**: 3 pods
- **Resources**: 512Mi memory, 500m CPU requests
- **Logging**: Google Cloud enabled
- **Monitoring**: Full metrics enabled
- **Autoscaling**: 1-10 pods

### 2. Staging
- **Replicas**: 2 pods
- **Resources**: 256Mi memory, 250m CPU requests
- **Logging**: Debug level
- **Monitoring**: Development metrics
- **Autoscaling**: 1-5 pods

### 3. Development
- **Replicas**: 1 pod
- **Resources**: 128Mi memory, 100m CPU requests
- **Logging**: Debug level
- **Monitoring**: Full profiling enabled
- **Autoscaling**: Disabled

## 📊 Monitoring

### 1. Prometheus Monitoring

#### Service Monitor
```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: obsidian-bot
  namespace: obsidian-system
spec:
  selector:
    matchLabels:
      app: obsidian-bot
  endpoints:
  - port: http
    path: /metrics
    interval: 30s
```

#### Dashboard Configuration
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: obsidian-bot-dashboard
  namespace: monitoring
data:
  obsidian-bot.json: |
    {
      "dashboard": {
        "title": "Obsidian Bot",
        "panels": [
          {
            "title": "Request Rate",
            "type": "graph",
            "targets": [{
              "expr": "rate(http_requests_total[5m])"
            }]
          }
        ]
      }
    }
```

### 2. Google Cloud Monitoring

#### Metrics Configuration
```bash
# Create custom metrics
gcloud monitoring metrics-descriptors create obsidian_bot_requests \
    --type=counter \
    --metric-kind=GAUGE \
    --unit=requests

# Create alert policies
gcloud alpha monitoring policies create \
    --notification-channels=projects/obsidian-bot-prod/notificationChannels/123 \
    --condition-filter='metric.type="logging.googleapis.com/user/obsidian_bot_requests"' \
    --condition-threshold-value=100 \
    --condition-threshold-duration=60s \
    --display-name="Obsidian Bot High Request Rate"
```

### 3. Health Checks

#### Custom Health Endpoints
```yaml
# In deployment.yaml
livenessProbe:
  httpGet:
    path: /api/services/status
    port: http
    scheme: HTTP
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3

readinessProbe:
  httpGet:
    path: /api/services/status
    port: http
    scheme: HTTP
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

## 📈 Scaling

### 1. Horizontal Pod Autoscaling

#### Resource-Based Scaling
```yaml
# k8s/base/autoscaling.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: obsidian-bot
  minReplicas: 2
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

#### Custom Metrics Scaling
```yaml
# Using custom metrics for intelligent scaling
metrics:
- type: External
  external:
    metric:
      name: obsidian_bot_queue_size
    target:
      type: Value
      value: 100
```

### 2. Cluster Autoscaling

#### Node Auto Scaling
```bash
# GKE node pools
gcloud container node-pools create obsidian-bot-pool \
    --cluster=obsidian-bot-cluster \
    --machine-type=e2-medium \
    --num-nodes=2 \
    --min-nodes=1 \
    --max-nodes=10 \
    --enable-autoscaling

# Update node pool
gcloud container node-pools update obsidian-bot-pool \
    --cluster=obsidian-bot-cluster \
    --min-nodes=2 \
    --max-nodes=20
```

### 3. Predictive Scaling

#### Custom Metrics
```yaml
# Define custom scaling metrics
apiVersion: v1
kind: ConfigMap
metadata:
  name: obsidian-bot-metrics
data:
  scaling-metrics.yaml: |
    metrics:
      queue_length:
        query: "obsidian_bot_queue_size"
        target: 50
        action: scale_up
      response_time:
        query: "obsidian_bot_avg_response_time"
        target: 1000
        action: scale_down
```

## 🔒 Security

### 1. Network Policies

#### Traffic Control
```yaml
# k8s/base/network-policy.yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
spec:
  podSelector:
    matchLabels:
      app: obsidian-bot
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
  egress:
  - to: []
    ports:
    - protocol: TCP
      port: 443  # HTTPS for external APIs
```

### 2. Pod Security Policies

#### Security Context
```yaml
# In deployment.yaml
securityContext:
  runAsNonRoot: true
  runAsUser: 1000
  runAsGroup: 1000
  fsGroup: 1000
```

#### Resource Limits
```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### 3. RBAC Configuration

#### Service Account
```yaml
# k8s/base/rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: obsidian-bot-sa
  namespace: obsidian-system
  annotations:
    iam.gke.io/gcp-service-account: obsidian-bot-sa@project.iam.gserviceaccount.com
```

#### Role-Based Access
```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
rules:
- apiGroups: [""]
  resources: ["configmaps", "secrets"]
  verbs: ["get", "list", "watch"]
```

## 🔧 Troubleshooting

### 1. Common Issues

#### Pod Not Starting
```bash
# Check pod status
kubectl describe pod -n obsidian-system -l app=obsidian-bot

# Check events
kubectl get events -n obsidian-system --sort-by='.lastTimestamp'

# Check logs
kubectl logs -n obsidian-system -l app=obsidian-bot --previous
```

#### Image Pull Issues
```bash
# Check image pull policy
kubectl get deployment obsidian-bot -n obsidian-system -o yaml | grep imagePullPolicy

# Test image pull
docker pull gcr.io/obsidian-bot-prod/obsidian-bot:v1.0.0

# Check registry access
gcloud auth configure-docker
```

#### Service Access Issues
```bash
# Check service
kubectl get service obsidian-bot-service -n obsidian-system -o yaml

# Check endpoints
kubectl get endpoints obsidian-bot-service -n obsidian-system

# Test service connectivity
kubectl run test-pod --image=busybox --rm -it --restart=Never -- \
  wget -qO- http://obsidian-bot-service.obsidian-system/api/services/status
```

### 2. Performance Issues

#### Resource Constraints
```bash
# Check resource usage
kubectl top pods -n obsidian-system

# Check node resources
kubectl top nodes

# Check resource limits
kubectl describe pod -n obsidian-system -l app=obsidian-bot | grep -A 10 Limits
```

#### Database Connection Issues
```bash
# Check secrets
kubectl get secrets obsidian-bot-secrets -n obsidian-system -o yaml

# Test database connectivity
kubectl exec -n obsidian-system deployment/obsidian-bot -- \
  curl -I "$TURSO_DATABASE_URL"

# Check network policies
kubectl get networkpolicy -n obsidian-system
```

### 3. Debugging Commands

#### Port Forwarding
```bash
# Forward local port to pod
kubectl port-forward -n obsidian-system deployment/obsidian-bot 8080:8080

# Forward service
kubectl port-forward -n obsidian-system service/obsidian-bot-service 8080:80
```

#### Debug Pods
```bash
# Start debug container
kubectl run debug-pod --image=busybox --rm -it --restart=Never -- \
  /bin/sh

# Mount secrets in debug pod
kubectl run debug-pod --image=busybox --rm -it --restart=Never -- \
  --mount=type=secret,secret-name=obsidian-bot-secrets,target-path=/secrets \
  -- /bin/sh
```

## 📚 Advanced Topics

### 1. Multi-Region Deployment

#### Geographic Distribution
```yaml
# Define regions
regions:
  - name: us-central1
    replicas: 2
  - name: europe-west1
    replicas: 1
  - name: asia-southeast1
    replicas: 1
```

### 2. Blue-Green Deployment

#### Strategy Implementation
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Rollout
metadata:
  name: obsidian-bot
spec:
  replicas: 3
  strategy:
    blueGreen:
      activeService: obsidian-bot-service
      previewService: obsidian-bot-preview
      autoPromotionEnabled: true
      scaleDownDelaySeconds: 30
      prePromotionAnalysis:
        templates:
        - templateName: success-rate
          args: ["success-rate"]
```

### 3. GitOps Integration

#### ArgoCD Application
```yaml
apiVersion: argoproj.io/v1alpha1
kind: Application
metadata:
  name: obsidian-bot
  namespace: argocd
spec:
  project: default
  source:
    repoURL: https://github.com/your-org/obsidian-vault.git
    targetRevision: HEAD
    path: k8s/overlays/production
  destination:
    server: https://kubernetes.default.svc
    namespace: obsidian-system
  syncPolicy:
    automated:
      prune: true
      selfHeal: true
    syncOptions:
    - CreateNamespace=true
```

---

## 🎯 Summary

This Kubernetes deployment provides:
- **Production-ready** configurations with HA and scaling
- **Multi-environment** support (dev, staging, prod)
- **Security best** practices with RBAC and network policies
- **Monitoring and** logging integration with Prometheus and Google Cloud
- **Auto-scaling** based on custom metrics and resource usage
- **GitOps-ready** manifests for continuous deployment

**Deploy Obsidian Bot to Kubernetes with enterprise-grade reliability and scalability! 🚀**