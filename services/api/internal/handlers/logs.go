package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/service"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/websocket"
)

// LogsHandler handles logging API requests
type LogsHandler struct {
	db             db.Querier
	loggingService *service.LoggingService
	wsHub          *websocket.Hub
}

// NewLogsHandler creates a new logs handler
func NewLogsHandler(database db.Querier, wsHub *websocket.Hub) *LogsHandler {
	return &LogsHandler{
		db:             database,
		loggingService: service.NewLoggingService(database),
		wsHub:          wsHub,
	}
}

// CreateLog implements POST /logs
func (h *LogsHandler) CreateLog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get org ID from context (set by auth middleware)
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Get employee ID from context (optional, can be nil)
	employeeID, _ := GetEmployeeID(ctx)

	// Parse request body
	var req api.CreateLogRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.EventType == "" {
		writeError(w, http.StatusBadRequest, "event_type is required")
		return
	}
	if req.EventCategory == "" {
		writeError(w, http.StatusBadRequest, "event_category is required")
		return
	}

	// Build log entry
	entry := service.LogEntry{
		OrgID:         orgID,
		EmployeeID:    employeeID,
		EventType:     string(req.EventType),
		EventCategory: string(req.EventCategory),
	}

	// Add optional fields
	if req.SessionId != nil {
		entry.SessionID = *req.SessionId
	}
	if req.AgentId != nil {
		entry.AgentID = *req.AgentId
	}
	if req.Content != nil {
		entry.Content = *req.Content
	}
	if req.Payload != nil {
		entry.Payload = *req.Payload
	}

	// Create log using service layer
	err = h.loggingService.CreateLog(ctx, entry)
	if err != nil {
		writeError(w, http.StatusInternalServerError, fmt.Sprintf("Failed to create log: %v", err))
		return
	}

	// Fetch the created log to return it
	// For now, we'll construct a response from the input
	// In a real implementation, the service would return the created log
	newID := uuid.New()

	// Broadcast log to WebSocket clients (if hub is configured)
	if h.wsHub != nil {
		wsMsg := websocket.LogMessage{
			ID:            newID,
			OrgID:         orgID,
			EmployeeID:    employeeID,
			EventType:     string(req.EventType),
			EventCategory: string(req.EventCategory),
			Timestamp:     time.Now(),
		}

		if req.SessionId != nil {
			wsMsg.SessionID = *req.SessionId
		}
		if req.AgentId != nil {
			wsMsg.AgentID = *req.AgentId
		}
		if req.Content != nil {
			wsMsg.Content = *req.Content
		}
		if req.Payload != nil {
			wsMsg.Payload = *req.Payload
		}

		h.wsHub.Broadcast(wsMsg)
	}
	response := api.ActivityLog{
		Id:            openapi_types.UUID(newID),
		OrgId:         openapi_types.UUID(orgID),
		EventType:     string(req.EventType),
		EventCategory: string(req.EventCategory),
		Payload:       map[string]interface{}{},
		CreatedAt:     time.Now(),
	}

	if employeeID != uuid.Nil {
		empAPIID := openapi_types.UUID(employeeID)
		response.EmployeeId = &empAPIID
	}
	if req.SessionId != nil {
		response.SessionId = req.SessionId
	}
	if req.AgentId != nil {
		response.AgentId = req.AgentId
	}
	if req.Content != nil {
		response.Content = req.Content
	}
	if req.Payload != nil {
		response.Payload = *req.Payload
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// ListLogs implements GET /logs
func (h *LogsHandler) ListLogs(w http.ResponseWriter, r *http.Request, params api.ListLogsParams) {
	ctx := r.Context()

	// Get org ID from context
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Handle session filter if provided
	if params.SessionId != nil {
		logs, err := h.loggingService.GetLogsBySession(ctx, *params.SessionId)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to fetch logs")
			return
		}

		// Convert to API response
		apiLogs := make([]api.ActivityLog, 0, len(logs))
		for _, log := range logs {
			apiLogs = append(apiLogs, dbLogToAPI(log))
		}

		response := api.ListLogsResponse{
			Logs: apiLogs,
			Pagination: api.PaginationMeta{
				Total:      len(apiLogs),
				Page:       1,
				PerPage:    100,
				TotalPages: 1,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Default list with pagination
	limit := int32(20)
	offset := int32(0)

	if params.PerPage != nil {
		limit = int32(*params.PerPage)
	}
	if params.Page != nil {
		offset = (int32(*params.Page) - 1) * limit
	}

	logs, err := h.db.ListActivityLogs(ctx, db.ListActivityLogsParams{
		OrgID:  orgID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch logs")
		return
	}

	// Get total count
	total, err := h.db.CountActivityLogs(ctx, orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to count logs")
		return
	}

	// Convert to API response
	apiLogs := make([]api.ActivityLog, 0, len(logs))
	for _, log := range logs {
		apiLogs = append(apiLogs, dbLogToAPI(log))
	}

	totalPages := int(total) / int(limit)
	if int(total)%int(limit) > 0 {
		totalPages++
	}

	page := 1
	if params.Page != nil {
		page = *params.Page
	}

	response := api.ListLogsResponse{
		Logs: apiLogs,
		Pagination: api.PaginationMeta{
			Total:      int(total),
			Page:       page,
			PerPage:    int(limit),
			TotalPages: totalPages,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ExportLogs implements GET /logs/export
func (h *LogsHandler) ExportLogs(w http.ResponseWriter, r *http.Request, params api.ExportLogsParams) {
	// TODO: Implement export functionality
	writeError(w, http.StatusNotImplemented, "Export not yet implemented")
}

// ListSessions implements GET /logs/sessions
func (h *LogsHandler) ListSessions(w http.ResponseWriter, r *http.Request, params api.ListSessionsParams) {
	// TODO: Implement sessions listing
	writeError(w, http.StatusNotImplemented, "Sessions listing not yet implemented")
}

// dbLogToAPI converts a database activity log to API format
func dbLogToAPI(log db.ActivityLog) api.ActivityLog {
	apiLog := api.ActivityLog{
		Id:            openapi_types.UUID(log.ID),
		OrgId:         openapi_types.UUID(log.OrgID),
		EventType:     log.EventType,
		EventCategory: log.EventCategory,
		Payload:       map[string]interface{}{},
		CreatedAt:     log.CreatedAt.Time,
	}

	if log.EmployeeID.Valid {
		empID := uuid.UUID(log.EmployeeID.Bytes)
		empAPIID := openapi_types.UUID(empID)
		apiLog.EmployeeId = &empAPIID
	}

	if log.SessionID.Valid {
		sessID := uuid.UUID(log.SessionID.Bytes)
		sessAPIID := openapi_types.UUID(sessID)
		apiLog.SessionId = &sessAPIID
	}

	if log.AgentID.Valid {
		agentID := uuid.UUID(log.AgentID.Bytes)
		agentAPIID := openapi_types.UUID(agentID)
		apiLog.AgentId = &agentAPIID
	}

	if log.Content != nil {
		apiLog.Content = log.Content
	}

	// Parse payload JSON
	if len(log.Payload) > 0 {
		var payload map[string]interface{}
		if err := json.Unmarshal(log.Payload, &payload); err == nil {
			apiLog.Payload = payload
		}
	}

	return apiLog
}

// apiLogToDB converts an API log request to service layer format
func apiLogToDB(req api.CreateLogRequest, orgID, employeeID uuid.UUID) db.CreateActivityLogParams {
	params := db.CreateActivityLogParams{
		OrgID:         orgID,
		EventType:     string(req.EventType),
		EventCategory: string(req.EventCategory),
		Payload:       []byte("{}"),
	}

	if employeeID != uuid.Nil {
		params.EmployeeID = pgtype.UUID{Bytes: employeeID, Valid: true}
	}

	if req.SessionId != nil {
		params.SessionID = pgtype.UUID{Bytes: *req.SessionId, Valid: true}
	}

	if req.AgentId != nil {
		params.AgentID = pgtype.UUID{Bytes: *req.AgentId, Valid: true}
	}

	if req.Content != nil {
		params.Content = req.Content
	}

	if req.Payload != nil {
		payloadJSON, _ := json.Marshal(req.Payload)
		params.Payload = payloadJSON
	}

	return params
}
