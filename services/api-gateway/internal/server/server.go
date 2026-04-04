package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/api-gateway/internal/client"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/api-gateway/internal/config"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/api-gateway/internal/handlers"
	"github.com/sirupsen/logrus"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
}

func New(cfg *config.Config) *Server {

	queueClient := client.NewQueueClient(cfg.Client.QueueServiceURL, cfg.Client.Timeout)

	rootHandler := handlers.RootHandler()
	healthHandler := handlers.HealthHandler()
	taskHandler := handlers.NewTaskHandler(queueClient)

	mux := http.NewServeMux()

	mux.HandleFunc("/", rootHandler)
	mux.HandleFunc("/healthz", healthHandler)

	mux.HandleFunc("POST /tasks", taskHandler.Create)
	mux.HandleFunc("GET /tasks/", taskHandler.Get)
	mux.HandleFunc("GET /tasks", taskHandler.List)

	httpServer := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mux,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	logger.WithFields("api-gateway", logrus.Fields{
		"port":          cfg.Port,
		"read_timeout":  cfg.Server.ReadTimeout,
		"write_timeout": cfg.Server.WriteTimeout,
		"idle_timeout":  cfg.Server.IdleTimeout,
	}).Info("HTTP server configured")

	return &Server{
		httpServer: httpServer,
		config:     cfg,
	}
}

func (s *Server) Start() error {
	log := logger.WithComponent("api-gateway")
	log.WithField("addr", s.httpServer.Addr).Info("Starting HTTP server")

	err := s.httpServer.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed: %w", err)
	}

	return nil
}

func (s *Server) Stop(ctx context.Context) error {
	log := logger.WithComponent("api-gateway")
	log.Info("Shutting down HTTP server gracefully...")

	err := s.httpServer.Shutdown(ctx)
	if err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	log.Info("HTTP server stopped")

	return nil
}
