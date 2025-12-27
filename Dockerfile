FROM golang:alpine AS builder

RUN apk add --no-cache git build-base tesseract-ocr-dev leptonica-dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o telegram-bot .

FROM alpine:latest

RUN apk add --no-cache \
    tesseract-ocr \
    tesseract-ocr-data-eng \
    tesseract-ocr-data-fra \
    tesseract-ocr-data-ara \
    poppler-utils \
    ca-certificates \
    tzdata

RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

COPY --from=builder /build/telegram-bot .

RUN mkdir -p attachments vault/Inbox vault/Attachments && \
    chown -R appuser:appuser /app

USER appuser

EXPOSE 8080

CMD ["./telegram-bot"]
