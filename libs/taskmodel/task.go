package taskmodel

import "time"

type TaskStatus string

const (
	StatusPending    TaskStatus = "pending"
	StatusProcessing TaskStatus = "processing"
	StatusCompleted  TaskStatus = "completed"
	StatusFailed     TaskStatus = "failed"
)

type Task struct {
	ID          string            `json:"id" db:"id"`
	Type        string            `json:"type" db:"type"`
	Payload     string            `json:"payload" db:"payload"`
	Status      TaskStatus        `json:"status" db:"status"`
	CreatedAt   time.Time         `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" db:"updated_at"`
	CompletedAt *time.Time        `json:"completed_at" db:"completed_at"`
	Error       string            `json:"error,omitempty" db:"error"`
	Labels      map[string]string `json:"labels,omitempty"`
}

func NewTask(id, taskType, payload string) *Task {
	now := time.Now().UTC()

	return &Task{
		ID:        id,
		Type:      taskType,
		Payload:   payload,
		Status:    StatusPending,
		CreatedAt: now,
		UpdatedAt: now,
		Labels:    make(map[string]string),
	}
}

func (t *Task) UpdateStatus(status TaskStatus, errMsg string) {
	t.Status = status
	t.UpdatedAt = time.Now().UTC()

	if status == StatusCompleted {
		now := time.Now().UTC()
		t.CompletedAt = &now
		t.Error = ""
	} else if status == StatusFailed {
		t.Error = errMsg
	}
}

func (t *Task) IsComplete() bool {
	return t.Status == StatusCompleted || t.Status == StatusFailed
}
