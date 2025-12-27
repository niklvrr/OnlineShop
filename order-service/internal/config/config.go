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

type ZookeeperConfig struct {
	Port string
}

type KafkaConfig struct {
	BrokerId            string
	ZookeeperConnect    string
	Listeners           []string
	AdvertisedListeners []string
	Brokers             []string
}

type Config struct {
	App       AppConfig
	Database  DatabaseConfig
	Zookeeper ZookeeperConfig
	Kafka     KafkaConfig
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
			Name:     getEnv("DB_NAME", "order-service"),
		},
		Zookeeper: ZookeeperConfig{
			Port: getEnv("ZOO_CLIENT_PORT", "9092"),
		},
		Kafka: KafkaConfig{
			BrokerId:            getEnv("KAFKA_BROKER_ID", "1"),
			ZookeeperConnect:    getEnv("KAFKA_ZOOKEEPER_CONNECT", "localhost:2181"),
			Listeners:           []string{getEnv("KAFKA_LISTENERS", "localhost:9092")},
			AdvertisedListeners: []string{getEnv("KAFKA_ADVERTISED_LISTENERS", "localhost:9092")},
			Brokers:             strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
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
