# Development Deployment Guide

This guide covers setting up a local development environment for Obsidian Vault.

## Prerequisites

### Required Software

| Software | Version | Installation |
|----------|---------|--------------|
| Go | 1.21+ | [go.dev/dl](https://go.dev/dl) |
| Node.js | 18+ | [nodejs.org](https://nodejs.org) |
| Python | 3.11+ | [python.org](https://python.org) |
| PostgreSQL | 15+ | [postgresql.org](https://postgresql.org) |
| Redis | 7+ | [redis.io](https://redis.io) |
| Git | 2.0+ | [git-scm.com](https://git-scm.com) |

### Recommended Tools

| Tool | Purpose | Installation |
|------|---------|--------------|
| Docker | Container runtime | [docker.com](https://docker.com) |
| make | Build automation | Built-in on macOS/Linux |
| curl | HTTP testing | Built-in |
| jq | JSON processing | `brew install jq` |

## Environment Setup

### 1. Clone Repository

```bash
git clone https://github.com/your-org/obsidian-vault.git
cd obsidian-vault
```

### 2. Install Go Dependencies

```bash
# Download dependencies
go mod download

# Verify installation
go version
go list -m all | head -20
```

### 3. Install Node Dependencies

```bash
# Install for each worker
cd workers/ai-manim-worker
npm install

cd ../worker-whatsapp
npm install

cd ../../apps/api-gateway
npm install
```

### 4. Install Python Dependencies

```bash
# For Manim renderer
cd workers/manim-renderer
pip install -r requirements.txt

# Verify Manim installation
manim --version
```

### 5. Database Setup

```bash
# Start PostgreSQL (macOS)
brew services start postgresql@15

# Create database
createdb obsidian_vault

# Run migrations
cd obsidian-vault
go run cmd/migrate/main.go up

# Seed test data (optional)
go run cmd/seed/main.go
```

### 6. Redis Setup

```bash
# Start Redis (macOS)
brew services start redis

# Test connection
redis-cli ping
# Should return: PONG
```

### 7. Cloudflare Workers Setup

```bash
# Install Wrangler CLI
npm install -g wrangler

# Login to Cloudflare
wrangler login

# Verify authentication
wrangler whoami
```

## Configuration

### Environment Variables

Create a `.env` file in the project root:

```bash
# .env

# Database
export DATABASE_URL="postgresql://localhost:5432/obsidian_vault?sslmode=disable"
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=obsidian_vault
export DB_USER=postgres
export DB_PASSWORD=your_password

# Redis
export REDIS_URL="redis://localhost:6379"

# Cloudflare
export CLOUDFLARE_ACCOUNT_ID="your-account-id"
export CLOUDFLARE_API_TOKEN="your-api-token"

# AI Providers
export OPENAI_API_KEY="sk-your-openai-key"
export ANTHROPIC_API_KEY="your-anthropic-key"
export GEMINI_API_KEY="your-gemini-key"
export DEEPSEEK_API_KEY="your-deepseek-key"
export GROQ_API_KEY="your-groq-key"
export HF_TOKEN="your-huggingface-token"

# Telegram
export TELEGRAM_BOT_TOKEN="your-bot-token"
export TELEGRAM_SECRET="your-webhook-secret"

# WhatsApp
export WHATSAPP_APP_ID="your-app-id"
export WHATSAPP_APP_SECRET="your-app-secret"
export WHATSAPP_PHONE_NUMBER_ID="your-phone-id"
export WHATSAPP_VERIFY_TOKEN="your-verify-token"

# R2 Storage
export R2_ENDPOINT="https://your-account.r2.cloudflarestorage.com"
export R2_ACCESS_KEY_ID="your-access-key"
export R2_SECRET_ACCESS_KEY="your-secret-key"
export R2_BUCKET_NAME="obsidian-vault-videos"

# App Config
export LOG_LEVEL="debug"
export ENVIRONMENT="development"
export SESSION_SECRET="your-session-secret-min-32-chars"
export JWT_SECRET="your-jwt-secret-min-32-chars"

# AI Service
export AI_DEFAULT_PROVIDER="openai"
export AI_MAX_TOKENS=4096
export AI_TEMPERATURE=0.7
```

### wrangler.toml Configuration

Create `workers/ai-manim-worker/wrangler.toml`:

```toml
name = "ai-manim-worker-dev"
main = "src/index.ts"
compatibility_date = "2024-09-23"
compatibility_flags = ["nodejs_compat"]

[vars]
LOG_LEVEL = "debug"
ENVIRONMENT = "development"
USE_MOCK_RENDERER = "true"

[[kv_namespaces]]
binding = "SESSIONS"
id = "your-dev-kv-id"

[env.development]
name = "ai-manim-worker-dev"
```

## Running Services

### 1. Start Database and Redis

```bash
# Terminal 1: PostgreSQL
brew services start postgresql@15

# Terminal 2: Redis
brew services start redis
```

### 2. Start API Gateway

```bash
cd apps/api-gateway
make run

# Or directly
go run cmd/main.go
```

### 3. Start Workers (Development Mode)

```bash
# Terminal 3: AI Worker
cd workers/ai-manim-worker
npm run dev

# Terminal 4: WhatsApp Worker
cd workers/worker-whatsapp
npm run dev
```

### 4. Start Manim Renderer (Optional)

```bash
# Terminal 5: Manim Renderer
cd workers/manim-renderer
python src/app.py

# With mock renderer (no Docker needed)
export USE_MOCK_RENDERER=true
python src/app.py
```

### 5. Verify All Services

```bash
# Check API Gateway
curl http://localhost:8080/health

# Check AI Worker
curl http://localhost:8787/health

# Check WhatsApp Worker
curl http://localhost:8788/health

# Check Manim Renderer
curl http://localhost:8080/health
```

## Testing the Setup

### 1. Test Database Connection

```bash
# Run database tests
go test ./internal/database/... -v
```

### 2. Test AI Service

```bash
# Test OpenAI integration
curl -X POST http://localhost:8080/api/v1/ai/chat \
  -H "Content-Type: application/json" \
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello"}]
  }'
```

### 3. Test WhatsApp Webhook

```bash
# Set up local tunnel for webhooks
ngrok http 8080

# Configure webhook in Meta dashboard to use ngrok URL
# https://your-ngrok.ngrok.io/api/v1/whatsapp/webhook

# Send test message from WhatsApp
# Check worker logs for incoming messages
```

### 4. Test Manim Generation

```bash
# Submit a test video generation
curl -X POST http://localhost:8787/api/v1/manim \
  -H "Content-Type: application/json" \
  -d '{
    "problem": "Show the Pythagorean theorem"
  }'

# Check job status
curl http://localhost:8787/api/v1/jobs/{job_id}
```

## Debugging

### View Logs

```bash
# API Gateway logs
tail -f logs/api-gateway.log

# Worker logs (in separate terminals)
npm run dev 2>&1 | grep -E "(ERROR|WARN|INFO)"

# Cloudflare Worker logs
npx wrangler tail --environment development
```

### Use Delve for Go Debugging

```bash
# Install Delve
go install github.com/go-delve/delve/cmd/dlv@latest

# Start API Gateway in debug mode
dlv debug cmd/main.go --listen=:40000 --headless=true --api-version=2

# Connect from VS Code
# Launch configuration: "Go: Connect to server"
```

### Use Node.js Inspector

```bash
# Start worker with inspector
cd workers/ai-manim-worker
npm run dev -- --inspect=9229

# Open Chrome DevTools
# chrome://inspect
```

## Troubleshooting

### Common Issues

#### Port Already in Use
```bash
# Find process using port
lsof -i :8080

# Kill process
kill -9 <PID>
```

#### Database Connection Refused
```bash
# Check PostgreSQL status
brew services list | grep postgresql

# Restart PostgreSQL
brew services restart postgresql@15
```

#### Module Not Found
```bash
# Clear and reinstall dependencies
go clean -modcache
rm -rf node_modules package-lock.json
go mod download
npm install
```

#### TypeScript Compilation Errors
```bash
# Check TypeScript errors
cd workers/ai-manim-worker
npx tsc --noEmit

# Fix type errors before running
```

## Code Style

### Go Formatting

```bash
# Format code
gofmt -w .

# Check formatting
gofmt -d .

# Run linter
golangci-lint run
```

### TypeScript Formatting

```bash
# Format code
npm run format

# Check linting
npm run lint
```

### Pre-commit Hooks

```bash
# Install pre-commit
pip install pre-commit
pre-commit install

# Run manually
pre-commit run --all-files
```

## Development Workflow

### 1. Create Feature Branch

```bash
git checkout -b feature/your-feature-name
```

### 2. Make Changes

Follow the coding standards:
- Go: Effective Go, Standard Go Project Structure
- TypeScript: Airbnb Style Guide
- Python: PEP 8

### 3. Run Tests

```bash
# Unit tests
go test ./... -short
npm test

# Integration tests
go test ./... -tags=integration
npm run test:integration
```

### 4. Commit Changes

```bash
# Stage changes
git add .

# Commit with conventional message
git commit -m "feat: add new feature"

# Push to remote
git push origin feature/your-feature-name
```

### 5. Create Pull Request

- Fill out PR template
- Link related issues
- Request reviews
- Address feedback

## Next Steps

Once development setup is complete:

1. Review [Architecture Documentation](../architecture/ARCHITECTURE.md)
2. Read [API Reference](../api-reference.md)
3. Check [Common Issues](../common-issues.md)
4. Set up [Staging Environment](DEPLOYMENT_STAGING.md)
