package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rastrigin-systems/ubik-enterprise/generated/api"
	"github.com/rastrigin-systems/ubik-enterprise/generated/db"
	"github.com/rastrigin-systems/ubik-enterprise/services/api/internal/middleware"
)

// WebhooksHandler handles webhook destination requests
type WebhooksHandler struct {
	db db.Querier
}

// NewWebhooksHandler creates a new webhooks handler
func NewWebhooksHandler(database db.Querier) *WebhooksHandler {
	return &WebhooksHandler{
		db: database,
	}
}

// ListWebhookDestinations handles GET /webhooks
func (h *WebhooksHandler) ListWebhookDestinations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, err := middleware.GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	destinations, err := h.db.ListWebhookDestinations(ctx, orgID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list webhook destinations")
		return
	}

	apiDestinations := make([]api.WebhookDestination, len(destinations))
	for i, dest := range destinations {
		apiDestinations[i] = dbWebhookDestinationRowToAPI(dest)
	}

	response := api.ListWebhookDestinationsResponse{
		Destinations: apiDestinations,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateWebhookDestination handles POST /webhooks
func (h *WebhooksHandler) CreateWebhookDestination(w http.ResponseWriter, r *http.Request) {
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

	// TODO: Check role (admin/manager) - for now allow all authenticated users

	var req api.CreateWebhookDestinationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" || req.Url == "" {
		writeError(w, http.StatusBadRequest, "Name and URL are required")
		return
	}

	// Generate signing secret
	signingSecret, err := generateSigningSecret()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate signing secret")
		return
	}

	// Convert auth_config to JSON
	var authConfigJSON []byte
	if req.AuthConfig != nil {
		authConfigJSON, _ = json.Marshal(req.AuthConfig)
	} else {
		authConfigJSON = []byte("{}")
	}

	// Convert event_types to []string
	var eventTypes []string
	if req.EventTypes != nil {
		for _, et := range *req.EventTypes {
			eventTypes = append(eventTypes, string(et))
		}
	}

	// Convert event_filter to JSON
	var eventFilterJSON []byte
	if req.EventFilter != nil {
		eventFilterJSON, _ = json.Marshal(req.EventFilter)
	} else {
		eventFilterJSON = []byte("{}")
	}

	// Set defaults
	authType := "none"
	if req.AuthType != nil {
		authType = string(*req.AuthType)
	}
	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	batchSize := int32(1)
	if req.BatchSize != nil {
		batchSize = int32(*req.BatchSize)
	}
	timeoutMs := int32(5000)
	if req.TimeoutMs != nil {
		timeoutMs = int32(*req.TimeoutMs)
	}
	retryMax := int32(3)
	if req.RetryMax != nil {
		retryMax = int32(*req.RetryMax)
	}

	params := db.CreateWebhookDestinationParams{
		OrgID:          orgID,
		Name:           req.Name,
		Url:            req.Url,
		AuthType:       authType,
		AuthConfig:     authConfigJSON,
		EventTypes:     eventTypes,
		EventFilter:    eventFilterJSON,
		Enabled:        enabled,
		BatchSize:      batchSize,
		TimeoutMs:      timeoutMs,
		RetryMax:       retryMax,
		RetryBackoffMs: 1000, // Default
		SigningSecret:  &signingSecret,
		CreatedBy:      pgtype.UUID{Bytes: employeeID, Valid: true},
	}

	dest, err := h.db.CreateWebhookDestination(ctx, params)
	if err != nil {
		// Check for unique constraint violation
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, "Webhook destination with this name already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to create webhook destination")
		return
	}

	response := dbWebhookDestinationToAPI(dest)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// GetWebhookDestination handles GET /webhooks/{webhookId}
func (h *WebhooksHandler) GetWebhookDestination(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, err := middleware.GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	webhookID, err := uuid.Parse(chi.URLParam(r, "webhookId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid webhook ID")
		return
	}

	dest, err := h.db.GetWebhookDestination(ctx, db.GetWebhookDestinationParams{
		ID:    webhookID,
		OrgID: orgID,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, "Webhook destination not found")
		return
	}

	response := dbWebhookDestinationToAPI(dest)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateWebhookDestination handles PATCH /webhooks/{webhookId}
func (h *WebhooksHandler) UpdateWebhookDestination(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, err := middleware.GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	webhookID, err := uuid.Parse(chi.URLParam(r, "webhookId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid webhook ID")
		return
	}

	// TODO: Check role (admin/manager)

	var req api.UpdateWebhookDestinationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Build update params
	params := db.UpdateWebhookDestinationParams{
		ID:    webhookID,
		OrgID: orgID,
	}

	if req.Name != nil {
		params.Name = req.Name
	}
	if req.Url != nil {
		params.Url = req.Url
	}
	if req.AuthType != nil {
		authType := string(*req.AuthType)
		params.AuthType = &authType
	}
	if req.AuthConfig != nil {
		authConfigJSON, _ := json.Marshal(req.AuthConfig)
		params.AuthConfig = authConfigJSON
	}
	if req.EventTypes != nil {
		params.EventTypes = *req.EventTypes
	}
	if req.EventFilter != nil {
		eventFilterJSON, _ := json.Marshal(req.EventFilter)
		params.EventFilter = eventFilterJSON
	}
	if req.Enabled != nil {
		params.Enabled = req.Enabled
	}
	if req.BatchSize != nil {
		bs := int32(*req.BatchSize)
		params.BatchSize = &bs
	}
	if req.TimeoutMs != nil {
		tm := int32(*req.TimeoutMs)
		params.TimeoutMs = &tm
	}
	if req.RetryMax != nil {
		rm := int32(*req.RetryMax)
		params.RetryMax = &rm
	}

	dest, err := h.db.UpdateWebhookDestination(ctx, params)
	if err != nil {
		writeError(w, http.StatusNotFound, "Webhook destination not found")
		return
	}

	response := dbWebhookDestinationToAPI(dest)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// DeleteWebhookDestination handles DELETE /webhooks/{webhookId}
func (h *WebhooksHandler) DeleteWebhookDestination(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, err := middleware.GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	webhookID, err := uuid.Parse(chi.URLParam(r, "webhookId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid webhook ID")
		return
	}

	// TODO: Check role (admin/manager)

	err = h.db.DeleteWebhookDestination(ctx, db.DeleteWebhookDestinationParams{
		ID:    webhookID,
		OrgID: orgID,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, "Webhook destination not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// TestWebhookDestination handles POST /webhooks/{webhookId}/test
func (h *WebhooksHandler) TestWebhookDestination(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, err := middleware.GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	webhookID, err := uuid.Parse(chi.URLParam(r, "webhookId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid webhook ID")
		return
	}

	dest, err := h.db.GetWebhookDestination(ctx, db.GetWebhookDestinationParams{
		ID:    webhookID,
		OrgID: orgID,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, "Webhook destination not found")
		return
	}

	// Send test event
	startTime := time.Now()
	success, statusCode, testErr := sendTestWebhook(dest)
	responseTimeMs := int(time.Since(startTime).Milliseconds())

	response := api.WebhookTestResult{
		Success:        success,
		ResponseStatus: &statusCode,
		ResponseTimeMs: &responseTimeMs,
	}
	if testErr != nil {
		errStr := testErr.Error()
		response.Error = &errStr
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ListWebhookDeliveries handles GET /webhooks/{webhookId}/deliveries
func (h *WebhooksHandler) ListWebhookDeliveries(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orgID, err := middleware.GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	webhookID, err := uuid.Parse(chi.URLParam(r, "webhookId"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid webhook ID")
		return
	}

	// Verify webhook belongs to org
	_, err = h.db.GetWebhookDestination(ctx, db.GetWebhookDestinationParams{
		ID:    webhookID,
		OrgID: orgID,
	})
	if err != nil {
		writeError(w, http.StatusNotFound, "Webhook destination not found")
		return
	}

	// Parse query params
	limit := int32(50)
	offset := int32(0)
	if l := r.URL.Query().Get("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = int32(parsed)
		}
	}
	if o := r.URL.Query().Get("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	deliveries, err := h.db.ListDeliveriesByDestination(ctx, db.ListDeliveriesByDestinationParams{
		DestinationID: webhookID,
		Limit:         limit,
		Offset:        offset,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to list deliveries")
		return
	}

	apiDeliveries := make([]api.WebhookDelivery, len(deliveries))
	for i, del := range deliveries {
		apiDeliveries[i] = dbWebhookDeliveryToAPI(del)
	}

	// Get total count for pagination
	statusCounts, _ := h.db.CountDeliveriesByStatus(ctx, webhookID)
	total := int64(0)
	for _, sc := range statusCounts {
		total += sc.Count
	}

	response := api.ListWebhookDeliveriesResponse{
		Deliveries: apiDeliveries,
		Pagination: api.PaginationMeta{
			Page:    int(offset/limit) + 1,
			PerPage: int(limit),
			Total:   int(total),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Helper functions

func generateSigningSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return "ubik_whsec_" + hex.EncodeToString(bytes), nil
}

func dbWebhookDestinationRowToAPI(dest db.ListWebhookDestinationsRow) api.WebhookDestination {
	id := openapi_types.UUID(dest.ID)
	authType := api.WebhookDestinationAuthType(dest.AuthType)

	result := api.WebhookDestination{
		Id:        id,
		Name:      dest.Name,
		Url:       dest.Url,
		AuthType:  authType,
		Enabled:   dest.Enabled,
		CreatedAt: dest.CreatedAt.Time,
	}

	if len(dest.EventTypes) > 0 {
		eventTypes := make([]api.WebhookDestinationEventTypes, len(dest.EventTypes))
		for i, et := range dest.EventTypes {
			eventTypes[i] = api.WebhookDestinationEventTypes(et)
		}
		result.EventTypes = &eventTypes
	}

	if len(dest.EventFilter) > 0 {
		var filter map[string]interface{}
		if err := json.Unmarshal(dest.EventFilter, &filter); err == nil {
			result.EventFilter = &filter
		}
	}

	result.BatchSize = intPtr(int(dest.BatchSize))
	result.TimeoutMs = intPtr(int(dest.TimeoutMs))
	result.RetryMax = intPtr(int(dest.RetryMax))

	if dest.CreatedBy.Valid {
		createdBy := openapi_types.UUID(dest.CreatedBy.Bytes)
		result.CreatedBy = &createdBy
	}

	if dest.UpdatedAt.Valid {
		result.UpdatedAt = &dest.UpdatedAt.Time
	}

	return result
}

func dbWebhookDestinationToAPI(dest db.WebhookDestination) api.WebhookDestination {
	id := openapi_types.UUID(dest.ID)
	authType := api.WebhookDestinationAuthType(dest.AuthType)

	result := api.WebhookDestination{
		Id:        id,
		Name:      dest.Name,
		Url:       dest.Url,
		AuthType:  authType,
		Enabled:   dest.Enabled,
		CreatedAt: dest.CreatedAt.Time,
	}

	if len(dest.EventTypes) > 0 {
		eventTypes := make([]api.WebhookDestinationEventTypes, len(dest.EventTypes))
		for i, et := range dest.EventTypes {
			eventTypes[i] = api.WebhookDestinationEventTypes(et)
		}
		result.EventTypes = &eventTypes
	}

	if len(dest.EventFilter) > 0 {
		var filter map[string]interface{}
		if err := json.Unmarshal(dest.EventFilter, &filter); err == nil {
			result.EventFilter = &filter
		}
	}

	result.BatchSize = intPtr(int(dest.BatchSize))
	result.TimeoutMs = intPtr(int(dest.TimeoutMs))
	result.RetryMax = intPtr(int(dest.RetryMax))

	if dest.CreatedBy.Valid {
		createdBy := openapi_types.UUID(dest.CreatedBy.Bytes)
		result.CreatedBy = &createdBy
	}

	if dest.UpdatedAt.Valid {
		result.UpdatedAt = &dest.UpdatedAt.Time
	}

	return result
}

func dbWebhookDeliveryToAPI(del db.WebhookDelivery) api.WebhookDelivery {
	id := openapi_types.UUID(del.ID)
	destID := openapi_types.UUID(del.DestinationID)
	logID := openapi_types.UUID(del.LogID)
	status := api.WebhookDeliveryStatus(del.Status)

	result := api.WebhookDelivery{
		Id:            id,
		DestinationId: destID,
		LogId:         logID,
		Status:        status,
		Attempts:      int(del.Attempts),
		CreatedAt:     del.CreatedAt.Time,
	}

	if del.LastAttemptAt.Valid {
		result.LastAttemptAt = &del.LastAttemptAt.Time
	}
	if del.NextRetryAt.Valid {
		result.NextRetryAt = &del.NextRetryAt.Time
	}
	if del.ResponseStatus != nil {
		rs := int(*del.ResponseStatus)
		result.ResponseStatus = &rs
	}
	if del.ResponseBody != nil {
		result.ResponseBody = del.ResponseBody
	}
	if del.ErrorMessage != nil {
		result.ErrorMessage = del.ErrorMessage
	}
	if del.DeliveredAt.Valid {
		result.DeliveredAt = &del.DeliveredAt.Time
	}

	return result
}

func sendTestWebhook(dest db.WebhookDestination) (bool, int, error) {
	// TODO: Implement actual HTTP request to dest.Url
	// For now, return success
	return true, 200, nil
}

func intPtr(i int) *int {
	return &i
}

func isUniqueViolation(err error) bool {
	// Check for PostgreSQL unique constraint violation
	return err != nil && strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate")
}
