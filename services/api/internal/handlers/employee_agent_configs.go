package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
)

// EmployeeAgentConfigsHandler handles employee-level agent configuration requests
type EmployeeAgentConfigsHandler struct {
	db db.Querier
}

// NewEmployeeAgentConfigsHandler creates a new employee agent configs handler
func NewEmployeeAgentConfigsHandler(database db.Querier) *EmployeeAgentConfigsHandler {
	return &EmployeeAgentConfigsHandler{
		db: database,
	}
}

// ListEmployeeAgentConfigs handles GET /employees/{employee_id}/agent-configs
// Returns list of agent configuration overrides for a specific employee
func (h *EmployeeAgentConfigsHandler) ListEmployeeAgentConfigs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse employee ID from URL
	employeeIDStr := chi.URLParam(r, "employee_id")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid employee ID")
		return
	}

	// Verify employee exists and belongs to current org
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	employee, err := h.db.GetEmployee(ctx, employeeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch employee")
		return
	}

	// Verify employee belongs to current org
	if employee.OrgID != orgID {
		writeError(w, http.StatusNotFound, "Employee not found")
		return
	}

	// Query database for all employee-level agent configs
	configs, err := h.db.ListEmployeeAgentConfigs(ctx, employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch employee agent configs")
		return
	}

	// Convert to API response
	apiConfigs := make([]api.EmployeeAgentConfig, len(configs))
	for i, cfg := range configs {
		apiConfigs[i] = dbEmployeeAgentConfigToAPI(cfg)
	}

	// Build response
	response := api.ListEmployeeAgentConfigsResponse{
		Configs: apiConfigs,
		Total:   len(apiConfigs),
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateEmployeeAgentConfig handles POST /employees/{employee_id}/agent-configs
// Creates a new employee-level agent configuration override
func (h *EmployeeAgentConfigsHandler) CreateEmployeeAgentConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse employee ID from URL
	employeeIDStr := chi.URLParam(r, "employee_id")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid employee ID")
		return
	}

	// Verify employee exists and belongs to current org
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	employee, err := h.db.GetEmployee(ctx, employeeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch employee")
		return
	}

	// Verify employee belongs to current org
	if employee.OrgID != orgID {
		writeError(w, http.StatusNotFound, "Employee not found")
		return
	}

	// Parse request body
	var req api.CreateEmployeeAgentConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate required fields
	if req.ConfigOverride == nil {
		writeError(w, http.StatusBadRequest, "config_override is required")
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
	exists, err := h.db.CheckEmployeeAgentExists(ctx, db.CheckEmployeeAgentExistsParams{
		EmployeeID: employeeID,
		AgentID:    agentID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to check existing config")
		return
	}

	if exists {
		writeError(w, http.StatusConflict, "Agent already configured for this employee")
		return
	}

	// Marshal config override to JSON
	configJSON, err := json.Marshal(req.ConfigOverride)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid config format")
		return
	}

	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}

	// Create config
	config, err := h.db.CreateEmployeeAgentConfig(ctx, db.CreateEmployeeAgentConfigParams{
		EmployeeID:     employeeID,
		AgentID:        agentID,
		ConfigOverride: configJSON,
		IsEnabled:      isEnabled,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create config")
		return
	}

	// Get agent details for response
	agent, _ := h.db.GetAgentByID(ctx, agentID)

	// Build response
	response := dbCreateEmployeeAgentConfigRowToAPI(config, agent)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetEmployeeAgentConfig handles GET /employees/{employee_id}/agent-configs/{config_id}
// Returns a specific employee agent configuration
func (h *EmployeeAgentConfigsHandler) GetEmployeeAgentConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse employee ID from URL
	employeeIDStr := chi.URLParam(r, "employee_id")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid employee ID")
		return
	}

	// Parse config ID from URL
	configIDStr := chi.URLParam(r, "config_id")
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid config ID")
		return
	}

	// Verify employee exists and belongs to current org
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	employee, err := h.db.GetEmployee(ctx, employeeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch employee")
		return
	}

	// Verify employee belongs to current org
	if employee.OrgID != orgID {
		writeError(w, http.StatusNotFound, "Employee not found")
		return
	}

	// Get config by ID
	config, err := h.db.GetEmployeeAgentConfig(ctx, configID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Config not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch config")
		return
	}

	// Verify config belongs to this employee
	if config.EmployeeID != employeeID {
		writeError(w, http.StatusNotFound, "Config not found")
		return
	}

	// Convert to API response (convert row type)
	listRow := db.ListEmployeeAgentConfigsRow{
		ID:             config.ID,
		EmployeeID:     config.EmployeeID,
		AgentID:        config.AgentID,
		ConfigOverride: config.ConfigOverride,
		IsEnabled:      config.IsEnabled,
		SyncToken:      config.SyncToken,
		LastSyncedAt:   config.LastSyncedAt,
		CreatedAt:      config.CreatedAt,
		UpdatedAt:      config.UpdatedAt,
		AgentName:      config.AgentName,
		AgentType:      config.AgentType,
		AgentProvider:  config.AgentProvider,
	}
	response := dbEmployeeAgentConfigToAPI(listRow)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateEmployeeAgentConfig handles PATCH /employees/{employee_id}/agent-configs/{config_id}
// Updates an existing employee agent configuration
func (h *EmployeeAgentConfigsHandler) UpdateEmployeeAgentConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse employee ID from URL
	employeeIDStr := chi.URLParam(r, "employee_id")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid employee ID")
		return
	}

	// Parse config ID from URL
	configIDStr := chi.URLParam(r, "config_id")
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid config ID")
		return
	}

	// Verify employee exists and belongs to current org
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	employee, err := h.db.GetEmployee(ctx, employeeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch employee")
		return
	}

	// Verify employee belongs to current org
	if employee.OrgID != orgID {
		writeError(w, http.StatusNotFound, "Employee not found")
		return
	}

	// Parse request body
	var req api.UpdateEmployeeAgentConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Build update params
	var configJSON []byte
	if req.ConfigOverride != nil {
		configJSON, err = json.Marshal(req.ConfigOverride)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid config format")
			return
		}
	}

	// Update config
	config, err := h.db.UpdateEmployeeAgentConfig(ctx, db.UpdateEmployeeAgentConfigParams{
		ID:             configID,
		ConfigOverride: configJSON,
		IsEnabled:      req.IsEnabled,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Config not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to update config")
		return
	}

	// Verify config belongs to this employee
	if config.EmployeeID != employeeID {
		writeError(w, http.StatusNotFound, "Config not found")
		return
	}

	// Get agent details for response
	agent, _ := h.db.GetAgentByID(ctx, config.AgentID)

	// Build response
	response := dbUpdateEmployeeAgentConfigRowToAPI(config, agent)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteEmployeeAgentConfig handles DELETE /employees/{employee_id}/agent-configs/{config_id}
// Deletes an employee agent configuration
func (h *EmployeeAgentConfigsHandler) DeleteEmployeeAgentConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse employee ID from URL
	employeeIDStr := chi.URLParam(r, "employee_id")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid employee ID")
		return
	}

	// Parse config ID from URL
	configIDStr := chi.URLParam(r, "config_id")
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid config ID")
		return
	}

	// Verify employee exists and belongs to current org
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	employee, err := h.db.GetEmployee(ctx, employeeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch employee")
		return
	}

	// Verify employee belongs to current org
	if employee.OrgID != orgID {
		writeError(w, http.StatusNotFound, "Employee not found")
		return
	}

	// Delete config
	err = h.db.DeleteEmployeeAgentConfig(ctx, configID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete config")
		return
	}

	// Write 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func dbEmployeeAgentConfigToAPI(config db.ListEmployeeAgentConfigsRow) api.EmployeeAgentConfig {
	var cfg map[string]interface{}
	if len(config.ConfigOverride) > 0 {
		json.Unmarshal(config.ConfigOverride, &cfg)
	}

	agentName := config.AgentName
	agentType := config.AgentType
	agentProvider := config.AgentProvider

	id := openapi_types.UUID(config.ID)
	employeeID := openapi_types.UUID(config.EmployeeID)
	agentID := openapi_types.UUID(config.AgentID)

	result := api.EmployeeAgentConfig{
		Id:             &id,
		EmployeeId:     employeeID,
		AgentId:        agentID,
		AgentName:      &agentName,
		AgentType:      &agentType,
		AgentProvider:  &agentProvider,
		ConfigOverride: cfg,
		IsEnabled:      config.IsEnabled,
		CreatedAt:      &config.CreatedAt.Time,
		UpdatedAt:      &config.UpdatedAt.Time,
	}

	// Handle nullable fields
	if config.SyncToken != nil {
		result.SyncToken = config.SyncToken
	}

	if config.LastSyncedAt.Valid {
		t := config.LastSyncedAt.Time
		result.LastSyncedAt = &t
	}

	return result
}

func dbEmployeeAgentConfigRowToAPI(config db.EmployeeAgentConfig, agent db.Agent) api.EmployeeAgentConfig {
	var cfg map[string]interface{}
	if len(config.ConfigOverride) > 0 {
		json.Unmarshal(config.ConfigOverride, &cfg)
	}

	agentName := agent.Name
	agentType := agent.Type
	agentProvider := agent.Provider

	id := openapi_types.UUID(config.ID)
	employeeID := openapi_types.UUID(config.EmployeeID)
	agentID := openapi_types.UUID(config.AgentID)

	result := api.EmployeeAgentConfig{
		Id:             &id,
		EmployeeId:     employeeID,
		AgentId:        agentID,
		AgentName:      &agentName,
		AgentType:      &agentType,
		AgentProvider:  &agentProvider,
		ConfigOverride: cfg,
		IsEnabled:      config.IsEnabled,
		CreatedAt:      &config.CreatedAt.Time,
		UpdatedAt:      &config.UpdatedAt.Time,
	}

	// Handle nullable fields
	if config.SyncToken != nil {
		result.SyncToken = config.SyncToken
	}

	if config.LastSyncedAt.Valid {
		t := config.LastSyncedAt.Time
		result.LastSyncedAt = &t
	}

	return result
}


// dbCreateEmployeeAgentConfigRowToAPI converts EmployeeAgentConfig (from Create) to api format
func dbCreateEmployeeAgentConfigRowToAPI(config db.EmployeeAgentConfig, agent db.Agent) api.EmployeeAgentConfig {
	return dbEmployeeAgentConfigRowToAPI(config, agent)
}

// dbUpdateEmployeeAgentConfigRowToAPI converts EmployeeAgentConfig (from Update) to api format
func dbUpdateEmployeeAgentConfigRowToAPI(config db.EmployeeAgentConfig, agent db.Agent) api.EmployeeAgentConfig {
	return dbEmployeeAgentConfigRowToAPI(config, agent)
}
