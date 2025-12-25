package handlers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rastrigin-systems/arfa/generated/api"
	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/generated/mocks"
	"github.com/rastrigin-systems/arfa/services/api/internal/handlers"
)

// ============================================================================
// ListWebhookDestinations Tests
// ============================================================================

func TestListWebhookDestinations_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	webhookID := uuid.New()
	createdBy := uuid.New()

	destinations := []db.ListWebhookDestinationsRow{
		{
			ID:             webhookID,
			OrgID:          orgID,
			Name:           "SIEM Export",
			Url:            "https://siem.example.com/events",
			AuthType:       "bearer",
			EventTypes:     []string{"tool_call", "permission_denied"},
			EventFilter:    []byte(`{}`),
			Enabled:        true,
			BatchSize:      10,
			TimeoutMs:      5000,
			RetryMax:       3,
			RetryBackoffMs: 1000,
			CreatedBy:      pgtype.UUID{Bytes: createdBy, Valid: true},
			CreatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
	}

	mockDB.EXPECT().
		ListWebhookDestinations(gomock.Any(), orgID).
		Return(destinations, nil)

	handler := handlers.NewWebhooksHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListWebhookDestinations(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListWebhookDestinationsResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response.Destinations, 1)
	assert.Equal(t, "SIEM Export", response.Destinations[0].Name)
	assert.Equal(t, "https://siem.example.com/events", response.Destinations[0].Url)
	assert.Equal(t, api.WebhookDestinationAuthTypeBearer, response.Destinations[0].AuthType)
	assert.True(t, response.Destinations[0].Enabled)
}

func TestListWebhookDestinations_Empty(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()

	mockDB.EXPECT().
		ListWebhookDestinations(gomock.Any(), orgID).
		Return([]db.ListWebhookDestinationsRow{}, nil)

	handler := handlers.NewWebhooksHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListWebhookDestinations(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListWebhookDestinationsResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response.Destinations, 0)
}

func TestListWebhookDestinations_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewWebhooksHandler(mockDB)

	// Request without org_id in context
	req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)
	rec := httptest.NewRecorder()

	handler.ListWebhookDestinations(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestListWebhookDestinations_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()

	mockDB.EXPECT().
		ListWebhookDestinations(gomock.Any(), orgID).
		Return(nil, assert.AnError)

	handler := handlers.NewWebhooksHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/webhooks", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListWebhookDestinations(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}

// ============================================================================
// CreateWebhookDestination Tests
// ============================================================================

func TestCreateWebhookDestination_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	employeeID := uuid.New()
	webhookID := uuid.New()

	// Expect create call with matching params
	mockDB.EXPECT().
		CreateWebhookDestination(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ interface{}, params db.CreateWebhookDestinationParams) (db.WebhookDestination, error) {
			assert.Equal(t, orgID, params.OrgID)
			assert.Equal(t, "SIEM Export", params.Name)
			assert.Equal(t, "https://siem.example.com/events", params.Url)
			assert.Equal(t, "bearer", params.AuthType)
			assert.True(t, params.Enabled)

			return db.WebhookDestination{
				ID:             webhookID,
				OrgID:          orgID,
				Name:           params.Name,
				Url:            params.Url,
				AuthType:       params.AuthType,
				AuthConfig:     params.AuthConfig,
				EventTypes:     params.EventTypes,
				EventFilter:    params.EventFilter,
				Enabled:        params.Enabled,
				BatchSize:      params.BatchSize,
				TimeoutMs:      params.TimeoutMs,
				RetryMax:       params.RetryMax,
				RetryBackoffMs: params.RetryBackoffMs,
				SigningSecret:  params.SigningSecret,
				CreatedBy:      params.CreatedBy,
				CreatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
				UpdatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
			}, nil
		})

	handler := handlers.NewWebhooksHandler(mockDB)

	authType := api.CreateWebhookDestinationRequestAuthTypeBearer
	body := api.CreateWebhookDestinationRequest{
		Name:     "SIEM Export",
		Url:      "https://siem.example.com/events",
		AuthType: &authType,
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := handlers.SetOrgIDInContext(req.Context(), orgID)
	ctx = handlers.SetEmployeeIDInContext(ctx, employeeID)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.CreateWebhookDestination(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var response api.WebhookDestination
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "SIEM Export", response.Name)
	assert.Equal(t, "https://siem.example.com/events", response.Url)
	assert.Equal(t, api.WebhookDestinationAuthTypeBearer, response.AuthType)
}

func TestCreateWebhookDestination_MissingName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	employeeID := uuid.New()

	handler := handlers.NewWebhooksHandler(mockDB)

	body := api.CreateWebhookDestinationRequest{
		Url: "https://siem.example.com/events",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := handlers.SetOrgIDInContext(req.Context(), orgID)
	ctx = handlers.SetEmployeeIDInContext(ctx, employeeID)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.CreateWebhookDestination(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)
	assert.Equal(t, "Name and URL are required", response["error"])
}

func TestCreateWebhookDestination_MissingURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	employeeID := uuid.New()

	handler := handlers.NewWebhooksHandler(mockDB)

	body := api.CreateWebhookDestinationRequest{
		Name: "Test Webhook",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	ctx := handlers.SetOrgIDInContext(req.Context(), orgID)
	ctx = handlers.SetEmployeeIDInContext(ctx, employeeID)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.CreateWebhookDestination(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestCreateWebhookDestination_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewWebhooksHandler(mockDB)

	body := api.CreateWebhookDestinationRequest{
		Name: "Test",
		Url:  "https://example.com",
	}
	bodyBytes, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateWebhookDestination(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestCreateWebhookDestination_InvalidBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	employeeID := uuid.New()

	handler := handlers.NewWebhooksHandler(mockDB)

	req := httptest.NewRequest(http.MethodPost, "/webhooks", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	ctx := handlers.SetOrgIDInContext(req.Context(), orgID)
	ctx = handlers.SetEmployeeIDInContext(ctx, employeeID)
	req = req.WithContext(ctx)
	rec := httptest.NewRecorder()

	handler.CreateWebhookDestination(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ============================================================================
// GetWebhookDestination Tests
// ============================================================================

func TestGetWebhookDestination_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	webhookID := uuid.New()

	mockDB.EXPECT().
		GetWebhookDestination(gomock.Any(), db.GetWebhookDestinationParams{
			ID:    webhookID,
			OrgID: orgID,
		}).
		Return(db.WebhookDestination{
			ID:             webhookID,
			OrgID:          orgID,
			Name:           "Test Webhook",
			Url:            "https://example.com/webhook",
			AuthType:       "none",
			AuthConfig:     []byte(`{}`),
			EventTypes:     []string{},
			EventFilter:    []byte(`{}`),
			Enabled:        true,
			BatchSize:      1,
			TimeoutMs:      5000,
			RetryMax:       3,
			RetryBackoffMs: 1000,
			CreatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
		}, nil)

	handler := handlers.NewWebhooksHandler(mockDB)

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("webhookId", webhookID.String())
	req := httptest.NewRequest(http.MethodGet, "/webhooks/"+webhookID.String(), nil)
	req = req.WithContext(handlers.WithChiContext(handlers.SetOrgIDInContext(req.Context(), orgID), chiCtx))
	rec := httptest.NewRecorder()

	handler.GetWebhookDestination(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.WebhookDestination
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Test Webhook", response.Name)
	assert.Equal(t, "https://example.com/webhook", response.Url)
}

func TestGetWebhookDestination_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	webhookID := uuid.New()

	mockDB.EXPECT().
		GetWebhookDestination(gomock.Any(), gomock.Any()).
		Return(db.WebhookDestination{}, assert.AnError)

	handler := handlers.NewWebhooksHandler(mockDB)

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("webhookId", webhookID.String())
	req := httptest.NewRequest(http.MethodGet, "/webhooks/"+webhookID.String(), nil)
	req = req.WithContext(handlers.WithChiContext(handlers.SetOrgIDInContext(req.Context(), orgID), chiCtx))
	rec := httptest.NewRecorder()

	handler.GetWebhookDestination(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestGetWebhookDestination_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()

	handler := handlers.NewWebhooksHandler(mockDB)

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("webhookId", "invalid-uuid")
	req := httptest.NewRequest(http.MethodGet, "/webhooks/invalid-uuid", nil)
	req = req.WithContext(handlers.WithChiContext(handlers.SetOrgIDInContext(req.Context(), orgID), chiCtx))
	rec := httptest.NewRecorder()

	handler.GetWebhookDestination(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ============================================================================
// UpdateWebhookDestination Tests
// ============================================================================

func TestUpdateWebhookDestination_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	webhookID := uuid.New()

	mockDB.EXPECT().
		UpdateWebhookDestination(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ interface{}, params db.UpdateWebhookDestinationParams) (db.WebhookDestination, error) {
			assert.Equal(t, webhookID, params.ID)
			assert.Equal(t, orgID, params.OrgID)
			assert.Equal(t, "Updated Name", *params.Name)

			return db.WebhookDestination{
				ID:             webhookID,
				OrgID:          orgID,
				Name:           *params.Name,
				Url:            "https://example.com/webhook",
				AuthType:       "none",
				AuthConfig:     []byte(`{}`),
				EventTypes:     []string{},
				EventFilter:    []byte(`{}`),
				Enabled:        true,
				BatchSize:      1,
				TimeoutMs:      5000,
				RetryMax:       3,
				RetryBackoffMs: 1000,
				CreatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
				UpdatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
			}, nil
		})

	handler := handlers.NewWebhooksHandler(mockDB)

	newName := "Updated Name"
	body := api.UpdateWebhookDestinationRequest{
		Name: &newName,
	}
	bodyBytes, _ := json.Marshal(body)

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("webhookId", webhookID.String())
	req := httptest.NewRequest(http.MethodPatch, "/webhooks/"+webhookID.String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.WithChiContext(handlers.SetOrgIDInContext(req.Context(), orgID), chiCtx))
	rec := httptest.NewRecorder()

	handler.UpdateWebhookDestination(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.WebhookDestination
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Updated Name", response.Name)
}

func TestUpdateWebhookDestination_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	webhookID := uuid.New()

	mockDB.EXPECT().
		UpdateWebhookDestination(gomock.Any(), gomock.Any()).
		Return(db.WebhookDestination{}, assert.AnError)

	handler := handlers.NewWebhooksHandler(mockDB)

	newName := "Updated Name"
	body := api.UpdateWebhookDestinationRequest{
		Name: &newName,
	}
	bodyBytes, _ := json.Marshal(body)

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("webhookId", webhookID.String())
	req := httptest.NewRequest(http.MethodPatch, "/webhooks/"+webhookID.String(), bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.WithChiContext(handlers.SetOrgIDInContext(req.Context(), orgID), chiCtx))
	rec := httptest.NewRecorder()

	handler.UpdateWebhookDestination(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ============================================================================
// DeleteWebhookDestination Tests
// ============================================================================

func TestDeleteWebhookDestination_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	webhookID := uuid.New()

	mockDB.EXPECT().
		DeleteWebhookDestination(gomock.Any(), db.DeleteWebhookDestinationParams{
			ID:    webhookID,
			OrgID: orgID,
		}).
		Return(nil)

	handler := handlers.NewWebhooksHandler(mockDB)

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("webhookId", webhookID.String())
	req := httptest.NewRequest(http.MethodDelete, "/webhooks/"+webhookID.String(), nil)
	req = req.WithContext(handlers.WithChiContext(handlers.SetOrgIDInContext(req.Context(), orgID), chiCtx))
	rec := httptest.NewRecorder()

	handler.DeleteWebhookDestination(rec, req)

	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestDeleteWebhookDestination_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	webhookID := uuid.New()

	mockDB.EXPECT().
		DeleteWebhookDestination(gomock.Any(), gomock.Any()).
		Return(assert.AnError)

	handler := handlers.NewWebhooksHandler(mockDB)

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("webhookId", webhookID.String())
	req := httptest.NewRequest(http.MethodDelete, "/webhooks/"+webhookID.String(), nil)
	req = req.WithContext(handlers.WithChiContext(handlers.SetOrgIDInContext(req.Context(), orgID), chiCtx))
	rec := httptest.NewRecorder()

	handler.DeleteWebhookDestination(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ============================================================================
// TestWebhookDestination Tests
// ============================================================================

func TestTestWebhookDestination_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	webhookID := uuid.New()

	mockDB.EXPECT().
		GetWebhookDestination(gomock.Any(), db.GetWebhookDestinationParams{
			ID:    webhookID,
			OrgID: orgID,
		}).
		Return(db.WebhookDestination{
			ID:             webhookID,
			OrgID:          orgID,
			Name:           "Test Webhook",
			Url:            "https://example.com/webhook",
			AuthType:       "none",
			AuthConfig:     []byte(`{}`),
			EventTypes:     []string{},
			EventFilter:    []byte(`{}`),
			Enabled:        true,
			BatchSize:      1,
			TimeoutMs:      5000,
			RetryMax:       3,
			RetryBackoffMs: 1000,
			CreatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
			UpdatedAt:      pgtype.Timestamp{Time: time.Now(), Valid: true},
		}, nil)

	handler := handlers.NewWebhooksHandler(mockDB)

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("webhookId", webhookID.String())
	req := httptest.NewRequest(http.MethodPost, "/webhooks/"+webhookID.String()+"/test", nil)
	req = req.WithContext(handlers.WithChiContext(handlers.SetOrgIDInContext(req.Context(), orgID), chiCtx))
	rec := httptest.NewRecorder()

	handler.TestWebhookDestination(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.WebhookTestResult
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// The current implementation returns success=true as a stub
	assert.True(t, response.Success)
}

func TestTestWebhookDestination_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	webhookID := uuid.New()

	mockDB.EXPECT().
		GetWebhookDestination(gomock.Any(), gomock.Any()).
		Return(db.WebhookDestination{}, assert.AnError)

	handler := handlers.NewWebhooksHandler(mockDB)

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("webhookId", webhookID.String())
	req := httptest.NewRequest(http.MethodPost, "/webhooks/"+webhookID.String()+"/test", nil)
	req = req.WithContext(handlers.WithChiContext(handlers.SetOrgIDInContext(req.Context(), orgID), chiCtx))
	rec := httptest.NewRecorder()

	handler.TestWebhookDestination(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// ============================================================================
// ListWebhookDeliveries Tests
// ============================================================================

func TestListWebhookDeliveries_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	webhookID := uuid.New()
	deliveryID := uuid.New()
	logID := uuid.New()

	// First, verify webhook exists
	mockDB.EXPECT().
		GetWebhookDestination(gomock.Any(), db.GetWebhookDestinationParams{
			ID:    webhookID,
			OrgID: orgID,
		}).
		Return(db.WebhookDestination{
			ID:    webhookID,
			OrgID: orgID,
			Name:  "Test Webhook",
		}, nil)

	// Then list deliveries
	mockDB.EXPECT().
		ListDeliveriesByDestination(gomock.Any(), db.ListDeliveriesByDestinationParams{
			DestinationID: webhookID,
			Limit:         50,
			Offset:        0,
		}).
		Return([]db.WebhookDelivery{
			{
				ID:            deliveryID,
				DestinationID: webhookID,
				LogID:         logID,
				Status:        "delivered",
				Attempts:      1,
				CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
				DeliveredAt:   pgtype.Timestamp{Time: time.Now(), Valid: true},
			},
		}, nil)

	// Count for pagination
	mockDB.EXPECT().
		CountDeliveriesByStatus(gomock.Any(), webhookID).
		Return([]db.CountDeliveriesByStatusRow{
			{Status: "delivered", Count: 1},
		}, nil)

	handler := handlers.NewWebhooksHandler(mockDB)

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("webhookId", webhookID.String())
	req := httptest.NewRequest(http.MethodGet, "/webhooks/"+webhookID.String()+"/deliveries", nil)
	req = req.WithContext(handlers.WithChiContext(handlers.SetOrgIDInContext(req.Context(), orgID), chiCtx))
	rec := httptest.NewRecorder()

	handler.ListWebhookDeliveries(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListWebhookDeliveriesResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response.Deliveries, 1)
	assert.Equal(t, api.WebhookDeliveryStatusDelivered, response.Deliveries[0].Status)
	assert.Equal(t, 1, response.Deliveries[0].Attempts)
	assert.Equal(t, 1, response.Pagination.Total)
}

func TestListWebhookDeliveries_WebhookNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	webhookID := uuid.New()

	mockDB.EXPECT().
		GetWebhookDestination(gomock.Any(), gomock.Any()).
		Return(db.WebhookDestination{}, assert.AnError)

	handler := handlers.NewWebhooksHandler(mockDB)

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("webhookId", webhookID.String())
	req := httptest.NewRequest(http.MethodGet, "/webhooks/"+webhookID.String()+"/deliveries", nil)
	req = req.WithContext(handlers.WithChiContext(handlers.SetOrgIDInContext(req.Context(), orgID), chiCtx))
	rec := httptest.NewRecorder()

	handler.ListWebhookDeliveries(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

func TestListWebhookDeliveries_CustomPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	webhookID := uuid.New()

	mockDB.EXPECT().
		GetWebhookDestination(gomock.Any(), gomock.Any()).
		Return(db.WebhookDestination{ID: webhookID, OrgID: orgID}, nil)

	// Expect custom pagination
	mockDB.EXPECT().
		ListDeliveriesByDestination(gomock.Any(), db.ListDeliveriesByDestinationParams{
			DestinationID: webhookID,
			Limit:         25,
			Offset:        50,
		}).
		Return([]db.WebhookDelivery{}, nil)

	mockDB.EXPECT().
		CountDeliveriesByStatus(gomock.Any(), webhookID).
		Return([]db.CountDeliveriesByStatusRow{}, nil)

	handler := handlers.NewWebhooksHandler(mockDB)

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add("webhookId", webhookID.String())
	req := httptest.NewRequest(http.MethodGet, "/webhooks/"+webhookID.String()+"/deliveries?limit=25&offset=50", nil)
	req = req.WithContext(handlers.WithChiContext(handlers.SetOrgIDInContext(req.Context(), orgID), chiCtx))
	rec := httptest.NewRecorder()

	handler.ListWebhookDeliveries(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
