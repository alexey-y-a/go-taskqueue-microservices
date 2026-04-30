package worker

import (
	"context"
	"fmt"
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/clickhouse"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/kafka"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/metrics"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/client"
	"github.com/sirupsen/logrus"
)

type KafkaWorker struct {
	queueClient      client.QueueServiceClient
	consumer         *kafka.Consumer
	metrics          *metrics.ServiceMetrics
	clickhouseClient *clickhouse.Client
	stopChan         chan struct{}
	doneChan         chan struct{}
}

func NewKafkaWorker(
	queueClient client.QueueServiceClient,
	consumer *kafka.Consumer,
	m *metrics.ServiceMetrics,
	chClient *clickhouse.Client,
) *KafkaWorker {
	return &KafkaWorker{
		queueClient:      queueClient,
		consumer:         consumer,
		stopChan:         make(chan struct{}),
		doneChan:         make(chan struct{}),
		metrics:          m,
		clickhouseClient: chClient,
	}
}

func (w *KafkaWorker) SetConsumer(consumer *kafka.Consumer) {
	w.consumer = consumer
}

func (w *KafkaWorker) HandleEvent(ctx context.Context, event kafka.Event) error {
	return w.handleEvent(ctx, event)
}

func (w *KafkaWorker) Start(ctx context.Context) error {
	log := logger.WithComponent("kafka-worker")
	log.Info("starting kafka worker")

	go w.run(ctx)

	return nil
}

func (w *KafkaWorker) run(ctx context.Context) {
	defer close(w.doneChan)

	log := logger.WithComponent("kafka-worker")
	log.Info("kafka worker running, waiting for message...")

	err := w.consumer.Start(ctx)
	if err != nil {
		log.WithError(err).Error("failed to start kafka consumer")
		return
	}

	select {
	case <-ctx.Done():
		log.Info("worker stopping due to context cancellation")
		return

	case <-w.stopChan:
		log.Info("worker stopping due to stop signal")
	}
}

func (w *KafkaWorker) handleEvent(ctx context.Context, event kafka.Event) error {
	log := logger.WithFields("kafka-worker", logrus.Fields{
		"event_type": event.EventType,
		"task_id":    event.TaskID,
	})

	log.Info("processing task from kafka")

	task, err := w.queueClient.GetTask(ctx, event.TaskID)
	if err != nil {
		log.WithError(err).Error("failed to get task")
		return err
	}

	startTime := time.Now()

	if err := w.executeTask(ctx, task); err != nil {
		log.WithError(err).Error("failed to execute task")
		w.metrics.TasksFailed.Inc()

		if updateErr := w.queueClient.UpdateTaskStatus(ctx, task.ID, taskmodel.StatusFailed, err.Error()); updateErr != nil {
			log.WithError(updateErr).Error("failed to update task status")
		}

		if w.clickhouseClient != nil {
			duration := uint32(time.Since(startTime).Milliseconds())
			task.Status = taskmodel.StatusFailed
			task.Error = err.Error()
			go w.sendAnalyticsEvent(context.Background(), task, duration)
		}
		return err
	}

	w.metrics.TasksProcessed.Inc()
	if err := w.queueClient.UpdateTaskStatus(ctx, task.ID, taskmodel.StatusCompleted, ""); err != nil {
		log.WithError(err).Error("failed to update task status")
		return err
	}

	if w.clickhouseClient != nil {
		duration := uint32(time.Since(startTime).Milliseconds())
		task.Status = taskmodel.StatusCompleted
		go w.sendAnalyticsEvent(context.Background(), task, duration)
	}

	log.Info("task successfully processed")
	return nil
}

func (w *KafkaWorker) sendAnalyticsEvent(ctx context.Context, task *taskmodel.Task, processingTimeMs uint32) {
	chEvent := clickhouse.TaskEvent{
		EventDate:        time.Now().UTC(),
		EventTime:        time.Now().UTC(),
		TaskID:           task.ID,
		TaskType:         task.Type,
		EventType:        string(task.Status),
		WorkerID:         "worker-1",
		ProcessingTimeMs: processingTimeMs,
		RetryCount:       0,
		ErrorMessage:     task.Error,
	}

	if err := w.clickhouseClient.InsertTaskEvent(ctx, chEvent); err != nil {
		log := logger.WithComponent("kafka-worker")
		log.WithError(err).Warn("Failed to send analytics event")
	}
}

func (w *KafkaWorker) executeTask(ctx context.Context, task *taskmodel.Task) error {
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

func (w *KafkaWorker) sendEmail(payload string) error {
	time.Sleep(100 * time.Millisecond)

	if payload == "" {
		return fmt.Errorf("email payload is empty")
	}

	return nil
}

func (w *KafkaWorker) generateReport(payload string) error {
	time.Sleep(200 * time.Millisecond)

	if payload == "" {
		return fmt.Errorf("report payload is empty")
	}

	return nil
}

func (w *KafkaWorker) Stop() {
	log := logger.WithComponent("worker")
	log.Info("stopping worker")

	close(w.stopChan)
	<-w.doneChan

	if w.consumer != nil {
		w.consumer.Stop()
	}

	log.Info("worker stopped")
}
