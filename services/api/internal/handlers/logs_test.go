package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

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
// POST /logs - Create Log Entry Tests
// ============================================================================

func TestCreateLog_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	employeeID := uuid.New()
	sessionID := uuid.New()
	agentID := uuid.New()

	// Test request payload
	requestBody := api.CreateLogRequest{
		SessionId:     &sessionID,
		AgentId:       &agentID,
		EventType:     "input",
		EventCategory: "io",
		Content:       stringPtr("User typed: write a test"),
		Payload: &map[string]interface{}{
			"command": "test",
			"tool":    "bash",
		},
	}

	// Expected database call
	mockDB.EXPECT().
		CreateActivityLog(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, params db.CreateActivityLogParams) (db.ActivityLog, error) {
			// Verify params
			assert.Equal(t, orgID, params.OrgID)
			assert.Equal(t, "input", params.EventType)
			assert.Equal(t, "io", params.EventCategory)

			// Return created log
			return db.ActivityLog{
				ID:            uuid.New(),
				OrgID:         orgID,
				EmployeeID:    pgtype.UUID{Bytes: employeeID, Valid: true},
				SessionID:     pgtype.UUID{Bytes: sessionID, Valid: true},
				AgentID:       pgtype.UUID{Bytes: agentID, Valid: true},
				EventType:     params.EventType,
				EventCategory: params.EventCategory,
				Content:       params.Content,
				Payload:       params.Payload,
				CreatedAt:     pgtype.Timestamp{Valid: true},
			}, nil
		})

	handler := handlers.NewLogsHandler(mockDB, nil)

	// Create request
	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewReader(bodyBytes))
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	req = req.WithContext(handlers.SetEmployeeIDInContext(req.Context(), employeeID))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	// Call handler
	handler.CreateLog(rec, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response api.ActivityLog
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.NotEmpty(t, response.Id.String())
	assert.Equal(t, orgID, uuid.UUID(response.OrgId))
	assert.Equal(t, "input", response.EventType)
	assert.Equal(t, "io", response.EventCategory)
}

func TestCreateLog_MissingRequiredFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	employeeID := uuid.New()

	// Invalid request - missing event_type
	requestBody := api.CreateLogRequest{
		EventCategory: "io",
	}

	handler := handlers.NewLogsHandler(mockDB, nil)

	// Create request
	bodyBytes, _ := json.Marshal(requestBody)
	req := httptest.NewRequest(http.MethodPost, "/logs", bytes.NewReader(bodyBytes))
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	req = req.WithContext(handlers.SetEmployeeIDInContext(req.Context(), employeeID))
	req.Header.Set("Content-Type", "application/json")

	rec := httptest.NewRecorder()

	// Call handler
	handler.CreateLog(rec, req)

	// Verify error response
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	// The handler returns plain text error for now
	// In production, it should return JSON error with proper format
	body := rec.Body.String()
	assert.Contains(t, body, "event_type")
}

// ============================================================================
// GET /logs - List Logs Tests
// ============================================================================

func TestListLogs_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	employeeID := uuid.New()
	sessionID := uuid.New()

	// Test logs
	logs := []db.ActivityLog{
		{
			ID:            uuid.New(),
			OrgID:         orgID,
			EmployeeID:    pgtype.UUID{Bytes: employeeID, Valid: true},
			SessionID:     pgtype.UUID{Bytes: sessionID, Valid: true},
			EventType:     "input",
			EventCategory: "io",
			Payload:       []byte("{}"),
			CreatedAt:     pgtype.Timestamp{Valid: true},
		},
		{
			ID:            uuid.New(),
			OrgID:         orgID,
			EmployeeID:    pgtype.UUID{Bytes: employeeID, Valid: true},
			SessionID:     pgtype.UUID{Bytes: sessionID, Valid: true},
			EventType:     "output",
			EventCategory: "io",
			Payload:       []byte("{}"),
			CreatedAt:     pgtype.Timestamp{Valid: true},
		},
	}

	// Expect database query with filtered method
	mockDB.EXPECT().
		ListActivityLogsFiltered(gomock.Any(), gomock.Any()).
		Return(logs, nil)

	// Expect count query with filtered method
	mockDB.EXPECT().
		CountActivityLogsFiltered(gomock.Any(), gomock.Any()).
		Return(int64(2), nil)

	handler := handlers.NewLogsHandler(mockDB, nil)

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/logs?page=1&per_page=20", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rec := httptest.NewRecorder()

	// Call handler
	handler.ListLogs(rec, req, api.ListLogsParams{})

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListLogsResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response.Logs, 2)
	assert.NotNil(t, response.Pagination)
}

func TestListLogs_WithSessionFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	sessionID := uuid.New()

	// Expect database query with session filter using filtered method
	mockDB.EXPECT().
		ListActivityLogsFiltered(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, params db.ListActivityLogsFilteredParams) ([]db.ActivityLog, error) {
			// Verify session_id is set
			assert.Equal(t, orgID, params.OrgID)
			assert.True(t, params.SessionID.Valid)
			assert.Equal(t, sessionID, uuid.UUID(params.SessionID.Bytes))
			return []db.ActivityLog{}, nil
		})

	// Expect count query
	mockDB.EXPECT().
		CountActivityLogsFiltered(gomock.Any(), gomock.Any()).
		Return(int64(0), nil)

	handler := handlers.NewLogsHandler(mockDB, nil)

	// Create request with session_id filter
	req := httptest.NewRequest(http.MethodGet, "/logs?session_id="+sessionID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rec := httptest.NewRecorder()

	// Call handler
	handler.ListLogs(rec, req, api.ListLogsParams{
		SessionId: &sessionID,
	})

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestListLogs_WithEmployeeFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	employeeID := uuid.New()

	// Test logs
	logs := []db.ActivityLog{
		{
			ID:            uuid.New(),
			OrgID:         orgID,
			EmployeeID:    pgtype.UUID{Bytes: employeeID, Valid: true},
			EventType:     "input",
			EventCategory: "io",
			Payload:       []byte("{}"),
			CreatedAt:     pgtype.Timestamp{Valid: true},
		},
	}

	// Expect database query with employee_id filter
	mockDB.EXPECT().
		ListActivityLogsFiltered(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, params db.ListActivityLogsFilteredParams) ([]db.ActivityLog, error) {
			// Verify that employee_id filter is applied
			assert.Equal(t, orgID, params.OrgID)
			assert.True(t, params.EmployeeID.Valid)
			assert.Equal(t, employeeID, uuid.UUID(params.EmployeeID.Bytes))
			return logs, nil
		})

	// Expect count query
	mockDB.EXPECT().
		CountActivityLogsFiltered(gomock.Any(), gomock.Any()).
		Return(int64(1), nil)

	handler := handlers.NewLogsHandler(mockDB, nil)

	// Create request with employee_id parameter
	req := httptest.NewRequest(http.MethodGet, "/logs?page=1&per_page=20&employee_id="+employeeID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rec := httptest.NewRecorder()

	// Call handler with employee filter
	handler.ListLogs(rec, req, api.ListLogsParams{
		EmployeeId: &employeeID,
	})

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListLogsResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Len(t, response.Logs, 1)
	assert.NotNil(t, response.Pagination)
}
