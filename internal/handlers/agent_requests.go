package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
)

type AgentRequestsHandler struct {
	queries *db.Queries
}

func NewAgentRequestsHandler(queries *db.Queries) *AgentRequestsHandler {
	return &AgentRequestsHandler{queries: queries}
}

// GetPendingCount gets the count of pending agent requests for the organization
func (h *AgentRequestsHandler) GetPendingCount(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	count, err := h.queries.CountPendingRequestsByOrg(ctx, orgID)
	if err != nil {
		http.Error(w, "Failed to get pending count", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"pending_count": count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListAgentRequests lists agent requests with optional filtering
func (h *AgentRequestsHandler) ListAgentRequests(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse query parameters
	status := r.URL.Query().Get("status")
	employeeIDStr := r.URL.Query().Get("employee_id")

	var employeeID sql.NullString
	if employeeIDStr != "" {
		employeeID = sql.NullString{String: employeeIDStr, Valid: true}
	}

	var statusFilter sql.NullString
	if status != "" {
		statusFilter = sql.NullString{String: status, Valid: true}
	}

	// TODO: Add pagination
	limit := int32(100)
	offset := int32(0)

	requests, err := h.queries.ListAgentRequests(ctx, db.ListAgentRequestsParams{
		Column1: statusFilter,
		Column2: employeeID,
		Limit:   limit,
		Offset:  offset,
	})
	if err != nil {
		http.Error(w, "Failed to list requests", http.StatusInternalServerError)
		return
	}

	count, err := h.queries.CountAgentRequests(ctx, db.CountAgentRequestsParams{
		Column1: statusFilter,
		Column2: employeeID,
	})
	if err != nil {
		count = int64(len(requests))
	}

	response := map[string]interface{}{
		"requests": requests,
		"total":    count,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
