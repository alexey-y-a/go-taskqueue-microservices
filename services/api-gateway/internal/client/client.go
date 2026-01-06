package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type QueueClient struct {
	baseURL string
	client  *http.Client
}

func NewQueueClient(baseURL string) *QueueClient {
	return &QueueClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

type CreateTaskRequest struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

type CreateTaskResponse struct {
	ID string `json:"id"`
}

func (c *QueueClient) CreateTask(req CreateTaskRequest) (CreateTaskResponse, error) {
	b, err := json.Marshal(req)
	if err != nil {
		return CreateTaskResponse{}, err
	}
	httpReq, err := http.NewRequest(http.MethodPost, c.baseURL+"/internal/tasks", bytes.NewReader(b))
	if err != nil {
		return CreateTaskResponse{}, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return CreateTaskResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return CreateTaskResponse{}, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var out CreateTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return CreateTaskResponse{}, err
	}
	return out, nil
}
