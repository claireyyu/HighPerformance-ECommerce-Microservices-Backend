#!/bin/bash

# Create results directory if it doesn't exist
mkdir -p tests/results/sync_vs_async

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
echo "Sync vs Async results are available in:"
echo "- tests/results/sync_vs_async/summary.html"
echo "- tests/results/sync_vs_async/summary.txt"