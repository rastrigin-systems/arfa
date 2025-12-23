package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/rastrigin-systems/ubik-enterprise/generated/api"
	"github.com/rastrigin-systems/ubik-enterprise/generated/db"
)

// ClaudeTokensHandler handles Claude API token management requests
type ClaudeTokensHandler struct {
	db db.Querier
}

// NewClaudeTokensHandler creates a new Claude tokens handler
func NewClaudeTokensHandler(database db.Querier) *ClaudeTokensHandler {
	return &ClaudeTokensHandler{
		db: database,
	}
}

// SetOrganizationClaudeToken handles PUT /organizations/current/claude-token
// Sets the company-wide Claude API token
func (h *ClaudeTokensHandler) SetOrganizationClaudeToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get org ID from context (set by auth middleware)
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var req api.SetClaudeTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate token
	if req.Token == "" || len(req.Token) < 10 {
		writeError(w, http.StatusBadRequest, "Token must be at least 10 characters")
		return
	}

	// Update database
	err = h.db.SetOrganizationClaudeToken(ctx, db.SetOrganizationClaudeTokenParams{
		ID:             orgID,
		ClaudeApiToken: &req.Token,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update token")
		return
	}

	// Write success response
	response := api.ClaudeTokenResponse{
		Success: true,
		Message: "Organization Claude API token updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteOrganizationClaudeToken handles DELETE /organizations/current/claude-token
// Removes the company-wide Claude API token
func (h *ClaudeTokensHandler) DeleteOrganizationClaudeToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get org ID from context
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Delete from database
	err = h.db.DeleteOrganizationClaudeToken(ctx, orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete token")
		return
	}

	// Write success response
	response := api.ClaudeTokenResponse{
		Success: true,
		Message: "Organization Claude API token deleted successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SetEmployeeClaudeToken handles PUT /employees/me/claude-token
// Sets the employee's personal Claude API token
func (h *ClaudeTokensHandler) SetEmployeeClaudeToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get employee ID from context
	employeeID, err := GetEmployeeID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var req api.SetClaudeTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate token
	if req.Token == "" || len(req.Token) < 10 {
		writeError(w, http.StatusBadRequest, "Token must be at least 10 characters")
		return
	}

	// Update database
	err = h.db.SetEmployeePersonalToken(ctx, db.SetEmployeePersonalTokenParams{
		ID:                  employeeID,
		PersonalClaudeToken: &req.Token,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update token")
		return
	}

	// Write success response
	response := api.ClaudeTokenResponse{
		Success: true,
		Message: "Personal Claude API token updated successfully",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteEmployeeClaudeToken handles DELETE /employees/me/claude-token
// Removes the employee's personal Claude API token
func (h *ClaudeTokensHandler) DeleteEmployeeClaudeToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get employee ID from context
	employeeID, err := GetEmployeeID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Delete from database
	err = h.db.DeleteEmployeePersonalToken(ctx, employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete token")
		return
	}

	// Write success response
	response := api.ClaudeTokenResponse{
		Success: true,
		Message: "Personal Claude API token deleted successfully (will fall back to company token)",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetClaudeTokenStatus handles GET /employees/me/claude-token/status
// Returns which Claude token is active (personal, company, or none)
func (h *ClaudeTokensHandler) GetClaudeTokenStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get employee ID from context
	employeeID, err := GetEmployeeID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Query database
	status, err := h.db.GetEmployeeTokenStatus(ctx, employeeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to get token status")
		return
	}

	// Map source string to API enum
	var activeSource api.ClaudeTokenStatusResponseActiveTokenSource
	switch status.ActiveTokenSource {
	case "personal":
		activeSource = api.ClaudeTokenStatusResponseActiveTokenSourcePersonal
	case "company":
		activeSource = api.ClaudeTokenStatusResponseActiveTokenSourceCompany
	case "none":
		activeSource = api.ClaudeTokenStatusResponseActiveTokenSourceNone
	default:
		activeSource = api.ClaudeTokenStatusResponseActiveTokenSourceNone
	}

	// Convert interface{} booleans from sqlc
	hasPersonal, _ := status.HasPersonalToken.(bool)
	hasCompany, _ := status.HasCompanyToken.(bool)

	// Build response
	response := api.ClaudeTokenStatusResponse{
		EmployeeId:        &status.EmployeeID,
		HasPersonalToken:  hasPersonal,
		HasCompanyToken:   hasCompany,
		ActiveTokenSource: activeSource,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetEffectiveClaudeToken handles GET /employees/me/claude-token/effective
// Returns the effective Claude API token (personal or company)
func (h *ClaudeTokensHandler) GetEffectiveClaudeToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get employee ID from context
	employeeID, err := GetEmployeeID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Query database for effective token
	tokenInfo, err := h.db.GetEffectiveClaudeToken(ctx, employeeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Employee not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to get effective token")
		return
	}

	// Check if token exists
	if tokenInfo.Token == nil || *tokenInfo.Token == "" {
		writeError(w, http.StatusNotFound, "No Claude API token configured. Please configure a token at organization or personal level")
		return
	}

	// Map source string to API enum
	var source api.EffectiveClaudeTokenResponseSource
	switch tokenInfo.Source {
	case "personal":
		source = api.Personal
	case "company":
		source = api.Company
	default:
		writeError(w, http.StatusNotFound, "No token available")
		return
	}

	// Build response
	response := api.EffectiveClaudeTokenResponse{
		Token:      *tokenInfo.Token,
		Source:     source,
		OrgId:      tokenInfo.OrgID,
		OrgName:    tokenInfo.OrgName,
		EmployeeId: &tokenInfo.EmployeeID,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
