package queue

import (
	"net/http"

	httpHandlers "github.com/alexey-y-a/go-taskqueue-microservices/services/queue-service/internal/http"
)

func main() {
    s := httpHandlers.NewServer()
    mux := s.Mux()

    addr := ":8081"

    err := http.ListenAndServe(addr, mux)
    if err != nil {
        panic(err)
    }
}