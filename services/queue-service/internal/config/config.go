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

	Store struct {
		MaxTasks int
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

	maxTasksStr := os.Getenv("MAX_TASKS")
	if maxTasksStr != "" {
		maxTasks, err := strconv.Atoi(maxTasksStr)
		if err == nil {
			cfg.Store.MaxTasks = maxTasks
		}
	}
	return cfg
}
