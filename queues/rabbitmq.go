package queues

import (
	"fmt"

	"github.com/streadway/amqp"
)

type RabbitMQClient struct {
	conn *amqp.Connection
	URL  string
}

// 创建 RabbitMQ 客户端
func NewRabbitMQClient(url string) *RabbitMQClient {
	return &RabbitMQClient{URL: url}
}

// 建立连接（只连一次）
func (r *RabbitMQClient) Connect() error {
	conn, err := amqp.Dial(r.URL)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	r.conn = conn
	return nil
}

// 发布消息（每次新建一个 channel，线程安全）
func (r *RabbitMQClient) Publish(queueName, message string) error {
	if r.conn == nil {
		return fmt.Errorf("RabbitMQ connection is not initialized")
	}

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

// 消费者（仅用于消费队列）
func (r *RabbitMQClient) Consume(queueName string, handler func(string)) error {
	ch, err := r.conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open channel: %v", err)
	}

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return err
	}

	msgs, err := ch.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		return err
	}

	go func() {
		for msg := range msgs {
			handler(string(msg.Body))
		}
	}()

	return nil
}

// 关闭连接
func (r *RabbitMQClient) Close() {
	if r.conn != nil {
		_ = r.conn.Close()
	}
}
