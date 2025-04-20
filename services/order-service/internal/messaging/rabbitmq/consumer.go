// internal/messaging/rabbitmq/consumer.go
package rabbitmq

import (
	"context"
	"encoding/json"
	"log"

	"github.com/ecommerce-platform/order-service/internal/config"
	"github.com/ecommerce-platform/order-service/internal/models"
	"github.com/ecommerce-platform/order-service/internal/repository"
	"github.com/streadway/amqp"
)

func StartRabbitMQConsumer(cfg config.RabbitMQConfig, orderRepo repository.OrderRepository) {
	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open RabbitMQ channel: %v", err)
	}

	msgs, err := ch.Consume(
		cfg.Queue,
		"",    // consumer tag
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		log.Fatalf("Failed to register RabbitMQ consumer: %v", err)
	}

	log.Println("✅ RabbitMQ consumer running...")

	go func() {
		for msg := range msgs {
			var order models.Order
			if err := json.Unmarshal(msg.Body, &order); err != nil {
				log.Printf("Invalid RabbitMQ message: %v", err)
				continue
			}

			err := orderRepo.Create(context.Background(), &order)
			if err != nil {
				log.Printf("❌ Failed to insert order from RabbitMQ: %v", err)
			} else {
				log.Printf("✅ Order inserted from RabbitMQ: ID=%d", order.ID)
			}
		}
	}()
}
