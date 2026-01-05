package main

import (
	"net/http"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
)

func main() {
    logger.Init()

    log := logger.L().With().Str("service", "api-gateway").Logger()

     mux := http.NewServeMux()

     mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
         w.WriteHeader(http.StatusOK)
         _, _ = w.Write([]byte("ok"))
     })

     addr := "8080"
     log.Info().Str("addr", addr).Msg("starting api-gateway")

     err := http.ListenAndServe(addr, mux)
     if err != nil {
         log.Error().Err(err).Msg("api-gateway stopped with error")
     }
}