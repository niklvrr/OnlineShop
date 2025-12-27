FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY api-gateway/go.mod api-gateway/go.sum* ./api-gateway/
COPY common_library ./common_library
COPY proto-contracts ./proto-contracts
COPY api-gateway ./api-gateway

WORKDIR /app/api-gateway

RUN go mod download
RUN go build -o api-gateway ./cmd/server

FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/api-gateway/api-gateway .

CMD ["./api-gateway"]

