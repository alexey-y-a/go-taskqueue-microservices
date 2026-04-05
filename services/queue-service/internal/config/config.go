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
		Type     string
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

	portStr := os.Getenv("PORT")
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err == nil {
			cfg.Port = port
		}
	}

	storeType := os.Getenv("STORE_TYPE")
	if storeType != "" {
		cfg.Store.Type = storeType
	}

	host := os.Getenv("POSTGRES_HOST")
	if host != "" {
		cfg.Postgres.Host = host
	}

	portStr = os.Getenv("POSTGRES_PORT")
	if portStr != "" {
		port, err := strconv.Atoi(portStr)
		if err == nil {
			cfg.Postgres.Port = port
		}
	}

	user := os.Getenv("POSTGRES_USER")
	if user != "" {
		cfg.Postgres.User = user
	}

	password := os.Getenv("POSTGRES_PASSWORD")
	if password != "" {
		cfg.Postgres.Password = password
	}

	dbname := os.Getenv("POSTGRES_DB")
	if dbname != "" {
		cfg.Postgres.DBName = dbname
	}

	return cfg
}
