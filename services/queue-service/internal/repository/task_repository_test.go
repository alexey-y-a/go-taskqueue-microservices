package repository

import (
	"context"
	"testing"
	"time"

	_ "github.com/lib/pq"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type TaskRepositoryTestSuite struct {
	suite.Suite
	container testcontainers.Container
	db        *sqlx.DB
	repo      *TaskRepository
	ctx       context.Context
}

func (s *TaskRepositoryTestSuite) SetupSuite() {
	s.ctx = context.Background()

	postgresContainer, err := postgres.RunContainer(s.ctx,
		testcontainers.WithImage("postgres:17-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpassword"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second)),
	)
	require.NoError(s.T(), err)

	s.container = postgresContainer

	connStr, err := postgresContainer.ConnectionString(s.ctx, "sslmode=disable")
	require.NoError(s.T(), err)

	db, err := sqlx.Connect("postgres", connStr)
	require.NoError(s.T(), err)

	s.db = db

	createTableSQL := `
CREATE TABLE IF NOT EXISTS tasks (
    id VARCHAR(36) PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    payload TEXT NOT NULL,
    status VARCHAR(20) NOT NULL,
    error TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP WITH TIME ZONE
);`
	_, err = s.db.ExecContext(s.ctx, createTableSQL)
	require.NoError(s.T(), err)

	s.repo = NewTaskRepository(s.db)
}

func (s *TaskRepositoryTestSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.container != nil {
		s.container.Terminate(s.ctx)
	}
}

func (s *TaskRepositoryTestSuite) SetupTest() {
	_, err := s.db.ExecContext(s.ctx, "TRUNCATE tasks")
	require.NoError(s.T(), err)
}

func (s *TaskRepositoryTestSuite) TestCreate() {
	task := taskmodel.NewTask(uuid.New().String(), "email", "test payload")
	err := s.repo.Create(s.ctx, task)
	require.NoError(s.T(), err)

	saved, err := s.repo.GetByID(s.ctx, task.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), task.ID, saved.ID)
	require.Equal(s.T(), task.Type, saved.Type)
	require.Equal(s.T(), task.Payload, saved.Payload)
	require.Equal(s.T(), task.Status, saved.Status)
}

func (s *TaskRepositoryTestSuite) TestGetByID() {
	task := taskmodel.NewTask(uuid.New().String(), "email", "test payload")
	err := s.repo.Create(s.ctx, task)
	require.NoError(s.T(), err)

	found, err := s.repo.GetByID(s.ctx, task.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), task.ID, found.ID)

	_, err = s.repo.GetByID(s.ctx, "non-existent-id")
	require.Error(s.T(), err)
	require.Contains(s.T(), err.Error(), "task not found")
}

func (s *TaskRepositoryTestSuite) TestUpdateStatus() {
	task := taskmodel.NewTask(uuid.New().String(), "email", "test payload")
	err := s.repo.Create(s.ctx, task)
	require.NoError(s.T(), err)

	err = s.repo.UpdateStatus(s.ctx, task.ID, taskmodel.StatusCompleted, "")
	require.NoError(s.T(), err)

	updated, err := s.repo.GetByID(s.ctx, task.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), taskmodel.StatusCompleted, updated.Status)
	require.NotNil(s.T(), updated.CompletedAt)

	err = s.repo.UpdateStatus(s.ctx, task.ID, taskmodel.StatusFailed, "something went wrong")
	require.NoError(s.T(), err)

	failed, err := s.repo.GetByID(s.ctx, task.ID)
	require.NoError(s.T(), err)
	require.Equal(s.T(), taskmodel.StatusFailed, failed.Status)
	require.Equal(s.T(), "something went wrong", failed.Error)
}

func (s *TaskRepositoryTestSuite) TestList() {
	for i := 0; i < 5; i++ {
		task := taskmodel.NewTask(uuid.New().String(), "email", "payload")
		err := s.repo.Create(s.ctx, task)
		require.NoError(s.T(), err)
	}

	tasks, err := s.repo.List(s.ctx, 3, 0)
	require.NoError(s.T(), err)
	require.Len(s.T(), tasks, 3)

	tasks, err = s.repo.List(s.ctx, 2, 3)
	require.NoError(s.T(), err)
	require.Len(s.T(), tasks, 2)
}

func (s *TaskRepositoryTestSuite) TestGetPending() {
	for i := 0; i < 3; i++ {
		task := taskmodel.NewTask(uuid.New().String(), "email", "payload")
		err := s.repo.Create(s.ctx, task)
		require.NoError(s.T(), err)
	}

	for i := 0; i < 2; i++ {
		task := taskmodel.NewTask(uuid.New().String(), "email", "payload")
		err := s.repo.Create(s.ctx, task)
		require.NoError(s.T(), err)
		err = s.repo.UpdateStatus(s.ctx, task.ID, taskmodel.StatusCompleted, "")
		require.NoError(s.T(), err)
	}

	pending, err := s.repo.GetPending(s.ctx, 10)
	require.NoError(s.T(), err)
	require.Len(s.T(), pending, 3)

	for _, task := range pending {
		require.Equal(s.T(), taskmodel.StatusPending, task.Status)
	}
}

func (s *TaskRepositoryTestSuite) TestConcurrentAccess() {
	task := taskmodel.NewTask(uuid.New().String(), "email", "payload")
	err := s.repo.Create(s.ctx, task)
	require.NoError(s.T(), err)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			err := s.repo.UpdateStatus(s.ctx, task.ID, taskmodel.StatusCompleted, "")
			require.NoError(s.T(), err)
			done <- true
		}()
	}

	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestTaskRepository(t *testing.T) {
	suite.Run(t, new(TaskRepositoryTestSuite))
}
