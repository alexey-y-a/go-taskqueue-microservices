package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/metrics"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/client"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/config"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/handlers"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/worker"
	"github.com/sirupsen/logrus"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
	worker     *worker.Worker
}

func New(cfg *config.Config) *Server {
	httpClient := client.NewHttpClient(cfg.Client.QueueServiceURL, cfg.Client.Timeout)

	metricsSvc := metrics.NewServiceMetrics("worker-service")

	w := worker.New(
		httpClient,
		cfg.Worker.BatchSize,
		cfg.Worker.PollInterval,
		cfg.Worker.RetryDelay,
		metricsSvc,
	)

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", handlers.HealthHandler())

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	logger.WithFields("worker-service", logrus.Fields{
		"port":          cfg.Port,
		"read_timeout":  cfg.Server.ReadTimeout,
		"write_timeout": cfg.Server.WriteTimeout,
		"idle_timeout":  cfg.Server.IdleTimeout,
		"poll_interval": cfg.Worker.PollInterval,
		"batch_size":    cfg.Worker.BatchSize,
		"queue_service": cfg.Client.QueueServiceURL,
	}).Info("HTTP server configured")

	return &Server{
		httpServer: httpServer,
		config:     cfg,
		worker:     w,
	}
}

func (s *Server) Start(ctx context.Context) error {
	log := logger.WithComponent("worker-service")
	log.WithField("addr", s.httpServer.Addr).Info("Starting HTTP server")

	s.worker.Start(ctx)

	errChan := make(chan error, 1)
	go func() {
		err := s.httpServer.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("server failed: %w", err)
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		return nil
	}
}

func (s *Server) Stop(ctx context.Context) error {
	log := logger.WithComponent("worker-service")
	log.Info("shutting down")

	s.worker.Stop()

	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Info("shutdown completed")

	return nil
}
