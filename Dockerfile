FROM golang:alpine AS builder

RUN apk add --no-cache git build-base tesseract-ocr-dev leptonica-dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o telegram-bot .

FROM alpine:latest

RUN apk add --no-cache tesseract-ocr tesseract-ocr-data-eng tesseract-ocr-data-fra tesseract-ocr-data-ara poppler-utils ca-certificates tzdata pandoc wget && \
    wget https://github.com/tectonic-typesetting/tectonic/releases/download/tectonic%400.9.0/tectonic-0.9.0-x86_64-unknown-linux-musl.tar.gz && \
    tar -xzf tectonic-0.9.0-x86_64-unknown-linux-musl.tar.gz && \
    mv tectonic /usr/local/bin/ && \
    rm tectonic-0.9.0-x86_64-unknown-linux-musl.tar.gz

USER root

RUN addgroup -S appgroup && adduser -S -G appgroup -u 1000 appuser

WORKDIR /app

COPY --from=builder /build/telegram-bot .

RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8080

CMD ["./telegram-bot"]
