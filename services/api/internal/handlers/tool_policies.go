package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rastrigin-systems/arfa/generated/api"
	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/services/api/internal/middleware"
)

// ToolPoliciesHandler handles tool policy-related requests
type ToolPoliciesHandler struct {
	db db.Querier
}

// NewToolPoliciesHandler creates a new tool policies handler
func NewToolPoliciesHandler(database db.Querier) *ToolPoliciesHandler {
	return &ToolPoliciesHandler{
		db: database,
	}
}

// GetEmployeeToolPolicies handles GET /employees/me/tool-policies
// Returns all tool policies that apply to the authenticated employee
func (h *ToolPoliciesHandler) GetEmployeeToolPolicies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get employee ID and org ID from JWT context
	employeeID, err := middleware.GetEmployeeID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	orgID, err := middleware.GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get session data for team_id
	sessionData, err := middleware.GetSessionData(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get session data")
		return
	}

	// Build query params
	params := db.GetToolPoliciesForEmployeeParams{
		OrgID:      orgID,
		TeamID:     sessionData.TeamID, // May be null
		EmployeeID: pgtype.UUID{Bytes: employeeID, Valid: true},
	}

	// Query database for policies
	policies, err := h.db.GetToolPoliciesForEmployee(ctx, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch tool policies")
		return
	}

	// Convert to API response
	apiPolicies := make([]api.ToolPolicy, len(policies))
	for i, policy := range policies {
		apiPolicies[i] = dbToolPolicyToAPI(policy)
	}

	// Build response with version and timestamp
	response := api.EmployeeToolPoliciesResponse{
		Policies: apiPolicies,
		Version:  int(time.Now().Unix()), // Simple version based on current time
		SyncedAt: time.Now(),
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// dbToolPolicyToAPI converts a database ToolPolicy to an API ToolPolicy
func dbToolPolicyToAPI(policy db.ToolPolicy) api.ToolPolicy {
	policyID := openapi_types.UUID(policy.ID)
	apiPolicy := api.ToolPolicy{
		Id:       &policyID,
		ToolName: policy.ToolName,
		Action:   api.ToolPolicyAction(policy.Action),
		Reason:   policy.Reason,
	}

	// Parse conditions JSON if present
	if len(policy.Conditions) > 0 {
		var conditions map[string]interface{}
		if err := json.Unmarshal(policy.Conditions, &conditions); err == nil {
			apiPolicy.Conditions = &conditions
		}
	}

	// Determine scope based on which IDs are set
	var scope api.ToolPolicyScope
	if policy.EmployeeID.Valid {
		scope = api.ToolPolicyScopeEmployee
	} else if policy.TeamID.Valid {
		scope = api.ToolPolicyScopeTeam
	} else {
		scope = api.ToolPolicyScopeOrganization
	}
	apiPolicy.Scope = &scope

	return apiPolicy
}
