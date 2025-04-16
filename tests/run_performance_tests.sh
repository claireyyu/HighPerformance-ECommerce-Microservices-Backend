#!/bin/bash

# Create results directory
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULTS_DIR="tests/results/$TIMESTAMP"
mkdir -p "$RESULTS_DIR"

# Function to run k6 test and save results
run_test() {
    local scenario=$1
    local test_file=$2
    local service_name=$3
    
    echo "Running $scenario test for $service_name..."
    
    # Create scenario-specific directory
    local scenario_dir="$RESULTS_DIR/${service_name}_${scenario}"
    mkdir -p "$scenario_dir"
    
    # Run k6 test with JSON output
    k6 run \
        --tag testid="${service_name}_${scenario}" \
        --out json="$scenario_dir/metrics.json" \
        --out csv="$scenario_dir/metrics.csv" \
        --scenario "$scenario" \
        "$test_file"
    
    # Capture Docker stats
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}\t{{.BlockIO}}" \
        > "$scenario_dir/docker_stats.txt"
    
    # Capture service logs
    docker-compose logs "$service_name" > "$scenario_dir/service_logs.txt"
}

# Ensure services are running
echo "Ensuring all services are up..."
docker-compose up -d

# Wait for services to be ready
echo "Waiting for services to be ready..."
sleep 30

# Product Service Tests
echo "Starting Product Service tests..."
for scenario in "load_test" "stress_test" "spike_test"; do
    run_test "$scenario" "tests/k6/scenarios/product_service_test.js" "product-service"
done

# Order Service Tests
echo "Starting Order Service tests..."
for scenario in "load_test" "stress_test" "async_spike_test"; do
    run_test "$scenario" "tests/k6/scenarios/order_service_test.js" "order-service"
done

# Generate summary report
echo "Generating summary report..."
cat > "$RESULTS_DIR/summary.md" << EOF
# Performance Test Results - $(date)

## Environment
- API Gateway: http://localhost:8080
- Product Service: http://localhost:8081
- Order Service: http://localhost:8082

## Test Scenarios
1. Load Test: Gradual ramp-up to normal load
2. Stress Test: Heavy sustained load
3. Spike Test: Sudden burst of traffic

## Results Summary
\`\`\`
$(cat "$RESULTS_DIR"/*/metrics.json | jq -r '.metrics.http_req_duration.avg')
\`\`\`

## Docker Resource Usage
\`\`\`
$(cat "$RESULTS_DIR"/*/docker_stats.txt)
\`\`\`

## Detailed Results
- Full metrics available in CSV format in each scenario directory
- Docker stats and service logs captured for each test
EOF

echo "Performance tests completed. Results available in: $RESULTS_DIR"
echo "View summary report at: $RESULTS_DIR/summary.md" 