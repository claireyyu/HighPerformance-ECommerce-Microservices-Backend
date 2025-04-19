package main

import (
	"log"

	"github.com/ecommerce-platform/order-service/internal/config"
	"github.com/ecommerce-platform/order-service/internal/messaging/kafka"
	"github.com/ecommerce-platform/order-service/internal/messaging/rabbitmq"
	"github.com/ecommerce-platform/order-service/internal/repository"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to DB
	orderRepo, err := repository.NewOrderRepository(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	// Start consumers
	log.Println("Starting Kafka and RabbitMQ consumers...")
	go kafka.StartKafkaConsumer(cfg.Kafka, orderRepo)
	go rabbitmq.StartRabbitMQConsumer(cfg.RabbitMQ, orderRepo)

	// Keep running
	select {}
}
