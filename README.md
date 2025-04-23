# High Performance E-Commerce Microservices

## Quick Start

1. Start services:
```bash
docker compose up -d --scale order=3
```

2. Run performance test:
```bash
export BASE_URL=http://<IP-Addr>:8081 && node test.js
```

3. Clean up data:
```bash
# Clear MySQL orders
docker exec -i highperformance-ecommerce-microservices-backend-mysql-1 \
  mysql -uroot -ppassword ecommerce -e "TRUNCATE TABLE orders;"

# Clear RabbitMQ queue
curl -u guest:guest -X DELETE http://localhost:15672/api/queues/%2F/orders/contents

# Reset Kafka topic
docker exec -it highperformance-ecommerce-microservices-backend-kafka-1 kafka-topics \
  --bootstrap-server localhost:9092 --delete --topic orders
docker exec -it highperformance-ecommerce-microservices-backend-kafka-1 kafka-topics \
  --bootstrap-server localhost:9092 --create --topic orders --partitions 3 --replication-factor 1
```

## Queue Comparison

| Feature | Kafka | RabbitMQ |
|---------|-------|----------|
| Message Persistence | Disk | Memory + Disk |
| Message Ordering | Per-partition | Per-queue |
| Message Routing | Topic-based | Exchange-based |
| Consumer Groups | Yes | No |
| Message Acknowledgment | Automatic | Manual/Automatic |
| Message TTL | Yes | Yes |
| Message Size | 1MB default | No limit |
| Throughput | High | Medium |
| Latency | Low | Very Low |
| Scalability | Horizontal | Vertical |
| Use Case | Stream processing | Task queues | 