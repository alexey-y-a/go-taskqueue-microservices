package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/queue-service/internal/config"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/queue-service/internal/server"
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

	srv := server.New(cfg)

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
