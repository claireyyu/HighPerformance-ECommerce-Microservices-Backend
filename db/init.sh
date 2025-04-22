#!/bin/bash

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

# Function to check if MySQL is ready
wait_for_mysql() {
    local max_attempts=30
    local attempt=1
    
    echo "Waiting for MySQL to be ready..."
    while [ $attempt -le $max_attempts ]; do
        if $DOCKER_COMPOSE exec mysql mysqladmin -uroot -ppassword ping &> /dev/null; then
            echo "MySQL is ready!"
            return 0
        fi
        echo "Attempt $attempt/$max_attempts: MySQL is not ready yet..."
        sleep 2
        attempt=$((attempt + 1))
    done
    
    echo "Error: MySQL failed to start within the expected time"
    return 1
}

# Function to initialize database
init_database() {
    echo "Initializing database..."
    
    # Wait for MySQL to be ready
    wait_for_mysql
    
    # Execute SQL script
    $DOCKER_COMPOSE exec -T mysql mysql -uroot -ppassword ecommerce < init.sql
    
    if [ $? -eq 0 ]; then
        echo "Database initialized successfully!"
    else
        echo "Error: Failed to initialize database"
        exit 1
    fi
}

# Main script
echo "Starting database initialization..."

# Check if we're in the right directory
if [ ! -f "init.sql" ]; then
    echo "Error: init.sql not found. Make sure you're in the db directory."
    exit 1
fi

# Initialize database
init_database 