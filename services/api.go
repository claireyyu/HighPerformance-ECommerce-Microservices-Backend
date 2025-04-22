package services

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

// StartAPIGateway starts the API Gateway service
func StartAPIGateway() error {
	productServiceURL := os.Getenv("PRODUCT_SERVICE_URL")
	orderServiceURL := os.Getenv("ORDER_SERVICE_URL")

	if productServiceURL == "" || orderServiceURL == "" {
		return fmt.Errorf("missing required environment variables: PRODUCT_SERVICE_URL, ORDER_SERVICE_URL")
	}

	// Set up routes
	http.HandleFunc("/products", func(w http.ResponseWriter, r *http.Request) {
		// Forward request to product service
		resp, err := http.Get(productServiceURL + "/products")
		if err != nil {
			log.Printf("Error forwarding request to product service: %v", err)
			http.Error(w, "Failed to get products", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Copy response to client
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Printf("Error copying response: %v", err)
			http.Error(w, "Failed to send response", http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		// Forward request to order service
		resp, err := http.Get(orderServiceURL + "/orders")
		if err != nil {
			log.Printf("Error forwarding request to order service: %v", err)
			http.Error(w, "Failed to get orders", http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()

		// Copy response to client
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(resp.StatusCode)
		if _, err := io.Copy(w, resp.Body); err != nil {
			log.Printf("Error copying response: %v", err)
			http.Error(w, "Failed to send response", http.StatusInternalServerError)
			return
		}
	})

	// Start server
	port := ":8080"
	log.Printf("API Gateway starting on port %s", port)
	return http.ListenAndServe(port, nil)
}
