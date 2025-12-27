FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY order-service/go.mod order-service/go.sum* ./order-service/
COPY common_library ./common_library
COPY proto-contracts ./proto-contracts
COPY order-service ./order-service

WORKDIR /app/order-service

RUN go mod download
RUN go build -o order-service ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/order-service/order-service .
COPY --from=builder /app/order-service/migrations ./migrations

CMD ["./order-service"]

