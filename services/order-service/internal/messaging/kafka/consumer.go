package kafka

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"os/signal"

	"github.com/IBM/sarama"
	"github.com/ecommerce-platform/order-service/internal/config"
	"github.com/ecommerce-platform/order-service/internal/models"
	"github.com/ecommerce-platform/order-service/internal/repository"
)

type OrderConsumer struct {
	repo repository.OrderRepository
}

func (consumer *OrderConsumer) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (consumer *OrderConsumer) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (consumer *OrderConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var order models.Order
		if err := json.Unmarshal(msg.Value, &order); err != nil {
			log.Printf("❌ Invalid Kafka message: %v", err)
			continue
		}

		err := consumer.repo.Create(context.Background(), &order)
		if err != nil {
			log.Printf("❌ Failed to insert order from Kafka: %v", err)
		} else {
			log.Printf("✅ Order inserted from Kafka: ID=%d", order.ID)
		}

		session.MarkMessage(msg, "") // Acknowledge message
	}
	return nil
}

func StartKafkaConsumer(cfg config.KafkaConfig, orderRepo repository.OrderRepository) {
	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	group, err := sarama.NewConsumerGroup(cfg.Brokers, "order-consumer-group", config)
	if err != nil {
		log.Fatalf("❌ Failed to create Kafka consumer group: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	orderConsumer := &OrderConsumer{repo: orderRepo}

	// Graceful shutdown on Ctrl+C
	go func() {
		sigchan := make(chan os.Signal, 1)
		signal.Notify(sigchan, os.Interrupt)
		<-sigchan
		cancel()
	}()

	log.Println("✅ Kafka consumer group is running...")

	for {
		if err := group.Consume(ctx, []string{cfg.Topic}, orderConsumer); err != nil {
			log.Fatalf("❌ Error consuming messages: %v", err)
		}
		if ctx.Err() != nil {
			log.Println("🛑 Kafka consumer shutting down")
			return
		}
	}
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
