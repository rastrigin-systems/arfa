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

// SetHTTPClient sets a custom HTTP client (for testing)
func (pc *PlatformClient) SetHTTPClient(client *http.Client) {
	pc.httpClient = client
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
	ID       string  `json:"id"`
	Email    string  `json:"email"`
	FullName string  `json:"full_name"`
	OrgID    string  `json:"org_id"`
	TeamID   *string `json:"team_id"` // nullable
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
	Provider      string                 `json:"provider"`
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
			Provider:      apiConfig.Provider,
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

// EffectiveClaudeTokenResponse represents the effective token response
type EffectiveClaudeTokenResponse struct {
	Token      string `json:"token"`
	Source     string `json:"source"`      // "personal" or "company"
	OrgID      string `json:"org_id"`
	OrgName    string `json:"org_name"`
	EmployeeID string `json:"employee_id"`
}

// GetEffectiveClaudeToken fetches the effective Claude token for the current employee
// Returns the actual token value (personal if set, otherwise company token)
func (pc *PlatformClient) GetEffectiveClaudeToken() (string, error) {
	var resp EffectiveClaudeTokenResponse
	endpoint := "/employees/me/claude-token/effective"
	if err := pc.doRequest("GET", endpoint, nil, &resp); err != nil {
		return "", fmt.Errorf("failed to get effective Claude token: %w", err)
	}
	return resp.Token, nil
}

// GetEffectiveClaudeTokenInfo fetches the effective Claude token with full metadata
func (pc *PlatformClient) GetEffectiveClaudeTokenInfo() (*EffectiveClaudeTokenResponse, error) {
	var resp EffectiveClaudeTokenResponse
	endpoint := "/employees/me/claude-token/effective"
	if err := pc.doRequest("GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get effective Claude token: %w", err)
	}
	return &resp, nil
}

// OrgAgentConfigResponse represents an org-level agent config
type OrgAgentConfigResponse struct {
	ID        string                 `json:"id"`
	AgentID   string                 `json:"agent_id"`
	AgentName string                 `json:"agent_name"`
	Config    map[string]interface{} `json:"config"`
	IsEnabled bool                   `json:"is_enabled"`
}

// TeamAgentConfigResponse represents a team-level agent config
type TeamAgentConfigResponse struct {
	ID             string                 `json:"id"`
	AgentID        string                 `json:"agent_id"`
	AgentName      string                 `json:"agent_name"`
	ConfigOverride map[string]interface{} `json:"config_override"`
	IsEnabled      bool                   `json:"is_enabled"`
}

// EmployeeAgentConfigResponse represents an employee-level agent config
type EmployeeAgentConfigResponse struct {
	ID             string                 `json:"id"`
	AgentID        string                 `json:"agent_id"`
	AgentName      string                 `json:"agent_name"`
	ConfigOverride map[string]interface{} `json:"config_override"`
	IsEnabled      bool                   `json:"is_enabled"`
}

// GetOrgAgentConfigs fetches organization-level agent configs
func (pc *PlatformClient) GetOrgAgentConfigs() ([]OrgAgentConfigResponse, error) {
	var resp struct {
		Configs []OrgAgentConfigResponse `json:"configs"`
	}
	if err := pc.doRequest("GET", "/organizations/current/agent-configs", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get org agent configs: %w", err)
	}
	return resp.Configs, nil
}

// GetTeamAgentConfigs fetches team-level agent configs
func (pc *PlatformClient) GetTeamAgentConfigs(teamID string) ([]TeamAgentConfigResponse, error) {
	var resp struct {
		Configs []TeamAgentConfigResponse `json:"configs"`
	}
	endpoint := fmt.Sprintf("/teams/%s/agent-configs", teamID)
	if err := pc.doRequest("GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get team agent configs: %w", err)
	}
	return resp.Configs, nil
}

// GetEmployeeAgentConfigs fetches employee-level agent configs
func (pc *PlatformClient) GetEmployeeAgentConfigs(employeeID string) ([]EmployeeAgentConfigResponse, error) {
	var resp struct {
		Configs []EmployeeAgentConfigResponse `json:"configs"`
	}
	endpoint := fmt.Sprintf("/employees/%s/agent-configs", employeeID)
	if err := pc.doRequest("GET", endpoint, nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get employee agent configs: %w", err)
	}
	return resp.Configs, nil
}

// GetCurrentEmployee fetches information about the currently authenticated employee
func (pc *PlatformClient) GetCurrentEmployee() (*EmployeeInfo, error) {
	var resp EmployeeInfo
	if err := pc.doRequest("GET", "/auth/me", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get current employee: %w", err)
	}
	return &resp, nil
}

// ClaudeCodeSyncResponse represents the complete Claude Code configuration bundle
type ClaudeCodeSyncResponse struct {
	Agents     []AgentConfigSync     `json:"agents"`
	Skills     []SkillConfigSync     `json:"skills"`
	MCPServers []MCPServerConfigSync `json:"mcp_servers"`
	Version    string                `json:"version"`
	SyncedAt   string                `json:"synced_at"`
}

// AgentConfigSync represents an agent configuration in the sync response
type AgentConfigSync struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Type      string                 `json:"type"`
	Filename  string                 `json:"filename"`
	Content   string                 `json:"content,omitempty"`
	Config    map[string]interface{} `json:"config"`
	Provider  string                 `json:"provider"`
	IsEnabled bool                   `json:"is_enabled"`
	Version   string                 `json:"version"`
}

// SkillConfigSync represents a skill configuration in the sync response
type SkillConfigSync struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	Category     string                 `json:"category,omitempty"`
	Version      string                 `json:"version"`
	Files        []map[string]string    `json:"files,omitempty"`
	Dependencies map[string]interface{} `json:"dependencies,omitempty"`
	IsEnabled    bool                   `json:"is_enabled"`
}

// MCPServerConfigSync represents an MCP server configuration in the sync response
type MCPServerConfigSync struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Provider        string                 `json:"provider"`
	Version         string                 `json:"version"`
	Description     string                 `json:"description,omitempty"`
	DockerImage     string                 `json:"docker_image"`
	Config          map[string]interface{} `json:"config"`
	RequiredEnvVars []string               `json:"required_env_vars,omitempty"`
	IsEnabled       bool                   `json:"is_enabled"`
}

// GetClaudeCodeConfig fetches the complete Claude Code configuration bundle
func (pc *PlatformClient) GetClaudeCodeConfig() (*ClaudeCodeSyncResponse, error) {
	var resp ClaudeCodeSyncResponse
	if err := pc.doRequest("GET", "/sync/claude-code", nil, &resp); err != nil {
		return nil, fmt.Errorf("failed to get Claude Code config: %w", err)
	}
	return &resp, nil
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
