package main

import (
	"log"

	"github.com/Shopify/sarama"
	"github.com/ecommerce-platform/product-service/internal/config"
	"github.com/ecommerce-platform/product-service/internal/handlers"
	"github.com/ecommerce-platform/product-service/internal/repository"
	"github.com/ecommerce-platform/product-service/internal/service"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize repository
	repo, err := repository.NewProductRepository(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	// Initialize Kafka producer
	kafkaConfig := sarama.NewConfig()
	kafkaConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer([]string{cfg.Kafka.Brokers}, kafkaConfig)
	if err != nil {
		log.Fatalf("Failed to create Kafka producer: %v", err)
	}
	defer producer.Close()

	// Initialize service
	productService := service.NewProductService(repo, producer, cfg.Kafka.Topic)

	// Initialize handler
	handler := handlers.NewProductHandler(productService)

	// Initialize router
	router := gin.Default()

	// Health check endpoint with database check
	router.GET("/health", func(c *gin.Context) {
		// Try to list products to check database connectivity
		_, err := repo.List()
		if err != nil {
			c.JSON(500, gin.H{
				"status": "error",
				"error":  "Database connection failed: " + err.Error(),
			})
			return
		}

		c.JSON(200, gin.H{
			"status": "ok",
			"checks": gin.H{
				"database": "ok",
				"server":   "ok",
			},
		})
	})

	// Routes
	router.POST("/products", handler.CreateProduct)
	router.GET("/products/:id", handler.GetProduct)
	router.PUT("/products/:id", handler.UpdateProduct)
	router.DELETE("/products/:id", handler.DeleteProduct)
	router.GET("/products", handler.ListProducts)

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
