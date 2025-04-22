# E-Commerce Microservices Backend

This is a high-performance e-commerce microservices backend system that consists of two main services: Product Service and Order Service.

## Architecture

The system consists of the following components:

- **Product Service** (Port 8080)
  - Manages product information
  - Exposes REST API endpoints for product operations
  - Uses MySQL for data persistence
  - Publishes product events to Kafka

- **Order Service** (Port 8081)
  - Manages order information
  - Exposes REST API endpoints for order operations
  - Uses MySQL for data persistence
  - Supports both synchronous and asynchronous order creation
  - Uses Kafka and RabbitMQ for asynchronous order processing

- **MySQL** (Port 3306)
  - Primary database for both services
  - Stores product and order information

- **Kafka** (Port 9092)
  - Message broker for asynchronous event processing
  - Used for product and order events

- **RabbitMQ** (Port 5672)
  - Alternative message broker for asynchronous order processing
  - Management interface available at port 15672

## API Endpoints

### Product Service (http://localhost:8080)

- `GET /products` - Get all products
- `POST /products` - Create a new product
  ```json
  {
    "name": "Product Name",
    "description": "Product Description",
    "price": 99.99
  }
  ```

### Order Service (http://localhost:8081)

- `GET /orders` - Get all orders
- `POST /orders/sync` - Create a new order synchronously
  ```json
  {
    "user_id": 1,
    "product_id": 1,
    "quantity": 2
  }
  ```
- `POST /orders/async/kafka` - Create a new order asynchronously using Kafka
- `POST /orders/async/rabbitmq` - Create a new order asynchronously using RabbitMQ

## Getting Started

1. Start the services:
   ```bash
   docker compose up -d
   ```

2. Initialize the database:
   ```bash
   cd db && ./init.sh
   ```

3. The services will be available at:
   - Product Service: http://localhost:8080
   - Order Service: http://localhost:8081
   - RabbitMQ Management: http://localhost:15672
   - Kafka: localhost:9092
   - MySQL: localhost:3306

## Development

- The project uses Go 1.22
- Dependencies are managed using Go modules
- Configuration is managed through environment variables and config.yaml
- Docker and Docker Compose are used for containerization and orchestration 