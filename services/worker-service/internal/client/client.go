package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
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

func (c *QueueClient) GetNextPending() (taskmodel.Task, bool, error) {
	resp, err := c.client.Get(c.baseURL + "/internal/next-pending")
	if err != nil {
		return taskmodel.Task{}, false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return taskmodel.Task{}, false, nil
	}
	if resp.StatusCode != http.StatusOK {
		return taskmodel.Task{}, false, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var t taskmodel.Task
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return taskmodel.Task{}, false, err
	}
	return t, true, nil
}

type updateStatusRequest struct {
	Status string `json:"status"`
}

func (c *QueueClient) UpdateStatus(id string, status taskmodel.Status) error {
	body, err := json.Marshal(updateStatusRequest{Status: string(status)})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/internal/tasks/"+id+"/status", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}
	return nil
}
