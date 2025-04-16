# Performance Testing Suite

This directory contains automated performance testing tools for the e-commerce microservices platform.

## Prerequisites

1. Install k6:
   ```bash
   # macOS
   brew install k6

   # Linux
   sudo apt-key adv --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
   echo "deb https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
   sudo apt-get update
   sudo apt-get install k6
   ```

2. Ensure Docker and Docker Compose are installed and running
3. Make sure jq is installed (for JSON processing in reports):
   ```bash
   # macOS
   brew install jq

   # Linux
   sudo apt-get install jq
   ```

## Directory Structure

```
tests/
├── k6/
│   └── scenarios/
│       ├── product_service_test.js   # Product service test scenarios
│       └── order_service_test.js     # Order service test scenarios
├── results/                          # Test results (created during test runs)
│   └── TIMESTAMP/                    # Results organized by timestamp
│       ├── product_service_load_test/
│       ├── product_service_stress_test/
│       ├── product_service_spike_test/
│       ├── order_service_load_test/
│       ├── order_service_stress_test/
│       ├── order_service_async_spike_test/
│       └── summary.md                # Test run summary
└── run_performance_tests.sh          # Test automation script
```

## Running Tests

1. Make the test script executable:
   ```bash
   chmod +x tests/run_performance_tests.sh
   ```

2. Run all tests:
   ```bash
   ./tests/run_performance_tests.sh
   ```

3. Run specific scenarios (modify script as needed):
   ```bash
   k6 run --scenario load_test tests/k6/scenarios/product_service_test.js
   k6 run --scenario stress_test tests/k6/scenarios/order_service_test.js
   ```

## Test Scenarios

### Product Service Tests
1. **Load Test**
   - Ramps up to 50 users over 1 minute
   - Maintains 50 users for 3 minutes
   - Ramps down over 1 minute

2. **Stress Test**
   - Starts with 50 users
   - Ramps up to 200 users in stages
   - Maintains heavy load for 5 minutes
   - Ramps down over 2 minutes

3. **Spike Test**
   - Sudden spike to 500 users
   - Maintains spike for 30 seconds
   - Quick ramp down

### Order Service Tests
1. **Load Test**
   - Similar to Product Service load test
   - Tests both sync and async order creation

2. **Stress Test**
   - Tests system under sustained heavy load
   - Focuses on order processing capacity

3. **Async Spike Test**
   - Tests async order processing
   - Spikes to 1000 concurrent users
   - Validates message broker performance

## Metrics Collected

1. **HTTP Metrics**
   - Request duration (avg, p90, p95)
   - Request rate
   - Error rate
   - Success/failure counts

2. **Custom Metrics**
   - Successful creates/updates
   - Failed operations
   - Async operation success rate

3. **System Metrics**
   - Docker container stats
   - CPU usage
   - Memory usage
   - Network I/O
   - Disk I/O

## Results Analysis

Test results are saved in the `results/` directory, organized by timestamp. Each test run includes:

1. **Summary Report** (`summary.md`)
   - Overview of test run
   - Key metrics summary
   - System resource usage

2. **Detailed Metrics**
   - JSON metrics file
   - CSV metrics file
   - Docker stats
   - Service logs

## Interpreting Results

1. **Performance Thresholds**
   - 95% of requests should complete within:
     - Product Service: 500ms
     - Order Service: 1000ms
   - Error rate should be below 1%

2. **Warning Signs**
   - Response times exceeding thresholds
   - Error rates above 1%
   - Memory usage approaching limits
   - High CPU utilization

3. **Success Criteria**
   - All scenarios complete without errors
   - Response times within thresholds
   - System resources stable
   - No service crashes

## Troubleshooting

1. **Common Issues**
   - Services not running: Check `docker-compose ps`
   - k6 not installed: Check installation
   - Permission denied: Run `chmod +x` on script

2. **Debugging**
   - Check service logs in results directory
   - Monitor Docker stats during tests
   - Review error messages in k6 output

## Contributing

1. Add new test scenarios in `k6/scenarios/`
2. Update thresholds as needed
3. Add custom metrics for new features
4. Improve reporting and analysis tools 