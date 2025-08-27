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

	// Логирование убрано - API работает корректно

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
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/pipeline", c.baseURL, c.projectID)

	reqBody := map[string]interface{}{
		"ref": ref,
		"variables": []map[string]string{
			{
				"key":   "PIPELINE_TYPE",
				"value": "domain_update",
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.accessToken)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("pipeline trigger failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	var pipelineResp PipelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&pipelineResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pipelineResp, nil
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
