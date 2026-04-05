package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/queue-service/internal/config"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/queue-service/internal/db"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/queue-service/internal/server"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/queue-service/internal/store"
	"github.com/sirupsen/logrus"
)

func main() {
	logger.Init()

	log := logger.WithComponent("queue-service")

	log.Info("Starting queue service")

	cfg := config.New()

	log.WithFields(logrus.Fields{
		"port":             cfg.Port,
		"shutdown_timeout": cfg.ShutdownTimeout,
		"max_tasks":        cfg.Store.MaxTasks,
	}).Info("Configuration loaded")

	var taskStore store.TaskStore

	if cfg.Store.Type == "postgres" {
		log.Info("Using PostgreSQL storage")

		dbConfig := db.Config{
			Host:            cfg.Postgres.Host,
			Port:            cfg.Postgres.Port,
			User:            cfg.Postgres.User,
			Password:        cfg.Postgres.Password,
			DBName:          cfg.Postgres.DBName,
			SSLMode:         cfg.Postgres.SSLMode,
			MaxOpenConns:    cfg.Postgres.MaxOpenConns,
			MaxIdleConns:    cfg.Postgres.MaxIdleConns,
			ConnMaxLifetime: cfg.Postgres.ConnMaxLifetime,
			ConnMaxIdleTime: cfg.Postgres.ConnMaxIdleTime,
		}

		dbConn, err := db.NewConnection(dbConfig)
		if err != nil {
			log.WithError(err).Fatal("Failed to connect to database")
		}

		taskStore = store.NewPostgresStore(dbConn)
		log.Info("Connected to database")
	} else {
		log.Info("Using in-memory storage")
		taskStore = store.NewMemoryStore(cfg.Store.MaxTasks)
	}

	srv := server.New(cfg, taskStore)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		err := srv.Start()
		if err != nil {
			log.WithError(err).Error("Failed to start server")
			sigChan <- syscall.SIGTERM
		}
	}()

	log.Info("Server is running. Press Ctrl+C to stop.")

	sig := <-sigChan
	log.WithField("signal", sig).Info("Received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	err := srv.Stop(ctx)
	if err != nil {
		log.WithError(err).Error("Graceful shutdown failed")
	} else {
		log.Info("Graceful shutdown succeeded")
	}

	time.Sleep(100 * time.Millisecond)
	log.Info("Queue service stopped")

}
