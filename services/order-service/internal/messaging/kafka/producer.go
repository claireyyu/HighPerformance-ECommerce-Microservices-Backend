package kafka

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/ecommerce-platform/order-service/internal/config"
	"github.com/ecommerce-platform/order-service/internal/models"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

// func NewKafkaProducer(cfg config.KafkaConfig) *KafkaProducer {
// 	config := sarama.NewConfig()
// 	config.Producer.Return.Successes = true
// 	config.Producer.RequiredAcks = sarama.WaitForAll

// 	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
// 	if err != nil {
// 		panic("Failed to start Kafka producer: " + err.Error())
// 	}

// 	return &KafkaProducer{
// 		producer: producer,
// 		topic:    cfg.Topic,
// 	}
// }

// func NewKafkaProducer(cfg config.KafkaConfig) *KafkaProducer {
// 	config := sarama.NewConfig()

// 	// // ✅ Reliability
// 	// config.Producer.Return.Successes = true
// 	// config.Producer.RequiredAcks = sarama.WaitForAll
// 	// config.Producer.Retry.Max = 5

// 	// // ✅ Performance tuning
// 	// config.Producer.Flush.Messages = 10                      // Flush batch after 10 messages
// 	// config.Producer.Flush.Frequency = 500 * time.Millisecond // ... or after 500ms
// 	// config.Producer.Flush.MaxMessages = 1000                 // Optional: upper cap on batch
// 	// config.Producer.Idempotent = true                        // Optional: safer for retrying
// 	// config.Net.MaxOpenRequests = 1

// 	// ❗ Optional for stress testing
// 	// config.Producer.RequiredAcks = sarama.WaitForLocal    // Lower latency, less durable
// 	// config.Producer.Compression = sarama.CompressionSnappy // Reduce payload size

// 	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
// 	if err != nil {
// 		panic("Failed to start Kafka producer: " + err.Error())
// 	}

// 	return &KafkaProducer{
// 		producer: producer,
// 		topic:    cfg.Topic,
// 	}
// }

func NewKafkaProducer(cfg config.KafkaConfig) *KafkaProducer {
	config := sarama.NewConfig()

	// ✅ Reliability (optional to tweak)
	config.Producer.RequiredAcks = sarama.WaitForLocal // Lower latency than WaitForAll
	config.Producer.Retry.Max = 3                      // Fewer retries for speed
	config.Producer.Return.Successes = true            // Required for SyncProducer

	// ✅ Performance tuning
	config.Producer.Flush.Messages = 100                     // Batch after 100 messages
	config.Producer.Flush.Frequency = 200 * time.Millisecond // Or flush every 200ms
	config.Producer.Flush.MaxMessages = 1000                 // Upper cap on batch size
	config.Producer.Compression = sarama.CompressionSnappy   // Reduce payload size
	config.Net.MaxOpenRequests = 5                           // Allow more concurrent requests

	// ❗ Optional - disable if not required for exact-once delivery
	config.Producer.Idempotent = false // Better throughput if retries are low

	producer, err := sarama.NewSyncProducer(cfg.Brokers, config)
	if err != nil {
		panic("Failed to start Kafka producer: " + err.Error())
	}

	return &KafkaProducer{
		producer: producer,
		topic:    cfg.Topic,
	}
}

func (p *KafkaProducer) ProduceOrderCreated(order *models.Order) error {
	msgBytes, err := json.Marshal(order)
	if err != nil {
		return err
	}

	msg := &sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(fmt.Sprintf("%d", order.UserID)), // Or order.ID
		Value: sarama.ByteEncoder(msgBytes),
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}
