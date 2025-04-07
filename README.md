# High-Performance E-Commerce Microservices Platform 

A high-performance, distributed e-commerce platform built with microservices architecture, focusing on asynchronous messaging patterns and performance optimization.

## Project Overview

This project implements a scalable, cloud-native e-commerce platform with a focus on asynchronous processing using Kafka and distributed systems architecture. The platform provides core product catalog and order processing capabilities while demonstrating the performance benefits of event-driven architecture.

### Key Features

- Product catalog management (CRUD operations)
- Order processing system
- Asynchronous event-driven architecture
- Kafka-based message processing
- Comprehensive performance testing and analysis
- AWS cloud deployment
- Horizontal scaling capabilities

## Documentation

- [System Architecture](docs/architecture/README.md)
- [API Documentation](docs/api/README.md)
- [Performance Testing Plan](docs/performance/README.md)

## Quick Start

### Prerequisites

- Docker and Docker Compose
- AWS CLI configured with appropriate permissions
- Go 1.16+
- Kafka tools

### Local Development Setup

1. Clone the repository
   ```
   git clone https://github.com/claireyyu/HighPerformance-ECommerce-Microservices-Backend.git
   cd HighPerformance-ECommerce-Microservices-Backend
   ```

2. Start local development environment
   ```
   docker compose up -d
   ```

3. Access services
   - Product Service: http://localhost:8081
   - Order Service: http://localhost:8082
   - API Gateway: http://localhost:8080
   - Kafka UI: http://localhost:8090

## Project Structure

```
e-commerce-platform/
├── README.md
├── docker-compose.yml
├── aws/
│   ├── cloudformation/
│   ├── ecs/
│   └── monitoring/
├── services/
│   ├── product-service/
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── order-service/
│   │   ├── src/
│   │   ├── pom.xml
│   │   └── Dockerfile
├── api-gateway/
│   ├── config/
│   └── Dockerfile
├── infrastructure/
│   ├── kafka/
│   └── database/
├── performance-tests/
│   ├── jmeter/
│   ├── scripts/
│   └── results/
└── docs/
    ├── architecture/
    ├── api/
    └── performance/
```

## Performance Testing

Detailed performance testing plans and results are available in the [Performance Testing Documentation](docs/performance/README.md).

## AWS Deployment

Detailed AWS deployment instructions will be provided in the `/aws/README.md` file.

## Performance Analysis

Performance test results and analysis will be available in the `/docs/performance/` directory after tests are completed.
