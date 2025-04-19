package kafka

import (
	"context"
	"encoding/json"
	"log"

	"github.com/IBM/sarama"
	"github.com/ecommerce-platform/order-service/internal/config"
	"github.com/ecommerce-platform/order-service/internal/models"
	"github.com/ecommerce-platform/order-service/internal/repository"
)

func StartKafkaConsumer(cfg config.KafkaConfig, orderRepo repository.OrderRepository) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	client, err := sarama.NewConsumer(cfg.Brokers, config)
	if err != nil {
		log.Fatalf("Failed to start Kafka consumer: %v", err)
	}

	partitionConsumer, err := client.ConsumePartition(cfg.Topic, 0, sarama.OffsetNewest)
	if err != nil {
		log.Fatalf("Failed to consume Kafka partition: %v", err)
	}

	log.Println("✅ Kafka consumer running...")

	go func() {
		for msg := range partitionConsumer.Messages() {
			var order models.Order
			if err := json.Unmarshal(msg.Value, &order); err != nil {
				log.Printf("Invalid Kafka message: %v", err)
				continue
			}

			err := orderRepo.Create(context.Background(), &order)
			if err != nil {
				log.Printf("❌ Failed to insert order from Kafka: %v", err)
			} else {
				log.Printf("✅ Order inserted from Kafka: ID=%d", order.ID)
			}
		}
	}()
}
