package queues

import (
	"fmt"

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

// ✅ 新增：并发消费队列（带 goroutine 池）
func (r *RabbitMQClient) ConsumeWithPool(queueName string, concurrency int, handler func(string)) error {
	ch, err := r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %v", err)
	}

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	err = ch.Qos(concurrency, 0, false) // 限制未ack消息数量
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	sem := make(chan struct{}, concurrency)

	go func() {
		for msg := range msgs {
			sem <- struct{}{}
			go func(m amqp.Delivery) {
				defer func() { <-sem }()
				handler(string(m.Body))
				_ = m.Ack(false)
			}(msg)
		}
	}()

	return nil
}
