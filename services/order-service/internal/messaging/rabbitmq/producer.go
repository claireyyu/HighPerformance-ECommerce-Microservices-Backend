package rabbitmq

import (
	"encoding/json"
	"log"
	"time"

	"github.com/ecommerce-platform/order-service/internal/config"
	"github.com/ecommerce-platform/order-service/internal/models"
	"github.com/streadway/amqp"
)

type RabbitMQProducer struct {
	channel *amqp.Channel
	queue   string
}

func NewRabbitMQProducer(cfg config.RabbitMQConfig) *RabbitMQProducer {
	conn, err := amqp.Dial(cfg.URL)
	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(cfg.URL)
		if err == nil {
			break
		}
		log.Printf("Retrying RabbitMQ connection... (%d/5)", i+1)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		panic("Failed to connect to RabbitMQ after retries: " + err.Error())
	}

	ch, err := conn.Channel()
	if err != nil {
		panic("Failed to open RabbitMQ channel: " + err.Error())
	}

	_, err = ch.QueueDeclare(
		cfg.Queue, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		panic("Failed to declare RabbitMQ queue: " + err.Error())
	}

	return &RabbitMQProducer{
		channel: ch,
		queue:   cfg.Queue,
	}
}

func (p *RabbitMQProducer) ProduceOrderCreated(order *models.Order) error {
	msgBytes, err := json.Marshal(order)
	if err != nil {
		return err
	}

	return p.channel.Publish(
		"",      // exchange
		p.queue, // routing key (queue name)
		false,   // mandatory
		false,   // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        msgBytes,
		},
	)
}
