package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
	"github.com/alexey-y-a/go-taskqueue-microservices/services/api-gateway/internal/client"
	"github.com/rs/zerolog"
)

type Server struct {
	log         zerolog.Logger
	queueClient *client.QueueClient
    queueBaseURL string
}

func NewServer(queueBaseURL string) *Server {
	logger.Init()
	log := logger.L().With().Str("service", "api-gateway").Logger()

	return &Server{
		log:          log,
		queueClient:  client.NewQueueClient(queueBaseURL),
		queueBaseURL: queueBaseURL,
	}
}

func (s *Server) Mux() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			s.handleCreateTask(w, r)
		case http.MethodGet:
			s.handleListTasks(w, r)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/tasks/", func(w http.ResponseWriter, r *http.Request) {
		id := strings.TrimPrefix(r.URL.Path, "/tasks/")
		if id == "" {
			http.Error(w, "missing id", http.StatusBadRequest)
			return
		}
		switch r.Method {
		case http.MethodGet:
			s.handleGetTask(w, r, id)
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		}
	})

	return mux
}

type publicCreateTaskRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type publicCreateTaskResponse struct {
	ID string `json:"id"`
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var req publicCreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	if req.Type == "" {
		http.Error(w, "type is required", http.StatusBadRequest)
		return
	}

	resp, err := s.queueClient.CreateTask(client.CreateTaskRequest{
		Type:    req.Type,
		Payload: req.Payload,
	})
	if err != nil {
		s.log.Error().Err(err).Msg("failed to create task in queue-service")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted)
	_ = json.NewEncoder(w).Encode(publicCreateTaskResponse{ID: resp.ID})
}

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Get(s.queueClientBase() + "/internal/tasks")
	if err != nil {
		s.log.Error().Err(err).Msg("failed to list tasks from queue-service")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (s *Server) handleGetTask(w http.ResponseWriter, r *http.Request, id string) {
	resp, err := http.Get(s.queueClientBase() + "/internal/tasks/" + id)
	if err != nil {
		s.log.Error().Err(err).Msg("failed to get task from queue-service")
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	_, _ = io.Copy(w, resp.Body)
}

func (s *Server) queueClientBase() string {
	return s.queueBaseURL
}
