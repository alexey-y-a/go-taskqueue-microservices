package handlers

import (
	"net/http"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
)

func NotFoundHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log := logger.WithRequest("api-gateway", r.Method, r.URL.Path, "")

		log.Warnf("Route not found: %s %s", r.Method, r.URL.Path)

		http.NotFound(w, r)
	}
}
