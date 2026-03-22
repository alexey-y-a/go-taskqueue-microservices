package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/alexey-y-a/go-taskqueue-microservices/libs/taskmodel"
)

type HTTPClient struct {
	baseURL    string
	httpClient *http.Client
}

func NewHttpClient(baseURL string, timeout time.Duration) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: timeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
			},
		},
	}
}

func (c *HTTPClient) GetPendingTasks(ctx context.Context, limit int) ([]*taskmodel.Task, error) {
	url := fmt.Sprintf("%s/tasks/pending?limit=%d", c.baseURL, limit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %s", resp.Status)
	}

	var tasks []*taskmodel.Task

	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return tasks, nil
}

func (c *HTTPClient) UpdateTaskStatus(ctx context.Context, taskID string, status taskmodel.TaskStatus, errMsg string) error {
	reqBody := map[string]interface{}{
		"status": status,
	}
	if errMsg != "" {
		reqBody["error"] = errMsg
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal request body: %w", err)
	}

	url := fmt.Sprintf("%s/tasks/%s", c.baseURL, taskID)

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("unexpected status: %s", resp.Status)
	}

	return nil
}
