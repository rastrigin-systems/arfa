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

// TeamAgentConfigsHandler handles team-level agent configuration requests
type TeamAgentConfigsHandler struct {
	db db.Querier
}

// NewTeamAgentConfigsHandler creates a new team agent configs handler
func NewTeamAgentConfigsHandler(database db.Querier) *TeamAgentConfigsHandler {
	return &TeamAgentConfigsHandler{
		db: database,
	}
}

// ListTeamAgentConfigs handles GET /teams/{team_id}/agent-configs
// Returns list of agent configurations for a specific team
func (h *TeamAgentConfigsHandler) ListTeamAgentConfigs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse team ID from URL
	teamIDStr := chi.URLParam(r, "team_id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	// Verify team exists and belongs to current org
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	team, err := h.db.GetTeam(ctx, db.GetTeamParams{
		ID:    teamID,
		OrgID: orgID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Team not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch team")
		return
	}

	// Query database for all team-level agent configs
	configs, err := h.db.ListTeamAgentConfigs(ctx, team.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch team agent configs")
		return
	}

	// Convert to API response
	apiConfigs := make([]api.TeamAgentConfig, len(configs))
	for i, cfg := range configs {
		apiConfigs[i] = dbTeamAgentConfigToAPI(cfg)
	}

	// Build response
	response := api.ListTeamAgentConfigsResponse{
		Configs: apiConfigs,
		Total:   len(apiConfigs),
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateTeamAgentConfig handles POST /teams/{team_id}/agent-configs
// Creates a new team-level agent configuration override
func (h *TeamAgentConfigsHandler) CreateTeamAgentConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse team ID from URL
	teamIDStr := chi.URLParam(r, "team_id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	// Verify team exists and belongs to current org
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	team, err := h.db.GetTeam(ctx, db.GetTeamParams{
		ID:    teamID,
		OrgID: orgID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Team not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch team")
		return
	}

	// Parse request body
	var req api.CreateTeamAgentConfigRequest
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
	exists, err := h.db.CheckTeamAgentConfigExists(ctx, db.CheckTeamAgentConfigExistsParams{
		TeamID:  team.ID,
		AgentID: agentID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to check existing config")
		return
	}

	if exists {
		writeError(w, http.StatusConflict, "Agent already configured for this team")
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
	config, err := h.db.CreateTeamAgentConfig(ctx, db.CreateTeamAgentConfigParams{
		TeamID:         team.ID,
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
	response := dbTeamAgentConfigRowToAPI(config, agent)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetTeamAgentConfig handles GET /teams/{team_id}/agent-configs/{config_id}
// Returns a specific team agent configuration
func (h *TeamAgentConfigsHandler) GetTeamAgentConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse team ID from URL
	teamIDStr := chi.URLParam(r, "team_id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	// Parse config ID from URL
	configIDStr := chi.URLParam(r, "config_id")
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid config ID")
		return
	}

	// Verify team exists and belongs to current org
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	_, err = h.db.GetTeam(ctx, db.GetTeamParams{
		ID:    teamID,
		OrgID: orgID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Team not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch team")
		return
	}

	// Get config by ID
	config, err := h.db.GetTeamAgentConfigByID(ctx, configID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Config not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch config")
		return
	}

	// Verify config belongs to this team
	if config.TeamID != teamID {
		writeError(w, http.StatusNotFound, "Config not found")
		return
	}

	// Convert to API response (convert row type)
	listRow := db.ListTeamAgentConfigsRow{
		ID:             config.ID,
		TeamID:         config.TeamID,
		AgentID:        config.AgentID,
		ConfigOverride: config.ConfigOverride,
		IsEnabled:      config.IsEnabled,
		CreatedAt:      config.CreatedAt,
		UpdatedAt:      config.UpdatedAt,
		AgentName:      config.AgentName,
		AgentType:      config.AgentType,
		AgentProvider:  config.AgentProvider,
	}
	response := dbTeamAgentConfigToAPI(listRow)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateTeamAgentConfig handles PATCH /teams/{team_id}/agent-configs/{config_id}
// Updates an existing team agent configuration
func (h *TeamAgentConfigsHandler) UpdateTeamAgentConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse team ID from URL
	teamIDStr := chi.URLParam(r, "team_id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	// Parse config ID from URL
	configIDStr := chi.URLParam(r, "config_id")
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid config ID")
		return
	}

	// Verify team exists and belongs to current org
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	_, err = h.db.GetTeam(ctx, db.GetTeamParams{
		ID:    teamID,
		OrgID: orgID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Team not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch team")
		return
	}

	// Parse request body
	var req api.UpdateTeamAgentConfigRequest
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
	config, err := h.db.UpdateTeamAgentConfig(ctx, db.UpdateTeamAgentConfigParams{
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

	// Verify config belongs to this team
	if config.TeamID != teamID {
		writeError(w, http.StatusNotFound, "Config not found")
		return
	}

	// Get agent details for response
	agent, _ := h.db.GetAgentByID(ctx, config.AgentID)

	// Build response
	response := dbTeamAgentConfigRowToAPI(config, agent)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteTeamAgentConfig handles DELETE /teams/{team_id}/agent-configs/{config_id}
// Deletes a team agent configuration
func (h *TeamAgentConfigsHandler) DeleteTeamAgentConfig(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse team ID from URL
	teamIDStr := chi.URLParam(r, "team_id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	// Parse config ID from URL
	configIDStr := chi.URLParam(r, "config_id")
	configID, err := uuid.Parse(configIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid config ID")
		return
	}

	// Verify team exists and belongs to current org
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	_, err = h.db.GetTeam(ctx, db.GetTeamParams{
		ID:    teamID,
		OrgID: orgID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Team not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch team")
		return
	}

	// Delete config
	err = h.db.DeleteTeamAgentConfig(ctx, configID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete config")
		return
	}

	// Write 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func dbTeamAgentConfigToAPI(config db.ListTeamAgentConfigsRow) api.TeamAgentConfig {
	var cfg map[string]interface{}
	if len(config.ConfigOverride) > 0 {
		json.Unmarshal(config.ConfigOverride, &cfg)
	}

	agentName := config.AgentName
	agentType := config.AgentType
	agentProvider := config.AgentProvider

	id := openapi_types.UUID(config.ID)
	teamID := openapi_types.UUID(config.TeamID)
	agentID := openapi_types.UUID(config.AgentID)

	return api.TeamAgentConfig{
		Id:             &id,
		TeamId:         teamID,
		AgentId:        agentID,
		AgentName:      &agentName,
		AgentType:      &agentType,
		AgentProvider:  &agentProvider,
		ConfigOverride: cfg,
		IsEnabled:      config.IsEnabled,
		CreatedAt:      &config.CreatedAt.Time,
		UpdatedAt:      &config.UpdatedAt.Time,
	}
}

func dbTeamAgentConfigRowToAPI(config db.TeamAgentConfig, agent db.Agent) api.TeamAgentConfig {
	var cfg map[string]interface{}
	if len(config.ConfigOverride) > 0 {
		json.Unmarshal(config.ConfigOverride, &cfg)
	}

	agentName := agent.Name
	agentType := agent.Type
	agentProvider := agent.Provider

	id := openapi_types.UUID(config.ID)
	teamID := openapi_types.UUID(config.TeamID)
	agentID := openapi_types.UUID(config.AgentID)

	return api.TeamAgentConfig{
		Id:             &id,
		TeamId:         teamID,
		AgentId:        agentID,
		AgentName:      &agentName,
		AgentType:      &agentType,
		AgentProvider:  &agentProvider,
		ConfigOverride: cfg,
		IsEnabled:      config.IsEnabled,
		CreatedAt:      &config.CreatedAt.Time,
		UpdatedAt:      &config.UpdatedAt.Time,
	}
}
