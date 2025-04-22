package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/claireyu/ecommerce/config"
	"github.com/claireyu/ecommerce/services"
)

func main() {
	// Parse command line flags
	serviceName := flag.String("service", "", "Service to run (product, order)")
	flag.Parse()

	if *serviceName == "" {
		log.Fatal("Service name is required. Use -service flag (product, order)")
	}

	// Load configuration for services that need it
	if *serviceName == "product" || *serviceName == "order" {
		if err := config.LoadConfig(); err != nil {
			log.Fatalf("Failed to load configuration: %v", err)
		}
	}

	// Start the appropriate service
	var err error
	switch *serviceName {
	case "product":
		err = services.StartProductService()
	case "order":
		err = services.StartOrderService()
	default:
		log.Fatalf("Unknown service: %s", *serviceName)
	}

	if err != nil {
		log.Fatalf("Failed to start service: %v", err)
	}

	// Wait for termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println("Shutting down...")
}
