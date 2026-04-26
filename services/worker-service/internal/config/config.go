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

	Worker struct {
		PollInterval time.Duration
		BatchSize    int
		RetryDelay   time.Duration
	}

	Client struct {
		QueueServiceURL string
		Timeout         time.Duration
		MaxRetries      int
	}

	WorkerMode string

	Kafka struct {
		Enabled       bool
		Brokers       []string
		Topic         string
		ConsumerGroup string
	}
}

func New() *Config {
	cfg := &Config{
		Port:            8082,
		ShutdownTimeout: 10 * time.Second,
	}

	cfg.Server.ReadTimeout = 5 * time.Second
	cfg.Server.WriteTimeout = 10 * time.Second
	cfg.Server.IdleTimeout = 120 * time.Second

	cfg.Worker.PollInterval = 5 * time.Second
	cfg.Worker.BatchSize = 10
	cfg.Worker.RetryDelay = 1 * time.Second

	cfg.Client.QueueServiceURL = "http://localhost:8081"
	cfg.Client.Timeout = 30 * time.Second
	cfg.Client.MaxRetries = 3

	cfg.WorkerMode = "polling"                     // По умолчанию polling режим
	cfg.Kafka.Enabled = false                      // Kafka выключена по умолчанию
	cfg.Kafka.Brokers = []string{"localhost:9092"} // Брокер по умолчанию
	cfg.Kafka.Topic = "tasks"                      // Топик по умолчанию
	cfg.Kafka.ConsumerGroup = "worker-group"       // Группа потребителей

	if portStr := os.Getenv("PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.Port = port
		}
	}

	if timeoutStr := os.Getenv("READ_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			cfg.Server.ReadTimeout = time.Duration(timeout) * time.Second
		}
	}

	if timeoutStr := os.Getenv("WRITE_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			cfg.Server.WriteTimeout = time.Duration(timeout) * time.Second
		}
	}

	if timeoutStr := os.Getenv("IDLE_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			cfg.Server.IdleTimeout = time.Duration(timeout) * time.Second
		}
	}

	if timeoutStr := os.Getenv("SHUTDOWN_TIMEOUT"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			cfg.ShutdownTimeout = time.Duration(timeout) * time.Second
		}
	}

	if timeoutStr := os.Getenv("POLL_INTERVAL"); timeoutStr != "" {
		if timeout, err := strconv.Atoi(timeoutStr); err == nil {
			cfg.Worker.PollInterval = time.Duration(timeout) * time.Second
		}
	}

	if batchSizeStr := os.Getenv("BATCH_SIZE"); batchSizeStr != "" {
		if batchSize, err := strconv.Atoi(batchSizeStr); err == nil {
			cfg.Worker.BatchSize = batchSize
		}
	}

	if queueURL := os.Getenv("QUEUE_SERVICE_URL"); queueURL != "" {
		cfg.Client.QueueServiceURL = queueURL
	}

	if mode := os.Getenv("WORKER_MODE"); mode != "" {
		cfg.WorkerMode = mode
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

	if group := os.Getenv("KAFKA_CONSUMER_GROUP"); group != "" {
		cfg.Kafka.ConsumerGroup = group
	}

	return cfg
}
