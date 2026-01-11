# Multi-stage Dockerfile for Obsidian Bot
# Optimized for production deployment with security best practices

# ============================================
# Build Arguments
# ============================================
ARG GO_VERSION=1.25.4
ARG ALPINE_VERSION=3.20

# ============================================
# Builder Stage
# ============================================
FROM --platform=$BUILDPLATFORM golang:${GO_VERSION}-alpine AS builder

# Install build dependencies
RUN apk add --no-cache \
    build-base \
    git \
    ca-certificates \
    tzdata

# Set working directory
WORKDIR /src

# Copy go.mod and go.sum first for better caching
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy rest of source code
COPY . .

# Generate templ files for dashboard UI
RUN go install github.com/a-h/templ/cmd/templ@v0.3.977
RUN /go/bin/templ generate

# Build arguments
ARG APP_VERSION=v0.0.0-dev
ARG TARGETOS
ARG TARGETARCH

# Build with optimizations
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} \
    go build \
    -ldflags="-w -s -X main.version=${APP_VERSION} -X main.buildTime=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    -tags=netgo \
    -installsuffix netgo \
    -o /app/bot \
    ./cmd/bot

# ============================================
# Production Stage
# ============================================
FROM alpine:${ALPINE_VERSION} AS production

# Install runtime dependencies
# - tesseract-ocr: OCR for image analysis
# - poppler-utils: PDF processing
# - git: vault sync
# - curl: health checks
RUN apk add --no-cache \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    poppler-utils \
    git \
    ca-certificates \
    tzdata \
    curl \
    && rm -rf /var/cache/apk/* \
    && rm -rf /tmp/*

# Create non-root user with proper permissions
RUN addgroup -S -g 1000 appgroup && \
    adduser -S -u 1000 -G appgroup appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bot /app/bot

# Copy dashboard static files
COPY --from=builder /src/internal/dashboard/static/ ./internal/dashboard/static

# Create required directories
RUN mkdir -p \
    attachments \
    vault \
    data \
    logs \
    && chown -R appuser:appgroup /app \
    && chmod 755 /app/bot

# Health check script
COPY --from=builder /src <<'EOF' /app/healthcheck.sh
#!/bin/sh
set -e

# Check if process is running
if ! pgrep -f "bot" > /dev/null 2>&1; then
    echo "Bot process not running"
    exit 1
fi

# Check HTTP endpoint
if ! curl -sf http://localhost:8080/api/services/status > /dev/null 2>&1; then
    echo "Health endpoint not responding"
    exit 1
fi

echo "OK"
exit 0
EOF

RUN chmod +x /app/healthcheck.sh && chown appuser:appgroup /app/healthcheck.sh

# Switch to non-root user
USER appuser

# Expose application port
EXPOSE 8080

# Environment variables
ENV ENVIRONMENT=production
ENV PORT=8080
ENV LOG_LEVEL=INFO
ENV GIN_MODE=release

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD /app/healthcheck.sh

# Entrypoint
ENTRYPOINT ["/app/bot"]

# Labels
LABEL maintainer="abdoullah.elvogani@example.com" \
      version="${APP_VERSION}" \
      description="Obsidian Bot - AI-powered Telegram assistant with Cloudflare integration" \
      org.opencontainers.image.source="https://github.com/Ablay19/obsidian-vault" \
      org.opencontainers.image.licenses="MIT"
