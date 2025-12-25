package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/generated/mocks"
	"github.com/rastrigin-systems/arfa/services/api/internal/handlers"
)

// ============================================================================
// ListActivityLogs Tests
// ============================================================================

// TDD Lesson: Testing activity logs retrieval with default pagination
func TestListActivityLogs_Success_DefaultPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	log1ID := uuid.New()
	log2ID := uuid.New()

	// Create test activity logs
	logs := []db.ActivityLog{
		{
			ID:            log1ID,
			OrgID:         orgID,
			EmployeeID:    pgtype.UUID{Valid: false},
			EventType:     "auth.login",
			EventCategory: "auth",
			Payload:       []byte(`{}`),
			CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
		{
			ID:            log2ID,
			OrgID:         orgID,
			EmployeeID:    pgtype.UUID{Valid: false},
			EventType:     "agent.installed",
			EventCategory: "agent",
			Payload:       []byte(`{"agent_id": "claude-code"}`),
			CreatedAt:     pgtype.Timestamp{Time: time.Now().Add(-10 * time.Minute), Valid: true},
		},
	}

	// Expect list query with default pagination (limit=10, offset=0)
	mockDB.EXPECT().
		ListActivityLogs(gomock.Any(), db.ListActivityLogsParams{
			OrgID:  orgID,
			Limit:  10, // default
			Offset: 0,  // default
		}).
		Return(logs, nil)

	// Expect count query
	mockDB.EXPECT().
		CountActivityLogs(gomock.Any(), orgID).
		Return(int64(2), nil)

	handler := handlers.NewActivityLogsHandler(mockDB)

	// Create request without pagination parameters (use defaults)
	req := httptest.NewRequest(http.MethodGet, "/activity-logs", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListActivityLogs(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response handlers.ActivityLogsListResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify response structure
	assert.Len(t, response.ActivityLogs, 2)
	assert.Equal(t, int64(2), response.Total)
	assert.Equal(t, int32(10), response.Limit)
	assert.Equal(t, int32(0), response.Offset)

	// Verify first log
	assert.Equal(t, log1ID.String(), response.ActivityLogs[0].ID)
	assert.Equal(t, "auth.login", response.ActivityLogs[0].EventType)
	assert.Equal(t, "auth", response.ActivityLogs[0].EventCategory)
	assert.Equal(t, "Logged in", response.ActivityLogs[0].Message)

	// Verify second log
	assert.Equal(t, log2ID.String(), response.ActivityLogs[1].ID)
	assert.Equal(t, "agent.installed", response.ActivityLogs[1].EventType)
	assert.Equal(t, "Installed new agent configuration", response.ActivityLogs[1].Message)
}

// TDD Lesson: Testing activity logs with custom pagination
func TestListActivityLogs_Success_CustomPagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()

	logs := []db.ActivityLog{
		{
			ID:            uuid.New(),
			OrgID:         orgID,
			EmployeeID:    pgtype.UUID{Valid: false},
			EventType:     "team.created",
			EventCategory: "admin",
			Payload:       []byte(`{}`),
			CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
	}

	// Expect list query with custom pagination (limit=5, offset=10)
	mockDB.EXPECT().
		ListActivityLogs(gomock.Any(), db.ListActivityLogsParams{
			OrgID:  orgID,
			Limit:  5,
			Offset: 10,
		}).
		Return(logs, nil)

	// Expect count query
	mockDB.EXPECT().
		CountActivityLogs(gomock.Any(), orgID).
		Return(int64(25), nil)

	handler := handlers.NewActivityLogsHandler(mockDB)

	// Create request with custom pagination
	req := httptest.NewRequest(http.MethodGet, "/activity-logs?limit=5&offset=10", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListActivityLogs(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response handlers.ActivityLogsListResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify pagination parameters
	assert.Equal(t, int32(5), response.Limit)
	assert.Equal(t, int32(10), response.Offset)
	assert.Equal(t, int64(25), response.Total)
}

// TDD Lesson: Testing activity logs with employee name fetching
func TestListActivityLogs_Success_WithEmployeeName(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	empID := uuid.New()
	logID := uuid.New()

	// Create activity log with employee
	logs := []db.ActivityLog{
		{
			ID:            logID,
			OrgID:         orgID,
			EmployeeID:    pgtype.UUID{Bytes: [16]byte(empID[:]), Valid: true},
			EventType:     "agent.updated",
			EventCategory: "agent",
			Payload:       []byte(`{}`),
			CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
	}

	// Expect list query
	mockDB.EXPECT().
		ListActivityLogs(gomock.Any(), gomock.Any()).
		Return(logs, nil)

	// Expect employee name fetch
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(db.GetEmployeeRow{
			ID:       empID,
			FullName: "Alice Smith",
		}, nil)

	// Expect count query
	mockDB.EXPECT().
		CountActivityLogs(gomock.Any(), orgID).
		Return(int64(1), nil)

	handler := handlers.NewActivityLogsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/activity-logs", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListActivityLogs(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response handlers.ActivityLogsListResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify employee name is included
	assert.Len(t, response.ActivityLogs, 1)
	assert.NotNil(t, response.ActivityLogs[0].EmployeeID)
	assert.NotNil(t, response.ActivityLogs[0].EmployeeName)
	assert.Equal(t, "Alice Smith", *response.ActivityLogs[0].EmployeeName)
}

// TDD Lesson: Testing activity logs with missing employee (graceful handling)
func TestListActivityLogs_Success_EmployeeNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	empID := uuid.New()

	// Create activity log with employee
	logs := []db.ActivityLog{
		{
			ID:            uuid.New(),
			OrgID:         orgID,
			EmployeeID:    pgtype.UUID{Bytes: [16]byte(empID[:]), Valid: true},
			EventType:     "auth.logout",
			EventCategory: "auth",
			Payload:       []byte(`{}`),
			CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
	}

	// Expect list query
	mockDB.EXPECT().
		ListActivityLogs(gomock.Any(), gomock.Any()).
		Return(logs, nil)

	// Employee fetch fails (deleted employee, etc.)
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(db.GetEmployeeRow{}, assert.AnError)

	// Expect count query
	mockDB.EXPECT().
		CountActivityLogs(gomock.Any(), orgID).
		Return(int64(1), nil)

	handler := handlers.NewActivityLogsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/activity-logs", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListActivityLogs(rec, req)

	// Assert HTTP 200 OK (graceful handling)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response handlers.ActivityLogsListResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify employee name is nil (not found)
	assert.Len(t, response.ActivityLogs, 1)
	assert.NotNil(t, response.ActivityLogs[0].EmployeeID)
	assert.Nil(t, response.ActivityLogs[0].EmployeeName)
}

// TDD Lesson: Testing unauthorized request (no org_id)
func TestListActivityLogs_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewActivityLogsHandler(mockDB)

	// Create request WITHOUT org_id in context
	req := httptest.NewRequest(http.MethodGet, "/activity-logs", nil)
	rec := httptest.NewRecorder()

	handler.ListActivityLogs(rec, req)

	// Assert HTTP 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Parse error response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Unauthorized", response["error"])
}

// TDD Lesson: Testing database error on list
func TestListActivityLogs_DatabaseError_OnList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()

	// Simulate database error
	mockDB.EXPECT().
		ListActivityLogs(gomock.Any(), gomock.Any()).
		Return(nil, assert.AnError)

	handler := handlers.NewActivityLogsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/activity-logs", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListActivityLogs(rec, req)

	// Assert HTTP 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// Parse error response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Failed to fetch activity logs", response["error"])
}

// TDD Lesson: Testing database error on count
func TestListActivityLogs_DatabaseError_OnCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()

	logs := []db.ActivityLog{
		{
			ID:            uuid.New(),
			OrgID:         orgID,
			EmployeeID:    pgtype.UUID{Valid: false},
			EventType:     "test.event",
			EventCategory: "test",
			Payload:       []byte(`{}`),
			CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
		},
	}

	mockDB.EXPECT().
		ListActivityLogs(gomock.Any(), gomock.Any()).
		Return(logs, nil)

	// Count query fails
	mockDB.EXPECT().
		CountActivityLogs(gomock.Any(), orgID).
		Return(int64(0), assert.AnError)

	handler := handlers.NewActivityLogsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/activity-logs", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListActivityLogs(rec, req)

	// Assert HTTP 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// Parse error response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Failed to count activity logs", response["error"])
}

// TDD Lesson: Testing empty result set
func TestListActivityLogs_Success_EmptyResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()

	// Return empty slice
	logs := []db.ActivityLog{}

	mockDB.EXPECT().
		ListActivityLogs(gomock.Any(), gomock.Any()).
		Return(logs, nil)

	mockDB.EXPECT().
		CountActivityLogs(gomock.Any(), orgID).
		Return(int64(0), nil)

	handler := handlers.NewActivityLogsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/activity-logs", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListActivityLogs(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response handlers.ActivityLogsListResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify empty results
	assert.Len(t, response.ActivityLogs, 0)
	assert.Equal(t, int64(0), response.Total)
}

// TDD Lesson: Testing invalid pagination parameters (use defaults)
func TestListActivityLogs_Success_InvalidPaginationDefaults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()

	logs := []db.ActivityLog{}

	// Expect list query with default values (invalid params ignored)
	mockDB.EXPECT().
		ListActivityLogs(gomock.Any(), db.ListActivityLogsParams{
			OrgID:  orgID,
			Limit:  10, // default (invalid "abc" ignored)
			Offset: 0,  // default (negative ignored)
		}).
		Return(logs, nil)

	mockDB.EXPECT().
		CountActivityLogs(gomock.Any(), orgID).
		Return(int64(0), nil)

	handler := handlers.NewActivityLogsHandler(mockDB)

	// Create request with invalid pagination parameters
	req := httptest.NewRequest(http.MethodGet, "/activity-logs?limit=abc&offset=-5", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListActivityLogs(rec, req)

	// Assert HTTP 200 OK (graceful handling)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response handlers.ActivityLogsListResponse
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify defaults were used
	assert.Equal(t, int32(10), response.Limit)
	assert.Equal(t, int32(0), response.Offset)
}

// TDD Lesson: Testing various event message generation
func TestListActivityLogs_Success_MessageGeneration(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()

	// Test various event types and their messages
	testCases := []struct {
		eventType     string
		eventCategory string
		expectedMsg   string
	}{
		{"auth.login", "auth", "Logged in"},
		{"auth.logout", "auth", "Logged out"},
		{"agent.installed", "agent", "Installed new agent configuration"},
		{"agent.updated", "agent", "Updated agent configuration"},
		{"agent.deleted", "agent", "Removed agent configuration"},
		{"mcp.configured", "mcp", "Configured MCP server"},
		{"mcp.updated", "mcp", "Updated MCP configuration"},
		{"employee.created", "admin", "Created new employee"},
		{"employee.updated", "admin", "Updated employee"},
		{"team.created", "admin", "Created new team"},
		{"unknown.event", "unknown", "unknown.event"}, // Fallback
	}

	for _, tc := range testCases {
		t.Run(tc.eventType, func(t *testing.T) {
			logs := []db.ActivityLog{
				{
					ID:            uuid.New(),
					OrgID:         orgID,
					EmployeeID:    pgtype.UUID{Valid: false},
					EventType:     tc.eventType,
					EventCategory: tc.eventCategory,
					Payload:       []byte(`{}`),
					CreatedAt:     pgtype.Timestamp{Time: time.Now(), Valid: true},
				},
			}

			mockDB.EXPECT().
				ListActivityLogs(gomock.Any(), gomock.Any()).
				Return(logs, nil)

			mockDB.EXPECT().
				CountActivityLogs(gomock.Any(), orgID).
				Return(int64(1), nil)

			handler := handlers.NewActivityLogsHandler(mockDB)

			req := httptest.NewRequest(http.MethodGet, "/activity-logs", nil)
			req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
			rec := httptest.NewRecorder()

			handler.ListActivityLogs(rec, req)

			// Parse response
			var response handlers.ActivityLogsListResponse
			err := json.NewDecoder(rec.Body).Decode(&response)
			require.NoError(t, err)

			// Verify message generation
			assert.Equal(t, tc.expectedMsg, response.ActivityLogs[0].Message)
		})
	}
}
