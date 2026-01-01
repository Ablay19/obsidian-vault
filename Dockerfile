# --- Python Model Converter Stage ---
FROM python:3.9-slim as converter

WORKDIR /convert

# Copy requirements and install dependencies
COPY scripts/requirements.txt /convert/scripts/
RUN pip install --no-cache-dir -r /convert/scripts/requirements.txt

# Copy converter script
COPY scripts/convert_to_onnx.py /convert/scripts/

# Run the conversion script, which will create the /convert/models directory
RUN python /convert/scripts/convert_to_onnx.py --output /convert/models/distilbert-onnx


# --- Go Builder Stage ---
FROM golang:alpine AS builder

RUN apk add --no-cache build-base tesseract-ocr-dev leptonica-dev

WORKDIR /build

# Copy the generated models from the converter stage
COPY --from=converter /convert/models ./models

# Copy only go mod files first (better cache)
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o telegram-bot ./cmd/bot


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

# Copy application files from the builder stage
COPY --from=builder /build/telegram-bot .
COPY --from=builder /build/models ./models
COPY config.yml .
COPY internal/database/schema.sql internal/database/schema.sql
COPY internal/database/migrations/ internal/database/migrations/

# Set correct ownership for all application files
RUN chown -R appuser:appgroup /app

# Switch to the non-root user
USER appuser

EXPOSE 8080
CMD ["./telegram-bot"]
