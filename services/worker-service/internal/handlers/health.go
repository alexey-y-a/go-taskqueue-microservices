package handlers

import (
	"net/http"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/sirupsen/logrus"
)

func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.WithFields("worker-service", logrus.Fields{
			"method": r.Method,
			"path":   r.URL.Path,
		})

		log.Debug("health check requested")

		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)

		_, err := w.Write([]byte("ok"))
		if err != nil {
			log.WithError(err).Error("failed to write response")
		}
	}
}
