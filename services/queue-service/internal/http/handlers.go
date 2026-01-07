package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/queue-service/internal/queue"
	"github.com/rs/zerolog"
)

type Server struct {
	log   zerolog.Logger
	store *queue.Store
}

func NewServer() *Server {
	logger.Init()
	log := logger.L().With().Str("service", "queue-service").Logger()

	return &Server{
		log:   log,
		store: queue.NewStore(),
	}
}

func (s *Server) Mux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/internal/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.handleCreateTask(w, r)
		case http.MethodGet:
			s.handleListTasks(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/internal/next-pending", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleNextPending(w, r)
	})

	mux.HandleFunc("/internal/tasks/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/internal/tasks/")
		if path == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}

		if strings.HasSuffix(path, "/status") {
			id := strings.TrimSuffix(path, "/status")
			if id == "" {
				http.Error(w, "missing id", http.StatusBadRequest)
				return
			}
			if r.Method != http.MethodPost {
				http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
				return
			}
			s.handleUpdateStatus(w, r, id)
			return
		}

		id := path
		if r.Method != http.MethodGet {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		s.handleGetTask(w, r, id)
	})

	return mux
}

type createTaskRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type createTaskResponse struct {
	ID string `json:"id"`
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		http.Error(w, "type is required", http.StatusBadRequest)
		return
	}

	task := s.store.CreateTask(req.Type, req.Payload)
	s.log.Info().Str("task_id", task.ID).Str("type", task.Type).Msg("task created")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(createTaskResponse{ID: task.ID})
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	tasks := s.store.ListTasks()

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(tasks)
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request, id string) {
	task, ok := s.store.GetTask(id)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func (s *Server) handleNextPending(w http.ResponseWriter, r *http.Request) {
	task, ok := s.store.NextPending()
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}

func (s *Server) handleUpdateStatus(w http.ResponseWriter, r *http.Request, id string) {
	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	status := taskmodel.Status(req.Status)
	switch status {
	case taskmodel.StatusPending, taskmodel.StatusProcessing, taskmodel.StatusCompleted, taskmodel.StatusFailed:
	default:
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	task, ok := s.store.UpdateStatus(id, status)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	s.log.Info().
		Str("task_id", task.ID).
		Str("status", string(task.Status)).
		Msg("status updated")

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(task)
}
