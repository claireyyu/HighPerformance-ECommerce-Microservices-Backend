package queues

import (
	"fmt"
	"strings"
	"time"

	"github.com/IBM/sarama"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
}

func NewKafkaProducer(brokers string) (*KafkaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Retry.Max = 3
	config.Producer.Timeout = 5 * time.Second

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

// StartKafkaConsumer starts a basic consumer for a topic and prints/logs message
func StartKafkaConsumer(brokers []string, topic string, handler func(msg string)) error {
	consumer, err := sarama.NewConsumer(brokers, nil)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %v", err)
	}

	partitions, err := consumer.Partitions(topic)
	if err != nil {
		return fmt.Errorf("failed to get partitions: %v", err)
	}

	for _, partition := range partitions {
		pc, err := consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
		if err != nil {
			return fmt.Errorf("failed to consume partition: %v", err)
		}

		go func(pc sarama.PartitionConsumer) {
			for msg := range pc.Messages() {
				handler(string(msg.Value))
			}
		}(pc)
	}

	return nil
}
