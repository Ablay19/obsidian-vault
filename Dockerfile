FROM golang:alpine AS builder

RUN apk add --no-cache git build-base tesseract-ocr-dev leptonica-dev

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o telegram-bot .

FROM alpine:latest

RUN apk add --no-cache tesseract-ocr tesseract-ocr-data-eng tesseract-ocr-data-fra tesseract-ocr-data-ara poppler-utils ca-certificates tzdata 

USER root

RUN addgroup -S appgroup && adduser -S -G appgroup -u 1000 appuser

WORKDIR /app

COPY --from=builder /build/telegram-bot .

RUN mkdir -p attachments vault/Inbox vault/Attachments && \
    chown -R appuser:appgroup attachments vault/Inbox vault/Attachments telegram-bot

USER appuser

ENV GEMINI_API_KEY ""

EXPOSE 8080

CMD ["./telegram-bot"]
