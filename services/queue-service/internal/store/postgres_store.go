package store

import (
	"context"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/queue-service/internal/repository"
	"github.com/jmoiron/sqlx"
)

type PostgresStore struct {
	repo *repository.TaskRepository
}

func NewPostgresStore(db *sqlx.DB) *PostgresStore {
	return &PostgresStore{
		repo: repository.NewTaskRepository(db),
	}
}

func (s *PostgresStore) Create(ctx context.Context, task *taskmodel.Task) error {
	return s.repo.Create(ctx, task)
}

func (s *PostgresStore) Get(ctx context.Context, id string) (*taskmodel.Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *PostgresStore) Update(ctx context.Context, task *taskmodel.Task) error {
	return s.repo.Update(ctx, task)
}

func (s *PostgresStore) List(ctx context.Context, limit, offset int) ([]*taskmodel.Task, error) {
	return s.repo.List(ctx, limit, offset)
}

func (s *PostgresStore) GetPending(ctx context.Context, limit int) ([]*taskmodel.Task, error) {
	return s.repo.GetPending(ctx, limit)
}

func (s *PostgresStore) Count(ctx context.Context) (int, error) {
	return s.repo.Count(ctx)
}

func (s *PostgresStore) UpdateStatus(ctx context.Context, id string, status taskmodel.TaskStatus, errMsg string) error {
	return s.repo.UpdateStatus(ctx, id, status, errMsg)
}

func (s *PostgresStore) Close() error {
	return nil
}
