package gitlab

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	baseURL      string
	accessToken  string
	triggerToken string
	projectID    string
	client       *http.Client
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

func NewClient(baseURL, accessToken, triggerToken, projectID string) *Client {
	return &Client{
		baseURL:      baseURL,
		accessToken:  accessToken,
		triggerToken: triggerToken,
		projectID:    projectID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TriggerPipeline triggers a new pipeline in GitLab using trigger token
// This method uses the trigger token API which doesn't require special permissions
func (c *Client) TriggerPipeline(req TriggerPipelineRequest) (*PipelineResponse, error) {
	apiURL := fmt.Sprintf("%s/api/v4/projects/%s/trigger/pipeline", c.baseURL, c.projectID)

	// Build form data with trigger token
	data := url.Values{}
	data.Set("ref", req.Ref)
	data.Set("token", c.triggerToken)
	
	// Add variables if provided
	for key, value := range req.Variables {
		data.Set(fmt.Sprintf("variables[%s]", key), value)
	}

	httpReq, err := http.NewRequest("POST", apiURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("pipeline trigger failed with status: %d, body: %s", resp.StatusCode, string(bodyBytes))
	}

	var pipelineResp PipelineResponse
	if err := json.Unmarshal(bodyBytes, &pipelineResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pipelineResp, nil
}

// TriggerDomainUpdate triggers a pipeline specifically for domain updates
func (c *Client) TriggerDomainUpdate(ref string) (*PipelineResponse, error) {
	// Simply call TriggerPipeline with the appropriate request
	req := TriggerPipelineRequest{
		Ref:       ref,
		Variables: nil, // No additional variables for domain updates
	}
	return c.TriggerPipeline(req)
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
