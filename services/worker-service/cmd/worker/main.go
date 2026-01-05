package main

import "github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/worker"

func main() {
    w := worker.New()
    w.Run()
}
