package queues

import (
	"fmt"
	"log"
	"time"

	"github.com/streadway/amqp"
)

// RabbitMQClient represents a RabbitMQ client
type RabbitMQClient struct {
	conn    *amqp.Connection
	channel *amqp.Channel
	config  *amqp.Config
	url     string
	done    chan struct{}
}

// NewRabbitMQClient creates a new RabbitMQ client
func NewRabbitMQClient(url string) *RabbitMQClient {
	return &RabbitMQClient{
		url: url,
		config: &amqp.Config{
			Dial: amqp.DefaultDial(30 * time.Second),
			Properties: amqp.Table{
				"connection_name": "ecommerce-service",
			},
			Locale:    "en_US",
			Heartbeat: 10 * time.Second,
		},
		done: make(chan struct{}),
	}
}

// Connect establishes a connection to RabbitMQ
func (c *RabbitMQClient) Connect() error {
	var err error
	c.conn, err = amqp.DialConfig(c.url, *c.config)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}

	c.channel, err = c.conn.Channel()
	if err != nil {
		c.conn.Close()
		return fmt.Errorf("failed to open channel: %v", err)
	}

	// Enable publisher confirms
	err = c.channel.Confirm(false)
	if err != nil {
		c.channel.Close()
		c.conn.Close()
		return fmt.Errorf("failed to enable publisher confirms: %v", err)
	}

	// Start reconnection goroutine
	go c.reconnect()

	return nil
}

// reconnect handles connection recovery
func (c *RabbitMQClient) reconnect() {
	for {
		select {
		case <-c.done:
			return
		case <-c.conn.NotifyClose(make(chan *amqp.Error)):
			log.Printf("Connection closed, attempting to reconnect...")
			for {
				select {
				case <-c.done:
					return
				default:
					err := c.Connect()
					if err == nil {
						log.Printf("Successfully reconnected to RabbitMQ")
						return
					}
					log.Printf("Failed to reconnect: %v", err)
					time.Sleep(5 * time.Second)
				}
			}
		}
	}
}

// Close closes the RabbitMQ connection
func (c *RabbitMQClient) Close() error {
	close(c.done)
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// PublishToRabbitMQ publishes a message to a RabbitMQ queue
func (r *RabbitMQClient) PublishToRabbitMQ(queue, message string) error {
	// Declare queue with durability
	q, err := r.channel.QueueDeclare(
		queue, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		amqp.Table{
			"x-message-ttl": int32(24 * 60 * 60 * 1000), // 24 hours
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	// Create confirmation channel
	confirms := r.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	// Publish message
	err = r.channel.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(message),
			Headers: amqp.Table{
				"timestamp": time.Now().Unix(),
			},
			DeliveryMode: amqp.Persistent,
		})
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	// Wait for confirmation
	if confirmed := <-confirms; !confirmed.Ack {
		return fmt.Errorf("failed to confirm message publication")
	}

	log.Printf("Message sent to queue %s", queue)
	return nil
}

// Consumer represents a RabbitMQ consumer
type Consumer struct {
	client    *RabbitMQClient
	queue     string
	handler   func([]byte) error
	stopChan  chan struct{}
	isRunning bool
	reconnect bool
}

// NewConsumer creates a new RabbitMQ consumer
func NewConsumer(client *RabbitMQClient, queue string, handler func([]byte) error) *Consumer {
	return &Consumer{
		client:    client,
		queue:     queue,
		handler:   handler,
		stopChan:  make(chan struct{}),
		isRunning: false,
		reconnect: true,
	}
}

// Start begins consuming messages from the queue
func (c *Consumer) Start() error {
	if c.isRunning {
		return fmt.Errorf("consumer is already running")
	}

	// Declare queue
	q, err := c.client.channel.QueueDeclare(
		c.queue, // name
		true,    // durable
		false,   // delete when unused
		false,   // exclusive
		false,   // no-wait
		amqp.Table{
			"x-message-ttl": int32(24 * 60 * 60 * 1000), // 24 hours
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	// Set QoS
	err = c.client.channel.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %v", err)
	}

	msgs, err := c.client.channel.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %v", err)
	}

	c.isRunning = true
	go c.consume(msgs)

	return nil
}

// Stop stops consuming messages
func (c *Consumer) Stop() {
	if !c.isRunning {
		return
	}

	close(c.stopChan)
	c.isRunning = false
}

// consume handles incoming messages
func (c *Consumer) consume(msgs <-chan amqp.Delivery) {
	for {
		select {
		case <-c.stopChan:
			return
		case msg, ok := <-msgs:
			if !ok {
				if !c.reconnect {
					return
				}
				log.Printf("Channel closed, attempting to reconnect...")
				for {
					select {
					case <-c.stopChan:
						return
					default:
						err := c.Start()
						if err == nil {
							log.Printf("Successfully reconnected consumer")
							return
						}
						log.Printf("Failed to reconnect consumer: %v", err)
						time.Sleep(5 * time.Second)
					}
				}
			}

			// Process message
			err := c.handler(msg.Body)
			if err != nil {
				log.Printf("Error processing message: %v", err)
				// Reject message and requeue
				msg.Nack(false, true)
				continue
			}

			// Acknowledge message
			err = msg.Ack(false)
			if err != nil {
				log.Printf("Error acknowledging message: %v", err)
			}
		}
	}
}

// QueueConfig holds configuration for queue declaration
type QueueConfig struct {
	Name       string
	Durable    bool
	AutoDelete bool
	Exclusive  bool
	NoWait     bool
	Args       amqp.Table
}

// DefaultQueueConfig returns a default queue configuration
func DefaultQueueConfig(name string) QueueConfig {
	return QueueConfig{
		Name:       name,
		Durable:    true,
		AutoDelete: false,
		Exclusive:  false,
		NoWait:     false,
		Args: amqp.Table{
			"x-message-ttl": int32(24 * 60 * 60 * 1000), // 24 hours
		},
	}
}

// DeclareQueue declares a queue with the given configuration
func (c *RabbitMQClient) DeclareQueue(config QueueConfig) (amqp.Queue, error) {
	if c.channel == nil {
		return amqp.Queue{}, fmt.Errorf("channel is not initialized")
	}

	queue, err := c.channel.QueueDeclare(
		config.Name,
		config.Durable,
		config.AutoDelete,
		config.Exclusive,
		config.NoWait,
		config.Args,
	)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to declare queue: %v", err)
	}

	return queue, nil
}

// PublishMessage publishes a message to the specified queue
func (c *RabbitMQClient) PublishMessage(queue string, message []byte) error {
	if c.channel == nil {
		return fmt.Errorf("channel is not initialized")
	}

	// Create confirmation channel
	confirmChan := c.channel.NotifyPublish(make(chan amqp.Confirmation, 1))

	// Publish message
	err := c.channel.Publish(
		"",    // exchange
		queue, // routing key
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        message,
			Headers: amqp.Table{
				"timestamp": time.Now().Unix(),
			},
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %v", err)
	}

	// Wait for confirmation
	if confirmed := <-confirmChan; !confirmed.Ack {
		return fmt.Errorf("failed to confirm message publication")
	}

	return nil
}

// PurgeQueue removes all messages from the specified queue
func (c *RabbitMQClient) PurgeQueue(queue string) error {
	if c.channel == nil {
		return fmt.Errorf("channel is not initialized")
	}

	_, err := c.channel.QueuePurge(queue, false)
	if err != nil {
		return fmt.Errorf("failed to purge queue: %v", err)
	}

	return nil
}

// DeleteQueue deletes the specified queue
func (c *RabbitMQClient) DeleteQueue(queue string) error {
	if c.channel == nil {
		return fmt.Errorf("channel is not initialized")
	}

	_, err := c.channel.QueueDelete(queue, false, false, false)
	if err != nil {
		return fmt.Errorf("failed to delete queue: %v", err)
	}

	return nil
}

// GetQueueInfo returns information about the specified queue
func (c *RabbitMQClient) GetQueueInfo(queue string) (amqp.Queue, error) {
	if c.channel == nil {
		return amqp.Queue{}, fmt.Errorf("channel is not initialized")
	}

	queueInfo, err := c.channel.QueueInspect(queue)
	if err != nil {
		return amqp.Queue{}, fmt.Errorf("failed to get queue info: %v", err)
	}

	return queueInfo, nil
}

// RabbitMQConsumerConfig holds configuration for consumer scaling
type RabbitMQConsumerConfig struct {
	PrefetchCount  int           // 每个 consumer 预取的消息数
	PrefetchSize   int           // 预取的消息大小（字节）
	Global         bool          // 是否全局应用 QoS
	WorkerCount    int           // worker 数量
	ReconnectDelay time.Duration // 重连延迟
}

// DefaultRabbitMQConsumerConfig returns default consumer configuration
func DefaultRabbitMQConsumerConfig() RabbitMQConsumerConfig {
	return RabbitMQConsumerConfig{
		PrefetchCount:  1,
		PrefetchSize:   0,
		Global:         false,
		WorkerCount:    1,
		ReconnectDelay: 5 * time.Second,
	}
}

// StartWorkers starts multiple consumer workers
func (c *Consumer) StartWorkers(config RabbitMQConsumerConfig) error {
	if c.isRunning {
		return fmt.Errorf("consumer is already running")
	}

	// Set QoS for the channel
	err := c.client.channel.Qos(
		config.PrefetchCount,
		config.PrefetchSize,
		config.Global,
	)
	if err != nil {
		return fmt.Errorf("failed to set QoS: %v", err)
	}

	// Declare queue
	q, err := c.client.channel.QueueDeclare(
		c.queue,
		true,
		false,
		false,
		false,
		amqp.Table{
			"x-message-ttl": int32(24 * 60 * 60 * 1000),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %v", err)
	}

	c.isRunning = true

	// Start multiple workers
	for i := 0; i < config.WorkerCount; i++ {
		workerID := fmt.Sprintf("worker-%d", i)
		go c.startWorker(q.Name, workerID, config)
	}

	return nil
}

// startWorker starts a single consumer worker
func (c *Consumer) startWorker(queueName, workerID string, config RabbitMQConsumerConfig) {
	msgs, err := c.client.channel.Consume(
		queueName,
		workerID, // 使用唯一的 consumer tag
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Printf("Worker %s failed to start: %v", workerID, err)
		return
	}

	for {
		select {
		case <-c.stopChan:
			return
		case msg, ok := <-msgs:
			if !ok {
				if !c.reconnect {
					return
				}
				log.Printf("Worker %s: Channel closed, attempting to reconnect...", workerID)
				time.Sleep(config.ReconnectDelay)
				continue
			}

			// Process message
			err := c.handler(msg.Body)
			if err != nil {
				log.Printf("Worker %s: Error processing message: %v", workerID, err)
				msg.Nack(false, true)
				continue
			}

			err = msg.Ack(false)
			if err != nil {
				log.Printf("Worker %s: Error acknowledging message: %v", workerID, err)
			}
		}
	}
}
