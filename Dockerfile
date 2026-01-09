# ARG for build-time variables
ARG GO_VERSION=1.25.4
ARG ALPINE_VERSION=3.19

# --- Builder Stage ---
FROM golang:${GO_VERSION}-alpine AS builder

# Install build dependencies
RUN apk add --no-cache build-base git

WORKDIR /src

# Copy go.mod and go.sum first to leverage Docker cache
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Generate templ files
RUN go run github.com/a-h/templ/cmd/templ@latest generate

# Build the application
# Use ldflags to embed version information
ARG APP_VERSION=v0.0.0-dev
RUN CGO_ENABLED=0 go build -ldflags="-w -s -X main.version=${APP_VERSION}" -o /app/main ./cmd/bot/


# --- Final Stage ---
FROM alpine:${ALPINE_VERSION}

# Install runtime dependencies
# Tesseract for OCR, Poppler for PDF handling (pdftotext), Git for vault sync
RUN apk add --no-cache \
  tesseract-ocr \
  poppler-utils \
  git \
  ca-certificates \
  tzdata

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

# Set up the application directory
WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/main /app/main

# Copy necessary assets
COPY internal/dashboard/static/ ./internal/dashboard/static

# Create directories for data and ensure correct permissions
RUN mkdir -p attachments vault && \
    chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

# Expose the dashboard port
EXPOSE 8080

# Set the entrypoint
ENTRYPOINT ["/app/main"]
