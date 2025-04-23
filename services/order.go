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

type Order struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ProductID int       `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

func StartOrderService() error {
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	rabbitURL := fmt.Sprintf(
		"amqp://%s:%s@%s:%s/",
		os.Getenv("RABBITMQ_USER"),
		os.Getenv("RABBITMQ_PASSWORD"),
		os.Getenv("RABBITMQ_HOST"),
		os.Getenv("RABBITMQ_PORT"),
	)

	kafkaProducer, err := queues.NewKafkaProducer(kafkaBrokers)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	rabbit := queues.NewRabbitMQClient(rabbitURL)
	if err := rabbit.Connect(); err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %v", err)
	}
	defer rabbit.Close()

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	db.SetMaxOpenConns(300)
	db.SetMaxIdleConns(100)

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping MySQL: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS orders (
		id INT AUTO_INCREMENT PRIMARY KEY,
		user_id INT NOT NULL,
		product_id INT NOT NULL,
		quantity INT NOT NULL,
		status VARCHAR(32),
		created_at DATETIME
	)`)
	if err != nil {
		return fmt.Errorf("failed to create orders table: %v", err)
	}

	consumeAndInsert := func(raw string) {
		log.Printf("[Consuming] %s", raw)
		var payload struct {
			Type  string `json:"type"`
			Order Order  `json:"order"`
		}
		if err := json.Unmarshal([]byte(raw), &payload); err != nil {
			log.Printf("❌ JSON error: %v", err)
			return
		}

		o := payload.Order
		_, err := db.Exec(
			"INSERT INTO orders (user_id, product_id, quantity, status, created_at) VALUES (?, ?, ?, ?, ?)",
			o.UserID, o.ProductID, o.Quantity, o.Status, o.CreatedAt,
		)
		if err != nil {
			log.Printf("❌ Insert error: %v", err)
			return
		}
		log.Printf("✅ Order inserted: user_id=%d product_id=%d", o.UserID, o.ProductID)
	}

	http.HandleFunc("/orders/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var o Order
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		o.Status = "pending"
		o.CreatedAt = time.Now()

		_, err := db.Exec(
			"INSERT INTO orders (user_id, product_id, quantity, status, created_at) VALUES (?, ?, ?, ?, ?)",
			o.UserID, o.ProductID, o.Quantity, o.Status, o.CreatedAt,
		)
		if err != nil {
			http.Error(w, "Failed to insert into DB", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Order inserted (sync)"})
	})

	http.HandleFunc("/orders/async/kafka", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var o Order
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		o.Status = "pending"
		o.CreatedAt = time.Now()

		event := map[string]interface{}{"type": "order_created", "order": o}
		eventJSON, _ := json.Marshal(event)

		if err := kafkaProducer.PublishToKafka(kafkaTopic, string(eventJSON)); err != nil {
			http.Error(w, "Failed to send to Kafka", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Order sent to Kafka"})
	})

	http.HandleFunc("/orders/async/rabbitmq", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}

		var o Order
		if err := json.NewDecoder(r.Body).Decode(&o); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}
		o.Status = "pending"
		o.CreatedAt = time.Now()

		event := map[string]interface{}{"type": "order_created", "order": o}
		eventJSON, _ := json.Marshal(event)

		if err := rabbit.Publish("orders", string(eventJSON)); err != nil {
			http.Error(w, "Failed to send to RabbitMQ", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Order sent to RabbitMQ"})
	})

	go queues.StartGroupConsumer(
		[]string{kafkaBrokers},
		kafkaTopic,
		func(msg string) {
			consumeAndInsert(msg)
		},
	)

	go func() {
		err := rabbit.ConsumeWithPool("orders", 50, func(msg string) error {
			log.Printf("🔧 Handling: %s", msg)

			var payload struct {
				Type  string `json:"type"`
				Order Order  `json:"order"`
			}
			if err := json.Unmarshal([]byte(msg), &payload); err != nil {
				return err
			}

			o := payload.Order
			_, err := db.Exec(
				"INSERT INTO orders (user_id, product_id, quantity, status, created_at) VALUES (?, ?, ?, ?, ?)",
				o.UserID, o.ProductID, o.Quantity, o.Status, o.CreatedAt,
			)
			return err
		})
		if err != nil {
			log.Printf("❌ Failed to start RabbitMQ consumer: %v", err)
		}
	}()

	port := os.Getenv("ORDER_SERVICE_PORT")
	if port == "" {
		port = "8081"
	}
	log.Printf("📦 Order Service running on port %s", port)
	return http.ListenAndServe(":"+port, nil)
}
