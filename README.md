# High Performance E-Commerce Microservices Architecture

## Overview

This project implements a high-performance e-commerce microservices architecture using modern technologies and best practices. The system is designed to handle high-throughput order processing with scalability and reliability.

## Architecture

The system consists of the following components:
- Order Service (Scalable)
- Message Brokers (Kafka & RabbitMQ)
- MySQL Database
- Performance Testing Tools

## Prerequisites

- Docker and Docker Compose
- Node.js (for running tests)
- curl (for API operations)

## Getting Started

### 1. Environment Setup

```bash
# Clone the repository
git clone [repository-url]

# Navigate to project directory
cd high-performance-ecommerce-microservices
```

### 2. Service Deployment

Deploy the services with multiple order service instances for load balancing:

```bash
docker compose up -d --scale order=3
```

### 3. Performance Testing

Configure and run performance tests:

```bash
# Set the base URL for your deployment
export BASE_URL=http://<IP-Addr>:8081

# Execute performance tests
node test.js
```

## System Maintenance

### Data Management

#### Clear Order Data
```bash
# Reset MySQL orders table
docker exec -i highperformance-ecommerce-microservices-backend-mysql-1 \
  mysql -uroot -ppassword ecommerce -e "TRUNCATE TABLE orders;"
```

#### Message Queue Management

##### RabbitMQ
```bash
# Clear RabbitMQ queue
curl -u guest:guest -X DELETE http://localhost:15672/api/queues/%2F/orders/contents
```

##### Kafka
```bash
# Reset Kafka topic
docker exec -it highperformance-ecommerce-microservices-backend-kafka-1 kafka-topics \
  --bootstrap-server localhost:9092 --delete --topic orders
docker exec -it highperformance-ecommerce-microservices-backend-kafka-1 kafka-topics \
  --bootstrap-server localhost:9092 --create --topic orders --partitions 3 --replication-factor 1
```

## Message Broker Comparison

### Feature Comparison Matrix

| Feature | Kafka | RabbitMQ |
|---------|-------|----------|
| Message Persistence | Disk | Memory + Disk |
| Message Ordering | Per-partition | Per-queue |
| Message Routing | Topic-based | Exchange-based |
| Consumer Groups | Yes | No |
| Message Acknowledgment | Automatic | Manual/Automatic |
| Message TTL | Yes | Yes |
| Message Size | 1MB default | No limit |
| Throughput | High | Medium |
| Latency | Low | Very Low |
| Scalability | Horizontal | Vertical |
| Use Case | Stream processing | Task queues |

### Selection Guidelines

- **Choose Kafka when:**
  - High throughput is required
  - Message ordering is critical
  - Stream processing is needed
  - Horizontal scaling is a priority

- **Choose RabbitMQ when:**
  - Low latency is crucial
  - Complex routing patterns are needed
  - Message acknowledgment is important
  - Queue-based processing is sufficient