package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	DB struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"db"`

	Kafka struct {
		Brokers string `yaml:"brokers"`
		Topic   string `yaml:"topic"`
	} `yaml:"kafka"`

	RabbitMQ struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"rabbitmq"`

	Services struct {
		Product struct {
			Port int `yaml:"port"`
		} `yaml:"product"`
		Order struct {
			Port int `yaml:"port"`
		} `yaml:"order"`
	} `yaml:"services"`
}

var cfg Config

// LoadConfig loads the configuration from config.yaml
func LoadConfig() error {
	f, err := os.Open("config.yaml")
	if err != nil {
		return fmt.Errorf("error opening config file: %v", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&cfg)
	if err != nil {
		return fmt.Errorf("error decoding config: %v", err)
	}

	// Set environment variables
	os.Setenv("DB_HOST", cfg.DB.Host)
	os.Setenv("DB_PORT", fmt.Sprintf("%d", cfg.DB.Port))
	os.Setenv("DB_USER", cfg.DB.User)
	os.Setenv("DB_PASSWORD", cfg.DB.Password)
	os.Setenv("DB_NAME", cfg.DB.Name)

	os.Setenv("KAFKA_BROKERS", cfg.Kafka.Brokers)
	os.Setenv("KAFKA_TOPIC", cfg.Kafka.Topic)

	os.Setenv("RABBITMQ_HOST", cfg.RabbitMQ.Host)
	os.Setenv("RABBITMQ_PORT", fmt.Sprintf("%d", cfg.RabbitMQ.Port))
	os.Setenv("RABBITMQ_USER", cfg.RabbitMQ.User)
	os.Setenv("RABBITMQ_PASSWORD", cfg.RabbitMQ.Password)

	// Set service port environment variables
	os.Setenv("PRODUCT_SERVICE_PORT", fmt.Sprintf("%d", cfg.Services.Product.Port))
	os.Setenv("ORDER_SERVICE_PORT", fmt.Sprintf("%d", cfg.Services.Order.Port))

	return nil
}

// GetConfig returns the loaded configuration
func GetConfig() *Config {
	return &cfg
}
