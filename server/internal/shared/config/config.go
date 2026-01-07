package config

import (
	"os"
	"strconv"
)

// Config holds application configuration
type Config struct {
	MongoDB  MongoDBConfig
	RabbitMQ RabbitMQConfig
	SMTP     SMTPConfig
	Server   ServerConfig
}

// MongoDBConfig holds MongoDB configuration
type MongoDBConfig struct {
	URI      string
	Database string
}

// RabbitMQConfig holds RabbitMQ configuration
type RabbitMQConfig struct {
	URL string
}

// SMTPConfig holds SMTP configuration
type SMTPConfig struct {
	Host      string
	Port      int
	Username  string
	Password  string
	FromEmail string
	FromName  string
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() (*Config, error) {
	smtpPort, _ := strconv.Atoi(getEnv("SMTP_PORT", "587"))

	return &Config{
		MongoDB: MongoDBConfig{
			URI:      getEnv("MONGODB_URI", "mongodb://localhost:27017"),
			Database: getEnv("MONGODB_DATABASE", "notification_service"),
		},
		RabbitMQ: RabbitMQConfig{
			URL: getEnv("RABBITMQ_URL", "amqp://guest:guest@localhost:5672/"),
		},
		SMTP: SMTPConfig{
			Host:      getEnv("SMTP_HOST", "smtp.gmail.com"),
			Port:      smtpPort,
			Username:  getEnv("SMTP_USERNAME", ""),
			Password:  getEnv("SMTP_PASSWORD", ""),
			FromEmail: getEnv("SMTP_FROM_EMAIL", "noreply@example.com"),
			FromName:  getEnv("SMTP_FROM_NAME", "Notification Service"),
		},
		Server: ServerConfig{
			Port: getEnv("NOTIFICATION_SERVICE_PORT", "8084"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
