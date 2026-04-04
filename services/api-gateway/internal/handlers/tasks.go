package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/api-gateway/internal/client"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/api-gateway/internal/models"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

type TaskHandler struct {
	queueClient *client.QueueClient
}

func NewTaskHandler(queueClient *client.QueueClient) *TaskHandler {
	return &TaskHandler{
		queueClient: queueClient,
	}
}

func (h *TaskHandler) Create(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()

	log := logger.WithFields("api-gateway", logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": requestID,
	})

	log.Debug("create task request received")

	if r.Method != http.MethodPost {
		log.Warn("method not allowed")
		h.sendError(w, "method not allowed", http.StatusMethodNotAllowed, requestID)
		return
	}

	var req models.CreateTaskRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.WithError(err).Warn("invalid JSON")
		h.sendError(w, "invalid JSON", http.StatusBadRequest, requestID)
		return
	}

	if req.Type == "" {
		log.Warn("missing task type")
		h.sendError(w, "missing task type", http.StatusBadRequest, requestID)
		return
	}

	task, err := h.queueClient.CreateTask(r.Context(), req.Type, req.Payload)
	if err != nil {
		log.WithError(err).Error("failed to create task in queue service")
		h.sendError(w, "failed to create task", http.StatusInternalServerError, requestID)
		return
	}

	log.WithFields(logrus.Fields{
		"task_id": task.ID,
		"type":    task.ID,
	}).Info("task created successfully")

	response := models.CreateTaskResponse{
		ID:     task.ID,
		Status: task.Status,
	}

	h.sendJSON(w, response, http.StatusCreated, requestID)
}

func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()

	log := logger.WithFields("api-gateway", logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": requestID,
	})

	log.Debug("get task request received")

	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 3 {
		log.Warn("missing task ID", http.StatusBadRequest, requestID)
		return
	}
	taskID := pathParts[2]

	log.WithField("task_id", taskID).Debug("fetching task")

	task, err := h.queueClient.GetTask(r.Context(), taskID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			log.WithField("task_id", taskID).Warn("task not found")
			h.sendError(w, "task not found", http.StatusNotFound, requestID)
		} else {
			log.WithError(err).Error("failed to get task")
			h.sendError(w, "failed to get task", http.StatusInternalServerError, requestID)
		}
		return
	}

	response := models.TaskResponse{
		ID:        task.ID,
		Type:      task.Type,
		Payload:   task.Payload,
		Status:    task.Status,
		CreatedAt: task.CreatedAt,
		UpdatedAt: task.UpdatedAt,
		Error:     task.Error,
	}
	h.sendJSON(w, response, http.StatusOK, requestID)
}

func (h *TaskHandler) List(w http.ResponseWriter, r *http.Request) {
	requestID := uuid.New().String()

	log := logger.WithFields("api-gateway", logrus.Fields{
		"method":     r.Method,
		"path":       r.URL.Path,
		"request_id": requestID,
	})

	log.Debug("list tasks request received")

	limit := 100
	offset := 0

	tasks, err := h.queueClient.ListTasks(r.Context(), limit, offset)
	if err != nil {
		log.WithError(err).Error("failed to list tasks")
		h.sendError(w, "failed to list tasks", http.StatusInternalServerError, requestID)
		return
	}

	responses := make([]models.TaskResponse, 0, len(tasks))
	for _, task := range tasks {
		responses = append(responses, models.TaskResponse{
			ID:        task.ID,
			Type:      task.Type,
			Payload:   task.Payload,
			Status:    task.Status,
			CreatedAt: task.CreatedAt,
			UpdatedAt: task.UpdatedAt,
			Error:     task.Error,
		})
	}

	h.sendJSON(w, responses, http.StatusOK, requestID)
}

func (h *TaskHandler) sendJSON(w http.ResponseWriter, data interface{}, status int, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		log := logger.WithComponent("api-gateway")
		log.WithError(err).Error("failed to encode JSON response")
	}
}

func (h *TaskHandler) sendError(w http.ResponseWriter, message string, code int, requestID string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", requestID)
	w.WriteHeader(code)

	response := models.ErrorResponse{
		Error:     message,
		Code:      code,
		RequestID: requestID,
	}

	err := json.NewEncoder(w).Encode(response)
	if err != nil {
		log := logger.WithComponent("api-gateway")
		log.WithError(err).Error("failed to encode error response")
	}
}
