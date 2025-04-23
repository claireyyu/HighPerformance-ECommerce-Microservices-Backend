package queues

import (
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

type RabbitMQClient struct {
	conn *amqp.Connection
	URL  string
}

func NewRabbitMQClient(url string) *RabbitMQClient {
	return &RabbitMQClient{URL: url}
}

func (r *RabbitMQClient) Connect() error {
	conn, err := amqp.Dial(r.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	r.conn = conn
	return nil
}

func (r *RabbitMQClient) Close() {
	if r.conn != nil {
		_ = r.conn.Close()
	}
}

func (r *RabbitMQClient) Publish(queueName, message string) error {
	ch, err := r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %v", err)
	}
	defer ch.Close()

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	return ch.Publish("", queueName, false, false, amqp.Publishing{
		ContentType: "application/json",
		Body:        []byte(message),
	})
}

// ✅ 并发消费队列（带 goroutine 池 + 日志）
func (r *RabbitMQClient) ConsumeWithPool(queueName string, concurrency int, handler func(string) error) error {
	log.Printf("🚀 Starting RabbitMQ consumer with %d workers on queue: %s", concurrency, queueName)

	for i := 0; i < concurrency; i++ {
		go func(workerID int) {
			ch, err := r.conn.Channel()
			if err != nil {
				log.Printf("❌ Worker %d: failed to open channel: %v", workerID, err)
				return
			}
			defer ch.Close()

			_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
			if err != nil {
				log.Printf("❌ Worker %d: queue declare error: %v", workerID, err)
				return
			}

			err = ch.Qos(1, 0, false) // 每个 goroutine 只处理一个未 ack 消息
			if err != nil {
				log.Printf("❌ Worker %d: QoS error: %v", workerID, err)
				return
			}

			msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
			if err != nil {
				log.Printf("❌ Worker %d: consume error: %v", workerID, err)
				return
			}

			for msg := range msgs {
				log.Printf("📥 [Worker %d] Received: %s", workerID, msg.Body)

				if err := handler(string(msg.Body)); err != nil {
					log.Printf("❌ [Worker %d] handler error: %v", workerID, err)
					_ = msg.Nack(false, true) // requeue
					continue
				}

				if err := msg.Ack(false); err != nil {
					log.Printf("❌ [Worker %d] ack error: %v", workerID, err)
				} else {
					log.Printf("✅ [Worker %d] message acked", workerID)
				}
			}
		}(i)
	}

	return nil
}
