package worker

import (
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/client"
	"github.com/rs/zerolog"
)

type Worker struct {
	log         zerolog.Logger
	queueClient *client.QueueClient
}

func New(queueBaseURL string) *Worker {
	logger.Init()
	log := logger.L().With().Str("service", "worker-service").Logger()

	return &Worker{
		log:         log,
		queueClient: client.NewQueueClient(queueBaseURL),
	}
}

func (w *Worker) Run() {
	w.log.Info().Msg("worker started")

	for {
		task, ok, err := w.queueClient.GetNextPending()
		if err != nil {
			w.log.Error().Err(err).Msg("failed to get next pending task")
			time.Sleep(2 * time.Second)
			continue
		}
		if !ok {
			time.Sleep(2 * time.Second)
			continue
		}

		w.log.Info().Str("task_id", task.ID).Msg("picked up task")

		if err := w.queueClient.UpdateStatus(task.ID, taskmodel.StatusProcessing); err != nil {
			w.log.Error().Err(err).Str("task_id", task.ID).Msg("failed to set status processing")
			time.Sleep(1 * time.Second)
			continue
		}

		time.Sleep(1 * time.Second)

		if err := w.queueClient.UpdateStatus(task.ID, taskmodel.StatusCompleted); err != nil {
			w.log.Error().Err(err).Str("task_id", task.ID).Msg("failed to set status completed")
			time.Sleep(1 * time.Second)
			continue
		}

		w.log.Info().Str("task_id", task.ID).Msg("task completed")
	}
}
