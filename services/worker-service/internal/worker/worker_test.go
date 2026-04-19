package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/metrics"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/mocks"
	"github.com/golang/mock/gomock"
	"github.com/prometheus/client_golang/prometheus"
)

func TestWorker_ProcessTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockQueueServiceClient(ctrl)

	task := taskmodel.NewTask("task-1", "email", "test payload")

	mockClient.EXPECT().GetPendingTasks(gomock.Any(), 10).Return([]*taskmodel.Task{task}, nil).MinTimes(1)

	mockClient.EXPECT().UpdateTaskStatus(gomock.Any(), "task-1", taskmodel.StatusCompleted, "").Return(nil).MinTimes(1)

	registry := prometheus.NewRegistry()
	metricsSvc := metrics.NewServiceMetricsWithRegistry("test_worker", registry)
	w := New(mockClient, 10, 100*time.Millisecond, 50*time.Millisecond, metricsSvc)

	ctx := context.Background()

	w.Start(ctx)

	time.Sleep(250 * time.Millisecond)

	w.Stop()
}

func TestWorker_ProcessTasks_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockQueueServiceClient(ctrl)

	task := taskmodel.NewTask("task-1", "email", "")

	mockClient.EXPECT().GetPendingTasks(gomock.Any(), 10).Return([]*taskmodel.Task{task}, nil).MinTimes(1)

	mockClient.EXPECT().UpdateTaskStatus(gomock.Any(), "task-1", taskmodel.StatusFailed, gomock.Any()).Return(nil).MinTimes(1)

	registry := prometheus.NewRegistry()
	metricsSvc := metrics.NewServiceMetricsWithRegistry("test_worker", registry)
	w := New(mockClient, 10, 100*time.Millisecond, 50*time.Millisecond, metricsSvc)
	ctx := context.Background()

	w.Start(ctx)
	time.Sleep(250 * time.Millisecond)
	w.Stop()
}

func TestWorker_ProcessTasks_ClientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockQueueServiceClient(ctrl)

	mockClient.EXPECT().GetPendingTasks(gomock.Any(), 10).Return(nil, errors.New("connection failed")).MinTimes(1)

	registry := prometheus.NewRegistry()
	metricsSvc := metrics.NewServiceMetricsWithRegistry("test_worker", registry)
	w := New(mockClient, 10, 100*time.Millisecond, 50*time.Millisecond, metricsSvc)

	ctx := context.Background()

	w.Start(ctx)

	time.Sleep(200 * time.Millisecond)

	w.Stop()
}
