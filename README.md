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
│   │   ├── cmd/
│   │   ├── internal/
│   │   ├── Dockerfile
│   │   └── go.mod
│   ├── api-gateway/
│   │   ├── config/
│   │   └── Dockerfile
│   ├── infrastructure/
│   │   ├── kafka/
│   │   └── database/
│   ├── performance-tests/
│   │   ├── jmeter/
│   │   ├── scripts/
│   │   └── results/
│   └── docs/
│       ├── architecture/
│       ├── api/
│       └── performance/
```

## Performance Testing

Detailed performance testing plans and results are available in the [Performance Testing Documentation](docs/performance/README.md).

## AWS Deployment

Detailed AWS deployment instructions will be provided in the `/aws/README.md` file.

## Performance Analysis

Performance test results and analysis will be available in the `/docs/performance/` directory after tests are completed.

## Services

### API Gateway (Port: 8080)
The API Gateway serves as the single entry point for all client requests. It routes requests to the appropriate microservices.

**Endpoints:**
- Products:
  - `POST /products` - Create a new product
  - `GET /products/:id` - Get a product by ID
  - `GET /products` - List all products
  - `PUT /products/:id` - Update a product
  - `DELETE /products/:id` - Delete a product

- Orders:
  - `POST /orders` - Create a new order
  - `POST /orders/async` - Create a new order asynchronously
  - `GET /orders/:id` - Get an order by ID
  - `GET /orders` - List all orders (with optional user_id filter)
  - `PUT /orders/:id/status` - Update order status

### Product Service (Port: 8081)
Handles product-related operations and maintains product inventory.

### Order Service (Port: 8082)
Manages order processing and maintains order history.

## Infrastructure

### Database
- MySQL (Port: 3306)
  - Product Database
  - Order Database

### Message Broker
- Kafka (Port: 9092)
  - Product Events Topic
  - Order Events Topic

### Monitoring
- Kafka UI (Port: 8083)
  - Monitor Kafka topics and messages

## Getting Started

1. Clone the repository
2. Create a `.env` file with the required environment variables
3. Run the services:
   ```bash
   docker compose up -d
   ```

## Environment Variables

Create a `.env` file with the following variables:

```env
# API Gateway
API_GATEWAY_PORT=8080
API_GATEWAY_HOST=0.0.0.0

# Product Service
PRODUCT_SERVICE_PORT=8081
PRODUCT_SERVICE_HOST=0.0.0.0
PRODUCT_DB_HOST=mysql
PRODUCT_DB_PORT=3306
PRODUCT_DB_USER=product_user
PRODUCT_DB_PASSWORD=product_password
PRODUCT_DB_NAME=product_db

# Order Service
ORDER_SERVICE_PORT=8082
ORDER_SERVICE_HOST=0.0.0.0
ORDER_DB_HOST=mysql
ORDER_DB_PORT=3306
ORDER_DB_USER=order_user
ORDER_DB_PASSWORD=order_password
ORDER_DB_NAME=order_db

# Kafka
KAFKA_BROKERS=kafka:9092
KAFKA_PRODUCT_TOPIC=product-events
KAFKA_ORDER_TOPIC=order-events

# MySQL
MYSQL_ROOT_PASSWORD=root_password
MYSQL_DATABASE=product_db
MYSQL_USER=product_user
MYSQL_PASSWORD=product_password
MYSQL_MULTIPLE_DATABASES=product_db,order_db
```

## API Documentation

### Product API

#### Create Product
```http
POST /products
Content-Type: application/json

{
  "name": "Product Name",
  "description": "Product Description",
  "price": 99.99,
  "stock": 100
}
```

#### Get Product
```http
GET /products/:id
```

#### List Products
```http
GET /products
```

#### Update Product
```http
PUT /products/:id
Content-Type: application/json

{
  "name": "Updated Product Name",
  "description": "Updated Description",
  "price": 149.99,
  "stock": 50
}
```

#### Delete Product
```http
DELETE /products/:id
```

### Order API

#### Create Order
```http
POST /orders
Content-Type: application/json

{
  "user_id": 1,
  "items": [
    {
      "product_id": 1,
      "quantity": 2
    }
  ]
}
```

#### Create Order Async
```http
POST /orders/async
Content-Type: application/json

{
  "user_id": 1,
  "items": [
    {
      "product_id": 1,
      "quantity": 2
    }
  ]
}
```

#### Get Order
```http
GET /orders/:id
```

#### List Orders
```http
GET /orders
GET /orders?user_id=1
```

#### Update Order Status
```http
PUT /orders/:id/status
Content-Type: application/json

{
  "status": "paid"
}
```

## Development

### Prerequisites
- Docker
- Docker Compose
- Go 1.21 or later

### Building Services
```bash
# Build all services
docker compose build

# Build specific service
docker compose build <service-name>
```

### Running Services
```bash
# Start all services
docker compose up -d

# Start specific service
docker compose up -d <service-name>
```

### Viewing Logs
```bash
# View all logs
docker compose logs -f

# View specific service logs
docker compose logs -f <service-name>
```

## Architecture

The system follows a microservices architecture with the following components:

1. **API Gateway**: Single entry point for all client requests
2. **Product Service**: Manages product catalog and inventory
3. **Order Service**: Handles order processing and management
4. **Message Broker**: Facilitates asynchronous communication between services
5. **Database**: Stores product and order data

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request
