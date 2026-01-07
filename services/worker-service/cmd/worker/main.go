package main

import "github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/worker"

func main() {
	queueBaseURL := "http://queue-service:8081"
	w := worker.New(queueBaseURL)
	w.Run()
}
