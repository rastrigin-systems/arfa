package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PlatformClient handles API communication with the platform server
type PlatformClient struct {
	baseURL    string
	httpClient *http.Client
	token      string
}

// NewPlatformClient creates a new platform API client
func NewPlatformClient(baseURL string) *PlatformClient {
	return &PlatformClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets the authentication token
func (pc *PlatformClient) SetToken(token string) {
	pc.token = token
}

// LoginRequest represents a login request
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token     string            `json:"token"`
	ExpiresAt string            `json:"expires_at"`
	Employee  LoginEmployeeInfo `json:"employee"`
}

// LoginEmployeeInfo contains employee info from login response
type LoginEmployeeInfo struct {
	ID       string `json:"id"`
	OrgID    string `json:"org_id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
}

// Login authenticates the user and returns a token
func (pc *PlatformClient) Login(email, password string) (*LoginResponse, error) {
	reqBody := LoginRequest{
		Email:    email,
		Password: password,
	}

	var resp LoginResponse
	if err := pc.doRequest("POST", "/auth/login", reqBody, &resp); err != nil {
		return nil, fmt.Errorf("login failed: %w", err)
	}

	// Store token for subsequent requests
	pc.token = resp.Token

	return &resp, nil
}

// GetEmployeeInfo gets information about the current employee
type EmployeeInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	FullName string `json:"full_name"`
	OrgID    string `json:"org_id"`
}

func (pc *PlatformClient) GetEmployeeInfo(employeeID string) (*EmployeeInfo, error) {
	var resp EmployeeInfo
	endpoint := fmt.Sprintf("/employees/%s", employeeID)
	if err := pc.doRequest("GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get employee info: %w", err)
	}
	return &resp, nil
}

// AgentConfigAPIResponse represents an agent config as returned by the API
type AgentConfigAPIResponse struct {
	AgentID      string                 `json:"agent_id"`
	AgentName    string                 `json:"agent_name"`
	AgentType    string                 `json:"agent_type"`
	IsEnabled    bool                   `json:"is_enabled"`
	Config       map[string]interface{} `json:"config"`
	Provider     string                 `json:"provider"`
	SyncToken    string                 `json:"sync_token"`
	SystemPrompt string                 `json:"system_prompt"`
	LastSyncedAt *string                `json:"last_synced_at"` // nullable timestamp
}

// AgentConfig represents a resolved agent configuration (internal use)
type AgentConfig struct {
	AgentID       string                 `json:"agent_id"`
	AgentName     string                 `json:"agent_name"`
	AgentType     string                 `json:"agent_type"`
	IsEnabled     bool                   `json:"is_enabled"`
	Configuration map[string]interface{} `json:"configuration"`
	MCPServers    []MCPServerConfig      `json:"mcp_servers"`
}

// MCPServerConfig represents an MCP server configuration
type MCPServerConfig struct {
	ServerID   string                 `json:"server_id"`
	ServerName string                 `json:"server_name"`
	ServerType string                 `json:"server_type"`
	IsEnabled  bool                   `json:"is_enabled"`
	Config     map[string]interface{} `json:"config"`
}

// ResolvedConfigsResponse represents the response from the resolved configs endpoint
type ResolvedConfigsResponse struct {
	Configs []AgentConfigAPIResponse `json:"configs"`
	Total   int                      `json:"total"`
}

// GetResolvedAgentConfigs fetches resolved agent configurations for an employee
func (pc *PlatformClient) GetResolvedAgentConfigs(employeeID string) ([]AgentConfig, error) {
	var resp ResolvedConfigsResponse
	endpoint := fmt.Sprintf("/employees/%s/agent-configs/resolved", employeeID)
	if err := pc.doRequest("GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get resolved configs: %w", err)
	}

	// Convert API response to internal format
	configs := make([]AgentConfig, len(resp.Configs))
	for i, apiConfig := range resp.Configs {
		configs[i] = AgentConfig{
			AgentID:       apiConfig.AgentID,
			AgentName:     apiConfig.AgentName,
			AgentType:     apiConfig.AgentType,
			IsEnabled:     apiConfig.IsEnabled,
			Configuration: apiConfig.Config,
			MCPServers:    []MCPServerConfig{}, // TODO: Fetch MCP servers separately if needed
		}
	}

	return configs, nil
}

// ClaudeTokenStatusResponse represents the Claude token status response
type ClaudeTokenStatusResponse struct {
	EmployeeID        string `json:"employee_id"`
	HasPersonalToken  bool   `json:"has_personal_token"`
	HasCompanyToken   bool   `json:"has_company_token"`
	ActiveTokenSource string `json:"active_token_source"` // "personal", "company", or "none"
}

// GetClaudeTokenStatus fetches the Claude token status for the current employee
func (pc *PlatformClient) GetClaudeTokenStatus() (*ClaudeTokenStatusResponse, error) {
	var resp ClaudeTokenStatusResponse
	endpoint := "/employees/me/claude-token/status"
	if err := pc.doRequest("GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get Claude token status: %w", err)
	}
	return &resp, nil
}

// GetEffectiveClaudeToken fetches the effective Claude token for the current employee
// This is a convenience wrapper that extracts just the token from the status
func (pc *PlatformClient) GetEffectiveClaudeToken() (string, error) {
	status, err := pc.GetClaudeTokenStatus()
	if err != nil {
		return "", err
	}

	// Check if any token is available
	if status.ActiveTokenSource == "none" {
		return "", fmt.Errorf("no Claude API token configured. Please configure a token at organization or personal level")
	}

	// Token exists but we need to make another API call to get the actual token value
	// For security reasons, the status endpoint doesn't return the actual token
	// In a real implementation, you would either:
	// 1. Have a separate endpoint to get the decrypted token (with proper auth)
	// 2. Return a reference/ID that can be used to fetch the token securely
	// For now, we'll return a placeholder indicating where the token comes from
	return fmt.Sprintf("token-source:%s", status.ActiveTokenSource), nil
}

// doRequest is a helper method to perform HTTP requests
func (pc *PlatformClient) doRequest(method, path string, body interface{}, result interface{}) error {
	// Add /api/v1 prefix to all API calls
	url := pc.baseURL + "/api/v1" + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authorization header if token is set
	if pc.token != "" {
		req.Header.Set("Authorization", "Bearer "+pc.token)
	}

	resp, err := pc.httpClient.Do(req)
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
