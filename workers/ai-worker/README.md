# AI Worker

The AI Worker is a Cloudflare Worker for handling AI-related requests at the edge.

## Quick Start

```bash
# Install dependencies
cd workers/ai-worker && npm install

# Run locally
npm run dev

# Deploy to staging
npm run deploy --env staging

# Deploy to production
npm run deploy --env production

# Run tests
npm test
```

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/workers` | List workers via API Gateway |
| GET | `/api/v1/workers/:id` | Get specific worker |

## Configuration

Environment variables:
- `LOG_LEVEL`: Logging level (debug, info, warn, error)
- `API_GATEWAY_URL`: URL of the API Gateway service

## Project Structure

```
workers/ai-worker/
├── src/
│   └── index.ts          # Worker entry point
├── tests/
│   └── index.test.ts     # Unit tests
├── package.json          # npm configuration
└── wrangler.toml         # Cloudflare Workers configuration
```

## Deployment

The worker is deployed to Cloudflare Workers using GitHub Actions:

1. Push to `005-architecture-separation` → Deploys to staging
2. Push to `main` → Deploys to production
