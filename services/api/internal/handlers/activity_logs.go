package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/services/api/internal/middleware"
)

// ActivityLogsHandler handles activity log requests
type ActivityLogsHandler struct {
	db db.Querier
}

// NewActivityLogsHandler creates a new activity logs handler
func NewActivityLogsHandler(database db.Querier) *ActivityLogsHandler {
	return &ActivityLogsHandler{
		db: database,
	}
}

// ActivityLogResponse represents a single activity log entry
type ActivityLogResponse struct {
	ID            string                 `json:"id"`
	OrgID         string                 `json:"org_id"`
	EmployeeID    *string                `json:"employee_id,omitempty"`
	EmployeeName  *string                `json:"employee_name,omitempty"`
	EventType     string                 `json:"event_type"`
	EventCategory string                 `json:"event_category"`
	Message       string                 `json:"message"`
	Time          string                 `json:"time"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	CreatedAt     time.Time              `json:"created_at"`
}

// ActivityLogsListResponse represents a list of activity logs with pagination
type ActivityLogsListResponse struct {
	ActivityLogs []ActivityLogResponse `json:"activity_logs"`
	Total        int64                 `json:"total"`
	Limit        int32                 `json:"limit"`
	Offset       int32                 `json:"offset"`
}

// ListActivityLogs handles GET /activity-logs
// Returns recent activity logs for the organization
func (h *ActivityLogsHandler) ListActivityLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get org ID from context (set by auth middleware)
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := int32(10) // default
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = int32(l)
		}
	}

	offset := int32(0) // default
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = int32(o)
		}
	}

	// Query activity logs
	logs, err := h.db.ListActivityLogs(ctx, db.ListActivityLogsParams{
		OrgID:  orgID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch activity logs")
		return
	}

	// Get total count
	total, err := h.db.CountActivityLogs(ctx, orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to count activity logs")
		return
	}

	// Convert to response format with employee names
	logResponses := make([]ActivityLogResponse, 0, len(logs))
	for _, log := range logs {
		logResponse := dbActivityLogToResponse(log)

		// Fetch employee name if employee_id is set
		if log.EmployeeID.Valid {
			employeeUUID, err := uuid.FromBytes(log.EmployeeID.Bytes[:])
			if err == nil {
				employee, err := h.db.GetEmployee(ctx, employeeUUID)
				if err == nil {
					logResponse.EmployeeName = &employee.FullName
				}
			}
		}

		logResponses = append(logResponses, logResponse)
	}

	// Write JSON response
	response := ActivityLogsListResponse{
		ActivityLogs: logResponses,
		Total:        total,
		Limit:        limit,
		Offset:       offset,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// dbActivityLogToResponse converts a database activity log to a response format
func dbActivityLogToResponse(log db.ActivityLog) ActivityLogResponse {
	// Parse payload JSON
	var payload map[string]interface{}
	if len(log.Payload) > 0 {
		json.Unmarshal(log.Payload, &payload)
	}

	// Convert employee ID
	var employeeID *string
	if log.EmployeeID.Valid {
		empUUID, err := uuid.FromBytes(log.EmployeeID.Bytes[:])
		if err == nil {
			id := empUUID.String()
			employeeID = &id
		}
	}

	// Generate user-friendly message
	message := generateActivityMessage(log.EventType, log.EventCategory, payload)

	// Format time
	timeStr := formatActivityTime(log.CreatedAt.Time)

	return ActivityLogResponse{
		ID:            log.ID.String(),
		OrgID:         log.OrgID.String(),
		EmployeeID:    employeeID,
		EventType:     log.EventType,
		EventCategory: log.EventCategory,
		Message:       message,
		Time:          timeStr,
		Payload:       payload,
		CreatedAt:     log.CreatedAt.Time,
	}
}

// generateActivityMessage creates a human-readable message from event data
func generateActivityMessage(eventType, eventCategory string, payload map[string]interface{}) string {
	// Simple message generation - can be enhanced
	switch eventCategory {
	case "agent":
		switch eventType {
		case "agent.installed":
			return "Installed new agent configuration"
		case "agent.updated":
			return "Updated agent configuration"
		case "agent.deleted":
			return "Removed agent configuration"
		}
	case "mcp":
		switch eventType {
		case "mcp.configured":
			return "Configured MCP server"
		case "mcp.updated":
			return "Updated MCP configuration"
		}
	case "auth":
		switch eventType {
		case "auth.login":
			return "Logged in"
		case "auth.logout":
			return "Logged out"
		}
	case "admin":
		switch eventType {
		case "employee.created":
			return "Created new employee"
		case "employee.updated":
			return "Updated employee"
		case "team.created":
			return "Created new team"
		}
	}

	// Default fallback
	return eventType
}

// formatActivityTime formats a timestamp for display
func formatActivityTime(t time.Time) string {
	now := time.Now()
	diff := now.Sub(t)

	switch {
	case diff < time.Minute:
		return "Just now"
	case diff < time.Hour:
		mins := int(diff.Minutes())
		if mins == 1 {
			return "1 minute ago"
		}
		return strconv.Itoa(mins) + " minutes ago"
	case diff < 24*time.Hour:
		hours := int(diff.Hours())
		if hours == 1 {
			return "1 hour ago"
		}
		return strconv.Itoa(hours) + " hours ago"
	case diff < 48*time.Hour:
		return "Yesterday"
	default:
		return t.Format("Jan 2")
	}
}

// CreateActivityLog is a helper function to create activity log entries
// This can be called from other handlers to log activities
func CreateActivityLog(ctx *http.Request, database db.Querier, eventType, eventCategory string, payload map[string]interface{}) error {
	orgID, err := GetOrgID(ctx.Context())
	if err != nil {
		return err
	}

	employeeID, _ := GetEmployeeID(ctx.Context())

	payloadJSON, _ := json.Marshal(payload)

	var empID pgtype.UUID
	if employeeID != uuid.Nil {
		empID = pgtype.UUID{
			Bytes: [16]byte(employeeID[:]),
			Valid: true,
		}
	}

	_, err = database.CreateActivityLog(ctx.Context(), db.CreateActivityLogParams{
		OrgID:         orgID,
		EmployeeID:    empID,
		EventType:     eventType,
		EventCategory: eventCategory,
		Payload:       payloadJSON,
	})

	return err
}

// GetEmployeeID wraps middleware.GetEmployeeID for convenience
func GetEmployeeID(ctx context.Context) (uuid.UUID, error) {
	return middleware.GetEmployeeID(ctx)
}
