package gitlab

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/skr1ms/mosaic/pkg/middleware"
)

type Client struct {
	baseURL      string
	APIToken     string
	TriggerToken string
	projectID    string
	client       *http.Client
	logger       *middleware.Logger
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

func NewClient(baseURL, APIToken, TriggerToken, projectID string, logger *middleware.Logger) *Client {
	return &Client{
		baseURL:      baseURL,
		APIToken:     APIToken,
		TriggerToken: TriggerToken,
		projectID:    projectID,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
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
	c.logger.GetZerologLogger().Info().
		Str("url", url).
		Str("form_data", formData).
		Msg("GitLab Trigger API Request")

	httpReq, err := http.NewRequest("POST", url, bytes.NewBufferString(formData))
	if err != nil {
		c.logger.GetZerologLogger().Error().
			Err(err).
			Str("url", url).
			Msg("GitLab Error: failed to create request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		c.logger.GetZerologLogger().Error().
			Err(err).
			Str("url", url).
			Msg("GitLab Error: failed to execute request")
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusCreated {
		c.logger.GetZerologLogger().Error().
			Int("status_code", resp.StatusCode).
			Str("response_body", string(body)).
			Str("url", url).
			Msg("GitLab Error: pipeline trigger failed with non-created status")
		return nil, fmt.Errorf("pipeline trigger failed with status: %d, body: %s", resp.StatusCode, string(body))
	}

	var pipelineResp PipelineResponse
	if err := json.Unmarshal(body, &pipelineResp); err != nil {
		c.logger.GetZerologLogger().Error().
			Err(err).
			Str("response_body", string(body)).
			Msg("GitLab Error: failed to decode response")
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

	// Clean domain from protocol if present
	cleanDomain := func(domain string) string {
		if domain == "" {
			return ""
		}
		// Remove http:// or https:// prefix
		if strings.HasPrefix(domain, "https://") {
			return strings.TrimPrefix(domain, "https://")
		}
		if strings.HasPrefix(domain, "http://") {
			return strings.TrimPrefix(domain, "http://")
		}
		return domain
	}

	if oldDomain != "" {
		cleanedOldDomain := cleanDomain(oldDomain)
		variables = append(variables, PipelineVariable{Key: "OLD_DOMAIN", Value: cleanedOldDomain})
	}

	if newDomain != "" {
		cleanedNewDomain := cleanDomain(newDomain)
		variables = append(variables, PipelineVariable{Key: "NEW_DOMAIN", Value: cleanedNewDomain})
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
		c.logger.GetZerologLogger().Error().
			Err(err).
			Str("url", url).
			Int("pipeline_id", pipelineID).
			Msg("GitLab Error: failed to create pipeline status request")
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.APIToken)

	resp, err := c.client.Do(req)
	if err != nil {
		c.logger.GetZerologLogger().Error().
			Err(err).
			Str("url", url).
			Int("pipeline_id", pipelineID).
			Msg("GitLab Error: failed to execute pipeline status request")
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.logger.GetZerologLogger().Error().
			Int("status_code", resp.StatusCode).
			Str("url", url).
			Int("pipeline_id", pipelineID).
			Msg("GitLab Error: get pipeline status failed with non-OK status")
		return nil, fmt.Errorf("get pipeline status failed with status: %d", resp.StatusCode)
	}

	var pipelineResp PipelineResponse
	if err := json.NewDecoder(resp.Body).Decode(&pipelineResp); err != nil {
		c.logger.GetZerologLogger().Error().
			Err(err).
			Int("pipeline_id", pipelineID).
			Msg("GitLab Error: failed to decode pipeline status response")
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &pipelineResp, nil
}
