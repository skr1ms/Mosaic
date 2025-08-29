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
	baseURL      string
	APIToken     string
	TriggerToken string
	projectID    string
	client       *http.Client
}

type TriggerPipelineRequest struct {
	Ref       string             `json:"ref"`
	Variables []PipelineVariable `json:"variables,omitempty"`
}

type PipelineVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type PipelineResponse struct {
	ID     int    `json:"id"`
	SHA    string `json:"sha"`
	Ref    string `json:"ref"`
	Status string `json:"status"`
	WebURL string `json:"web_url"`
}

func NewClient(baseURL, APIToken, TriggerToken, projectID string) *Client {
	return &Client{
		baseURL:      baseURL,
		APIToken:     APIToken,
		TriggerToken: TriggerToken,
		projectID:    projectID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// TriggerPipeline triggers a new pipeline in GitLab using trigger endpoint
func (c *Client) TriggerPipeline(req TriggerPipelineRequest) (*PipelineResponse, error) {

	url := fmt.Sprintf("%s/api/v4/projects/%s/trigger/pipeline", c.baseURL, c.projectID)

	// Build form data for trigger endpoint
	formData := fmt.Sprintf("token=%s&ref=%s", c.TriggerToken, req.Ref)

	// Add variables if provided
	if req.Variables != nil {
		for _, v := range req.Variables {
			formData += fmt.Sprintf("&variables[%s]=%s", v.Key, v.Value)
		}
	}

	// Log the request for debugging
	fmt.Printf("GitLab Trigger API Request URL: %s\n", url)
	fmt.Printf("GitLab Trigger API Request Data: %s\n", formData)

	httpReq, err := http.NewRequest("POST", url, bytes.NewBufferString(formData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("pipeline trigger failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	var pipelineResp PipelineResponse
	if err := json.Unmarshal(body, &pipelineResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pipelineResp, nil
}

// TriggerDomainUpdate triggers a pipeline specifically for domain updates
func (c *Client) TriggerDomainUpdate(ref string) (*PipelineResponse, error) {
	return c.TriggerDomainUpdateWithDetails(ref, "refresh", "", "")
}

// TriggerDomainUpdateWithDetails triggers a pipeline with specific domain operation details
func (c *Client) TriggerDomainUpdateWithDetails(ref, operation, oldDomain, newDomain string) (*PipelineResponse, error) {
	variables := []PipelineVariable{
		{Key: "DOMAIN_UPDATE", Value: "true"},
		{Key: "DOMAIN_OPERATION", Value: operation},
	}

	if oldDomain != "" {
		variables = append(variables, PipelineVariable{Key: "OLD_DOMAIN", Value: oldDomain})
	}

	if newDomain != "" {
		variables = append(variables, PipelineVariable{Key: "NEW_DOMAIN", Value: newDomain})
	}

	return c.TriggerPipeline(TriggerPipelineRequest{
		Ref:       ref,
		Variables: variables,
	})
}

// GetPipelineStatus gets the status of a pipeline by ID
func (c *Client) GetPipelineStatus(pipelineID int) (*PipelineResponse, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%s/pipelines/%d", c.baseURL, c.projectID, pipelineID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIToken)

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
