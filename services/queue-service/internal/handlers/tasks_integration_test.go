package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/metrics"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/queue-service/internal/store"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

type TaskHandlerIntegrationSuite struct {
	suite.Suite
	handler *TaskHandler
	store   *store.MemoryStore
	ctx     context.Context
	metrics *metrics.ServiceMetrics
}

func (s *TaskHandlerIntegrationSuite) SetupTest() {
	s.ctx = context.Background()
	s.store = store.NewMemoryStore(100)

	registry := prometheus.NewRegistry()
	s.metrics = metrics.NewServiceMetricsWithRegistry("queue_service", registry)

	s.handler = NewTaskHandler(s.store, s.metrics)
}

func (s *TaskHandlerIntegrationSuite) TestCreateTask() {
	reqBody := CreateTaskRequest{
		Type:    "email",
		Payload: "test payload",
	}
	bodyBytes, err := json.Marshal(reqBody)
	require.NoError(s.T(), err)

	req := httptest.NewRequest(http.MethodPost, "/tasks", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	s.handler.Create(rr, req)

	require.Equal(s.T(), http.StatusCreated, rr.Code)

	var response CreateTaskResponse

	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(s.T(), err)
	require.NotEmpty(s.T(), response.ID)
	require.Equal(s.T(), "email", response.Task.Type)
}

func (s *TaskHandlerIntegrationSuite) TestGetTask() {
	testTask := taskmodel.NewTask(uuid.New().String(), "email", "payload")
	err := s.store.Create(s.ctx, testTask)
	require.NoError(s.T(), err)

	req := httptest.NewRequest(http.MethodGet, "/tasks/"+testTask.ID, nil)
	rr := httptest.NewRecorder()

	s.handler.Get(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	var response taskmodel.Task

	err = json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(s.T(), err)
	require.Equal(s.T(), testTask.ID, response.ID)
	require.Equal(s.T(), testTask.Type, response.Type)
}

func (s *TaskHandlerIntegrationSuite) TestGetTaskNotFound() {
	req := httptest.NewRequest(http.MethodGet, "/tasks/non-existent-id", nil)
	rr := httptest.NewRecorder()

	s.handler.Get(rr, req)

	require.Equal(s.T(), http.StatusNotFound, rr.Code)
}

func (s *TaskHandlerIntegrationSuite) TestListTasks() {
	for i := 0; i < 3; i++ {
		testTask := taskmodel.NewTask(uuid.New().String(), "email", "payload")
		_ = s.store.Create(s.ctx, testTask)
	}

	req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
	rr := httptest.NewRecorder()

	s.handler.List(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	var tasks []*taskmodel.Task
	err := json.NewDecoder(rr.Body).Decode(&tasks)
	require.NoError(s.T(), err)
	require.Len(s.T(), tasks, 3)
}

func (s *TaskHandlerIntegrationSuite) TestGetPending() {
	for i := 0; i < 2; i++ {
		task := taskmodel.NewTask(uuid.New().String(), "email", "payload")
		s.store.Create(s.ctx, task)
	}

	completed := taskmodel.NewTask(uuid.New().String(), "email", "payload")
	s.store.Create(s.ctx, completed)
	completed.UpdateStatus(taskmodel.StatusCompleted, "")
	s.store.Update(s.ctx, completed)

	req := httptest.NewRequest(http.MethodGet, "/tasks/pending", nil)
	rr := httptest.NewRecorder()

	s.handler.GetPending(rr, req)

	require.Equal(s.T(), http.StatusOK, rr.Code)

	var tasks []*taskmodel.Task
	err := json.NewDecoder(rr.Body).Decode(&tasks)
	require.NoError(s.T(), err)
	require.Len(s.T(), tasks, 2)

	for _, task := range tasks {
		require.Equal(s.T(), taskmodel.StatusPending, task.Status)
	}
}

func TestTaskHandlerIntegration(t *testing.T) {
	suite.Run(t, new(TaskHandlerIntegrationSuite))
}
