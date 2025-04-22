# Service URLs
PRODUCT_SERVICE_URL = "http://localhost:8080"
ORDER_SERVICE_URL = "http://localhost:8081"

# Test Configuration
THREAD_GROUPS = 4
THREADS_PER_GROUP = 100
REQUESTS_PER_THREAD = 10

# Test Data
PRODUCT_ID = 1
USER_ID = 1

# Test Types
TEST_TYPES = {
    "sync": "/orders/sync",
    "kafka": "/orders/async/kafka",
    "rabbitmq": "/orders/async/rabbitmq"
}

# Results Configuration
RESULTS_DIR = "test_results"