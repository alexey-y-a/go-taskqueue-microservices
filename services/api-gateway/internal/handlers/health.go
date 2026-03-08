package handlers

import (
	"net/http"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
)

func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.WithRequest("api-gateway", r.Method, r.URL.Path, "")

		log.Debug("Health check requested")

		w.Header().Set("Content-Type", "text/plain")

		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte("ok"))
		if err != nil {
			log.WithError(err).Errorf("Failed to write health check response")
		}

		log.Debug("Health check completed")

	}
}
