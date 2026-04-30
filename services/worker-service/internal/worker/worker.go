package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/metrics"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/client"
	"github.com/sirupsen/logrus"
)

type Worker struct {
	queueClient  client.QueueServiceClient
	batchSize    int
	pollInterval time.Duration
	retryDelay   time.Duration
	metrics      *metrics.ServiceMetrics
	stopChan     chan struct{}
	doneChan     chan struct{}
}

func New(
	queueClient client.QueueServiceClient,
	batchSize int,
	pollInterval time.Duration,
	retryDelay time.Duration,
	m *metrics.ServiceMetrics,
) *Worker {
	return &Worker{
		queueClient:  queueClient,
		batchSize:    batchSize,
		pollInterval: pollInterval,
		retryDelay:   retryDelay,
		metrics:      m,
		stopChan:     make(chan struct{}),
		doneChan:     make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context) {
	log := logger.WithComponent("worker")
	log.Info("starting polling worker")

	go w.run(ctx)
}

func (w *Worker) run(ctx context.Context) {
	defer close(w.doneChan)
	log := logger.WithComponent("worker")

	log.Debug("initial poll (immediate)")
	w.poll(ctx)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("worker stopping due to context cancellation")
			return

		case <-w.stopChan:
			log.Info("worker stopping due to stop signal")
			return

		case <-ticker.C:
			log.Debug("poll triggered by ticker")
			w.poll(ctx)
		}
	}
}

func (w *Worker) poll(ctx context.Context) {
	log := logger.WithComponent("worker")

	tasks, err := w.queueClient.GetPendingTasks(ctx, w.batchSize)
	if err != nil {
		log.WithError(err).Error("failed to get pending tasks")
		return
	}

	if len(tasks) == 0 {
		log.Debug("no pending tasks found")
		return
	}

	log.WithField("count", len(tasks)).Info("processing tasks")

	for _, task := range tasks {
		select {
		case <-ctx.Done():
			log.Warn("context cancelled while processing batch, stopping")
			return
		default:
		}

		w.processTask(ctx, task)
	}
}

func (w *Worker) processTask(ctx context.Context, task *taskmodel.Task) {
	log := logger.WithFields("worker", logrus.Fields{
		"task_id": task.ID,
		"type":    task.Type,
	})

	startTime := time.Now()

	log.WithField("status", task.Status).Info("processing task")

	if err := w.executeTask(ctx, task); err != nil {
		log.WithError(err).Error("task execution failed")
		w.metrics.TasksFailed.Inc()

		if updateErr := w.queueClient.UpdateTaskStatus(ctx, task.ID, taskmodel.StatusFailed, err.Error()); updateErr != nil {
			log.WithError(updateErr).Error("failed to update task status to failed")
		} else {
			log.Info("task status updated to failed")
		}

		return
	}

	w.metrics.TasksProcessed.Inc()
	if err := w.queueClient.UpdateTaskStatus(ctx, task.ID, taskmodel.StatusCompleted, ""); err != nil {
		log.WithError(err).Error("failed to update task status to completed")
	} else {
		log.Info("task status updated to completed")
	}

	log.WithField("duration_ms", time.Since(startTime).Milliseconds()).Info("task completed")
}

func (w *Worker) executeTask(ctx context.Context, task *taskmodel.Task) error {
	log := logger.WithFields("worker", logrus.Fields{
		"task_id": task.ID,
		"type":    task.Type,
	})

	log.Debug("executing task logic")

	switch task.Type {
	case "email":
		return w.sendEmail(task.Payload)
	case "report":
		return w.generateReport(task.Payload)
	default:
		return fmt.Errorf("unknown task type: %s", task.Type)
	}
}

func (w *Worker) sendEmail(payload string) error {
	time.Sleep(100 * time.Millisecond)

	if payload == "" {
		return fmt.Errorf("email payload is empty")
	}

	return nil
}

func (w *Worker) generateReport(payload string) error {
	time.Sleep(200 * time.Millisecond)

	if payload == "" {
		return fmt.Errorf("report payload is empty")
	}

	return nil
}

func (w *Worker) Stop() {
	log := logger.WithComponent("worker")
	log.Info("stopping worker")

	close(w.stopChan)
	<-w.doneChan

	log.Info("worker stopped")
}
