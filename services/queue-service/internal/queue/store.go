package queue

import (
	"sync"
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
	"github.com/google/uuid"
)

type Store struct {
    mu sync.RWMutex
    tasks map[string]taskmodel.Task
}

func NewStore() *Store {
    return &Store{
        tasks: make(map[string]taskmodel.Task ),
    }
}

func (s *Store) CreateTask(taskType, payload string) taskmodel.Task {
    s.mu.Lock()
    defer s.mu.Unlock()

    id := uuid.NewString()
    now := time.Now()

    t := taskmodel.Task{
        ID: id,
        Type: taskType,
        Payload: payload,
        Status: taskmodel.StatusPending,
        CreatedAt: now,
        UpdatedAt: now,
    }

    s.tasks[id] = t
    return t
}

func (s *Store) GetTask(id string) (taskmodel.Task, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    t, ok := s.tasks[id]
    return t, ok
}

func (s *Store) ListTasks() []taskmodel.Task {
    s.mu.RLock()
    defer s.mu.RUnlock()

    res := make([]taskmodel.Task, 0, len(s.tasks))
    for _, t := range s.tasks {
        res = append(res, t)
    }
    return res
}

func (s *Store) UpdateStatus(id string, status taskmodel.Status) (taskmodel.Task, bool) {
    s.mu.Lock()
    defer s.mu.Unlock()

    t, ok := s.tasks[id]
    if !ok {
        return taskmodel.Task{}, false
    }

    t.Status = status
    t.UpdatedAt = time.Now()
    s.tasks[id] = t
    return t, true
}

func (s *Store) NextPending() (taskmodel.Task, bool) {
    s.mu.RLock()
    defer s.mu.RUnlock()

    for _, t := range s.tasks {
        if t.Status == taskmodel.StatusPending {
            return t, true
        }
    }
    return taskmodel.Task{}, false
}

