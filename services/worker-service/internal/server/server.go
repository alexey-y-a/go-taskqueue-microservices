package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/metrics"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/config"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/handlers"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/worker"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
	worker     *worker.Worker
	metrics    *metrics.ServiceMetrics
}

func New(cfg *config.Config, w *worker.Worker) *Server {
	return NewWithWorker(cfg, w)
}

func NewWithWorker(cfg *config.Config, w *worker.Worker) *Server {
	serviceMetrics := metrics.NewServiceMetrics("worker_service")

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", handlers.HealthHandler())
	mux.Handle("/metrics", promhttp.Handler())

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
		metrics:    serviceMetrics,
	}
}

func (s *Server) Start(ctx context.Context) error {
	log := logger.WithComponent("worker-service")
	log.WithField("addr", s.httpServer.Addr).Info("Starting HTTP server")

	s.worker.Start(ctx)

	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Error("HTTP server failed")
		}
	}()

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	log := logger.WithComponent("worker-service")
	log.Info("Shutting down HTTP server...")

	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Info("HTTP server stopped")
	return nil
}
