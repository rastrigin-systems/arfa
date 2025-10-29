package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
)

type UsageStatsHandler struct {
	queries *db.Queries
}

func NewUsageStatsHandler(queries *db.Queries) *UsageStatsHandler {
	return &UsageStatsHandler{queries: queries}
}

// GetEmployeeUsageStats gets usage statistics for a specific employee
func (h *UsageStatsHandler) GetEmployeeUsageStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	employeeIDStr := chi.URLParam(r, "employee_id")
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		http.Error(w, "Invalid employee ID", http.StatusBadRequest)
		return
	}

	// Verify employee belongs to org
	employee, err := h.queries.GetEmployee(ctx, db.GetEmployeeParams{
		ID:    employeeID,
		OrgID: orgID,
	})
	if err != nil {
		http.Error(w, "Employee not found", http.StatusNotFound)
		return
	}

	// Get stats for last 30 days
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)

	stats, err := h.queries.GetEmployeeUsageStats(ctx, db.GetEmployeeUsageStatsParams{
		EmployeeID: sql.NullString{String: employee.ID.String(), Valid: true},
		Column2:    startTime,
		Column3:    endTime,
	})
	if err != nil {
		http.Error(w, "Failed to get usage stats", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"employee_id":     employee.ID,
		"employee_name":   employee.FullName,
		"period_start":    startTime,
		"period_end":      endTime,
		"total_api_calls": stats.TotalApiCalls,
		"total_tokens":    stats.TotalTokens,
		"total_cost_usd":  stats.TotalCostUsd,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetOrgUsageStats gets overall usage statistics for the organization
func (h *UsageStatsHandler) GetOrgUsageStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get stats for last 30 days
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)

	stats, err := h.queries.GetOrgUsageStats(ctx, db.GetOrgUsageStatsParams{
		OrgID:   orgID,
		Column2: startTime,
		Column3: endTime,
	})
	if err != nil {
		http.Error(w, "Failed to get usage stats", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"org_id":          orgID,
		"period_start":    startTime,
		"period_end":      endTime,
		"total_api_calls": stats.TotalApiCalls,
		"total_tokens":    stats.TotalTokens,
		"total_cost_usd":  stats.TotalCostUsd,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetCurrentEmployeeUsageStats gets usage stats for the authenticated employee
func (h *UsageStatsHandler) GetCurrentEmployeeUsageStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	employeeID, err := GetEmployeeID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get stats for last 30 days
	endTime := time.Now()
	startTime := endTime.AddDate(0, 0, -30)

	stats, err := h.queries.GetEmployeeUsageStats(ctx, db.GetEmployeeUsageStatsParams{
		EmployeeID: sql.NullString{String: employeeID.String(), Valid: true},
		Column2:    startTime,
		Column3:    endTime,
	})
	if err != nil {
		http.Error(w, "Failed to get usage stats", http.StatusInternalServerError)
		return
	}

	response := map[string]interface{}{
		"employee_id":     employeeID,
		"period_start":    startTime,
		"period_end":      endTime,
		"total_api_calls": stats.TotalApiCalls,
		"total_tokens":    stats.TotalTokens,
		"total_cost_usd":  stats.TotalCostUsd,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
