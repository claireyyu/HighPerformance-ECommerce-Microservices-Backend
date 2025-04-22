package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/claireyu/ecommerce/queues"
	_ "github.com/go-sql-driver/mysql"
)

// Order represents an order in the system
type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// StartOrderService starts the Order Service
func StartOrderService() error {
	// Connect to MySQL
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	// Test database connection
	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	// Initialize Kafka producer
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	kafkaProducer, err := queues.NewKafkaProducer(kafkaBrokers)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// Initialize RabbitMQ client
	rabbitmqURL := fmt.Sprintf("amqp://%s:%s@%s:%s/%%2F",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)
	rabbitmqClient := queues.NewRabbitMQClient(rabbitmqURL)
	if err := rabbitmqClient.Connect(); err != nil {
		return fmt.Errorf("failed to create RabbitMQ client: %v", err)
	}
	defer rabbitmqClient.Close()

	// Set up routes
	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			// Get all orders
			rows, err := db.Query("SELECT id, user_id, product_id, quantity, status, created_at FROM orders")
			if err != nil {
				http.Error(w, "Failed to get orders", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var orders []Order
			for rows.Next() {
				var o Order
				if err := rows.Scan(&o.ID, &o.UserID, &o.ProductID, &o.Quantity, &o.Status, &o.CreatedAt); err != nil {
					http.Error(w, "Failed to scan order", http.StatusInternalServerError)
					return
				}
				orders = append(orders, o)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(orders)
		}
	})

	// Synchronous order creation
	http.HandleFunc("/orders/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var o Order
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Set default values
		o.Status = "pending"
		o.CreatedAt = time.Now()

		result, err := db.Exec(
			"INSERT INTO orders (user_id, product_id, quantity, status, created_at) VALUES (?, ?, ?, ?, ?)",
			o.UserID, o.ProductID, o.Quantity, o.Status, o.CreatedAt,
		)
		if err != nil {
			http.Error(w, "Failed to create order", http.StatusInternalServerError)
			return
		}

		id, _ := result.LastInsertId()
		o.ID = int(id)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(o)
	})

	// Asynchronous order creation with Kafka
	http.HandleFunc("/orders/async/kafka", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var o Order
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Set default values
		o.Status = "pending"
		o.CreatedAt = time.Now()

		// Publish order created event to Kafka
		event := map[string]interface{}{
			"type":  "order_created",
			"order": o,
		}
		eventJSON, _ := json.Marshal(event)
		if err := kafkaProducer.PublishToKafka(kafkaTopic, string(eventJSON)); err != nil {
			http.Error(w, "Failed to publish to Kafka", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Order creation request accepted and queued in Kafka",
			"order":   string(eventJSON),
		})
	})

	// Asynchronous order creation with RabbitMQ
	http.HandleFunc("/orders/async/rabbitmq", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var o Order
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Set default values
		o.Status = "pending"
		o.CreatedAt = time.Now()

		// Publish order created event to RabbitMQ
		event := map[string]interface{}{
			"type":  "order_created",
			"order": o,
		}
		eventJSON, _ := json.Marshal(event)
		if err := rabbitmqClient.PublishToRabbitMQ("orders", string(eventJSON)); err != nil {
			http.Error(w, "Failed to publish to RabbitMQ", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "Order creation request accepted and queued in RabbitMQ",
			"order":   string(eventJSON),
		})
	})

	// Start server
	port := ":8082"
	log.Printf("Order Service starting on port %s", port)
	return http.ListenAndServe(port, nil)
}
