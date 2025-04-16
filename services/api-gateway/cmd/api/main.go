package main

import (
	"log"

	"github.com/ecommerce-platform/api-gateway/internal/config"
	"github.com/ecommerce-platform/api-gateway/internal/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize handlers
	productHandler := handlers.NewProductHandler(cfg.ProductService.URL)
	orderHandler := handlers.NewOrderHandler(cfg.OrderService.URL)

	// Initialize router
	router := gin.Default()

	// Product routes
	router.POST("/products", productHandler.CreateProduct)
	router.GET("/products/:id", productHandler.GetProduct)
	router.GET("/products", productHandler.ListProducts)
	router.PUT("/products/:id", productHandler.UpdateProduct)
	router.DELETE("/products/:id", productHandler.DeleteProduct)

	// Order routes
	router.POST("/orders", orderHandler.CreateOrder)
	router.POST("/orders/async", orderHandler.CreateOrderAsync)
	router.GET("/orders/:id", orderHandler.GetOrder)
	router.GET("/orders", orderHandler.ListOrders)
	router.PUT("/orders/:id/status", orderHandler.UpdateOrderStatus)

	// Start server
	addr := cfg.Server.Host + ":" + cfg.Server.Port
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
