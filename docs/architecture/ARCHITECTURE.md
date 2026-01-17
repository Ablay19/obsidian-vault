# Architecture Documentation

This document describes the high-level architecture of the Obsidian Vault system.

## System Overview

Obsidian Vault is a multi-service platform providing:
- WhatsApp Business API integration
- AI-powered automation and chat
- Mathematical visualization generation (Manim)
- Knowledge base and RAG capabilities
- Real-time processing with Cloudflare Workers

## Architecture Diagram

```mermaid
graph TB
    subgraph "User Layer"
        WA[WhatsApp Users] --> WAB[WhatsApp Cloud API]
        TG[Telegram Users] --> TGB[Telegram Bot API]
        WEB[Web Clients] --> LB[Load Balancer]
    end

    subgraph "Edge Layer"
        LB --> CF[Cloudflare CDN]
        WAB --> WH[Webhook Handler]
        TGB --> TH[Telegram Handler]
        CF --> W[Cloudflare Workers]
    end

    subgraph "Application Layer"
        W --> API[API Gateway]
        API --> AS[AI Service]
        API --> WS[WhatsApp Service]
        API --> VS[Visualizer Service]
        API --> RS[RAG Service]
        API --> MS[Manim Service]
    end

    subgraph "AI Services"
        AS --> OA[OpenAI]
        AS --> AN[Anthropic]
        AS --> GM[Google Gemini]
        AS --> DF[DeepSeek]
        AS --> GR[Groq]
        AS --> HF[HuggingFace]
        AS --> CFW[Cloudflare AI]
    end

    subgraph "Storage Layer"
        WS --> KV[KV Namespace]
        VS --> R2[R2 Storage]
        MS --> R2
        RS --> VS[Vector Store]
        AS --> RC[Redis Cache]
        API --> PG[(PostgreSQL)]
    end

    subgraph "Rendering Layer"
        MS --> MR[Manim Renderer]
        MR --> R2
        MR --> WH
    end
```

---

## Component Architecture

### 1. Cloudflare Workers (Edge Computing)

**Purpose**: Handle incoming requests at the edge for low latency

**Workers**
- `ai-manim-worker`: Main worker for AI video generation
- `worker-whatsapp`: WhatsApp webhook processing
- `worker-ai`: General AI request handling

**Characteristics**
- Serverless execution
- Global distribution
- Automatic scaling
- 10ms cold start typical

**Configuration**
```toml
name = "ai-manim-worker"
main = "src/index.ts"
compatibility_date = "2024-09-23"
compatibility_flags = ["nodejs_compat"]

[vars]
LOG_LEVEL = "info"

[[kv_namespaces]]
binding = "SESSIONS"
id = "your-kv-id"
```

---

### 2. API Gateway

**Purpose**: Central entry point for all API requests

**Responsibilities**
- Request routing
- Authentication
- Rate limiting
- Request validation
- Response formatting

**Flow**
```
Client Request
    ↓
Authentication Check
    ↓
Rate Limit Check
    ↓
Route to Service
    ↓
Process Request
    ↓
Format Response
    ↓
Return to Client
```

---

### 3. Services Architecture

#### WhatsApp Service

```
┌─────────────────┐
│  Webhook Handler│
└────────┬────────┘
         ↓
┌─────────────────┐
│  Message Parser │
└────────┬────────┘
         ↓
┌─────────────────┐
│  Pipeline       │────→ AI Processing
│  Orchestrator   │
└────────┬────────┘
         ↓
┌─────────────────┐
│  Response       │
│  Sender         │
└─────────────────┘
```

**Key Components**
- Webhook verification
- Message parsing (text, media, buttons)
- Session management
- Template message handling

#### AI Service

```
┌─────────────────┐
│  Request Router │
└────────┬────────┘
         ↓
┌─────────────────┐
│  Provider       │
│  Selector       │
└────────┬────────┘
         ↓
┌─────────────────┐
│  Fallback Chain │
│  ├─ OpenAI      │
│  ├─ Anthropic   │
│  └─ ...         │
└────────┬────────┘
         ↓
┌─────────────────┐
│  Response       │
│  Validator      │
└─────────────────┘
```

**Features**
- Automatic provider fallback
- Cost tracking per user
- Request caching
- Token usage monitoring

#### Manim Service

```
┌─────────────────┐
│  Problem Input  │
└────────┬────────┘
         ↓
┌─────────────────┐
│  AI Code Gen    │
└────────┬────────┘
         ↓
┌─────────────────┐
│  Code Validator │
└────────┬────────┘
         ↓
┌─────────────────┐
│  Renderer       │
│  (Docker/Cloud) │
└────────┬────────┘
         ↓
┌─────────────────┐
│  R2 Upload      │
└────────┬────────┘
         ↓
┌─────────────────┐
│  Presigned URL  │
│  Generation     │
└─────────────────┘
```

---

## Data Flow Diagrams

### WhatsApp Message Flow

```mermaid
sequenceDiagram
    participant U as User
    participant W as WhatsApp
    participant WH as Webhook Worker
    participant KV as KV Store
    participant AI as AI Service
    participant R as Response

    U->>W: Send message
    W->>WH: POST webhook
    WH->>WH: Verify signature
    WH->>KV: Create/Update session
    WH->>AI: Process message
    AI->>WH: Generate response
    WH->>W: Send message via API
    W->>U: Deliver message
```

### Manim Video Generation Flow

```mermaid
sequenceDiagram
    participant U as User
    participant T as Telegram Bot
    participant W as Worker
    participant AI as AI Service
    participant R as Renderer
    participant R2 as R2 Storage
    participant U2 as User

    U->>T: /video Explain derivative
    T->>W: Webhook
    W->>W: Create job
    W->>AI: Generate Manim code
    AI->>W: Return code
    W->>R: Submit render job
    R->>R: Render animation
    R->>R2: Upload video
    R->>W: Callback complete
    W->>T: Send video link
    T->>U: Video ready
    U->>U2: Click link
    U2->>W: Access video
    W->>R2: Get presigned URL
    R2->>U2: Stream video
    W->>R2: Delete video
```

---

## Storage Architecture

### Cloudflare KV

**Purpose**: Session storage and job state management

**Structure**
```
sessions:{session_id} → JSON (UserSession)
jobs:{job_id} → JSON (ProcessingJob)
sessions:by_chat:{chat_id} → session_id (index)
```

**TTL Strategy**
- Active sessions: 7 days
- Queued jobs: 24 hours
- Processing jobs: 1 hour
- Ready videos: 5 minutes
- Failed jobs: 1 hour

### Cloudflare R2

**Purpose**: Media storage (videos, images, documents)

**Structure**
```
videos/{job_id}.mp4
images/{media_id}
documents/{doc_id}
```

**Lifecycle**
- Videos: Auto-delete after 24 hours
- Images: Auto-delete after 7 days
- Documents: Auto-delete after 30 days

### Vector Store (Memory/SQLite)

**Purpose**: RAG document embeddings

**Schema**
```sql
CREATE TABLE vector_documents (
    id TEXT PRIMARY KEY,
    content TEXT NOT NULL,
    metadata TEXT,
    vector TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_content ON vector_documents(content);
```

---

## Security Architecture

### Authentication Layers

```
┌─────────────────────────────────┐
│  Layer 1: Edge Security         │
│  - WAF rules                    │
│  - DDoS protection              │
│  - Bot detection                │
└─────────────────────────────────┘
         ↓
┌─────────────────────────────────┐
│  Layer 2: API Authentication    │
│  - Bearer token validation      │
│  - API key verification         │
│  - JWT verification             │
└─────────────────────────────────┘
         ↓
┌─────────────────────────────────┐
│  Layer 3: Service Authorization │
│  - RBAC permissions             │
│  - Resource ownership           │
│  - Rate limits per user         │
└─────────────────────────────────┘
```

### Secrets Management

**Strategy**: Doppler + Environment Variables

```
┌──────────────┐     ┌──────────────┐     ┌──────────────┐
│  Doppler     │────▶│  Environment │────▶│  Application│
│  Dashboard   │     │  Variables   │     │  Code        │
└──────────────┘     └──────────────┘     └──────────────┘
```

**Secrets Categories**
- API Keys (external services)
- Database credentials
- Encryption keys
- JWT secrets

---

## Scaling Strategy

### Horizontal Scaling

**Cloudflare Workers**
- Automatic scaling per request
- No configuration needed
- Global distribution

**Stateless Services**
- All state in KV/R2/database
- No sticky sessions
- Load balanced globally

### Vertical Scaling

**Database**
- Connection pooling
- Read replicas for queries
- Connection limits per service

**AI Services**
- Request queuing
- Provider load balancing
- Cache expensive operations

---

## Monitoring Architecture

### Metrics Collection

```mermaid
graph LR
    App[Application] --> M[Metrics]
    M --> P[Prometheus]
    P --> G[Grafana]
    
    App --> L[Logs]
    L --> Loki[Loki]
    Loki --> G
    
    App --> T[Traces]
    T --> Jaeger[Jaeger]
    Jaeger --> G
```

### Key Metrics

**Application Metrics**
- Request latency (p50, p95, p99)
- Request rate
- Error rate
- Active sessions

**Business Metrics**
- AI token usage
- Video renders per day
- Message throughput
- Cost per user

**Infrastructure Metrics**
- KV operation latency
- R2 upload/download speed
- AI provider response times
- Cache hit rate

---

## Deployment Architecture

### Environments

```
┌─────────────────────────────────────────────────────────┐
│  Development                                           │
│  - Local machine                                       │
│  - Local databases                                     │
│  - Debug logging enabled                               │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│  Staging                                               │
│  - Cloudflare preview deployments                      │
│  - Staging data stores                                 │
│  - Full feature parity                                 │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│  Production                                            │
│  - Live Cloudflare workers                             │
│  - Production data stores                              │
│  - Optimized performance                               │
└─────────────────────────────────────────────────────────┘
```

### CI/CD Pipeline

```
Git Push
    ↓
GitHub Actions
    ├─ Lint & Type Check
    ├─ Unit Tests
    ├─ Build Workers
    └─ Integration Tests
            ↓
Deploy Preview (Staging)
    ↓
Manual Approval
    ↓
Deploy to Production
```

---

## Cost Architecture

### Cost Breakdown

| Component | Cost Factor | Estimated Monthly |
|-----------|-------------|-------------------|
| Cloudflare Workers | Requests × Compute | $50-200 |
| Cloudflare R2 | Storage + Egress | $20-100 |
| Cloudflare KV | Reads/Writes | $10-50 |
| AI Providers | Token usage | $100-500 |
| Database | Compute + Storage | $50-200 |

### Cost Optimization Strategies

1. **Caching**
   - Cache AI responses
   - Cache RAG queries
   - CDN caching for static assets

2. **Request Batching**
   - Batch RAG queries
   - Combine database operations

3. **Provider Selection**
   - Use cheaper providers for simple tasks
   - Implement fallback to cost-effective options

4. **Resource Limits**
   - Maximum video length
   - Maximum prompt size
   - Rate limits per user
