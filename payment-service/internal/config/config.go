package config

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strings"
)

var (
	dbUserEmptyError = errors.New("DB User is Empty")
	dbNameEmptyError = errors.New("DB Name is Empty")
)

type AppConfig struct {
	EnvLevel string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	URL      string
}

type KafkaConfig struct {
	Brokers []string
}

type GrpcConfig struct {
	Port string
}

type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Kafka    KafkaConfig
	Grpc     GrpcConfig
}

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()
	c := &Config{
		App: AppConfig{
			EnvLevel: getEnv("APP_ENV_LEVEL", "debug"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			Name:     getEnv("DB_NAME", "payment-service"),
		},
		Kafka: KafkaConfig{
			Brokers: strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		},
		Grpc: GrpcConfig{
			Port: getEnv("GRPC_PORT", "50052"),
		},
	}

	err := makeDbUrl(c)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func makeDbUrl(cfg *Config) error {
	if cfg.Database.URL == "" {
		if cfg.Database.User == "" {
			return dbUserEmptyError
		}
		if cfg.Database.Name == "" {
			return dbNameEmptyError
		}
		cfg.Database.URL = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
			cfg.Database.User,
			cfg.Database.Password,
			cfg.Database.Host,
			cfg.Database.Port,
			cfg.Database.Name,
		)
	}
	return nil
}

