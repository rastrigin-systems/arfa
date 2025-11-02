package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	mockdb "github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/handlers"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/middleware"
)

// TestGetClaudeCodeSync_Success tests successful sync with all resources
func TestGetClaudeCodeSync_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)
	handler := handlers.NewSyncHandler(mockDB)

	employeeID := uuid.New()
	orgID := uuid.New()
	agentID := uuid.New()
	skillID := uuid.New()
	mcpID := uuid.New()

	// Mock agent configs
	mockDB.EXPECT().
		ListEmployeeAgentConfigs(gomock.Any(), employeeID).
		Return([]db.ListEmployeeAgentConfigsRow{
			{
				ID:                 uuid.New(),
				EmployeeID:         employeeID,
				AgentID:            agentID,
				ConfigOverride:     []byte(`{"model": "claude-3-5-sonnet-20241022"}`),
				IsEnabled:          true,
				AgentName:          "go-backend-developer",
				AgentType:          "claude-code",
				AgentProvider:      "anthropic",
				AgentDefaultConfig: []byte(`{"temperature": 0.7}`),
			},
		}, nil)

	// Mock skills
	description := "Release management skill"
	category := "project-management"
	isEnabledSkill := true
	mockDB.EXPECT().
		ListEmployeeSkills(gomock.Any(), employeeID).
		Return([]db.ListEmployeeSkillsRow{
			{
				ID:           skillID,
				Name:         "release-manager",
				Description:  &description,
				Category:     &category,
				Version:      "1.0.0",
				Files:        []byte(`[{"path": "SKILL.md", "content": "# Release Manager"}]`),
				Dependencies: []byte(`{"mcp_servers": ["github-mcp-server"]}`),
				IsActive:     &isEnabledSkill,
				IsEnabled:    &isEnabledSkill,
			},
		}, nil)

	// Mock MCP servers
	dockerImage := "ghcr.io/github/github-mcp-server"
	isEnabledMCP := true
	mockDB.EXPECT().
		ListEmployeeMCPConfigs(gomock.Any(), employeeID).
		Return([]db.ListEmployeeMCPConfigsRow{
			{
				ID:               mcpID,
				Name:             "github-mcp-server",
				Provider:         "github",
				Version:          "1.0.0",
				Description:      "GitHub MCP server",
				DockerImage:      &dockerImage,
				ConfigTemplate:   []byte(`{"env": {"GITHUB_PERSONAL_ACCESS_TOKEN": ""}}`),
				RequiredEnvVars:  []byte(`["GITHUB_PERSONAL_ACCESS_TOKEN"]`),
				ConnectionConfig: []byte(`{"env": {"GITHUB_PERSONAL_ACCESS_TOKEN": "ghp_xxx"}}`),
				IsEnabled:        &isEnabledMCP,
			},
		}, nil)

	// Create request with auth context
	req := httptest.NewRequest(http.MethodGet, "/api/v1/sync/claude-code", nil)
	ctx := middleware.WithTestAuth(req.Context(), employeeID, orgID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	// Execute
	handler.GetClaudeCodeSync(rr, req)

	// Assert
	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "agents")
	assert.Contains(t, response, "skills")
	assert.Contains(t, response, "mcp_servers")
	assert.Contains(t, response, "version")
	assert.Contains(t, response, "synced_at")

	// Verify agents
	agents := response["agents"].([]interface{})
	assert.Len(t, agents, 1)
	agent := agents[0].(map[string]interface{})
	assert.Equal(t, "go-backend-developer", agent["name"])
	assert.Equal(t, "claude-code", agent["type"])

	// Verify skills
	skills := response["skills"].([]interface{})
	assert.Len(t, skills, 1)
	skill := skills[0].(map[string]interface{})
	assert.Equal(t, "release-manager", skill["name"])

	// Verify MCP servers
	mcpServers := response["mcp_servers"].([]interface{})
	assert.Len(t, mcpServers, 1)
	mcp := mcpServers[0].(map[string]interface{})
	assert.Equal(t, "github-mcp-server", mcp["name"])
}

// TestGetClaudeCodeSync_EmptyConfigurations tests sync with no resources assigned
func TestGetClaudeCodeSync_EmptyConfigurations(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)
	handler := handlers.NewSyncHandler(mockDB)

	employeeID := uuid.New()
	orgID := uuid.New()

	// Mock empty responses
	mockDB.EXPECT().
		ListEmployeeAgentConfigs(gomock.Any(), employeeID).
		Return([]db.ListEmployeeAgentConfigsRow{}, nil)

	mockDB.EXPECT().
		ListEmployeeSkills(gomock.Any(), employeeID).
		Return([]db.ListEmployeeSkillsRow{}, nil)

	mockDB.EXPECT().
		ListEmployeeMCPConfigs(gomock.Any(), employeeID).
		Return([]db.ListEmployeeMCPConfigsRow{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sync/claude-code", nil)
	ctx := middleware.WithTestAuth(req.Context(), employeeID, orgID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.GetClaudeCodeSync(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	// Empty arrays should still be present
	agents := response["agents"].([]interface{})
	skills := response["skills"].([]interface{})
	mcpServers := response["mcp_servers"].([]interface{})

	assert.Empty(t, agents)
	assert.Empty(t, skills)
	assert.Empty(t, mcpServers)
}

// TestGetClaudeCodeSync_MissingAuthContext tests request without auth context
func TestGetClaudeCodeSync_MissingAuthContext(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)
	handler := handlers.NewSyncHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sync/claude-code", nil)
	rr := httptest.NewRecorder()

	handler.GetClaudeCodeSync(rr, req)

	assert.Equal(t, http.StatusUnauthorized, rr.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Missing authentication", response["error"])
}

// TestGetClaudeCodeSync_DatabaseError tests database failure handling
func TestGetClaudeCodeSync_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)
	handler := handlers.NewSyncHandler(mockDB)

	employeeID := uuid.New()
	orgID := uuid.New()

	// Mock database error
	mockDB.EXPECT().
		ListEmployeeAgentConfigs(gomock.Any(), employeeID).
		Return(nil, assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sync/claude-code", nil)
	ctx := middleware.WithTestAuth(req.Context(), employeeID, orgID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.GetClaudeCodeSync(rr, req)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}

// TestGetClaudeCodeSync_PartialConfiguration tests sync with only some resources
func TestGetClaudeCodeSync_PartialConfiguration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)
	handler := handlers.NewSyncHandler(mockDB)

	employeeID := uuid.New()
	orgID := uuid.New()
	agentID := uuid.New()

	// Mock only agents (no skills or MCPs)
	mockDB.EXPECT().
		ListEmployeeAgentConfigs(gomock.Any(), employeeID).
		Return([]db.ListEmployeeAgentConfigsRow{
			{
				ID:                 uuid.New(),
				EmployeeID:         employeeID,
				AgentID:            agentID,
				ConfigOverride:     []byte(`{}`),
				IsEnabled:          true,
				AgentName:          "go-backend-developer",
				AgentType:          "claude-code",
				AgentProvider:      "anthropic",
				AgentDefaultConfig: []byte(`{}`),
			},
		}, nil)

	mockDB.EXPECT().
		ListEmployeeSkills(gomock.Any(), employeeID).
		Return([]db.ListEmployeeSkillsRow{}, nil)

	mockDB.EXPECT().
		ListEmployeeMCPConfigs(gomock.Any(), employeeID).
		Return([]db.ListEmployeeMCPConfigsRow{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sync/claude-code", nil)
	ctx := middleware.WithTestAuth(req.Context(), employeeID, orgID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.GetClaudeCodeSync(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	agents := response["agents"].([]interface{})
	skills := response["skills"].([]interface{})
	mcpServers := response["mcp_servers"].([]interface{})

	assert.Len(t, agents, 1)
	assert.Empty(t, skills)
	assert.Empty(t, mcpServers)
}

// TestGetClaudeCodeSync_VersionAndTimestamp tests version and timestamp in response
func TestGetClaudeCodeSync_VersionAndTimestamp(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mockdb.NewMockQuerier(ctrl)
	handler := handlers.NewSyncHandler(mockDB)

	employeeID := uuid.New()
	orgID := uuid.New()

	mockDB.EXPECT().ListEmployeeAgentConfigs(gomock.Any(), employeeID).Return([]db.ListEmployeeAgentConfigsRow{}, nil)
	mockDB.EXPECT().ListEmployeeSkills(gomock.Any(), employeeID).Return([]db.ListEmployeeSkillsRow{}, nil)
	mockDB.EXPECT().ListEmployeeMCPConfigs(gomock.Any(), employeeID).Return([]db.ListEmployeeMCPConfigsRow{}, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/sync/claude-code", nil)
	ctx := middleware.WithTestAuth(req.Context(), employeeID, orgID)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()

	handler.GetClaudeCodeSync(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rr.Body).Decode(&response)
	require.NoError(t, err)

	// Check version
	assert.Equal(t, "1.0.0", response["version"])

	// Check timestamp is recent (within 1 second window)
	syncedAt, err := time.Parse(time.RFC3339, response["synced_at"].(string))
	require.NoError(t, err)

	// Allow for slight timing differences - synced_at should be within 1 second of test execution
	assert.WithinDuration(t, time.Now().UTC(), syncedAt, time.Second)
}
