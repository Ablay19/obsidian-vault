# --- Go Builder Stage ---
FROM golang:alpine AS builder

# Install build dependencies
RUN apk add --no-cache build-base

WORKDIR /build

# Copy only go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Install golangci-lint
RUN go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Copy the rest of the source code
COPY . .

# Generate Go code from Templ files
RUN go install github.com/a-h/templ/cmd/templ@latest
RUN templ generate ./...

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o telegram-bot ./cmd/bot

# =====================
# Runtime Stage
# =====================
FROM alpine:latest

# Install runtime dependencies
# Tesseract for OCR, Poppler for PDF handling, and standard tools
RUN apk add --no-cache tesseract-ocr poppler-utils ca-certificates tzdata

# Security: Create a non-root user
RUN addgroup -S appgroup && adduser -S -G appgroup -u 1000 appuser

WORKDIR /app

# Copy application binary from the builder stage
COPY --from=builder /build/telegram-bot .

# Copy configuration and database schema
COPY config.yml .
COPY internal/database/schema.sql internal/database/schema.sql
COPY internal/database/migrations/ internal/database/migrations/

# Set correct ownership for all application files
RUN chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

# Expose the dashboard port
EXPOSE 8080

# Run the application
CMD ["./telegram-bot"]
