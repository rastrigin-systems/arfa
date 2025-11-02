package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
)

// MCPServersHandler handles MCP server-related requests
type MCPServersHandler struct {
	db db.Querier
}

// NewMCPServersHandler creates a new MCP servers handler
func NewMCPServersHandler(database db.Querier) *MCPServersHandler {
	return &MCPServersHandler{
		db: database,
	}
}

// ListMCPServers handles GET /mcp-servers
// Returns list of approved MCP servers from the catalog
func (h *MCPServersHandler) ListMCPServers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Query database for approved MCP servers
	servers, err := h.db.ListMCPServers(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch MCP servers")
		return
	}

	// Convert db.McpCatalog to api.MCPServer
	apiServers := make([]api.MCPServer, len(servers))
	for i, server := range servers {
		apiServers[i] = dbMCPServerToAPI(server)
	}

	// Build response
	response := api.ListMCPServersResponse{
		Servers: apiServers,
		Total:   len(apiServers),
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetMCPServer handles GET /mcp-servers/{id}
// Returns a specific MCP server by ID
func (h *MCPServersHandler) GetMCPServer(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract id from URL path using Chi's URLParam
	serverIDStr := chi.URLParam(r, "id")
	if serverIDStr == "" {
		writeError(w, http.StatusBadRequest, "Missing server id")
		return
	}

	// Parse UUID
	serverID, err := uuid.Parse(serverIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid server id format")
		return
	}

	// Query database
	server, err := h.db.GetMCPServer(ctx, serverID)
	if err != nil {
		writeError(w, http.StatusNotFound, "MCP server not found")
		return
	}

	// Convert to API format
	apiServer := dbMCPServerToAPI(server)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiServer)
}

// ListEmployeeMCPServers handles GET /employees/me/mcp-servers
// Returns list of MCP servers configured for the authenticated employee
func (h *MCPServersHandler) ListEmployeeMCPServers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get employee ID from context (set by JWT middleware)
	employeeID, err := GetEmployeeID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Query database for employee's MCP configurations
	empMCPs, err := h.db.ListEmployeeMCPConfigs(ctx, employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch employee MCP servers")
		return
	}

	// Convert to API format
	apiServers := make([]api.EmployeeMCPServer, len(empMCPs))
	for i, empMCP := range empMCPs {
		apiServers[i] = dbEmployeeMCPConfigToAPI(empMCP)
	}

	// Build response
	response := api.ListEmployeeMCPServersResponse{
		Servers: apiServers,
		Total:   len(apiServers),
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// dbMCPServerToAPI converts db.McpCatalog to api.MCPServer
func dbMCPServerToAPI(server db.McpCatalog) api.MCPServer {
	apiServer := api.MCPServer{
		Id:          (*openapi_types.UUID)(&server.ID),
		Name:        server.Name,
		Provider:    server.Provider,
		Version:     server.Version,
		Description: server.Description,
		IsApproved:  server.IsApproved,
	}

	// Handle nullable/optional fields
	if server.CreatedAt.Valid {
		apiServer.CreatedAt = &server.CreatedAt.Time
	}

	if server.UpdatedAt.Valid {
		apiServer.UpdatedAt = &server.UpdatedAt.Time
	}

	if server.DockerImage != nil && *server.DockerImage != "" {
		apiServer.DockerImage = server.DockerImage
	}

	if server.CategoryID.Valid {
		categoryUUID := (openapi_types.UUID)(server.CategoryID.Bytes)
		apiServer.CategoryId = &categoryUUID
	}

	// Handle JSON fields
	if len(server.ConnectionSchema) > 0 && string(server.ConnectionSchema) != "null" {
		var schema map[string]interface{}
		if err := json.Unmarshal(server.ConnectionSchema, &schema); err == nil {
			apiServer.ConnectionSchema = &schema
		}
	}

	if len(server.Capabilities) > 0 && string(server.Capabilities) != "null" {
		var capabilities []string
		if err := json.Unmarshal(server.Capabilities, &capabilities); err == nil {
			apiServer.Capabilities = &capabilities
		}
	}

	if len(server.ConfigTemplate) > 0 && string(server.ConfigTemplate) != "null" {
		var template map[string]interface{}
		if err := json.Unmarshal(server.ConfigTemplate, &template); err == nil {
			apiServer.ConfigTemplate = &template
		}
	}

	if len(server.RequiredEnvVars) > 0 && string(server.RequiredEnvVars) != "null" {
		var envVars []string
		if err := json.Unmarshal(server.RequiredEnvVars, &envVars); err == nil {
			apiServer.RequiredEnvVars = &envVars
		}
	}

	// Set requires_credentials default if not set
	apiServer.RequiresCredentials = &server.RequiresCredentials

	return apiServer
}

// dbEmployeeMCPConfigToAPI converts db.ListEmployeeMCPConfigsRow to api.EmployeeMCPServer
func dbEmployeeMCPConfigToAPI(empMCP db.ListEmployeeMCPConfigsRow) api.EmployeeMCPServer {
	isEnabled := false
	if empMCP.IsEnabled != nil {
		isEnabled = *empMCP.IsEnabled
	}

	apiServer := api.EmployeeMCPServer{
		Id:          (openapi_types.UUID)(empMCP.ID),
		Name:        empMCP.Name,
		Provider:    empMCP.Provider,
		Version:     empMCP.Version,
		Description: empMCP.Description,
		IsEnabled:   isEnabled,
	}

	// Handle configured_at timestamp
	if empMCP.CreatedAt.Valid {
		apiServer.ConfiguredAt = empMCP.CreatedAt.Time
	}

	// Handle nullable fields
	if empMCP.DockerImage != nil && *empMCP.DockerImage != "" {
		apiServer.DockerImage = empMCP.DockerImage
	}

	// Handle JSON fields
	if len(empMCP.ConfigTemplate) > 0 && string(empMCP.ConfigTemplate) != "null" {
		var template map[string]interface{}
		if err := json.Unmarshal(empMCP.ConfigTemplate, &template); err == nil {
			apiServer.ConfigTemplate = &template
		}
	}

	if len(empMCP.RequiredEnvVars) > 0 && string(empMCP.RequiredEnvVars) != "null" {
		var envVars []string
		if err := json.Unmarshal(empMCP.RequiredEnvVars, &envVars); err == nil {
			apiServer.RequiredEnvVars = &envVars
		}
	}

	if len(empMCP.ConnectionConfig) > 0 && string(empMCP.ConnectionConfig) != "null" {
		var config map[string]interface{}
		if err := json.Unmarshal(empMCP.ConnectionConfig, &config); err == nil {
			apiServer.ConnectionConfig = &config
		}
	}

	return apiServer
}
