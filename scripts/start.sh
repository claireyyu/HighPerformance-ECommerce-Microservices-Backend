#!/bin/bash

# Exit on error
set -e

# Function to check if Docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        echo "Error: Docker is not running"
        exit 1
    fi
}

# Function to determine which Docker Compose command to use
get_docker_compose_cmd() {
    if command -v docker-compose &> /dev/null; then
        echo "docker-compose"
    elif docker compose version &> /dev/null; then
        echo "docker compose"
    else
        echo "Error: Neither docker-compose nor docker compose is available"
        exit 1
    fi
}

# Get the Docker Compose command
DOCKER_COMPOSE=$(get_docker_compose_cmd)
echo "Using Docker Compose command: $DOCKER_COMPOSE"

# Function to wait for a service to be ready
wait_for_service() {
    local service=$1
    local max_attempts=30
    local attempt=1
    
    echo "Waiting for $service to be ready..."
    while [ $attempt -le $max_attempts ]; do
        if $DOCKER_COMPOSE ps $service | grep -q "Up"; then
            echo "$service is ready!"
            return 0
        fi
        echo "Attempt $attempt/$max_attempts: $service is not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "Error: $service failed to start within the expected time"
    return 1
}

# Main script
echo "Starting E-Commerce Microservices..."

# Check prerequisites
check_docker

# Build and start services
echo "Building and starting services..."
$DOCKER_COMPOSE up --build -d

# Wait for services to be ready in the correct order
echo "Waiting for services to be ready..."
wait_for_service zookeeper || exit 1
wait_for_service mysql || exit 1
wait_for_service kafka || exit 1
wait_for_service rabbitmq || exit 1

# Initialize database
echo "Initializing database..."
if [ ! -f "db/init.sh" ]; then
    echo "Error: db/init.sh not found"
    exit 1
fi

cd db && ./init.sh
if [ $? -ne 0 ]; then
    echo "Error: Database initialization failed"
    exit 1
fi
cd ..

echo "Services are ready!"
echo "Product Service: http://localhost:8080"
echo "Order Service: http://localhost:8081"
echo "RabbitMQ Management: http://localhost:15672"
echo "Kafka: localhost:9092"
echo "MySQL: localhost:3306"

# Show only error logs
echo "Showing error logs (press Ctrl+C to stop)..."
echo "To see all logs, run: $DOCKER_COMPOSE logs -f"
echo "To see logs for a specific service, run: $DOCKER_COMPOSE logs -f [service_name]"
$DOCKER_COMPOSE logs -f --tail=0 | grep -i "error\|exception\|failed\|fatal" 