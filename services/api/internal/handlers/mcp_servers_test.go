package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/handlers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestListMCPServers_Success tests successful retrieval of MCP servers
func TestListMCPServers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Mock MCP server data
	dockerImage1 := "modelcontextprotocol/server-github:latest"
	server1 := db.McpCatalog{
		ID:                  uuid.New(),
		Name:                "GitHub",
		Provider:            "Anthropic",
		Version:             "1.0.0",
		Description:         "GitHub integration MCP server",
		ConnectionSchema:    []byte(`{"token":"string"}`),
		Capabilities:        []byte(`["repository_access","issue_management"]`),
		RequiresCredentials: true,
		IsApproved:          true,
		DockerImage:         &dockerImage1,
		ConfigTemplate:      []byte(`{"token":"${GITHUB_TOKEN}"}`),
		RequiredEnvVars:     []byte(`["GITHUB_TOKEN"]`),
		CreatedAt:           pgtype.Timestamp{Valid: true},
		UpdatedAt:           pgtype.Timestamp{Valid: true},
	}

	dockerImage2 := "modelcontextprotocol/server-filesystem:latest"
	server2 := db.McpCatalog{
		ID:                  uuid.New(),
		Name:                "Filesystem",
		Provider:            "Anthropic",
		Version:             "1.0.0",
		Description:         "Local filesystem access",
		ConnectionSchema:    []byte(`{}`),
		Capabilities:        []byte(`["read","write"]`),
		RequiresCredentials: false,
		IsApproved:          true,
		DockerImage:         &dockerImage2,
		ConfigTemplate:      []byte(`{}`),
		RequiredEnvVars:     []byte(`[]`),
		CreatedAt:           pgtype.Timestamp{Valid: true},
		UpdatedAt:           pgtype.Timestamp{Valid: true},
	}

	// Expect ListMCPServers to be called
	mockDB.EXPECT().
		ListMCPServers(gomock.Any()).
		Return([]db.McpCatalog{server1, server2}, nil)

	handler := handlers.NewMCPServersHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/mcp-servers", nil)
	rec := httptest.NewRecorder()

	handler.ListMCPServers(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListMCPServersResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Servers))
	assert.Equal(t, 2, response.Total)

	// Verify first server
	assert.Equal(t, "GitHub", response.Servers[0].Name)
	assert.Equal(t, "Anthropic", response.Servers[0].Provider)
	assert.Equal(t, true, response.Servers[0].IsApproved)

	// Verify second server
	assert.Equal(t, "Filesystem", response.Servers[1].Name)
	assert.NotNil(t, response.Servers[1].RequiresCredentials)
	assert.Equal(t, false, *response.Servers[1].RequiresCredentials)
}

// TestListMCPServers_EmptyResult tests when no MCP servers are available
func TestListMCPServers_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Return empty list
	mockDB.EXPECT().
		ListMCPServers(gomock.Any()).
		Return([]db.McpCatalog{}, nil)

	handler := handlers.NewMCPServersHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/mcp-servers", nil)
	rec := httptest.NewRecorder()

	handler.ListMCPServers(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListMCPServersResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 0, len(response.Servers))
	assert.Equal(t, 0, response.Total)
}

// TestListMCPServers_DatabaseError tests database error handling
func TestListMCPServers_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Simulate database error
	mockDB.EXPECT().
		ListMCPServers(gomock.Any()).
		Return(nil, assert.AnError)

	handler := handlers.NewMCPServersHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/mcp-servers", nil)
	rec := httptest.NewRecorder()

	handler.ListMCPServers(rec, req)

	// Verify error response
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var errorResponse api.Error
	err := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Failed to fetch MCP servers", errorResponse.Error)
}

// TestGetMCPServer_Success tests successful retrieval of a specific MCP server
func TestGetMCPServer_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	serverID := uuid.New()
	dockerImage := "modelcontextprotocol/server-github:latest"
	server := db.McpCatalog{
		ID:                  serverID,
		Name:                "GitHub",
		Provider:            "Anthropic",
		Version:             "1.0.0",
		Description:         "GitHub integration MCP server",
		ConnectionSchema:    []byte(`{"token":"string"}`),
		Capabilities:        []byte(`["repository_access","issue_management"]`),
		RequiresCredentials: true,
		IsApproved:          true,
		DockerImage:         &dockerImage,
		ConfigTemplate:      []byte(`{"token":"${GITHUB_TOKEN}"}`),
		RequiredEnvVars:     []byte(`["GITHUB_TOKEN"]`),
		CreatedAt:           pgtype.Timestamp{Valid: true},
		UpdatedAt:           pgtype.Timestamp{Valid: true},
	}

	// Expect GetMCPServer to be called
	mockDB.EXPECT().
		GetMCPServer(gomock.Any(), serverID).
		Return(server, nil)

	handler := handlers.NewMCPServersHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/mcp-servers/"+serverID.String(), nil)
	rec := httptest.NewRecorder()

	handler.GetMCPServer(rec, req, serverID)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.MCPServer
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, "GitHub", response.Name)
	assert.Equal(t, "Anthropic", response.Provider)
	assert.Equal(t, "1.0.0", response.Version)
	assert.Equal(t, true, response.IsApproved)
	assert.NotNil(t, response.RequiresCredentials)
	assert.Equal(t, true, *response.RequiresCredentials)
}

// TestGetMCPServer_NotFound tests MCP server not found
func TestGetMCPServer_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	serverID := uuid.New()

	// Simulate not found
	mockDB.EXPECT().
		GetMCPServer(gomock.Any(), serverID).
		Return(db.McpCatalog{}, pgx.ErrNoRows)

	handler := handlers.NewMCPServersHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/mcp-servers/"+serverID.String(), nil)
	rec := httptest.NewRecorder()

	handler.GetMCPServer(rec, req, serverID)

	// Verify error response
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var errorResponse api.Error
	err := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "MCP server not found", errorResponse.Error)
}

// TestGetMCPServer_DatabaseError tests database error handling
func TestGetMCPServer_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	serverID := uuid.New()

	// Simulate database error
	mockDB.EXPECT().
		GetMCPServer(gomock.Any(), serverID).
		Return(db.McpCatalog{}, assert.AnError)

	handler := handlers.NewMCPServersHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/mcp-servers/"+serverID.String(), nil)
	rec := httptest.NewRecorder()

	handler.GetMCPServer(rec, req, serverID)

	// Verify error response
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var errorResponse api.Error
	err := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Failed to fetch MCP server", errorResponse.Error)
}

// TestListEmployeeMCPServers_Success tests successful retrieval of employee's MCP configs
func TestListEmployeeMCPServers_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	employeeID := uuid.New()
	orgID := uuid.New()

	// Mock employee MCP config data
	enabled := true
	dockerImage1 := "modelcontextprotocol/server-github:latest"
	dockerImage2 := "modelcontextprotocol/server-filesystem:latest"
	config1 := db.ListEmployeeMCPConfigsRow{
		ID:               uuid.New(),
		Name:             "GitHub",
		Provider:         "Anthropic",
		Version:          "1.0.0",
		Description:      "GitHub integration",
		DockerImage:      &dockerImage1,
		ConfigTemplate:   []byte(`{"token":"${GITHUB_TOKEN}"}`),
		RequiredEnvVars:  []byte(`["GITHUB_TOKEN"]`),
		ConnectionConfig: []byte(`{"token":"ghp_xxx"}`),
		IsEnabled:        &enabled,
		CreatedAt:        pgtype.Timestamp{Valid: true},
	}

	config2 := db.ListEmployeeMCPConfigsRow{
		ID:               uuid.New(),
		Name:             "Filesystem",
		Provider:         "Anthropic",
		Version:          "1.0.0",
		Description:      "Local filesystem",
		DockerImage:      &dockerImage2,
		ConfigTemplate:   []byte(`{}`),
		RequiredEnvVars:  []byte(`[]`),
		ConnectionConfig: []byte(`{}`),
		IsEnabled:        &enabled,
		CreatedAt:        pgtype.Timestamp{Valid: true},
	}

	// Expect ListEmployeeMCPConfigs to be called
	mockDB.EXPECT().
		ListEmployeeMCPConfigs(gomock.Any(), employeeID).
		Return([]db.ListEmployeeMCPConfigsRow{config1, config2}, nil)

	handler := handlers.NewMCPServersHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/employees/me/mcp-servers", nil)
	req = req.WithContext(handlers.SetEmployeeIDInContext(req.Context(), employeeID))
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListEmployeeMCPServers(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListEmployeeMCPServersResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 2, len(response.Servers))
	assert.Equal(t, 2, response.Total)

	// Verify first config
	assert.Equal(t, "GitHub", response.Servers[0].Name)
	assert.Equal(t, true, response.Servers[0].IsEnabled)

	// Verify second config
	assert.Equal(t, "Filesystem", response.Servers[1].Name)
}

// TestListEmployeeMCPServers_Unauthorized tests missing employee ID
func TestListEmployeeMCPServers_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	handler := handlers.NewMCPServersHandler(mockDB)

	// Request without employee ID in context
	req := httptest.NewRequest(http.MethodGet, "/employees/me/mcp-servers", nil)
	rec := httptest.NewRecorder()

	handler.ListEmployeeMCPServers(rec, req)

	// Verify error response
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var errorResponse api.Error
	err := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Unauthorized", errorResponse.Error)
}

// TestListEmployeeMCPServers_EmptyResult tests when employee has no MCP configs
func TestListEmployeeMCPServers_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	employeeID := uuid.New()
	orgID := uuid.New()

	// Return empty list
	mockDB.EXPECT().
		ListEmployeeMCPConfigs(gomock.Any(), employeeID).
		Return([]db.ListEmployeeMCPConfigsRow{}, nil)

	handler := handlers.NewMCPServersHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/employees/me/mcp-servers", nil)
	req = req.WithContext(handlers.SetEmployeeIDInContext(req.Context(), employeeID))
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListEmployeeMCPServers(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListEmployeeMCPServersResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, 0, len(response.Servers))
	assert.Equal(t, 0, response.Total)
}

// TestListEmployeeMCPServers_DatabaseError tests database error handling
func TestListEmployeeMCPServers_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	employeeID := uuid.New()
	orgID := uuid.New()

	// Simulate database error
	mockDB.EXPECT().
		ListEmployeeMCPConfigs(gomock.Any(), employeeID).
		Return(nil, assert.AnError)

	handler := handlers.NewMCPServersHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/employees/me/mcp-servers", nil)
	req = req.WithContext(handlers.SetEmployeeIDInContext(req.Context(), employeeID))
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListEmployeeMCPServers(rec, req)

	// Verify error response
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var errorResponse api.Error
	err := json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Equal(t, "Failed to fetch employee MCP servers", errorResponse.Error)
}
