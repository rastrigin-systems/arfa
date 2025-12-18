package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rastrigin-systems/ubik-enterprise/generated/api"
	"github.com/rastrigin-systems/ubik-enterprise/generated/db"
)

// TeamsHandler handles team management requests
type TeamsHandler struct {
	db db.Querier
}

// NewTeamsHandler creates a new teams handler
func NewTeamsHandler(database db.Querier) *TeamsHandler {
	return &TeamsHandler{
		db: database,
	}
}

// ListTeams handles GET /teams
// Returns list of teams for the current organization
func (h *TeamsHandler) ListTeams(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Query database for all teams in org
	teams, err := h.db.ListTeams(ctx, orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch teams")
		return
	}

	// Convert to API response with aggregate counts
	apiTeams := make([]api.Team, len(teams))
	for i, team := range teams {
		apiTeam := dbTeamToAPI(team)

		// Get member count for this team
		memberCount, err := h.db.CountEmployeesByTeam(ctx, pgtype.UUID{Bytes: team.ID, Valid: true})
		if err == nil {
			count := int(memberCount)
			apiTeam.MemberCount = &count
		}

		// Get agent config count for this team
		agentCount, err := h.db.CountTeamAgentConfigs(ctx, team.ID)
		if err == nil {
			count := int(agentCount)
			apiTeam.AgentConfigCount = &count
		}

		apiTeams[i] = apiTeam
	}

	// Build response
	response := api.ListTeamsResponse{
		Teams: apiTeams,
		Total: len(apiTeams),
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateTeam handles POST /teams
// Creates a new team in the organization (admin only)
func (h *TeamsHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var req api.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Validate required fields
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Create team
	team, err := h.db.CreateTeam(ctx, db.CreateTeamParams{
		OrgID:       orgID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create team")
		return
	}

	// Build response
	response := dbTeamToAPI(team)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetTeam handles GET /teams/{team_id}
// Returns a specific team by ID
func (h *TeamsHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse team ID from URL
	teamIDStr := chi.URLParam(r, "team_id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	// Query database
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

	// Convert to API response
	response := dbTeamToAPI(team)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateTeam handles PATCH /teams/{team_id}
// Updates an existing team (admin only)
// SECURITY: Verifies team belongs to authenticated user's organization
func (h *TeamsHandler) UpdateTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get org_id from context (set by JWT middleware) - SECURITY FIX
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse team ID from URL
	teamIDStr := chi.URLParam(r, "team_id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	// Parse request body
	var req api.UpdateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid JSON")
		return
	}

	// Build update params - use empty string if not provided (COALESCE in SQL will keep existing)
	name := ""
	if req.Name != nil {
		name = *req.Name
	}

	// Update team - SECURITY: Query includes org_id for isolation
	team, err := h.db.UpdateTeam(ctx, db.UpdateTeamParams{
		ID:          teamID,
		Name:        name,
		Description: req.Description,
		OrgID:       orgID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			// Return 404 for security - don't reveal if team exists in another org
			writeError(w, http.StatusNotFound, "Team not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to update team")
		return
	}

	// Build response
	response := dbTeamToAPI(team)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteTeam handles DELETE /teams/{team_id}
// Deletes a team (admin only)
// SECURITY: Verifies team belongs to authenticated user's organization
func (h *TeamsHandler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get org_id from context (set by JWT middleware) - SECURITY FIX
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse team ID from URL
	teamIDStr := chi.URLParam(r, "team_id")
	teamID, err := uuid.Parse(teamIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	// Delete team - SECURITY: Query includes org_id for isolation
	err = h.db.DeleteTeam(ctx, db.DeleteTeamParams{
		ID:    teamID,
		OrgID: orgID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete team")
		return
	}

	// Write 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// Helper functions

func dbTeamToAPI(team db.Team) api.Team {
	id := openapi_types.UUID(team.ID)
	orgID := openapi_types.UUID(team.OrgID)
	createdAt := team.CreatedAt.Time
	updatedAt := team.UpdatedAt.Time

	result := api.Team{
		Id:          &id,
		OrgId:       orgID,
		Name:        team.Name,
		Description: team.Description,
		CreatedAt:   &createdAt,
		UpdatedAt:   &updatedAt,
	}

	return result
}
