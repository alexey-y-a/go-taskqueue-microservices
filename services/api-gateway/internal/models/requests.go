package models

import (
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
)

type CreateTaskRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type CreateTaskResponse struct {
	ID     string               `json:"id"`
	Status taskmodel.TaskStatus `json:"status"`
}

type TaskResponse struct {
	ID        string               `json:"id"`
	Type      string               `json:"type"`
	Payload   string               `json:"payload"`
	Status    taskmodel.TaskStatus `json:"status"`
	CreatedAt time.Time            `json:"created_at"`
	UpdatedAt time.Time            `json:"updated_at"`
	Error     string               `json:"error,omitempty"`
}

type ErrorResponse struct {
	Error     string `json:"error"`
	Code      int    `json:"code"`
	RequestID string `json:"request_id,omitempty"`
}
