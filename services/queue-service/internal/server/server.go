package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/metrics"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/queue-service/internal/config"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/queue-service/internal/handlers"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/queue-service/internal/store"
	"github.com/sirupsen/logrus"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
	store      *store.MemoryStore
	metrics    *metrics.ServiceMetrics
}

func New(cfg *config.Config) *Server {

	serviceMetrics := metrics.NewServiceMetrics("queue-service")

	store := store.NewMemoryStore(cfg.Store.MaxTasks)

	taskHandler := handlers.NewTaskHandler(store, serviceMetrics)

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", handlers.HealthHandler())
	mux.HandleFunc("POST /tasks", taskHandler.Create)
	mux.HandleFunc("GET /tasks/", taskHandler.Get)
	mux.HandleFunc("GET /tasks", taskHandler.List)
	mux.HandleFunc("GET /tasks/pending", taskHandler.GetPending)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	logger.WithFields("queue-service", logrus.Fields{
		"port":          cfg.Port,
		"read_timeout":  cfg.Server.ReadTimeout,
		"write_timeout": cfg.Server.WriteTimeout,
		"idle_timeout":  cfg.Server.IdleTimeout,
		"max_tasks":     cfg.Store.MaxTasks,
	}).Info("HTTP server configured")

	return &Server{
		httpServer: httpServer,
		config:     cfg,
		store:      store,
		metrics:    serviceMetrics,
	}
}

func (s *Server) Start() error {
	log := logger.WithComponent("queue-service")
	log.WithField("addr", s.httpServer.Addr).Info("starting HTTP server")

	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	log := logger.WithComponent("queue-service")
	log.Info("shutting down HTTP server gracefully")

	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Info("HTTP server stopped")

	return nil
}
