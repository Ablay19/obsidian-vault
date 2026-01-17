# Migration Guide: From Monolithic to Microservices Architecture

This guide provides a step-by-step approach to migrating from a monolithic application to the AI Platform's microservices architecture with architectural separation.

## 📋 Overview

### Current State (Monolithic)
- Single large application handling all functionality
- Shared database and codebase
- Tight coupling between components
- Difficult to scale individual features
- Long deployment cycles

### Target State (Microservices)
- Independent Go applications and Cloudflare Workers
- Separate databases per service (future)
- Loose coupling with API contracts
- Independent scaling and deployment
- Shared packages for common functionality

## 🎯 Migration Strategy

### Phase 1: Assessment and Planning (2-4 weeks)

#### 1.1 Application Analysis

**Identify Components**:
- **API Gateway**: Request routing and authentication
- **AI Worker**: AI/ML processing logic
- **Manim Renderer**: Video generation (if exists)
- **Shared Logic**: Common utilities and types

**Analyze Dependencies**:
```bash
# Find code dependencies
grep -r "import.*internal" --include="*.go" .

# Identify database tables usage
grep -r "SELECT.*FROM" --include="*.go" .

# Find shared functions
grep -r "func.*common" --include="*.go" .
```

**Map Data Flows**:
- User requests → API Gateway → AI processing → Response
- Session management and persistence
- File storage and retrieval

#### 1.2 Infrastructure Assessment

**Current Infrastructure**:
- Single server/cluster
- Monolithic database
- Shared file storage
- Unified monitoring

**Target Infrastructure**:
- Kubernetes cluster with namespaces
- Service-specific databases (future)
- Cloudflare Workers + R2 storage
- Distributed monitoring stack

#### 1.3 Risk Assessment

**High-Risk Areas**:
- Data consistency during migration
- Session management across services
- Breaking API changes
- Performance regression

**Mitigation Strategies**:
- Feature flags for gradual rollout
- Database migration scripts
- Comprehensive testing
- Rollback procedures

### Phase 2: Foundation Setup (4-6 weeks)

#### 2.1 Project Structure Setup

```bash
# Create new directory structure
mkdir -p ai-platform/{apps,workers,packages,deploy,tests}

# Initialize Go workspaces
cd ai-platform
go work init

# Setup shared packages first
mkdir -p packages/{shared-types,communication,api-contracts}
```

#### 2.2 Shared Packages Migration

**Extract Common Types**:
```go
// Before (monolithic)
type User struct {
    ID    string `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
}

// After (shared)
package types

type User struct {
    ID    string `json:"id"`
    Email string `json:"email"`
    Name  string `json:"name"`
}
```

**Create Communication Layer**:
```go
// packages/communication/go/client.go
package communication

type HttpClient struct {
    // HTTP client with circuit breaker
}

func NewHttpClient(baseURL string) *HttpClient {
    // Implementation with fail-fast behavior
}
```

#### 2.3 Database Migration Planning

**Current Database Schema**:
```sql
-- Monolithic schema
CREATE TABLE users (...);
CREATE TABLE sessions (...);
CREATE TABLE jobs (...);
```

**Target Schema (Per Service)**:
```sql
-- API Gateway schema
CREATE TABLE sessions (...);

-- AI Worker schema (future)
CREATE TABLE jobs (...);
```

**Migration Strategy**:
1. **Dual-Write**: Write to both old and new schemas
2. **Backfill**: Migrate historical data
3. **Verification**: Ensure data consistency
4. **Cutover**: Switch to new schema

### Phase 3: Component Extraction (8-12 weeks)

#### 3.1 API Gateway Extraction

**Identify API Gateway Code**:
```bash
# Find HTTP handlers
grep -r "http.HandleFunc" --include="*.go" .

# Find middleware
grep -r "middleware" --include="*.go" .

# Find authentication logic
grep -r "auth" --include="*.go" .
```

**Extract API Gateway**:
```go
// apps/api-gateway/cmd/main.go
package main

func main() {
    // Start API Gateway service
    // Route requests to appropriate handlers
    // No AI processing logic here
}
```

**Migration Steps**:
1. **Create separate main.go** for API Gateway
2. **Move HTTP handlers** to new package
3. **Extract middleware** (auth, logging, CORS)
4. **Update imports** to use shared packages

#### 3.2 AI Worker Extraction

**Identify AI Processing Code**:
```bash
# Find AI/ML functions
grep -r "openai\|anthropic\|gemini" --include="*.go" .

# Find job processing logic
grep -r "job.*process" --include="*.go" .

# Find file upload/download logic
grep -r "upload\|download" --include="*.go" .
```

**Extract AI Worker**:
```typescript
// workers/ai-worker/src/index.ts
export default {
    async fetch(request: Request, env: Env): Promise<Response> {
        // Handle AI processing requests
        // Use API Gateway for coordination
    }
}
```

**Migration Steps**:
1. **Port AI logic** from Go to TypeScript
2. **Implement API calls** to API Gateway
3. **Move file handling** to Cloudflare R2
4. **Update session management** via API Gateway

#### 3.3 Testing During Migration

**Create Integration Tests**:
```go
// tests/integration/migration_test.go
func TestAPIGatewayIndependence(t *testing.T) {
    // Test API Gateway works without AI logic
}

func TestAIWorkerViaAPI(t *testing.T) {
    // Test AI Worker communicates via API
}
```

**Feature Flags**:
```go
// Use feature flags for gradual migration
type FeatureFlags struct {
    UseNewAPIGateway bool
    UseNewAIWorker   bool
}

func (f *FeatureFlags) ShouldUseNewAPI() bool {
    return f.UseNewAPIGateway
}
```

### Phase 4: Infrastructure Migration (6-8 weeks)

#### 4.1 Kubernetes Setup

**Create Namespaces**:
```yaml
# deploy/k8s/namespaces.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: arch-separation
  labels:
    name: arch-separation
---
apiVersion: v1
kind: Namespace
metadata:
  name: monitoring
  labels:
    name: monitoring
```

**Deploy API Gateway**:
```yaml
# deploy/k8s/go-services/api-gateway.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: api-gateway
  namespace: arch-separation
spec:
  replicas: 1
  template:
    spec:
      containers:
      - name: api-gateway
        image: api-gateway:latest
        ports:
        - containerPort: 8080
```

#### 4.2 Cloudflare Workers Setup

**Deploy AI Worker**:
```bash
# Deploy to Cloudflare
cd workers/ai-worker
wrangler deploy --env staging

# Configure routing
# Update DNS and load balancer
```

#### 4.3 Monitoring Setup

**Deploy Monitoring Stack**:
```bash
cd monitoring
docker-compose up -d

# Configure dashboards and alerts
```

### Phase 5: Data Migration (2-4 weeks)

#### 5.1 Database Migration

**Create Migration Scripts**:
```sql
-- migrations/001_initial_schema.sql
-- Create new schema for microservices

-- migrations/002_migrate_sessions.sql
-- Migrate session data from monolithic to API Gateway

-- migrations/003_migrate_jobs.sql
-- Migrate job data to AI Worker (future)
```

**Execute Migration**:
```bash
# Run migrations
kubectl apply -f migrations/

# Verify data integrity
./scripts/verify-migration.sh
```

#### 5.2 File Storage Migration

**Migrate Files to R2**:
```bash
# Export files from current storage
./scripts/export-files.sh

# Import to Cloudflare R2
./scripts/import-to-r2.sh

# Update references in database
./scripts/update-file-references.sh
```

### Phase 6: Testing and Validation (4-6 weeks)

#### 6.1 Comprehensive Testing

**End-to-End Tests**:
```typescript
// tests/e2e/migration-e2e.test.ts
describe('Migration E2E Tests', () => {
    it('should handle requests through new architecture', async () => {
        // Test complete request flow
        const response = await request('https://api.new-domain.com/process')
            .post('/ai/generate')
            .send({ prompt: 'test' });

        expect(response.status).toBe(200);
    });
});
```

**Performance Testing**:
```bash
# Load testing
hey -n 1000 -c 10 https://api.new-domain.com/health

# Compare performance metrics
./scripts/compare-performance.sh
```

#### 6.2 Gradual Rollout

**Canary Deployment**:
```yaml
# Deploy new architecture to 10% of traffic
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: api-gateway-canary
  annotations:
    nginx.ingress.kubernetes.io/canary: "true"
    nginx.ingress.kubernetes.io/canary-weight: "10"
spec:
  # Canary ingress configuration
```

**Feature Toggles**:
```go
// Gradual feature rollout
if featureFlags.UseNewArchitecture {
    return handleNewArchitecture(request)
} else {
    return handleLegacyArchitecture(request)
}
```

### Phase 7: Production Cutover (1-2 weeks)

#### 7.1 Final Validation

**Production Readiness Checklist**:
- [ ] All tests passing
- [ ] Performance benchmarks met
- [ ] Monitoring configured
- [ ] Rollback plan documented
- [ ] Team trained on new architecture

**Data Consistency Check**:
```bash
# Verify no data loss
./scripts/verify-data-consistency.sh

# Check API compatibility
./scripts/test-api-compatibility.sh
```

#### 7.2 Cutover Execution

**DNS Cutover**:
```bash
# Update DNS records
# api.your-domain.com -> new load balancer

# Monitor traffic migration
./scripts/monitor-traffic-cutover.sh
```

**Service Decommissioning**:
```bash
# After successful cutover
kubectl delete deployment monolithic-app
kubectl delete service monolithic-service

# Archive old codebase
git tag v1-monolithic-archive
```

### Phase 8: Post-Migration Optimization (Ongoing)

#### 8.1 Performance Optimization

**Identify Bottlenecks**:
```bash
# Monitor new architecture performance
# Compare with baseline metrics
# Optimize slow components
```

**Resource Optimization**:
```yaml
# Adjust resource limits based on actual usage
apiVersion: v1
kind: LimitRange
metadata:
  name: resource-limits
  namespace: arch-separation
spec:
  limits:
  - type: Container
    default:
      cpu: 500m
      memory: 512Mi
    defaultRequest:
      cpu: 100m
      memory: 128Mi
```

#### 8.2 Cost Optimization

**Resource Rightsizing**:
- Monitor actual resource usage
- Adjust Kubernetes resource requests/limits
- Optimize Cloudflare Workers usage

**Storage Optimization**:
- Implement data lifecycle policies
- Archive old data
- Optimize backup strategies

## ⚠️ Risk Mitigation

### Rollback Strategy

**Immediate Rollback**:
```bash
# DNS rollback
# Revert to old load balancer IP

# Feature flag rollback
# Disable new architecture features
```

**Gradual Rollback**:
- Reduce traffic to new architecture
- Monitor for issues
- Complete rollback if needed

### Data Recovery

**Backup Strategy**:
- Daily backups of monolithic database
- Point-in-time recovery capability
- Cross-region backup replication

**Data Validation**:
- Automated data consistency checks
- Manual verification scripts
- Alerting for data anomalies

## 📊 Success Metrics

### Technical Metrics
- **Latency**: <500ms p99 response time
- **Availability**: >99.9% uptime
- **Error Rate**: <1% error rate
- **Resource Usage**: Within 20% of baseline

### Business Metrics
- **Deployment Frequency**: Daily deployments
- **Time to Recovery**: <15 minutes MTTR
- **Development Velocity**: 2x faster feature delivery
- **Cost Efficiency**: 20% cost reduction

## 📚 Resources and Support

### Documentation
- [Architecture Decision Records](./docs/adr/)
- [API Contracts](./packages/api-contracts/)
- [Deployment Guide](./DEPLOYMENT-GUIDE.md)
- [Troubleshooting Guide](./TROUBLESHOOTING-GUIDE.md)

### Training
- Team training on microservices concepts
- Hands-on workshops for new architecture
- Documentation review sessions

### Support
- Migration working group
- 24/7 on-call support during cutover
- Post-migration support team

## 📋 Migration Checklist

### Pre-Migration
- [ ] Architecture assessment complete
- [ ] Migration plan approved
- [ ] Team training completed
- [ ] Infrastructure ready

### During Migration
- [ ] Shared packages created
- [ ] Components extracted
- [ ] Testing comprehensive
- [ ] Data migration successful

### Post-Migration
- [ ] Performance validated
- [ ] Monitoring operational
- [ ] Documentation updated
- [ ] Team feedback collected

### Timeline Estimate: 6-9 months

**Phase 1-2**: 1-2 months (Planning & Foundation)
**Phase 3-4**: 3-4 months (Extraction & Infrastructure)
**Phase 5-6**: 1-2 months (Data & Testing)
**Phase 7-8**: 1-2 months (Cutover & Optimization)

Success depends on thorough planning, comprehensive testing, and gradual rollout with proper rollback capabilities.