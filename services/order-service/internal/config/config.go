package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Kafka    KafkaConfig    `mapstructure:"kafka"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Host string `mapstructure:"host"`
}

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
}

type KafkaConfig struct {
	Broker string `mapstructure:"broker"`
}

func LoadConfig() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Map environment variables
	viper.BindEnv("server.port", "ORDER_SERVICE_PORT")
	viper.BindEnv("server.host", "ORDER_SERVICE_HOST")
	viper.BindEnv("database.host", "ORDER_DB_HOST")
	viper.BindEnv("database.port", "ORDER_DB_PORT")
	viper.BindEnv("database.user", "ORDER_DB_USER")
	viper.BindEnv("database.password", "ORDER_DB_PASSWORD")
	viper.BindEnv("database.dbname", "ORDER_DB_NAME")
	viper.BindEnv("kafka.broker", "KAFKA_BROKERS")

	// Set default values
	viper.SetDefault("server.port", 8082)
	viper.SetDefault("server.host", "0.0.0.0")
	viper.SetDefault("database.host", "mysql")
	viper.SetDefault("database.port", 3306)
	viper.SetDefault("database.user", "order_user")
	viper.SetDefault("database.password", "order_password")
	viper.SetDefault("database.dbname", "order_db")
	viper.SetDefault("kafka.broker", "kafka:9092")

	// Read environment variables
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
