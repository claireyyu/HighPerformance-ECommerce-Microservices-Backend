FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install curl for healthcheck
RUN apk add --no-cache curl

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go build -o main ./cmd/api

EXPOSE 8081

CMD ["./main"] 