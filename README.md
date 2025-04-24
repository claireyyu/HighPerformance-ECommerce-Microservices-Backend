# High Performance E-Commerce Microservices Architecture

## Overview

This project implements a high-performance e-commerce microservices architecture using modern technologies and best practices. The system is designed to handle high-throughput order processing with scalability and reliability.

## Architecture

The system consists of the following components:
- Order Service (Scalable)
- Message Brokers (Kafka & RabbitMQ)
- MySQL Database
- Performance Testing Tools

## Project Structure

```
.
├── performance-test/           # Performance testing tools and scripts
│   ├── public/                # Static files for test results visualization
│   ├── test.js               # Main test script
│   └── package.json          # Node.js dependencies
│
├── queues/                    # Message queue implementations
│   ├── kafka.go             # Kafka producer/consumer implementation
│   └── rabbitmq.go          # RabbitMQ producer/consumer implementation
│
├── services/                  # Core microservices
│   ├── order.go             # Order service implementation
│   └── product.go           # Product service implementation
│
├── main.go                   # Main application entry point
├── go.mod                    # Go module definition
├── go.sum                    # Go module checksums
└── Dockerfile               # Container definition
```

## Architecture Details

### Service Layer
- **Order Service**: Handles order processing with support for both synchronous and asynchronous operations
  - Implements both Kafka and RabbitMQ message queues
  - Supports horizontal scaling
  - Provides order status tracking and management

- **Product Service**: Manages product information and inventory
  - Product CRUD operations
  - Inventory management
  - Product search and filtering

### Message Queue Layer
- **Kafka Implementation**
  - Topic-based message routing
  - Partition-based message ordering
  - Consumer group support for load balancing
  - High throughput message processing

- **RabbitMQ Implementation**
  - Exchange-based message routing
  - Queue-based message ordering
  - Message acknowledgment support
  - Low latency message processing

### Data Layer
- **MySQL Database**
  - Stores order and product information
  - Supports ACID transactions
  - Provides data persistence and reliability

### Performance Testing
- Load testing tools for both synchronous and asynchronous operations
- Real-time performance metrics collection
- Comparative analysis between different message queue implementations
- Visualization of test results

## Deployment Architecture

### Container Orchestration
- **Docker Compose** configuration for local development and testing
- Service health checks for all components
- Automatic service dependency management
- Network isolation using Docker bridge networks

### Service Configuration
- Environment-based configuration using `.env` files
- Service-specific port mapping
- Health check endpoints for monitoring
- Graceful shutdown handling

### Infrastructure Components
- **MySQL 8.0**: Primary data store
- **Kafka**: High-throughput message broker
  - Zookeeper for cluster coordination
  - Configurable partitions and replication
- **RabbitMQ**: Low-latency message broker
  - Management interface for monitoring
  - Configurable queues and exchanges

## Technical Implementation Details

### Message Queue Optimizations

#### Kafka Implementation
- **Producer Optimizations**
  - Idempotent producer configuration
  - Hash-based partitioning using user ID
  - Batch processing with configurable thresholds
  - Snappy compression for message payload
  - Retry mechanism with max attempts

- **Consumer Optimizations**
  - Consumer group with sticky partition assignment
  - Goroutine pool for concurrent message processing
  - Configurable fetch size and wait time
  - Offset management for message tracking
  - Error handling and recovery mechanisms

#### RabbitMQ Implementation
- **Producer Features**
  - Persistent queues for message durability
  - JSON content type for structured messages
  - Channel management for connection pooling
  - Error handling and reconnection logic

- **Consumer Features**
  - Concurrent message processing with worker pools
  - Configurable QoS settings
  - Manual acknowledgment for message reliability
  - Error handling with message requeuing
  - Detailed logging for monitoring

### Performance Considerations

#### Message Processing
- **Throughput Optimization**
  - Batch processing for high-volume scenarios
  - Compression for network efficiency
  - Connection pooling for resource management
  - Concurrent processing with worker pools

- **Reliability Features**
  - Message acknowledgment mechanisms
  - Error handling and recovery
  - Message persistence
  - Idempotent processing

#### Monitoring and Debugging
- **Logging**
  - Structured logging for message processing
  - Error tracking and reporting
  - Performance metrics collection
  - Worker status monitoring

- **Metrics**
  - Message processing rates
  - Error rates and types
  - Processing latency
  - Queue depths and consumer lag

### Configuration Guidelines

#### Kafka Configuration
```yaml
producer:
  required_acks: "all"
  retry_max: 5
  idempotent: true
  compression: "snappy"
  batch:
    size: 500
    frequency: 50ms
    max_messages: 500

consumer:
  group_rebalance_strategy: "sticky"
  rebalance_timeout: 60s
  fetch:
    min: 1
    default: 1MB
  max_wait_time: 100ms
```

#### RabbitMQ Configuration
```yaml
connection:
  url: "amqp://guest:guest@localhost:5672/"
  heartbeat: 30s

consumer:
  concurrency: 10
  prefetch_count: 1
  queue:
    durable: true
    auto_delete: false
```

## Development Guidelines

### Environment Setup
1. Required environment variables:
   ```bash
   PRODUCT_SERVICE_PORT=8080
   ORDER_SERVICE_PORT=8081
   MYSQL_HOST=mysql
   KAFKA_BROKERS=kafka:29092
   RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
   ```

2. Service startup:
   ```bash
   # Product Service
   ./main -service product

   # Order Service
   ./main -service order
   ```

### Best Practices
- Follow Go project layout standards
- Implement proper error handling and logging
- Use health checks for service monitoring
- Maintain proper service isolation
- Follow microservices design principles

## Prerequisites

- Docker and Docker Compose
- Node.js (for running tests)
- curl (for API operations)

## Getting Started

### 1. Environment Setup

```bash
# Clone the repository
git clone https://github.com/claireyyu/HighPerformance-ECommerce-Microservices-Backend.git

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