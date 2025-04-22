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

// Product represents a product in the system
type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
}

// StartProductService starts the Product Service
func StartProductService() error {
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
	kafkaTopic := "product-events"
	kafkaProducer, err := queues.NewKafkaProducer(kafkaBrokers)
	if err != nil {
		return fmt.Errorf("failed to create Kafka producer: %v", err)
	}
	defer kafkaProducer.Close()

	// Set up routes
	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			// Get all products
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
			// Create new product
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

			// Publish product created event to Kafka
			event := map[string]interface{}{
				"type":    "product_created",
				"product": p,
			}
			eventJSON, _ := json.Marshal(event)
			if err := kafkaProducer.PublishToKafka(kafkaTopic, string(eventJSON)); err != nil {
				log.Printf("Failed to publish to Kafka: %v", err)
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
		port = "8080" // Default port if environment variable is not set
	}
	port = ":" + port
	log.Printf("Product Service starting on port %s", port)
	return http.ListenAndServe(port, nil)
}
