package client

import (
	"context"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
)

//go:generate mockgen -source=$GOFILE -destination=../mocks/client_mock.gen.go -package=mocks
type QueueServiceClient interface {
	GetPendingTasks(ctx context.Context, limit int) ([]*taskmodel.Task, error)

	UpdateTaskStatus(ctx context.Context, taskID string, status taskmodel.TaskStatus, errMsg string) error
}
