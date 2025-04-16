FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/order-service ./cmd/api

FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/order-service .
EXPOSE 8082
CMD ["./order-service"] 