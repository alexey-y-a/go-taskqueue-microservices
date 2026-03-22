package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/config"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/server"
	"github.com/sirupsen/logrus"
)

func main() {
	logger.Init()

	log := logger.WithComponent("worker-service")

	log.Info("starting worker-service")

	cfg := config.New()

	log.WithFields(logrus.Fields{
		"port":              cfg.Port,
		"shutdown_timeout":  cfg.ShutdownTimeout,
		"poll_interval":     cfg.Worker.PollInterval,
		"batch_size":        cfg.Worker.BatchSize,
		"queue_service_url": cfg.Client.QueueServiceURL,
	}).Info("configuration loaded")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv := server.New(cfg)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)

	go func() {
		err := srv.Start(ctx)
		if err != nil {
			errChan <- err
		}
	}()

	log.Info("server is running. Press Ctrl+C to stop.")

	select {
	case sig := <-sigChan:
		log.WithField("signal", sig.String()).Info("received shutdown signal")
	case err := <-errChan:
		log.WithError(err).Error("server error")
	}

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	err := srv.Stop(shutdownCtx)
	if err != nil {
		log.WithError(err).Error("graceful shutdown failed")
	} else {
		log.Info("graceful shutdown completed")
	}

	time.Sleep(100 * time.Millisecond)
	log.Info("worker service stopped")
}
