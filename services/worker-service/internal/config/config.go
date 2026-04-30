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

	ClickHouse struct {
		Enabled  bool
		Host     string
		Port     int
		Database string
		User     string
		Password string
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

	cfg.WorkerMode = "polling"
	cfg.Kafka.Enabled = false
	cfg.Kafka.Brokers = []string{"localhost:9092"}
	cfg.Kafka.Topic = "tasks"
	cfg.Kafka.ConsumerGroup = "worker-group"

	cfg.ClickHouse.Enabled = false
	cfg.ClickHouse.Host = "localhost"
	cfg.ClickHouse.Port = 9000
	cfg.ClickHouse.Database = "analytics"
	cfg.ClickHouse.User = "default"
	cfg.ClickHouse.Password = ""

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

	if enabled := os.Getenv("CLICKHOUSE_ENABLED"); enabled == "true" {
		cfg.ClickHouse.Enabled = true
	}

	if host := os.Getenv("CLICKHOUSE_HOST"); host != "" {
		cfg.ClickHouse.Host = host
	}
	if portStr := os.Getenv("CLICKHOUSE_PORT"); portStr != "" {
		if port, err := strconv.Atoi(portStr); err == nil {
			cfg.ClickHouse.Port = port
		}
	}
	if db := os.Getenv("CLICKHOUSE_DATABASE"); db != "" {
		cfg.ClickHouse.Database = db
	}
	if user := os.Getenv("CLICKHOUSE_USER"); user != "" {
		cfg.ClickHouse.User = user
	}
	if pass := os.Getenv("CLICKHOUSE_PASSWORD"); pass != "" {
		cfg.ClickHouse.Password = pass
	}

	return cfg
}
