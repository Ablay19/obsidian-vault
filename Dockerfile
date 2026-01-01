FROM golang:alpine AS builder

RUN apk add --no-cache build-base tesseract-ocr-dev leptonica-dev

WORKDIR /build

# Copy only go mod files first (better cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 \
    go build -o telegram-bot ./cmd/bot

# =====================
# Runtime stage
# =====================
FROM alpine:latest

# Runtime dependencies only
RUN apk add --no-cache tesseract-ocr tesseract-ocr-data-eng tesseract-ocr-data-fra tesseract-ocr-data-ara poppler-utils ca-certificates tzdata pandoc tar wget && \
    wget -q https://github.com/tectonic-typesetting/tectonic/releases/download/tectonic%400.9.0/tectonic-0.9.0-x86_64-unknown-linux-musl.tar.gz \
    && tar -xzf tectonic-0.9.0-x86_64-unknown-linux-musl.tar.gz \
    && mv tectonic /usr/local/bin/ \
    && rm tectonic-0.9.0-x86_64-unknown-linux-musl.tar.gz \
    && apk del wget tar

# Security: non-root user
RUN addgroup -S appgroup && adduser -S -G appgroup -u 1000 appuser

WORKDIR /app

COPY --from=builder /build/telegram-bot .
COPY config.yml .
COPY internal/database/schema.sql internal/database/schema.sql
COPY internal/database/migrations/ internal/database/migrations/

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080
CMD ["./telegram-bot"]
