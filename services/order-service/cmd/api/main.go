package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ecommerce-platform/order-service/internal/config"
	"github.com/ecommerce-platform/order-service/internal/models"
	"github.com/ecommerce-platform/order-service/internal/repository"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize repository
	orderRepo, err := repository.NewOrderRepository(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	// Initialize router
	router := gin.Default()

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// Order routes
	router.POST("/orders", func(c *gin.Context) {
		var order models.Order
		if err := c.ShouldBindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		order.Status = models.OrderStatusPending
		if err := orderRepo.Create(c.Request.Context(), &order); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, order)
	})

	// Async order creation
	router.POST("/orders/async", func(c *gin.Context) {
		var order models.Order
		if err := c.ShouldBindJSON(&order); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		order.Status = models.OrderStatusPending

		// Process order asynchronously
		go func() {
			ctx := context.Background()
			if err := orderRepo.Create(ctx, &order); err != nil {
				log.Printf("Failed to create async order: %v", err)
				return
			}
			log.Printf("Async order created successfully: %d", order.ID)
		}()

		// Return immediately with 202 Accepted
		c.JSON(http.StatusAccepted, gin.H{
			"message": "Order is being processed",
			"status":  "accepted",
		})
	})

	router.GET("/orders/:id", func(c *gin.Context) {
		id := c.Param("id")
		var orderID uint
		if _, err := fmt.Sscanf(id, "%d", &orderID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
			return
		}

		order, err := orderRepo.GetByID(c.Request.Context(), orderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if order == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}

		c.JSON(http.StatusOK, order)
	})

	router.GET("/orders", func(c *gin.Context) {
		userIDStr := c.Query("user_id")
		var userID *uint
		if userIDStr != "" {
			var id uint
			if _, err := fmt.Sscanf(userIDStr, "%d", &id); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
				return
			}
			userID = &id
		}

		orders, err := orderRepo.List(c.Request.Context(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, orders)
	})

	router.PUT("/orders/:id/status", func(c *gin.Context) {
		// Parse order ID
		id := c.Param("id")
		var orderID uint
		if _, err := fmt.Sscanf(id, "%d", &orderID); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
			return
		}

		// Check if order exists
		order, err := orderRepo.GetByID(c.Request.Context(), orderID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if order == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}

		// Parse status update
		var status struct {
			Status models.OrderStatus `json:"status" binding:"required"`
		}
		if err := c.ShouldBindJSON(&status); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Validate status
		switch status.Status {
		case models.OrderStatusPending, models.OrderStatusPaid, models.OrderStatusShipped,
			models.OrderStatusDelivered, models.OrderStatusCancelled:
			// Valid status
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order status"})
			return
		}

		if err := orderRepo.UpdateStatus(c.Request.Context(), orderID, status.Status); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "updated"})
	})

	// Create server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Give outstanding requests 5 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
