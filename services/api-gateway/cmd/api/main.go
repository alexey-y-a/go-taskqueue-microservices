package main

import (
	"net/http"

	httpHandlers "github.com/alexey-y-a/go-taskqueue-microservices/services/api-gateway/internal/http"
)

func main() {
	queueBaseURL := "http://localhost:8081"
	s := httpHandlers.NewServer(queueBaseURL)
	mux := s.Mux()

	addr := ":8080"
	if err := http.ListenAndServe(addr, mux); err != nil {
		panic(err)
	}
}
