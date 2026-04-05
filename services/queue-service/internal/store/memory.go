package store

import (
	"context"
	"fmt"
	"sync"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
)

var (
	ErrTaskNotFound = fmt.Errorf("task not found")
	ErrTaskExists   = fmt.Errorf("task already exists")
	ErrStoreFull    = fmt.Errorf("store is full")
)

type MemoryStore struct {
	mu     sync.RWMutex
	tasks  map[string]*taskmodel.Task
	maxLen int
}

func NewMemoryStore(maxLen int) *MemoryStore {
	return &MemoryStore{
		tasks:  make(map[string]*taskmodel.Task),
		maxLen: maxLen,
	}
}

func (s *MemoryStore) Create(ctx context.Context, task *taskmodel.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.tasks) >= s.maxLen {
		return ErrStoreFull
	}

	_, exists := s.tasks[task.ID]
	if exists {
		return ErrTaskExists
	}

	s.tasks[task.ID] = task

	return nil
}

func (s *MemoryStore) Get(ctx context.Context, id string) (*taskmodel.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exist := s.tasks[id]
	if !exist {
		return nil, ErrTaskNotFound
	}

	return task, nil
}

func (s *MemoryStore) Update(ctx context.Context, task *taskmodel.Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.tasks[task.ID]
	if !exists {
		return ErrTaskNotFound
	}

	s.tasks[task.ID] = task

	return nil
}

func (s *MemoryStore) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, exists := s.tasks[id]
	if !exists {
		return ErrTaskNotFound
	}

	delete(s.tasks, id)

	return nil
}

func (s *MemoryStore) List(ctx context.Context, limit, offset int) ([]*taskmodel.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit > len(s.tasks) {
		limit = len(s.tasks)
	}

	tasks := make([]*taskmodel.Task, 0, limit)

	i := 0
	for _, task := range s.tasks {
		if i < offset {
			i++
			continue
		}

		tasks = append(tasks, task)

		if len(tasks) >= limit {
			break
		}
		i++
	}

	return tasks, nil

}

func (s *MemoryStore) GetPending(ctx context.Context, limit int) ([]*taskmodel.Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 {
		limit = 100
	}

	tasks := make([]*taskmodel.Task, 0, limit)

	for _, task := range s.tasks {
		if task.Status == taskmodel.StatusPending {
			tasks = append(tasks, task)

			if len(tasks) >= limit {
				break
			}
		}
	}

	return tasks, nil
}

func (s *MemoryStore) Count(ctx context.Context) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.tasks), nil
}

func (s *MemoryStore) UpdateStatus(ctx context.Context, id string, status taskmodel.TaskStatus, errMsg string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task, exists := s.tasks[id]
	if !exists {
		return ErrTaskNotFound
	}

	task.Status = status
	if errMsg != "" {
		task.Error = errMsg
	}

	return nil
}
