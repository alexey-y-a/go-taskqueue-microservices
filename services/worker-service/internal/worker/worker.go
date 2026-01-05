package worker

import (
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/rs/zerolog"
)

type Worker struct {
    log zerolog.Logger
}

func New() *Worker {
    logger.Init()
    log := logger.L().With().Str("service", "worker-service").Logger()
    return &Worker{log: log}
}

func (w *Worker) Run() {
    w.log.Info().Msg("worker started")
    for {
        w.log.Info().Msg("worker tick: here we'll poll queue-service for tasks")
        time.Sleep(5 * time.Second)
    }
}