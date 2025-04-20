package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server         ServerConfig
	ProductService ServiceConfig
	OrderService   ServiceConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type ServiceConfig struct {
	URL string
}

func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	// Set default values
	viper.SetDefault("API_GATEWAY_PORT", "8080")
	viper.SetDefault("API_GATEWAY_HOST", "0.0.0.0")
	viper.SetDefault("PRODUCT_SERVICE_HOST", "product-service")
	viper.SetDefault("PRODUCT_SERVICE_PORT", "8081")
	viper.SetDefault("ORDER_SERVICE_HOST", "order-service")
	viper.SetDefault("ORDER_SERVICE_PORT", "8082")

	config := &Config{
		Server: ServerConfig{
			Port: viper.GetString("API_GATEWAY_PORT"),
			Host: viper.GetString("API_GATEWAY_HOST"),
		},
		ProductService: ServiceConfig{
			URL: "http://" + viper.GetString("PRODUCT_SERVICE_HOST") + ":" + viper.GetString("PRODUCT_SERVICE_PORT"),
		},
		OrderService: ServiceConfig{
			URL: "http://" + viper.GetString("ORDER_SERVICE_HOST") + ":" + viper.GetString("ORDER_SERVICE_PORT"),
		},
	}

	return config, nil
}
