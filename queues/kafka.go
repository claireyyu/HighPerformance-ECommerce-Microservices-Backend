package queues

import (
	"context"
	"encoding/json"
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

	// 分区策略配置
	config.Producer.Partitioner = sarama.NewHashPartitioner
	config.Producer.Return.Successes = true

	// 🚀 批处理配置
	config.Producer.Flush.Messages = 500
	config.Producer.Flush.Frequency = 50 * time.Millisecond
	config.Producer.Flush.MaxMessages = 500

	config.Producer.Compression = sarama.CompressionSnappy

	producer, err := sarama.NewSyncProducer(strings.Split(brokers, ","), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %v", err)
	}

	return &KafkaProducer{producer}, nil
}

func (kp *KafkaProducer) PublishToKafka(topic, message string) error {
	// 使用用户ID作为分区key，确保同一用户的订单进入同一分区
	var key string
	var order struct {
		UserID int `json:"user_id"`
	}
	if err := json.Unmarshal([]byte(message), &order); err == nil {
		key = fmt.Sprintf("%d", order.UserID)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
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
	handler func(string) error
	pool    chan struct{}
}

func (h *kafkaHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (h *kafkaHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

func (h *kafkaHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.pool <- struct{}{} // block if pool is full
		go func(m *sarama.ConsumerMessage) {
			defer func() { <-h.pool }()
			if err := h.handler(string(m.Value)); err != nil {
				log.Printf("❌ Handler error: %v", err)
			}
			session.MarkMessage(m, "")
		}(msg)
	}
	return nil
}

// StartGroupConsumer creates a consumer group with limited concurrency
func StartGroupConsumer(brokers []string, topic string, handler func(string) error) {
	groupID := "ecommerce-group"

	config := sarama.NewConfig()
	config.Version = sarama.V2_1_0_0

	// 消费者组再平衡策略：粘性分区分配
	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategySticky
	config.Consumer.Group.Rebalance.Timeout = 60 * time.Second
	config.Consumer.Offsets.Initial = sarama.OffsetNewest

	// 优化的Fetch配置
	config.Consumer.Fetch.Min = 1
	config.Consumer.Fetch.Default = 1024 * 1024 // 1MB
	config.Consumer.MaxWaitTime = 100 * time.Millisecond

	consumerGroup, err := sarama.NewConsumerGroup(brokers, groupID, config)
	if err != nil {
		log.Fatalf("❌ Failed to create Kafka consumer group: %v", err)
	}

	ctx := context.Background()
	h := &kafkaHandler{
		handler: handler,
		pool:    make(chan struct{}, 1000),
	}

	go func() {
		for {
			if err := consumerGroup.Consume(ctx, []string{topic}, h); err != nil {
				log.Printf("❌ Error from consumer group: %v", err)
				time.Sleep(2 * time.Second)
			}
		}
	}()
}
