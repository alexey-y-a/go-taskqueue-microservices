package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/kafka"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/metrics"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/queue-service/internal/store"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type TaskHandler struct {
	store    store.TaskStore
	metrics  *metrics.ServiceMetrics
	producer *kafka.Producer
}

func NewTaskHandler(store store.TaskStore, m *metrics.ServiceMetrics, producer *kafka.Producer) *TaskHandler {
	return &TaskHandler{
		store:    store,
		metrics:  m,
		producer: producer,
	}
}

type CreateTaskRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type CreateTaskResponse struct {
	ID     string               `json:"id"`
	Status taskmodel.TaskStatus `json:"status"`
	Task   *taskmodel.Task      `json:"task"`
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	log := logger.WithFields("queue-service", logrus.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	log.Debug("Create task request received")

	if r.Method != http.MethodPost {
		log.Warn("Method not allowed")
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateTaskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.WithError(err).Warn("Invalid JSON")
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		log.Warn("Missing task type")
		http.Error(w, "Missing task type", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()

	task := taskmodel.NewTask(id, req.Type, req.Payload)

	err = h.store.Create(r.Context(), task)
	if err != nil {
		log.WithError(err).Error("Failed to save task")

		switch err {
		case store.ErrStoreFull:
			http.Error(w, "Store is full", http.StatusServiceUnavailable)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

		return
	}

	if h.producer != nil {
		eventData := map[string]interface{}{
			"type":    task.Type,
			"payload": task.Payload,
		}
		err := h.producer.SendEvent(r.Context(), "task_created", task.ID, eventData)
		if err != nil {
			log.WithError(err).Warn("Failed to send Kafka event")
		}
	}

	log.WithFields(logrus.Fields{
		"task_id": task.ID,
		"type":    task.Type,
	}).Info("Task created successfully")

	response := CreateTaskResponse{
		ID:     task.ID,
		Status: task.Status,
		Task:   task,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	err = json.NewEncoder(w).Encode(response)
	if err != nil {
		log.WithError(err).Error("Failed to encode response")
	}
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	log := logger.WithFields("queue-service", logrus.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		log.Warn("Missing task ID")
		http.Error(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	id := pathParts[2]

	log.WithField("task_id", id).Debug("Get task request")

	task, err := h.store.Get(r.Context(), id)
	if err != nil {
		if err == store.ErrTaskNotFound {
			log.WithField("task_id", id).Warn("Task not found")
			http.Error(w, "Task not found", http.StatusNotFound)
		} else {
			log.WithError(err).Error("Failed to get task")
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(task)
	if err != nil {
		log.WithError(err).Error("Failed to encode response")
	}
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	log := logger.WithFields("queue-service", logrus.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	log.Debug("List task request")

	tasks, err := h.store.List(r.Context(), 100, 0)
	if err != nil {
		log.WithError(err).Error("Failed to list tasks")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(tasks)
	if err != nil {
		log.WithError(err).Error("Failed to encode response")
	}
}

func (h *TaskHandler) GetPending(w http.ResponseWriter, r *http.Request) {
	log := logger.WithFields("queue-service", logrus.Fields{
		"method": r.Method,
		"path":   r.URL.Path,
	})

	log.Debug("Get pending task request")

	tasks, err := h.store.GetPending(r.Context(), 100)
	if err != nil {
		log.WithError(err).Error("Failed to get pending tasks")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(tasks)
	if err != nil {
		log.WithError(err).Error("Failed to encode response")
	}
}
