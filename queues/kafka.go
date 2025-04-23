package queues

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func NewKafkaProducer(brokers string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Idempotent = true

	// 🚀 批处理配置
	config.Producer.Flush.Messages = 500
	config.Producer.Flush.Frequency = 50 * time.Millisecond
	config.Producer.Flush.MaxMessages = 500

	config.Producer.Return.Successes = true
	config.Producer.Compression = sarama.CompressionSnappy

	producer, err := sarama.NewSyncProducer(strings.Split(brokers, ","), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %v", err)
	}

	return &KafkaProducer{producer}, nil
}

func (kp *KafkaProducer) PublishToKafka(topic, message string) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	}
	_, _, err := kp.producer.SendMessage(msg)
	return err
}

func (kp *KafkaProducer) Close() error {
	return kp.producer.Close()
}

// ----- ✅ Kafka Consumer Group with Goroutine Pool -----

type kafkaHandler struct {
	handler func(string)
	pool    chan struct{}
}

func (h *kafkaHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *kafkaHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *kafkaHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.pool <- struct{}{} // block if pool is full
		go func(m *sarama.ConsumerMessage) {
			defer func() { <-h.pool }()
			h.handler(string(m.Value))
			session.MarkMessage(m, "")
		}(msg)
	}
	return nil
}

// StartGroupConsumer creates a consumer group with limited concurrency
func StartGroupConsumer(brokers []string, topic string, handler func(string)) {
	groupID := "ecommerce-group"
	concurrency := 10

	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		log.Fatalf("❌ Failed to create Kafka consumer group: %v", err)
	}

	ctx := context.Background()
	h := &kafkaHandler{
		handler: handler,
		pool:    make(chan struct{}, concurrency),
	}

	go func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{topic}, h); err != nil {
				log.Printf("❌ Error from consumer group: %v", err)
				time.Sleep(2 * time.Second) // retry delay
			}
		}
	}()
}
