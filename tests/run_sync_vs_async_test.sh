#!/bin/bash

# Create results directory if it doesn't exist
mkdir -p results

# Ensure services are running
echo "Ensuring all services are up..."
docker compose up -d

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 10

# Run the test
echo "Running sync vs async comparison test..."
k6 run tests/sync_vs_async_test.js

# Print the location of the results
echo "Test results are available in:"
echo "- results/sync_vs_async_summary.html"
echo "- results/sync_vs_async_summary.txt" 