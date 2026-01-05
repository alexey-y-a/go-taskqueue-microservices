package http

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/logger"
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

    mux.HandleFunc("/internal/tasks",func(w http.ResponseWriter, r *http.Request) {
      switch r.Method {
          case http.MethodPost:
            s.handleCreateTask(w, r)
          case http.MethodGet:
            s.handleListTasks(w, r)
          default:
            http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
      }
    })

    mux.HandleFunc("/internal/tasks/", func(w http.ResponseWriter, r *http.Request) {
        id := strings.TrimPrefix(r.URL.Path, "/internal/tasks/")
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

type createTaskRequest struct {
    Type string `json:"type"`
    Payload string `json:"payload"`
}

type createTaskResponse struct {
    ID string `json:"id"`
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
    var req createTaskRequest
    err := json.NewDecoder(r.Body).Decode(&req)
    if err != nil {
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

