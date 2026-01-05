package http

import (
	"net/http"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
)

func NewMux() *http.ServeMux {
    logger.Init()
    log := logger.L().With().Str("service", "queue-service").Logger()
    mux := http.NewServeMux()

    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        log.Info().Str("path", r.URL.Path).Msg("Health check")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("ok"))
    })

    return mux
}
