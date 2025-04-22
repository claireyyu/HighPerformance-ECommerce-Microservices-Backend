package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/claireyu/ecommerce/queues"
	_ "github.com/go-sql-driver/mysql"
)

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

func StartProductService() error {
	// Connect to MySQL
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_NAME"),
	)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %v", err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS products (
		id INT AUTO_INCREMENT PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		price DOUBLE
	)`)
	if err != nil {
		return fmt.Errorf("failed to create products table: %v", err)
	}

	// Connect to Kafka
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	producer, err := queues.NewKafkaProducer(kafkaBrokers)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Setup HTTP routes
	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			rows, err := db.Query("SELECT id, name, description, price FROM products")
			if err != nil {
				http.Error(w, "Failed to get products", http.StatusInternalServerError)
				return
			}
			defer rows.Close()

			var products []Product
			for rows.Next() {
				var p Product
				if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Price); err != nil {
					http.Error(w, "Failed to scan product", http.StatusInternalServerError)
					return
				}
				products = append(products, p)
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(products)

		case "POST":
			var p Product
			if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
				http.Error(w, "Invalid request body", http.StatusBadRequest)
				return
			}

			result, err := db.Exec("INSERT INTO products (name, description, price) VALUES (?, ?, ?)",
				p.Name, p.Description, p.Price)
			if err != nil {
				http.Error(w, "Failed to create product", http.StatusInternalServerError)
				return
			}

			id, _ := result.LastInsertId()
			p.ID = int(id)

			// Create event & publish
			event := map[string]interface{}{
				"type":    "product_created",
				"product": p,
			}
			eventJSON, _ := json.Marshal(event)
			if err := producer.PublishToKafka(kafkaTopic, string(eventJSON)); err != nil {
				log.Printf("[ERROR] Failed to publish to Kafka: %v", err)
			} else if os.Getenv("DEBUG") == "true" {
				log.Printf("[Kafka] Sent: %s", eventJSON)
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(p)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Start server
	port := os.Getenv("PRODUCT_SERVICE_PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("[INFO] Product service running on port %s", port)
	return http.ListenAndServe(":"+port, nil)
}
