package taskmodel

import "time"

type Status string

const (
    StatusPending    Status = "pending"
    StatusProcessing Status = "processing"
    StatusCompleted  Status = "completed"
    StatusFailed     Status = "failed"
)

type Task struct {
    ID        string    `json:"id"`
    Type      string    `json:"type"`
    Payload   string    `json:"payload"`
    Status    Status    `json:"status"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}


