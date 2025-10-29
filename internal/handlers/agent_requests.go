package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
		writeError(w, http.StatusInternalServerError, "Failed to get pending count")
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

	var employeeID pgtype.UUID
	if employeeIDStr != "" {
		empUUID, err := uuid.Parse(employeeIDStr)
		if err == nil {
			employeeID = pgtype.UUID{Bytes: empUUID, Valid: true}
		}
	}

	var statusFilter *string
	if status != "" {
		statusFilter = &status
	}

	// TODO: Add pagination
	limit := int32(100)
	offset := int32(0)

	requests, err := h.queries.ListAgentRequests(ctx, db.ListAgentRequestsParams{
		Status:      statusFilter,
		EmployeeID:  employeeID,
		QueryOffset: offset,
		QueryLimit:  limit,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list requests")
		return
	}

	count, err := h.queries.CountAgentRequests(ctx, db.CountAgentRequestsParams{
		Status:     statusFilter,
		EmployeeID: employeeID,
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
