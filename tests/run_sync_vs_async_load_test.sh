#!/bin/bash

# Create results directory if it doesn't exist
mkdir -p tests/results/load_test

# Ensure services are running
echo "Ensuring all services are up..."
docker compose up -d

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 10

# Run the load test
echo "Running load test (multi-stage)..."
k6 run tests/sync_vs_async_load_test.js

# Print the location of the results
echo "Load test results are available in:"
echo "- tests/results/load_test/summary.html"
echo "- tests/results/load_test/summary.txt"