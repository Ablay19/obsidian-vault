# Multi-stage Dockerfile for Enhanced Cloudflare Workers AI Proxy
# Optimized for development and deployment with Wrangler CLI

# ============================================
# Build Arguments
# ============================================
ARG NODE_VERSION=18
ARG ALPINE_VERSION=3.20

# ============================================
# Builder Stage
# ============================================
FROM --platform=$BUILDPLATFORM node:${NODE_VERSION}-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
     git \
     ca-certificates \
     tzdata \
     python3 \
     make \
     g++

# Set working directory
WORKDIR /src

# Copy package files first for better caching
COPY package*.json ./
RUN npm ci --only=production && npm cache clean --force

# Copy source code
COPY . .

# Build the project (if build script exists)
RUN npm run build 2>/dev/null || echo "No build script found, skipping"

# ============================================
# Production Stage
# ============================================
FROM node:${NODE_VERSION}-alpine AS production

# Install runtime dependencies
# - wrangler: Cloudflare Workers CLI
# - curl: health checks and API calls
# - git: version control for deployments
RUN apk add --no-cache \
     wrangler \
     curl \
     git \
     ca-certificates \
     tzdata \
     && rm -rf /var/cache/apk/* \
     && rm -rf /tmp/*

# Create non-root user with proper permissions
RUN addgroup -S -g 1000 appgroup && \
     adduser -S -u 1000 -G appgroup appuser

# Set working directory
WORKDIR /app

# Copy built application from builder
COPY --from=builder /src /app

# Copy package files and install production dependencies
COPY package*.json ./
RUN npm ci --only=production && npm cache clean --force

# Create required directories
RUN mkdir -p \
     logs \
     .wrangler \
     && chown -R appuser:appgroup /app

# Health check script for workers deployment
RUN cat <<'EOF' > /app/healthcheck.sh
#!/bin/sh
set -e

# Check if wrangler is available
if ! command -v wrangler >/dev/null 2>&1; then
     echo "Wrangler CLI not found"
     exit 1
fi

# Check if package.json exists
if [ ! -f "package.json" ]; then
     echo "package.json not found"
     exit 1
fi

# Check if wrangler.toml exists
if [ ! -f "wrangler.toml" ]; then
     echo "wrangler.toml not found"
     exit 1
fi

echo "OK"
exit 0
EOF

RUN chmod +x /app/healthcheck.sh && chown appuser:appgroup /app/healthcheck.sh

# Switch to non-root user
USER appuser

# Expose port for local development (if needed)
EXPOSE 8787

# Environment variables
ENV NODE_ENV=production
ENV ENVIRONMENT=production
ENV LOG_LEVEL=info

# Health check
HEALTHCHECK --interval=60s --timeout=10s --start-period=10s --retries=3 \
     CMD /app/healthcheck.sh

# Default command - deploy workers
CMD ["npm", "run", "deploy"]

# Labels
LABEL maintainer="abdoullah.elvogani@example.com" \
       version="2.0.0" \
       description="Enhanced Cloudflare Workers AI Proxy - High-performance AI proxy with analytics and caching" \
       org.opencontainers.image.source="https://github.com/Ablay19/obsidian-vault" \
       org.opencontainers.image.licenses="MIT"
