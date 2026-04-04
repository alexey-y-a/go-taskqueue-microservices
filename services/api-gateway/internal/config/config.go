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

	Client struct {
		QueueServiceURL string
		Timeout         time.Duration
	}
}

func New() *Config {
	cfg := &Config{
		Port:            8080,
		ShutdownTimeout: 10 * time.Second,
	}

	cfg.Server.ReadTimeout = 5 * time.Second
	cfg.Server.WriteTimeout = 10 * time.Second
	cfg.Server.IdleTimeout = 120 * time.Second

	cfg.Client.QueueServiceURL = "http://localhost:8081"
	cfg.Client.Timeout = 30 * time.Second

	portStr := os.Getenv("PORT")
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err == nil {
			cfg.Port = port
		}
	}

	queueURL := os.Getenv("QUEUE_SERVICE_URL")
	if queueURL != "" {
		cfg.Client.QueueServiceURL = queueURL
	}

	timeoutStr := os.Getenv("CLIENT_TIMEOUT")
	if timeoutStr != "" {
		timeout, err := strconv.Atoi(timeoutStr)
		if err == nil {
			cfg.Client.Timeout = time.Duration(timeout) * time.Second
		}
	}

	return cfg
}
