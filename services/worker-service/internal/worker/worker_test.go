package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/worker-service/internal/mocks"
	"github.com/golang/mock/gomock"
)

func TestWorker_ProcessTasks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockQueueServiceClient(ctrl)

	task := taskmodel.NewTask("task-1", "email", "test payload")

	mockClient.EXPECT().GetPendingTasks(gomock.Any(), 10).Return([]*taskmodel.Task{task}, nil).Times(1)

	mockClient.EXPECT().UpdateTaskStatus(gomock.Any(), "task-1", taskmodel.StatusCompleted, "").Return(nil).Times(1)

	w := New(mockClient, 10, time.Second, time.Second)

	ctx := context.Background()

	w.Start(ctx)

	time.Sleep(500 * time.Millisecond)

	w.Stop()
}

func TestWorker_ProcessTasks_Failure(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockQueueServiceClient(ctrl)

	task := taskmodel.NewTask("task-1", "email", "")

	mockClient.EXPECT().GetPendingTasks(gomock.Any(), 10).Return([]*taskmodel.Task{task}, nil).Times(1)

	mockClient.EXPECT().UpdateTaskStatus(gomock.Any(), "task-1", taskmodel.StatusFailed, gomock.Any()).Return(nil).Times(1)

	w := New(mockClient, 10, time.Second, time.Second)
	ctx := context.Background()

	w.Start(ctx)
	time.Sleep(500 * time.Millisecond)
	w.Stop()
}

func TestWorker_ProcessTasks_ClientError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockClient := mocks.NewMockQueueServiceClient(ctrl)

	mockClient.EXPECT().GetPendingTasks(gomock.Any(), 10).Return(nil, errors.New("connection failed")).Times(1)

	w := New(mockClient, 10, time.Millisecond*100, time.Millisecond*50)

	ctx := context.Background()

	w.Start(ctx)

	time.Sleep(200 * time.Millisecond)

	w.Stop()
}
