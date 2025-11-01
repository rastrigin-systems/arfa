package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/service"
)

// OrgAgentConfigsHandler handles org-level agent configuration requests
type OrgAgentConfigsHandler struct {
	db       db.Querier
	resolver *service.ConfigResolver
}

// NewOrgAgentConfigsHandler creates a new org agent configs handler
func NewOrgAgentConfigsHandler(database db.Querier) *OrgAgentConfigsHandler {
	return &OrgAgentConfigsHandler{
		db:       database,
		resolver: service.NewConfigResolver(database),
	}
}

// ListOrgAgentConfigs handles GET /organizations/current/agent-configs
// Returns list of agent configurations for the current organization
func (h *OrgAgentConfigsHandler) ListOrgAgentConfigs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Query database for all org-level agent configs
	configs, err := h.db.ListOrgAgentConfigs(ctx, orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch org agent configs")
		return
	}

	// Convert to API response
	apiConfigs := make([]api.OrgAgentConfig, len(configs))
	for i, cfg := range configs {
		apiConfigs[i] = dbOrgAgentConfigToAPI(cfg)
	}

	// Build response
	response := api.ListOrgAgentConfigsResponse{
		Configs: apiConfigs,
		Total:   len(apiConfigs),
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateOrgAgentConfig handles POST /organizations/current/agent-configs
// Creates a new agent configuration for the organization
func (h *OrgAgentConfigsHandler) CreateOrgAgentConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var req api.CreateOrgAgentConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate required fields
	if req.Config == nil {
		writeError(w, http.StatusBadRequest, "config is required")
		return
	}

	agentID := uuid.UUID(req.AgentId)

	// Check if agent exists
	_, err = h.db.GetAgentByID(ctx, agentID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Agent not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to verify agent")
		return
	}

	// Check if already configured
	exists, err := h.db.CheckOrgAgentConfigExists(ctx, db.CheckOrgAgentConfigExistsParams{
		OrgID:   orgID,
		AgentID: agentID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to check existing config")
		return
	}

	if exists {
		writeError(w, http.StatusConflict, "Agent already configured for this organization")
		return
	}

	// Marshal config to JSON
	configJSON, err := json.Marshal(req.Config)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid config format")
		return
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	// Create config
	config, err := h.db.CreateOrgAgentConfig(ctx, db.CreateOrgAgentConfigParams{
		OrgID:     orgID,
		AgentID:   agentID,
		Config:    configJSON,
		IsEnabled: isEnabled,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create config")
		return
	}

	// Get agent details for response
	agent, _ := h.db.GetAgentByID(ctx, agentID)

	// Build response
	response := dbOrgAgentConfigRowToAPI(config, agent)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetOrgAgentConfig handles GET /organizations/current/agent-configs/{config_id}
// Returns a specific org agent configuration
func (h *OrgAgentConfigsHandler) GetOrgAgentConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse config ID from URL
	configIDStr := chi.URLParam(r, "config_id")
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid config ID")
		return
	}

	// Query database (Note: We don't have a direct GetOrgAgentConfigByID query)
	// For now, list all and filter. In production, add a proper query.
	configs, err := h.db.ListOrgAgentConfigs(ctx, orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch config")
		return
	}

	// Find matching config
	var foundConfig *db.ListOrgAgentConfigsRow
	for _, cfg := range configs {
		if cfg.ID == configID {
			foundConfig = &cfg
			break
		}
	}

	if foundConfig == nil {
		writeError(w, http.StatusNotFound, "Config not found")
		return
	}

	// Convert to API response
	response := dbOrgAgentConfigToAPI(*foundConfig)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateOrgAgentConfig handles PATCH /organizations/current/agent-configs/{config_id}
// Updates an existing org agent configuration
func (h *OrgAgentConfigsHandler) UpdateOrgAgentConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse config ID from URL
	configIDStr := chi.URLParam(r, "config_id")
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid config ID")
		return
	}

	// Parse request body
	var req api.UpdateOrgAgentConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Build update params
	var configJSON []byte
	if req.Config != nil {
		configJSON, err = json.Marshal(req.Config)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid config format")
			return
		}
	}

	// Build isEnabled nullable bool
	var isEnabled *bool
	if req.IsEnabled != nil {
		isEnabled = req.IsEnabled
	}

	// Update config
	config, err := h.db.UpdateOrgAgentConfig(ctx, db.UpdateOrgAgentConfigParams{
		ID:        configID,
		Config:    configJSON,
		IsEnabled: isEnabled,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Config not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to update config")
		return
	}

	// Get agent details for response
	agent, _ := h.db.GetAgentByID(ctx, config.AgentID)

	// Build response
	response := dbOrgAgentConfigRowToAPI(config, agent)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteOrgAgentConfig handles DELETE /organizations/current/agent-configs/{config_id}
// Deletes an org agent configuration
func (h *OrgAgentConfigsHandler) DeleteOrgAgentConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse config ID from URL
	configIDStr := chi.URLParam(r, "config_id")
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid config ID")
		return
	}

	// Delete config
	err = h.db.DeleteOrgAgentConfig(ctx, configID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete config")
		return
	}

	// Write 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// GetEmployeeResolvedAgentConfigs handles GET /employees/{employee_id}/agent-configs/resolved
// Returns fully resolved agent configs for an employee (org → team → employee)
// This is the PRIMARY endpoint used by the CLI for sync
func (h *OrgAgentConfigsHandler) GetEmployeeResolvedAgentConfigs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse employee ID from URL
	employeeIDStr := chi.URLParam(r, "employee_id")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid employee ID")
		return
	}

	// Verify employee exists
	_, err = h.db.GetEmployee(ctx, employeeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch employee")
		return
	}

	// Resolve all agent configs for this employee
	resolvedConfigs, err := h.resolver.ResolveEmployeeAgents(ctx, employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to resolve agent configs")
		return
	}

	// Convert to API response
	apiConfigs := make([]api.ResolvedAgentConfig, len(resolvedConfigs))
	for i, cfg := range resolvedConfigs {
		apiConfigs[i] = resolvedAgentConfigToAPI(cfg)
	}

	// Build response
	response := api.ListResolvedAgentConfigsResponse{
		Configs: apiConfigs,
		Total:   len(apiConfigs),
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper functions

func dbOrgAgentConfigToAPI(config db.ListOrgAgentConfigsRow) api.OrgAgentConfig {
	var cfg map[string]interface{}
	if len(config.Config) > 0 {
		json.Unmarshal(config.Config, &cfg)
	}

	agentName := config.AgentName
	agentType := config.AgentType
	agentProvider := config.AgentProvider

	id := openapi_types.UUID(config.ID)
	orgId := openapi_types.UUID(config.OrgID)
	agentId := openapi_types.UUID(config.AgentID)

	return api.OrgAgentConfig{
		Id:            &id,
		OrgId:         orgId,
		AgentId:       agentId,
		AgentName:     &agentName,
		AgentType:     &agentType,
		AgentProvider: &agentProvider,
		Config:        cfg,
		IsEnabled:     config.IsEnabled,
		CreatedAt:     &config.CreatedAt.Time,
		UpdatedAt:     &config.UpdatedAt.Time,
	}
}

func dbOrgAgentConfigRowToAPI(config db.OrgAgentConfig, agent db.Agent) api.OrgAgentConfig {
	var cfg map[string]interface{}
	if len(config.Config) > 0 {
		json.Unmarshal(config.Config, &cfg)
	}

	agentName := agent.Name
	agentType := agent.Type
	agentProvider := agent.Provider

	id := openapi_types.UUID(config.ID)
	orgId := openapi_types.UUID(config.OrgID)
	agentId := openapi_types.UUID(config.AgentID)

	return api.OrgAgentConfig{
		Id:            &id,
		OrgId:         orgId,
		AgentId:       agentId,
		AgentName:     &agentName,
		AgentType:     &agentType,
		AgentProvider: &agentProvider,
		Config:        cfg,
		IsEnabled:     config.IsEnabled,
		CreatedAt:     &config.CreatedAt.Time,
		UpdatedAt:     &config.UpdatedAt.Time,
	}
}

func resolvedAgentConfigToAPI(config service.ResolvedAgentConfig) api.ResolvedAgentConfig {
	agentId := openapi_types.UUID(config.AgentID)

	var lastSyncedAt *time.Time
	if config.LastSyncedAt != nil {
		t, _ := time.Parse(time.RFC3339, *config.LastSyncedAt)
		lastSyncedAt = &t
	}

	return api.ResolvedAgentConfig{
		AgentId:      agentId,
		AgentName:    config.AgentName,
		AgentType:    config.AgentType,
		Provider:     config.Provider,
		Config:       config.Config,
		SystemPrompt: config.SystemPrompt,
		IsEnabled:    config.IsEnabled,
		SyncToken:    config.SyncToken,
		LastSyncedAt: lastSyncedAt,
	}
}
