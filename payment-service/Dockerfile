FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY payment-service/go.mod payment-service/go.sum* ./payment-service/
COPY common_library ./common_library
COPY proto-contracts ./proto-contracts
COPY payment-service ./payment-service

WORKDIR /app/payment-service

RUN go mod download
RUN go build -o payment-service ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/payment-service/payment-service .
COPY --from=builder /app/payment-service/migrations ./migrations

CMD ["./payment-service"]

