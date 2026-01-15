# Manim Renderer Deployment Options

## Option 1: Railway (Recommended)

### Quick Deploy
```bash
# Install Railway CLI
npm install -g @railway/cli

# Login
railway login

# Initialize project (select "Empty Project" and link to this repo)
railway init

# Set environment variables
railway variables set PORT=8080
railway variables set MANIM_QUALITY=medium

# Deploy
railway up

# Get URL
railway domain
```

### Manual Deploy with Docker
```bash
# Build and push to Railway registry
railway login
railway init
railway build -t .
railway push
```

## Option 2: Fly.io

```bash
# Install Fly CLI
curl -L https://fly.io/install.sh | sh

# Launch
fly launch --image manim-renderer:latest

# Set secrets
fly secrets set PORT=8080 MANIM_QUALITY=medium
```

## Option 3: Render.com

1. Create new Web Service on Render.com
2. Connect GitHub repository
3. Build command: `pip install -r requirements.txt`
4. Start command: `gunicorn src.app:app --bind 0.0.0.0:$PORT --workers 2 --timeout 300`
5. Set environment variables in Render dashboard

## Option 4: Local Testing

```bash
# Build minimal image (without LaTeX)
docker build -f Dockerfile.minimal -t manim-renderer:minimal .

# Run
docker run -p 8080:8080 -e PORT=8080 manim-renderer:minimal

# Test
curl http://localhost:8080/health
```

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| PORT | Yes | 8080 | HTTP port |
| MANIM_QUALITY | No | medium | Video quality (low/medium/high) |
| MANIM_COLOR | No | BLUE | Default Manim color |
| MANIM_PREVIEW | No | false | Preview mode |

## Testing

```bash
# Health check
curl http://localhost:8080/health

# Submit render job
curl -X POST http://localhost:8080/render \
  -H "Content-Type: application/json" \
  -d '{
    "job_id": "test-123",
    "code": "from manim import *\nclass Scene(Scene):\n    pass",
    "problem": "Test",
    "output_format": "mp4",
    "quality": "medium"
  }'

# Check status
curl http://localhost:8080/status/test-123
```

## Production Notes

- Manim requires ~2GB RAM for rendering
- Initial container boot takes ~30 seconds (LaTeX compilation if enabled)
- Consider using a pre-built image from Docker Hub for faster deployment
- Enable health checks for container orchestration
