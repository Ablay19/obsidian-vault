FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api-gateway ./apps/api-gateway/cmd/main.go

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/api-gateway .
COPY --from=builder /app/packages ./packages

EXPOSE 8080

ENV LOG_LEVEL=info
ENV PORT=8080

CMD ["./api-gateway"]