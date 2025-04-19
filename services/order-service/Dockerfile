FROM golang:1.22-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
# Build API (producer)
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/order-service ./cmd/api
# Build consumer
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/order-consumer ./cmd/api/consumer

FROM alpine:latest
WORKDIR /app
# Copy both binaries
COPY --from=builder /app/order-service .
COPY --from=builder /app/order-consumer .
EXPOSE 8082
CMD ["./order-service"] 