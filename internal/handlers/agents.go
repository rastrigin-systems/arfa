package handlers

import (
	"encoding/json"
	"net/http"

	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
)

// AgentsHandler handles agent-related requests
type AgentsHandler struct {
	db db.Querier
}

// NewAgentsHandler creates a new agents handler
func NewAgentsHandler(database db.Querier) *AgentsHandler {
	return &AgentsHandler{
		db: database,
	}
}

// ListAgents handles GET /agents
// Returns list of available AI agents from the catalog
func (h *AgentsHandler) ListAgents(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Query database for all public agents
	agents, err := h.db.ListAgents(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch agents")
		return
	}

	// Convert db.Agent to api.Agent
	apiAgents := make([]api.Agent, len(agents))
	for i, agent := range agents {
		apiAgents[i] = dbAgentToAPI(agent)
	}

	// Build response
	response := api.ListAgentsResponse{
		Agents: apiAgents,
		Total:  len(apiAgents),
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// dbAgentToAPI converts db.Agent to api.Agent
func dbAgentToAPI(agent db.Agent) api.Agent {
	apiAgent := api.Agent{
		Id:          (*openapi_types.UUID)(&agent.ID),
		Name:        agent.Name,
		Type:        agent.Type,
		Description: agent.Description,
		Provider:    agent.Provider,
		LlmProvider: agent.LlmProvider,
		LlmModel:    agent.LlmModel,
		IsPublic:    agent.IsPublic,
	}

	// Handle nullable fields
	if agent.CreatedAt.Valid {
		apiAgent.CreatedAt = &agent.CreatedAt.Time
	}

	if agent.UpdatedAt.Valid {
		apiAgent.UpdatedAt = &agent.UpdatedAt.Time
	}

	// Handle default_config JSON
	if len(agent.DefaultConfig) > 0 && string(agent.DefaultConfig) != "null" {
		var config map[string]interface{}
		if err := json.Unmarshal(agent.DefaultConfig, &config); err == nil {
			apiAgent.DefaultConfig = &config
		}
	}

	// Handle capabilities JSON
	if len(agent.Capabilities) > 0 && string(agent.Capabilities) != "null" {
		var capabilities []string
		if err := json.Unmarshal(agent.Capabilities, &capabilities); err == nil {
			apiAgent.Capabilities = &capabilities
		}
	}

	return apiAgent
}
