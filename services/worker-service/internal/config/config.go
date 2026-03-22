package config

import (
	"os"
	"strconv"
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

	portStr := os.Getenv("PORT")
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err == nil {
			cfg.Port = port
		}
	}

	timeoutStr := os.Getenv("READ_TIMEOUT")
	if timeoutStr != "" {
		timeout, err := strconv.Atoi(timeoutStr)
		if err == nil {
			cfg.Server.ReadTimeout = time.Duration(timeout) * time.Second
		}
	}

	timeoutStr = os.Getenv("WRITE_TIMEOUT")
	if timeoutStr != "" {
		timeout, err := strconv.Atoi(timeoutStr)
		if err == nil {
			cfg.Server.WriteTimeout = time.Duration(timeout) * time.Second
		}
	}

	timeoutStr = os.Getenv("IDLE_TIMEOUT")
	if timeoutStr != "" {
		timeout, err := strconv.Atoi(timeoutStr)
		if err == nil {
			cfg.Server.IdleTimeout = time.Duration(timeout) * time.Second
		}
	}

	timeoutStr = os.Getenv("SHUTDOWN_TIMEOUT")
	if timeoutStr != "" {
		timeout, err := strconv.Atoi(timeoutStr)
		if err == nil {
			cfg.ShutdownTimeout = time.Duration(timeout) * time.Second
		}
	}

	timeoutStr = os.Getenv("POLL_INTERVAL")
	if timeoutStr != "" {
		timeout, err := strconv.Atoi(timeoutStr)
		if err == nil {
			cfg.Worker.PollInterval = time.Duration(timeout) * time.Second
		}
	}

	batchSizeStr := os.Getenv("BATCH_SIZE")
	if batchSizeStr != "" {
		batchSize, err := strconv.Atoi(batchSizeStr)
		if err == nil {
			cfg.Worker.BatchSize = batchSize
		}
	}

	queueURL := os.Getenv("QUEUE_SERVICE_URL")
	if queueURL != "" {
		cfg.Client.QueueServiceURL = queueURL
	}

	return cfg
}
