package kafka

import (
	"encoding/json"

	"github.com/IBM/sarama"
	"github.com/ecommerce-platform/order-service/internal/config"
	"github.com/ecommerce-platform/order-service/internal/models"
)

type KafkaProducer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewKafkaProducer(cfg config.KafkaConfig) *KafkaProducer {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Producer.RequiredAcks = sarama.WaitForAll

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
		Value: sarama.ByteEncoder(msgBytes),
	}

	_, _, err = p.producer.SendMessage(msg)
	return err
}
