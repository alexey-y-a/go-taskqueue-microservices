package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Port int

	Server struct {
		ReadTimeout  time.Duration
		WriteTimeout time.Duration
		IdleTimeout  time.Duration
	}

	ShutdownTimeout time.Duration

	Store struct {
		MaxTasks int
		Type     string
	}

	Kafka struct {
		Enabled bool
		Brokers []string
		Topic   string
	}

	Postgres struct {
		Host     string
		Port     int
		User     string
		Password string
		DBName   string
		SSLMode  string

		MaxOpenConns    int
		MaxIdleConns    int
		ConnMaxLifetime time.Duration
		ConnMaxIdleTime time.Duration
	}
}

func New() *Config {
	cfg := &Config{
		Port:            8081,
		ShutdownTimeout: 10 * time.Second,
	}

	cfg.Server.ReadTimeout = 5 * time.Second
	cfg.Server.WriteTimeout = 10 * time.Second
	cfg.Server.IdleTimeout = 120 * time.Second

	cfg.Store.MaxTasks = 10000
	cfg.Store.Type = "memory"

	cfg.Kafka.Enabled = false
	cfg.Kafka.Brokers = []string{"localhost:9092"}
	cfg.Kafka.Topic = "tasks"

	cfg.Postgres.Host = "localhost"
	cfg.Postgres.Port = 5432
	cfg.Postgres.User = "postgres"
	cfg.Postgres.Password = "postgres"
	cfg.Postgres.DBName = "queue_service"
	cfg.Postgres.SSLMode = "disable"

	cfg.Postgres.MaxOpenConns = 25
	cfg.Postgres.MaxIdleConns = 10
	cfg.Postgres.ConnMaxLifetime = 5 * time.Minute
	cfg.Postgres.ConnMaxIdleTime = 2 * time.Minute

	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = port
		}
	}

	if enabled := os.Getenv("KAFKA_ENABLED"); enabled == "true" {
		cfg.Kafka.Enabled = true
	}
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		cfg.Kafka.Brokers = strings.Split(brokers, ",")
	}
	if topic := os.Getenv("KAFKA_TOPIC"); topic != "" {
		cfg.Kafka.Topic = topic
	}

	if storeType := os.Getenv("STORE_TYPE"); storeType != "" {
		cfg.Store.Type = storeType
	}

	if host := os.Getenv("POSTGRES_HOST"); host != "" {
		cfg.Postgres.Host = host
	}

	if portStr := os.Getenv("POSTGRES_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Postgres.Port = port
		}
	}

	if user := os.Getenv("POSTGRES_USER"); user != "" {
		cfg.Postgres.User = user
	}

	if password := os.Getenv("POSTGRES_PASSWORD"); password != "" {
		cfg.Postgres.Password = password
	}

	if dbname := os.Getenv("POSTGRES_DB"); dbname != "" {
		cfg.Postgres.DBName = dbname
	}

	return cfg
}
