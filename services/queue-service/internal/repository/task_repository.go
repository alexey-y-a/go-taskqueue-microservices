package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
	"github.com/jmoiron/sqlx"
)

type TaskRepository struct {
	db *sqlx.DB
}

func NewTaskRepository(db *sqlx.DB) *TaskRepository {
	return &TaskRepository{db: db}
}

func (r *TaskRepository) Create(ctx context.Context, task *taskmodel.Task) error {
	query := `
              IBSERT INTO tasks (id, type, payload, status, error, created_at, updated_at, completed_at)
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `

	_, err := r.db.ExecContext(ctx, query,
		task.ID,
		task.Type,
		task.Payload,
		task.Status,
		task.Error,
		task.CreatedAt,
		task.UpdatedAt,
		task.CompletedAt,
	)
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}

	return nil
}

func (r *TaskRepository) GetByID(ctx context.Context, id string) (*taskmodel.Task, error) {
	query := `SELECT id, type, payload, status, error, created_at, updated_at, completed_at FROM tasks WHERE id = $1`

	var task taskmodel.Task
	err := r.db.QueryRowxContext(ctx, query, id).StructScan(&task)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("task not found: %w", err)
		}
		return nil, fmt.Errorf("get task: %w", err)
	}

	return &task, nil
}

func (r *TaskRepository) Update(ctx context.Context, task *taskmodel.Task) error {
	query := `
              UPDATE tasks
              SET type = $1, payload = $2, status = $3, error = $4, updated_at = $5, completed_at = $6
              WHERE id = $7
    `

	result, err := r.db.ExecContext(ctx, query,
		task.Type,
		task.Payload,
		task.Status,
		task.Error,
		task.UpdatedAt,
		task.CompletedAt,
		task.ID,
	)

	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}

func (r *TaskRepository) List(ctx context.Context, limit, offset int) ([]*taskmodel.Task, error) {
	query := `
              SELECT id, type, payload, status, error, created_at, updated_at, completed_at
              FROM tasks
              ORDER BY created_at DESC 
              LIMIT $1 OFFSET $2
    `

	tasks := []*taskmodel.Task{}
	err := r.db.SelectContext(ctx, &tasks, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get pending tasks: %w", err)
	}

	return tasks, nil
}

func (r *TaskRepository) GetPending(ctx context.Context, limit int) ([]*taskmodel.Task, error) {
	query := `
		SELECT id, type, payload, status, error, created_at, updated_at, completed_at 
		FROM tasks 
		WHERE status = 'pending' 
		ORDER BY created_at ASC 
		LIMIT $1
	`

	tasks := []*taskmodel.Task{}
	err := r.db.SelectContext(ctx, &tasks, query, limit)
	if err != nil {
		return nil, fmt.Errorf("get pending tasks: %w", err)
	}

	return tasks, nil
}

func (r *TaskRepository) Count(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM tasks`

	var count int
	err := r.db.QueryRowxContext(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count tasks: %w", err)
	}

	return count, nil
}

func (r *TaskRepository) UpdateStatus(ctx context.Context, id string, status taskmodel.TaskStatus, errMsg string) error {
	query := `
              UPDATE tasks
              SET status = $1, error = $2, updated_at = $3,
                  completed_at = CASE WHEN $1 IN ('completed', 'failed') THEN $3 ELSE NULL END
              WHERE id = $4
   `

	now := time.Now().UTC()
	result, err := r.db.ExecContext(ctx, query, status, errMsg, now, id)
	if err != nil {
		return fmt.Errorf("update task: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected: %w", err)
	}

	if rows == 0 {
		return fmt.Errorf("task not found")
	}

	return nil
}
