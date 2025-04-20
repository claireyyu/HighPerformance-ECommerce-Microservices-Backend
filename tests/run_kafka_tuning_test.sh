#!/bin/bash

# Create results directory if it doesn't exist
mkdir -p tests/results/kafka_tuning

# Rebuild order-service (include latest producer/consumer tuning)
echo "Rebuilding order-service with latest Kafka tuning code..."
docker compose build order-service

# Restart the updated order-service container
echo "Restarting order-service..."
docker compose up -d order-service

# Wait for services to be ready (you can also curl health endpoint instead)
echo "Waiting for services to be ready..."
sleep 10

# Run the Kafka tuning test
echo "Running Kafka tuning test..."
k6 run tests/kafka_tuning_test.js

# Print the location of the results
echo ""
echo "Kafka tuning results are available in:"
echo "- tests/results/kafka_tuning/summary.html"
echo "- tests/results/kafka_tuning/summary.txt"