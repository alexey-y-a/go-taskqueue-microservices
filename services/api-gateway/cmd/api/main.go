package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/clickhouse"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/api-gateway/internal/config"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/api-gateway/internal/server"
	"github.com/sirupsen/logrus"
)

func main() {
	logger.Init()

	log := logger.WithComponent("api-gateway")

	log.Info("Starting API Gateway service")

	cfg := config.New()

	var chClient *clickhouse.Client
	var err error
	if cfg.ClickHouse.Enabled {
		chClient, err = clickhouse.NewClient(clickhouse.Config{
			Host:     cfg.ClickHouse.Host,
			Port:     cfg.ClickHouse.Port,
			Database: cfg.ClickHouse.Database,
			User:     cfg.ClickHouse.User,
			Password: cfg.ClickHouse.Password,
		})
		if err != nil {
			log.WithError(err).Warn("Failed to connect to ClickHouse, analytics disabled")
		}
	}

	log.WithFields(logrus.Fields{
		"port":             cfg.Port,
		"shutdown_timeout": cfg.ShutdownTimeout,
	}).Info("Configuration loaded")

	srv := server.New(cfg, chClient)

	sigChan := make(chan os.Signal, 1)

	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		err := srv.Start()
		if err != nil {
			log.WithError(err).Error("Server failed")
			sigChan <- syscall.SIGTERM
		}
	}()

	log.Info("Server is running. Press Ctrl+C to stop.")

	sig := <-sigChan
	log.WithField("signal", sig.String()).Info("Received shutdown signal")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	err = srv.Stop(ctx)
	if err != nil {
		log.WithError(err).Error("Graceful shutdown failed")
	} else {
		log.Info("Graceful shutdown completed")
	}

	time.Sleep(100 * time.Millisecond)

	log.Info("Api Gateway stopped")

}
