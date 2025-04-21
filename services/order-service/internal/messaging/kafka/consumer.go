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

// package kafka

// import (
// 	"context"
// 	"encoding/json"
// 	"log"
// 	"time"

// 	"github.com/IBM/sarama"
// 	"github.com/ecommerce-platform/order-service/internal/config"
// 	"github.com/ecommerce-platform/order-service/internal/models"
// 	"github.com/ecommerce-platform/order-service/internal/repository"
// )

// func StartKafkaConsumer(cfg config.KafkaConfig, orderRepo repository.OrderRepository) {
// 	config := sarama.NewConfig()

// 	// ✅ Tuning for better throughput and lower latency
// 	config.Consumer.Fetch.Min = 1                          // Respond quickly with available data
// 	config.Consumer.Fetch.Default = 1024 * 1024            // 1MB fetch size
// 	config.Consumer.MaxWaitTime = 100 * time.Millisecond   // Broker waits max this before responding
// 	config.Consumer.Return.Errors = true

// 	// 🔌 Create Kafka client
// 	client, err := sarama.NewConsumer(cfg.Brokers, config)
// 	if err != nil {
// 		log.Fatalf("❌ Failed to start Kafka consumer: %v", err)
// 	}

// 	partitions, err := client.Partitions(cfg.Topic)
// 	if err != nil {
// 		log.Fatalf("❌ Failed to get Kafka partitions: %v", err)
// 	}
// 	log.Printf("✅ Kafka consumer listening to %d partitions\n", len(partitions))

// 	// 🔄 Worker pool for concurrent database inserts
// 	const workerCount = 10
// 	jobs := make(chan []byte, 1000)

// 	for i := 0; i < workerCount; i++ {
// 		go func(id int) {
// 			for msg := range jobs {
// 				var order models.Order
// 				if err := json.Unmarshal(msg, &order); err != nil {
// 					continue // skip bad messages silently to reduce log pressure
// 				}
// 				if err := orderRepo.Create(context.Background(), &order); err != nil {
// 					// Optionally log failed DB insertions
// 				}
// 			}
// 		}(i)
// 	}

// 	// 🚀 Spawn a goroutine per partition to feed messages into the worker pool
// 	for _, partition := range partitions {
// 		pc, err := client.ConsumePartition(cfg.Topic, partition, sarama.OffsetNewest)
// 		if err != nil {
// 			log.Fatalf("❌ Failed to consume partition %d: %v", partition, err)
// 		}

// 		go func(pc sarama.PartitionConsumer, p int32) {
// 			for msg := range pc.Messages() {
// 				jobs <- msg.Value
// 			}
// 		}(pc, partition)
// 	}
// }
