// Package api provides HTTP client for platform API communication.
package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
)

// Client handles HTTP API communication with the platform server.
// This is the single HTTP client for all platform API calls.
type Client struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewClient creates a new platform API client.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets the authentication token.
func (c *Client) SetToken(token string) {
	c.token = token
}

// SetBaseURL sets the base URL for API requests.
// This allows overriding the URL at runtime (e.g., during login).
func (c *Client) SetBaseURL(url string) {
	c.baseURL = url
}

// SetHTTPClient sets a custom HTTP client (for testing).
func (c *Client) SetHTTPClient(client *http.Client) {
	c.httpClient = client
}

// BaseURL returns the current base URL.
func (c *Client) BaseURL() string {
	return c.baseURL
}

// Token returns the current token.
func (c *Client) Token() string {
	return c.token
}

// ============================================================================
// Authentication
// ============================================================================

// Login authenticates the user and returns a token.
func (c *Client) Login(ctx context.Context, email, password string) (*LoginResponse, error) {
	reqBody := LoginRequest{
		Email:    email,
		Password: password,
	}

	var resp LoginResponse
	if err := c.DoRequest(ctx, "POST", "/auth/login", reqBody, &resp); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// Store token for subsequent requests
	c.token = resp.Token

	return &resp, nil
}

// GetCurrentEmployee fetches information about the currently authenticated employee.
func (c *Client) GetCurrentEmployee(ctx context.Context) (*EmployeeInfo, error) {
	var resp EmployeeInfo
	if err := c.DoRequest(ctx, "GET", "/auth/me", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get current employee: %w", err)
	}
	return &resp, nil
}

// GetEmployeeInfo gets information about a specific employee.
func (c *Client) GetEmployeeInfo(ctx context.Context, employeeID string) (*EmployeeInfo, error) {
	var resp EmployeeInfo
	endpoint := fmt.Sprintf("/employees/%s", employeeID)
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get employee info: %w", err)
	}
	return &resp, nil
}

// ============================================================================
// Agent Configuration
// ============================================================================

// GetResolvedAgentConfigs fetches resolved agent configurations for an employee.
func (c *Client) GetResolvedAgentConfigs(ctx context.Context, employeeID string) ([]AgentConfig, error) {
	var resp ResolvedConfigsResponse
	endpoint := fmt.Sprintf("/employees/%s/agent-configs/resolved", employeeID)
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get resolved configs: %w", err)
	}
	return convertAPIConfigsToAgentConfigs(resp.Configs), nil
}

// GetMyResolvedAgentConfigs fetches resolved agent configurations for the current employee (JWT-based).
// Uses /employees/me/agent-configs/resolved endpoint which derives employee from JWT token.
func (c *Client) GetMyResolvedAgentConfigs(ctx context.Context) ([]AgentConfig, error) {
	var resp ResolvedConfigsResponse
	endpoint := "/employees/me/agent-configs/resolved"
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get resolved configs: %w", err)
	}
	return convertAPIConfigsToAgentConfigs(resp.Configs), nil
}

// GetOrgAgentConfigs fetches organization-level agent configs.
func (c *Client) GetOrgAgentConfigs(ctx context.Context) ([]OrgAgentConfigResponse, error) {
	var resp struct {
		Configs []OrgAgentConfigResponse `json:"configs"`
	}
	if err := c.DoRequest(ctx, "GET", "/organizations/current/agent-configs", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get org agent configs: %w", err)
	}
	return resp.Configs, nil
}

// GetTeamAgentConfigs fetches team-level agent configs.
func (c *Client) GetTeamAgentConfigs(ctx context.Context, teamID string) ([]TeamAgentConfigResponse, error) {
	var resp struct {
		Configs []TeamAgentConfigResponse `json:"configs"`
	}
	endpoint := fmt.Sprintf("/teams/%s/agent-configs", teamID)
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get team agent configs: %w", err)
	}
	return resp.Configs, nil
}

// GetEmployeeAgentConfigs fetches employee-level agent configs.
func (c *Client) GetEmployeeAgentConfigs(ctx context.Context, employeeID string) ([]EmployeeAgentConfigResponse, error) {
	var resp struct {
		Configs []EmployeeAgentConfigResponse `json:"configs"`
	}
	endpoint := fmt.Sprintf("/employees/%s/agent-configs", employeeID)
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get employee agent configs: %w", err)
	}
	return resp.Configs, nil
}

// ============================================================================
// Claude Token Management
// ============================================================================

// GetClaudeTokenStatus fetches the Claude token status for the current employee.
func (c *Client) GetClaudeTokenStatus(ctx context.Context) (*ClaudeTokenStatusResponse, error) {
	var resp ClaudeTokenStatusResponse
	endpoint := "/employees/me/claude-token/status"
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get Claude token status: %w", err)
	}
	return &resp, nil
}

// GetEffectiveClaudeToken fetches the effective Claude token for the current employee.
// Returns the actual token value (personal if set, otherwise company token).
func (c *Client) GetEffectiveClaudeToken(ctx context.Context) (string, error) {
	var resp EffectiveClaudeTokenResponse
	endpoint := "/employees/me/claude-token/effective"
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return "", fmt.Errorf("failed to get effective Claude token: %w", err)
	}
	return resp.Token, nil
}

// GetEffectiveClaudeTokenInfo fetches the effective Claude token with full metadata.
func (c *Client) GetEffectiveClaudeTokenInfo(ctx context.Context) (*EffectiveClaudeTokenResponse, error) {
	var resp EffectiveClaudeTokenResponse
	endpoint := "/employees/me/claude-token/effective"
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get effective Claude token: %w", err)
	}
	return &resp, nil
}

// ============================================================================
// Sync
// ============================================================================

// GetClaudeCodeConfig fetches the complete Claude Code configuration bundle.
func (c *Client) GetClaudeCodeConfig(ctx context.Context) (*ClaudeCodeSyncResponse, error) {
	var resp ClaudeCodeSyncResponse
	if err := c.DoRequest(ctx, "GET", "/sync/claude-code", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get Claude Code config: %w", err)
	}
	return &resp, nil
}

// ============================================================================
// Skills
// ============================================================================

// ListSkills fetches all available skills from the catalog.
func (c *Client) ListSkills(ctx context.Context) (*ListSkillsResponse, error) {
	var resp ListSkillsResponse
	if err := c.DoRequest(ctx, "GET", "/skills", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list skills: %w", err)
	}
	return &resp, nil
}

// GetSkill fetches details for a specific skill by ID.
func (c *Client) GetSkill(ctx context.Context, skillID string) (*Skill, error) {
	var skill Skill
	endpoint := fmt.Sprintf("/skills/%s", skillID)
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &skill); err != nil {
		return nil, fmt.Errorf("failed to get skill: %w", err)
	}
	return &skill, nil
}

// ListEmployeeSkills fetches skills assigned to the authenticated employee.
func (c *Client) ListEmployeeSkills(ctx context.Context) (*ListEmployeeSkillsResponse, error) {
	var resp ListEmployeeSkillsResponse
	if err := c.DoRequest(ctx, "GET", "/employees/me/skills", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list employee skills: %w", err)
	}
	return &resp, nil
}

// GetEmployeeSkill fetches a specific skill assigned to the authenticated employee.
func (c *Client) GetEmployeeSkill(ctx context.Context, skillID string) (*EmployeeSkill, error) {
	var skill EmployeeSkill
	endpoint := fmt.Sprintf("/employees/me/skills/%s", skillID)
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &skill); err != nil {
		return nil, fmt.Errorf("failed to get employee skill: %w", err)
	}
	return &skill, nil
}

// ============================================================================
// Tool Policies
// ============================================================================

// GetMyToolPolicies fetches tool policies applicable to the current employee.
// Uses JWT-based /employees/me/tool-policies endpoint.
func (c *Client) GetMyToolPolicies(ctx context.Context) (*EmployeeToolPoliciesResponse, error) {
	var resp EmployeeToolPoliciesResponse
	endpoint := "/employees/me/tool-policies"
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get tool policies: %w", err)
	}
	return &resp, nil
}

// ============================================================================
// Logging
// ============================================================================

// CreateLog sends a single log entry to the platform API.
// isValidUUID checks if a string is a valid UUID.
func isValidUUID(s string) bool {
	_, err := uuid.Parse(s)
	return err == nil
}

func (c *Client) CreateLog(ctx context.Context, entry LogEntry) error {
	req := CreateLogRequest{
		EventType:     entry.EventType,
		EventCategory: entry.EventCategory,
	}

	// Only include session_id and agent_id if they are valid UUIDs
	// The API validates these as UUID format
	if entry.SessionID != "" && isValidUUID(entry.SessionID) {
		req.SessionID = &entry.SessionID
	}
	if entry.AgentID != "" && isValidUUID(entry.AgentID) {
		req.AgentID = &entry.AgentID
	}
	if entry.Content != "" {
		req.Content = &entry.Content
	}
	if entry.Payload != nil {
		req.Payload = &entry.Payload
	}

	if err := c.DoRequest(ctx, "POST", "/logs", req, nil); err != nil {
		return fmt.Errorf("failed to create log: %w", err)
	}

	return nil
}

// CreateLogBatch sends multiple log entries in a single request.
func (c *Client) CreateLogBatch(ctx context.Context, entries []LogEntry) error {
	for _, entry := range entries {
		if err := c.CreateLog(ctx, entry); err != nil {
			return err
		}
	}
	return nil
}

// GetLogsParams contains parameters for fetching logs.
type GetLogsParams struct {
	SessionID     string
	EventCategory string
	PerPage       int
}

// GetLogs fetches logs from the API with optional filters.
func (c *Client) GetLogs(ctx context.Context, params GetLogsParams) (*LogsResponse, error) {
	query := url.Values{}
	if params.SessionID != "" {
		query.Set("session_id", params.SessionID)
	}
	if params.EventCategory != "" {
		query.Set("event_category", params.EventCategory)
	}
	if params.PerPage > 0 {
		query.Set("per_page", fmt.Sprintf("%d", params.PerPage))
	} else {
		query.Set("per_page", "1000")
	}

	endpoint := "/logs"
	if len(query) > 0 {
		endpoint = fmt.Sprintf("/logs?%s", query.Encode())
	}

	var resp LogsResponse
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get logs: %w", err)
	}
	return &resp, nil
}

// ============================================================================
// Webhooks
// ============================================================================

// ListWebhooks fetches all webhook destinations for the organization.
func (c *Client) ListWebhooks(ctx context.Context) (*ListWebhooksResponse, error) {
	var resp ListWebhooksResponse
	if err := c.DoRequest(ctx, "GET", "/webhooks", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to list webhooks: %w", err)
	}
	return &resp, nil
}

// CreateWebhook creates a new webhook destination.
func (c *Client) CreateWebhook(ctx context.Context, req CreateWebhookRequest) (*WebhookDestination, error) {
	var resp WebhookDestination
	if err := c.DoRequest(ctx, "POST", "/webhooks", req, &resp); err != nil {
		return nil, fmt.Errorf("failed to create webhook: %w", err)
	}
	return &resp, nil
}

// GetWebhook fetches a specific webhook by ID.
func (c *Client) GetWebhook(ctx context.Context, id string) (*WebhookDestination, error) {
	var resp WebhookDestination
	endpoint := fmt.Sprintf("/webhooks/%s", id)
	if err := c.DoRequest(ctx, "GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get webhook: %w", err)
	}
	return &resp, nil
}

// UpdateWebhook updates an existing webhook.
func (c *Client) UpdateWebhook(ctx context.Context, id string, req UpdateWebhookRequest) (*WebhookDestination, error) {
	var resp WebhookDestination
	endpoint := fmt.Sprintf("/webhooks/%s", id)
	if err := c.DoRequest(ctx, "PATCH", endpoint, req, &resp); err != nil {
		return nil, fmt.Errorf("failed to update webhook: %w", err)
	}
	return &resp, nil
}

// DeleteWebhook deletes a webhook by ID.
func (c *Client) DeleteWebhook(ctx context.Context, id string) error {
	endpoint := fmt.Sprintf("/webhooks/%s", id)
	if err := c.DoRequest(ctx, "DELETE", endpoint, nil, nil); err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	return nil
}

// TestWebhook tests a webhook by sending a test event.
func (c *Client) TestWebhook(ctx context.Context, id string) (*WebhookTestResult, error) {
	var resp WebhookTestResult
	endpoint := fmt.Sprintf("/webhooks/%s/test", id)
	if err := c.DoRequest(ctx, "POST", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to test webhook: %w", err)
	}
	return &resp, nil
}

// ============================================================================
// Internal Helpers
// ============================================================================

// convertAPIConfigsToAgentConfigs converts API response to internal format.
func convertAPIConfigsToAgentConfigs(apiConfigs []AgentConfigAPIResponse) []AgentConfig {
	configs := make([]AgentConfig, len(apiConfigs))
	for i, api := range apiConfigs {
		configs[i] = AgentConfig{
			AgentID:       api.AgentID,
			AgentName:     api.AgentName,
			AgentType:     api.AgentType,
			Provider:      api.Provider,
			DockerImage:   getDockerImage(api),
			IsEnabled:     api.IsEnabled,
			Configuration: api.Config,
			MCPServers:    []MCPServerConfig{},
		}
	}
	return configs
}

// getDockerImage returns the Docker image from API response or constructs a default.
func getDockerImage(api AgentConfigAPIResponse) string {
	if api.DockerImage != nil && *api.DockerImage != "" {
		return *api.DockerImage
	}
	return fmt.Sprintf("ubik/%s:latest", api.AgentType)
}

// DoRequest is a helper method to perform HTTP requests.
// This is exported to allow other packages to make custom API calls.
func (c *Client) DoRequest(ctx context.Context, method, path string, body interface{}, result interface{}) error {
	apiURL := c.baseURL + "/api/v1" + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequestWithContext(ctx, method, apiURL, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
