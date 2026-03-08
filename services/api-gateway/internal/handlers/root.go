package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
)

type RootResponse struct {
	Service string `json:"service"`
	Status  string `json:"status"`
	Version string `json:"version"`
}

func RootHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.WithRequest("api-gateway", r.Method, r.URL.Path, "")

		log.Debug("Root endpoint requested")

		response := RootResponse{
			Service: "api-gateway",
			Status:  "running",
			Version: "v1",
		}

		w.Header().Set("Content-Type", "application/json")

		err := json.NewEncoder(w).Encode(response)
		if err != nil {
			log.WithError(err).Error("Failed to encode JSON response")
			http.Error(w, "Internal server error", http.StatusInternalServerError)

			return
		}

		log.Debug("Root response sent")
	}
}
