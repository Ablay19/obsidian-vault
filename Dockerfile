# --- Go Builder Stage ---
FROM golang:alpine AS builder

# Install build dependencies
RUN apk add --no-cache build-base git

WORKDIR /build

# Copy only go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Install templ
RUN go install github.com/a-h/templ/cmd/templ@latest

COPY . .

# Generate templ files
RUN templ generate ./

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -o telegram-bot ./cmd/bot

# =====================
# Runtime Stage
# =====================
FROM alpine:latest

# Install runtime dependencies
# Tesseract for OCR, Poppler for PDF handling (pdftotext), Git for vault sync, Pandoc for PDF conversion
RUN apk add --no-cache \
    tesseract-ocr \
    poppler-utils \
    ca-certificates \
    tzdata \
    git \
    pandoc \
    curl \
    fontconfig \
    font-noto

# Install Tectonic (required by converter for high-fidelity PDF rendering)
RUN curl --proto '=https' --tlsv1.2 -fsSL https://drop-sh.tectonic-typesetting.org | sh && \
    mv tectonic /usr/local/bin/

# Security: Create a non-root user
RUN addgroup -S appgroup && adduser -S -G appgroup -u 1000 appuser

WORKDIR /app

# Copy application binary from the builder stage
COPY --from=builder /build/telegram-bot .

# Copy configuration and migrations
COPY config.yml .
COPY internal/database/migrations/ internal/database/migrations/
COPY internal/dashboard/static/ internal/dashboard/static/

# Ensure necessary directories exist and have correct permissions
RUN mkdir -p attachments vault pdfs data && \
    chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

# Expose the dashboard port
EXPOSE 8080

# Run the application
CMD ["./telegram-bot"]