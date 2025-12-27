package config

import (
	"github.com/joho/godotenv"
	"os"
)

type Config struct {
	HTTPPort        string
	OrderServiceURL string
	PaymentServiceURL string
}

func LoadConfig() *Config {
	_ = godotenv.Load()
	return &Config{
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		OrderServiceURL: getEnv("ORDER_SERVICE_URL", "localhost:50051"),
		PaymentServiceURL: getEnv("PAYMENT_SERVICE_URL", "localhost:50052"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

