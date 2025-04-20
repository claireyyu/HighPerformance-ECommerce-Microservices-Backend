package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
}

type ServerConfig struct {
	Port string
	Host string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

type KafkaConfig struct {
	Brokers string
	Topic   string
}

func LoadConfig() (*Config, error) {
	viper.AutomaticEnv()

	config := &Config{
		Server: ServerConfig{
			Port: viper.GetString("PRODUCT_SERVICE_PORT"),
			Host: viper.GetString("PRODUCT_SERVICE_HOST"),
		},
		Database: DatabaseConfig{
			Host:     viper.GetString("PRODUCT_DB_HOST"),
			Port:     viper.GetString("PRODUCT_DB_PORT"),
			User:     viper.GetString("PRODUCT_DB_USER"),
			Password: viper.GetString("PRODUCT_DB_PASSWORD"),
			DBName:   viper.GetString("PRODUCT_DB_NAME"),
		},
		Kafka: KafkaConfig{
			Brokers: viper.GetString("KAFKA_BROKERS"),
			Topic:   viper.GetString("KAFKA_PRODUCT_TOPIC"),
		},
	}

	return config, nil
}
