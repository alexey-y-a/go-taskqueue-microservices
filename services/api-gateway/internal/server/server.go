package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/api-gateway/internal/config"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/api-gateway/internal/handlers"
	"github.com/sirupsen/logrus"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
}

func New(cfg *config.Config) *Server {
	log := logger.WithComponent("api-gateway")

	mux := http.NewServeMux()

	mux.HandleFunc("/", handlers.RootHandler())
	mux.HandleFunc("/healthz", handlers.HealthHandler())

	mux.HandleFunc("/404", handlers.NotFoundHandler())

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,

		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	log.WithFields(logrus.Fields{
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
