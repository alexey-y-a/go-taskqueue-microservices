package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/kafka"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/metrics"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/client"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/config"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/server"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/worker"
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
		"worker_mode":       cfg.WorkerMode,
		"kafka_enabled":     cfg.Kafka.Enabled,
	}).Info("configuration loaded")

	httpClient := client.NewHTTPClient(cfg.Client.QueueServiceURL, cfg.Client.Timeout)

	serviceMetrics := metrics.NewServiceMetrics("worker_service")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kafkaWorker := worker.NewKafkaWorker(httpClient, nil, serviceMetrics)

	if cfg.WorkerMode == "kafka" && cfg.Kafka.Enabled {
		log.Info("Creating Kafka consumer...")

		consumer, err := kafka.NewConsumer(
			kafka.Config{
				Brokers: cfg.Kafka.Brokers,
				Topic:   cfg.Kafka.Topic,
			},
			func(ctx context.Context, event kafka.Event) error {
				return kafkaWorker.HandleEvent(ctx, event)
			},
		)
		if err != nil {
			log.WithError(err).Fatal("Failed to create Kafka consumer")
		}

		kafkaWorker.SetConsumer(consumer)

		if err := kafkaWorker.Start(ctx); err != nil {
			log.WithError(err).Fatal("Failed to start Kafka worker")
		}

		log.Info("Kafka worker started successfully")
	} else {
		log.Info("Kafka mode disabled, worker will run with HTTP server only")
	}

	srv := server.New(cfg)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	errChan := make(chan error, 1)

	go func() {
		if err := srv.Start(ctx); err != nil {
			errChan <- err
		}
	}()

	log.Info("Server is running. Press Ctrl+C to stop.")

	select {
	case sig := <-sigChan:
		log.WithField("signal", sig.String()).Info("Received shutdown signal")
	case err := <-errChan:
		log.WithError(err).Error("Server error")
	}

	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer shutdownCancel()

	if cfg.WorkerMode == "kafka" && cfg.Kafka.Enabled {
		kafkaWorker.Stop()
	}

	if err := srv.Stop(shutdownCtx); err != nil {
		log.WithError(err).Error("Graceful shutdown failed")
	} else {
		log.Info("Graceful shutdown completed")
	}

	time.Sleep(100 * time.Millisecond)
	log.Info("Worker service stopped")
}
