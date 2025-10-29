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
	Token      string `json:"token"`
	EmployeeID string `json:"employee_id"`
	OrgID      string `json:"org_id"`
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

// AgentConfig represents a resolved agent configuration
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

// GetResolvedAgentConfigs fetches resolved agent configurations for an employee
func (pc *PlatformClient) GetResolvedAgentConfigs(employeeID string) ([]AgentConfig, error) {
	var configs []AgentConfig
	endpoint := fmt.Sprintf("/employees/%s/agent-configs/resolved", employeeID)
	if err := pc.doRequest("GET", endpoint, nil, &configs); err != nil {
		return nil, fmt.Errorf("failed to get resolved configs: %w", err)
	}
	return configs, nil
}

// doRequest is a helper method to perform HTTP requests
func (pc *PlatformClient) doRequest(method, path string, body interface{}, result interface{}) error {
	url := pc.baseURL + path

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
