package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
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

// ListToolPolicies handles GET /policies
// Returns all tool policies for the organization with optional filters
func (h *ToolPoliciesHandler) ListToolPolicies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, err := middleware.GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// TODO: Check role (admin/manager)

	// Parse optional query filters
	var teamID, employeeID pgtype.UUID
	var scope *string

	if tid := r.URL.Query().Get("team_id"); tid != "" {
		if parsed, err := uuid.Parse(tid); err == nil {
			teamID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}
	if eid := r.URL.Query().Get("employee_id"); eid != "" {
		if parsed, err := uuid.Parse(eid); err == nil {
			employeeID = pgtype.UUID{Bytes: parsed, Valid: true}
		}
	}
	if s := r.URL.Query().Get("scope"); s != "" {
		scope = &s
	}

	// Query with filters
	params := db.ListToolPoliciesFilteredParams{
		OrgID:      orgID,
		TeamID:     teamID,
		EmployeeID: employeeID,
		Scope:      scope,
	}

	policies, err := h.db.ListToolPoliciesFiltered(ctx, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list policies")
		return
	}

	// Convert to API response
	apiPolicies := make([]api.ToolPolicy, len(policies))
	for i, policy := range policies {
		apiPolicies[i] = dbToolPolicyToAPI(policy)
	}

	response := api.ListToolPoliciesResponse{
		Policies: apiPolicies,
		Total:    len(apiPolicies),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// CreateToolPolicy handles POST /policies
func (h *ToolPoliciesHandler) CreateToolPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, err := middleware.GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	employeeID, err := middleware.GetEmployeeID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// TODO: Check role (admin for org-level, manager for team/employee-level)

	var req api.CreateToolPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.ToolName == "" {
		writeError(w, http.StatusBadRequest, "tool_name is required")
		return
	}
	if req.Action == "" {
		writeError(w, http.StatusBadRequest, "action is required")
		return
	}

	// Build create params
	params := db.CreateToolPolicyParams{
		OrgID:     orgID,
		ToolName:  req.ToolName,
		Action:    string(req.Action),
		Reason:    req.Reason,
		CreatedBy: pgtype.UUID{Bytes: employeeID, Valid: true},
	}

	// Set optional team_id
	if req.TeamId != nil {
		params.TeamID = pgtype.UUID{Bytes: uuid.UUID(*req.TeamId), Valid: true}
	}

	// Set optional employee_id
	if req.EmployeeId != nil {
		params.EmployeeID = pgtype.UUID{Bytes: uuid.UUID(*req.EmployeeId), Valid: true}
	}

	// Convert conditions to JSON
	if req.Conditions != nil {
		conditionsJSON, err := json.Marshal(req.Conditions)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid conditions format")
			return
		}
		params.Conditions = conditionsJSON
	}

	policy, err := h.db.CreateToolPolicy(ctx, params)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create policy")
		return
	}

	response := dbToolPolicyToAPI(policy)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(response)
}

// GetToolPolicy handles GET /policies/{policy_id}
func (h *ToolPoliciesHandler) GetToolPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, err := middleware.GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	policyID, err := uuid.Parse(chi.URLParam(r, "policy_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid policy ID")
		return
	}

	policy, err := h.db.GetToolPolicyByIdAndOrg(ctx, db.GetToolPolicyByIdAndOrgParams{
		ID:    policyID,
		OrgID: orgID,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, "Policy not found")
		return
	}

	response := dbToolPolicyToAPI(policy)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// UpdateToolPolicy handles PATCH /policies/{policy_id}
func (h *ToolPoliciesHandler) UpdateToolPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, err := middleware.GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	policyID, err := uuid.Parse(chi.URLParam(r, "policy_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid policy ID")
		return
	}

	// TODO: Check role (admin for org-level, manager for team/employee-level)

	var req api.UpdateToolPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Build update params
	params := db.UpdateToolPolicyByOrgParams{
		ID:    policyID,
		OrgID: orgID,
	}

	if req.ToolName != nil {
		params.ToolName = req.ToolName
	}
	if req.Action != nil {
		action := string(*req.Action)
		params.Action = &action
	}
	if req.Reason != nil {
		params.Reason = req.Reason
	}
	if req.Conditions != nil {
		conditionsJSON, err := json.Marshal(req.Conditions)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid conditions format")
			return
		}
		params.Conditions = conditionsJSON
	}

	policy, err := h.db.UpdateToolPolicyByOrg(ctx, params)
	if err != nil {
		writeError(w, http.StatusNotFound, "Policy not found")
		return
	}

	response := dbToolPolicyToAPI(policy)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// DeleteToolPolicy handles DELETE /policies/{policy_id}
func (h *ToolPoliciesHandler) DeleteToolPolicy(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, err := middleware.GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	policyID, err := uuid.Parse(chi.URLParam(r, "policy_id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid policy ID")
		return
	}

	// TODO: Check role (admin for org-level, manager for team/employee-level)

	err = h.db.DeleteToolPolicyByOrg(ctx, db.DeleteToolPolicyByOrgParams{
		ID:    policyID,
		OrgID: orgID,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, "Policy not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// dbToolPolicyToAPI converts a database ToolPolicy to an API ToolPolicy
func dbToolPolicyToAPI(policy db.ToolPolicy) api.ToolPolicy {
	policyID := openapi_types.UUID(policy.ID)
	orgID := openapi_types.UUID(policy.OrgID)

	// CreatedAt is a pointer in API
	var createdAt *time.Time
	if policy.CreatedAt.Valid {
		createdAt = &policy.CreatedAt.Time
	}

	apiPolicy := api.ToolPolicy{
		Id:        &policyID,
		OrgId:     orgID,
		ToolName:  policy.ToolName,
		Action:    api.ToolPolicyAction(policy.Action),
		Reason:    policy.Reason,
		CreatedAt: createdAt,
	}

	// Set optional IDs
	if policy.TeamID.Valid {
		teamID := openapi_types.UUID(policy.TeamID.Bytes)
		apiPolicy.TeamId = &teamID
	}
	if policy.EmployeeID.Valid {
		employeeID := openapi_types.UUID(policy.EmployeeID.Bytes)
		apiPolicy.EmployeeId = &employeeID
	}
	if policy.UpdatedAt.Valid {
		apiPolicy.UpdatedAt = &policy.UpdatedAt.Time
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
	apiPolicy.Scope = scope

	return apiPolicy
}
