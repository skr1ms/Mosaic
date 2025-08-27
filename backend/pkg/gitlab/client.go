package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	baseURL     string
	accessToken string
	projectID   string
	client      *http.Client
}

type TriggerPipelineRequest struct {
	Ref       string            `json:"ref"`
	Variables map[string]string `json:"variables,omitempty"`
}

type PipelineResponse struct {
	ID     int    `json:"id"`
	SHA    string `json:"sha"`
	Ref    string `json:"ref"`
	Status string `json:"status"`
	WebURL string `json:"web_url"`
}

func NewClient(baseURL, accessToken, projectID string) *Client {
	return &Client{
		baseURL:     baseURL,
		accessToken: accessToken,
		projectID:   projectID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TriggerPipeline triggers a new pipeline in GitLab
func (c *Client) TriggerPipeline(req TriggerPipelineRequest) (*PipelineResponse, error) {
	// Используем обычный API для запуска пайплайна
	url := fmt.Sprintf("%s/api/v4/projects/%s/pipeline", c.baseURL, c.projectID)

	jsonBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	fmt.Printf("GitLab API Request: %s\n", url)
	fmt.Printf("GitLab API Body: %s\n", string(jsonBody))

	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)
	fmt.Printf("GitLab API Response Status: %d\n", resp.StatusCode)
	fmt.Printf("GitLab API Response Body: %s\n", string(bodyBytes))

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("pipeline trigger failed with status: %d", resp.StatusCode)
	}

	var pipelineResp PipelineResponse
	if err := json.Unmarshal(bodyBytes, &pipelineResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pipelineResp, nil
}

// TriggerDomainUpdate triggers a pipeline specifically for domain updates
func (c *Client) TriggerDomainUpdate(ref string) (*PipelineResponse, error) {
	return c.TriggerPipeline(TriggerPipelineRequest{
		Ref: ref,
		Variables: map[string]string{
			"TRIGGER_TYPE": "domain_update",
		},
	})
}

// GetPipelineStatus gets the status of a pipeline by ID
func (c *Client) GetPipelineStatus(pipelineID int) (*PipelineResponse, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%s/pipelines/%d", c.baseURL, c.projectID, pipelineID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.accessToken)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("get pipeline status failed with status: %d", resp.StatusCode)
	}

	var pipelineResp PipelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&pipelineResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pipelineResp, nil
}
