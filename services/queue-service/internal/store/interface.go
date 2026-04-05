package store

import (
	"context"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
)

type TaskStore interface {
	Create(ctx context.Context, task *taskmodel.Task) error
	Get(ctx context.Context, id string) (*taskmodel.Task, error)
	Update(ctx context.Context, task *taskmodel.Task) error
	List(ctx context.Context, limit, offset int) ([]*taskmodel.Task, error)
	GetPending(ctx context.Context, limit int) ([]*taskmodel.Task, error)
	Count(ctx context.Context) (int, error)
	UpdateStatus(ctx context.Context, id string, status taskmodel.TaskStatus, errMsg string) error
}
