package queues

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

// KafkaProducer represents a Kafka producer
type KafkaProducer struct {
	producer sarama.SyncProducer
}

// NewKafkaProducer creates a new Kafka producer
func NewKafkaProducer(brokers string) (*KafkaProducer, error) {
	config := sarama.NewConfig()

	// Producer config
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Retry.Backoff = 100 * time.Millisecond

	// Compression
	config.Producer.Compression = sarama.CompressionSnappy

	// Timeouts
	config.Producer.Timeout = 5 * time.Second

	// Create producer
	producer, err := sarama.NewSyncProducer(strings.Split(brokers, ","), config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka producer: %v", err)
	}

	return &KafkaProducer{producer: producer}, nil
}

// Close closes the Kafka producer
func (kp *KafkaProducer) Close() error {
	return kp.producer.Close()
}

// PublishToKafka publishes a message to a Kafka topic
func (kp *KafkaProducer) PublishToKafka(topic, message string) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
		Key:   sarama.StringEncoder(fmt.Sprintf("%d", time.Now().UnixNano())), // Add message key for partitioning
	}

	partition, offset, err := kp.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to Kafka: %v", err)
	}

	log.Printf("Message sent to partition %d at offset %d", partition, offset)
	return nil
}

// KafkaConsumer represents a Kafka consumer
type KafkaConsumer struct {
	consumer sarama.Consumer
	client   sarama.Client
	config   *sarama.Config
	done     chan struct{}
}

// KafkaConsumerConfig holds configuration for consumer scaling
type KafkaConsumerConfig struct {
	GroupID           string
	InitialOffset     int64
	MaxProcessingTime time.Duration
	Workers           int
	ReconnectDelay    time.Duration
}

// DefaultKafkaConsumerConfig returns default consumer configuration
func DefaultKafkaConsumerConfig() KafkaConsumerConfig {
	return KafkaConsumerConfig{
		GroupID:           "ecommerce-group",
		InitialOffset:     sarama.OffsetNewest,
		MaxProcessingTime: 500 * time.Millisecond,
		Workers:           1,
		ReconnectDelay:    5 * time.Second,
	}
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(brokers string, config KafkaConsumerConfig) (*KafkaConsumer, error) {
	saramaConfig := sarama.NewConfig()

	// Consumer config
	saramaConfig.Consumer.Return.Errors = true
	saramaConfig.Consumer.Offsets.Initial = config.InitialOffset
	saramaConfig.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRoundRobin
	saramaConfig.Consumer.Group.Session.Timeout = 20 * time.Second
	saramaConfig.Consumer.Group.Heartbeat.Interval = 6 * time.Second
	saramaConfig.Consumer.MaxProcessingTime = config.MaxProcessingTime

	// Create client
	client, err := sarama.NewClient(strings.Split(brokers, ","), saramaConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %v", err)
	}

	// Create consumer
	consumer, err := sarama.NewConsumerFromClient(client)
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create Kafka consumer: %v", err)
	}

	return &KafkaConsumer{
		consumer: consumer,
		client:   client,
		config:   saramaConfig,
		done:     make(chan struct{}),
	}, nil
}

// StartConsuming starts consuming messages from the specified topic
func (kc *KafkaConsumer) StartConsuming(topic string, handler func([]byte) error) error {
	// Get partitions
	partitions, err := kc.consumer.Partitions(topic)
	if err != nil {
		return fmt.Errorf("failed to get partitions: %v", err)
	}

	// Start a consumer for each partition
	for _, partition := range partitions {
		pc, err := kc.consumer.ConsumePartition(topic, partition, kc.config.Consumer.Offsets.Initial)
		if err != nil {
			return fmt.Errorf("failed to create partition consumer: %v", err)
		}

		go func(pc sarama.PartitionConsumer) {
			defer pc.Close()

			for {
				select {
				case <-kc.done:
					return
				case msg := <-pc.Messages():
					if err := handler(msg.Value); err != nil {
						log.Printf("Error processing message: %v", err)
						continue
					}
					// The offset is automatically committed by the consumer group
					log.Printf("Message processed successfully at offset %d", msg.Offset)
				case err := <-pc.Errors():
					log.Printf("Error consuming from partition: %v", err)
					time.Sleep(kc.config.Consumer.Group.Heartbeat.Interval)
				}
			}
		}(pc)
	}

	return nil
}

// Close closes the Kafka consumer
func (kc *KafkaConsumer) Close() error {
	close(kc.done)
	if err := kc.consumer.Close(); err != nil {
		return fmt.Errorf("failed to close consumer: %v", err)
	}
	if err := kc.client.Close(); err != nil {
		return fmt.Errorf("failed to close client: %v", err)
	}
	return nil
}
