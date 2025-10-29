package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"github.com/sergeirastrigin/ubik-enterprise/internal/handlers"
)

// Helper to create string pointer
func strPtr(s string) *string {
	return &s
}

// ============================================================================
// GetPendingCount Tests
// ============================================================================

// TDD Lesson: Testing pending count retrieval with org isolation
func TestGetPendingCount_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()

	// Expect count query for this organization
	mockDB.EXPECT().
		CountPendingRequestsByOrg(gomock.Any(), orgID).
		Return(int64(5), nil)

	handler := handlers.NewAgentRequestsHandler(mockDB)

	// Create request with org_id in context (set by middleware)
	req := httptest.NewRequest(http.MethodGet, "/agent-requests/pending/count", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.GetPendingCount(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, float64(5), response["pending_count"])
}

// TDD Lesson: Testing unauthorized request (no org_id in context)
func TestGetPendingCount_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewAgentRequestsHandler(mockDB)

	// Create request WITHOUT org_id in context (simulates missing auth)
	req := httptest.NewRequest(http.MethodGet, "/agent-requests/pending/count", nil)
	rec := httptest.NewRecorder()

	handler.GetPendingCount(rec, req)

	// Assert HTTP 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// Parse error response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Unauthorized", response["error"])
}

// TDD Lesson: Testing database error handling
func TestGetPendingCount_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()

	// Simulate database error
	mockDB.EXPECT().
		CountPendingRequestsByOrg(gomock.Any(), orgID).
		Return(int64(0), assert.AnError)

	handler := handlers.NewAgentRequestsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/agent-requests/pending/count", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.GetPendingCount(rec, req)

	// Assert HTTP 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// Parse error response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Failed to get pending count", response["error"])
}

// TDD Lesson: Testing zero pending requests
func TestGetPendingCount_ZeroCount(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()

	// Return zero count
	mockDB.EXPECT().
		CountPendingRequestsByOrg(gomock.Any(), orgID).
		Return(int64(0), nil)

	handler := handlers.NewAgentRequestsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/agent-requests/pending/count", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.GetPendingCount(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify zero count in response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, float64(0), response["pending_count"])
}

// ============================================================================
// ListAgentRequests Tests
// ============================================================================

// TDD Lesson: Testing list all agent requests (no filters)
func TestListAgentRequests_Success_NoFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	req1ID := uuid.New()
	req2ID := uuid.New()
	emp1ID := uuid.New()
	emp2ID := uuid.New()

	// Create test agent requests
	requests := []db.AgentRequest{
		{
			ID:          req1ID,
			EmployeeID:  emp1ID,
			RequestType: "agent_access",
			RequestData: []byte(`{"agent_id": "claude-code"}`),
			Status:      "pending",
			Reason:      strPtr("Need for project"),
			CreatedAt:   pgtype.Timestamp{Valid: true},
			ResolvedAt:  pgtype.Timestamp{Valid: false},
		},
		{
			ID:          req2ID,
			EmployeeID:  emp2ID,
			RequestType: "mcp_access",
			RequestData: []byte(`{"mcp_id": "github-mcp"}`),
			Status:      "approved",
			Reason:      strPtr("Required for development"),
			CreatedAt:   pgtype.Timestamp{Valid: true},
			ResolvedAt:  pgtype.Timestamp{Valid: true},
		},
	}

	// Expect list query with no filters, default pagination
	mockDB.EXPECT().
		ListAgentRequests(gomock.Any(), db.ListAgentRequestsParams{
			Status:      nil, // No status filter
			EmployeeID:  pgtype.UUID{Valid: false}, // No employee filter
			QueryLimit:  100, // Default limit from handler
			QueryOffset: 0,   // Default offset
		}).
		Return(requests, nil)

	// Expect count query with same filters
	mockDB.EXPECT().
		CountAgentRequests(gomock.Any(), db.CountAgentRequestsParams{
			Status:     nil,
			EmployeeID: pgtype.UUID{Valid: false},
		}).
		Return(int64(2), nil)

	handler := handlers.NewAgentRequestsHandler(mockDB)

	// Create request with no query parameters
	req := httptest.NewRequest(http.MethodGet, "/agent-requests", nil)
	rec := httptest.NewRecorder()

	handler.ListAgentRequests(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify response structure
	assert.NotNil(t, response["requests"])
	assert.Equal(t, float64(2), response["total"])

	// Verify requests array
	requestsArray, ok := response["requests"].([]interface{})
	require.True(t, ok)
	assert.Len(t, requestsArray, 2)
}

// TDD Lesson: Testing list with status filter
func TestListAgentRequests_Success_WithStatusFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	req1ID := uuid.New()
	emp1ID := uuid.New()

	// Only pending requests
	requests := []db.AgentRequest{
		{
			ID:          req1ID,
			EmployeeID:  emp1ID,
			RequestType: "agent_access",
			RequestData: []byte(`{"agent_id": "claude-code"}`),
			Status:      "pending",
			Reason:      strPtr("Need for project"),
			CreatedAt:   pgtype.Timestamp{Valid: true},
			ResolvedAt:  pgtype.Timestamp{Valid: false},
		},
	}

	statusFilter := "pending"

	// Expect list query with status filter
	mockDB.EXPECT().
		ListAgentRequests(gomock.Any(), db.ListAgentRequestsParams{
			Status:      &statusFilter,
			EmployeeID:  pgtype.UUID{Valid: false},
			QueryLimit:  100,
			QueryOffset: 0,
		}).
		Return(requests, nil)

	// Expect count query with status filter
	mockDB.EXPECT().
		CountAgentRequests(gomock.Any(), db.CountAgentRequestsParams{
			Status:     &statusFilter,
			EmployeeID: pgtype.UUID{Valid: false},
		}).
		Return(int64(1), nil)

	handler := handlers.NewAgentRequestsHandler(mockDB)

	// Create request with status query parameter
	req := httptest.NewRequest(http.MethodGet, "/agent-requests?status=pending", nil)
	rec := httptest.NewRecorder()

	handler.ListAgentRequests(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify filtered results
	assert.Equal(t, float64(1), response["total"])

	requestsArray, ok := response["requests"].([]interface{})
	require.True(t, ok)
	assert.Len(t, requestsArray, 1)
}

// TDD Lesson: Testing list with employee filter
func TestListAgentRequests_Success_WithEmployeeFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	req1ID := uuid.New()
	emp1ID := uuid.New()

	requests := []db.AgentRequest{
		{
			ID:          req1ID,
			EmployeeID:  emp1ID,
			RequestType: "agent_access",
			RequestData: []byte(`{"agent_id": "claude-code"}`),
			Status:      "pending",
			Reason:      strPtr("Need for project"),
			CreatedAt:   pgtype.Timestamp{Valid: true},
			ResolvedAt:  pgtype.Timestamp{Valid: false},
		},
	}

	// Expect list query with employee filter
	mockDB.EXPECT().
		ListAgentRequests(gomock.Any(), db.ListAgentRequestsParams{
			Status:      nil,
			EmployeeID:  pgtype.UUID{Bytes: emp1ID, Valid: true},
			QueryLimit:  100,
			QueryOffset: 0,
		}).
		Return(requests, nil)

	// Expect count query with employee filter
	mockDB.EXPECT().
		CountAgentRequests(gomock.Any(), db.CountAgentRequestsParams{
			Status:     nil,
			EmployeeID: pgtype.UUID{Bytes: emp1ID, Valid: true},
		}).
		Return(int64(1), nil)

	handler := handlers.NewAgentRequestsHandler(mockDB)

	// Create request with employee_id query parameter
	req := httptest.NewRequest(http.MethodGet, "/agent-requests?employee_id="+emp1ID.String(), nil)
	rec := httptest.NewRecorder()

	handler.ListAgentRequests(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify filtered results
	assert.Equal(t, float64(1), response["total"])

	requestsArray, ok := response["requests"].([]interface{})
	require.True(t, ok)
	assert.Len(t, requestsArray, 1)
}

// TDD Lesson: Testing list with both status and employee filters
func TestListAgentRequests_Success_WithMultipleFilters(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	req1ID := uuid.New()
	emp1ID := uuid.New()
	statusFilter := "approved"

	requests := []db.AgentRequest{
		{
			ID:          req1ID,
			EmployeeID:  emp1ID,
			RequestType: "agent_access",
			RequestData: []byte(`{"agent_id": "claude-code"}`),
			Status:      "approved",
			Reason:      strPtr("Need for project"),
			CreatedAt:   pgtype.Timestamp{Valid: true},
			ResolvedAt:  pgtype.Timestamp{Valid: true},
		},
	}

	// Expect list query with both filters
	mockDB.EXPECT().
		ListAgentRequests(gomock.Any(), db.ListAgentRequestsParams{
			Status:      &statusFilter,
			EmployeeID:  pgtype.UUID{Bytes: emp1ID, Valid: true},
			QueryLimit:  100,
			QueryOffset: 0,
		}).
		Return(requests, nil)

	// Expect count query with both filters
	mockDB.EXPECT().
		CountAgentRequests(gomock.Any(), db.CountAgentRequestsParams{
			Status:     &statusFilter,
			EmployeeID: pgtype.UUID{Bytes: emp1ID, Valid: true},
		}).
		Return(int64(1), nil)

	handler := handlers.NewAgentRequestsHandler(mockDB)

	// Create request with multiple query parameters
	req := httptest.NewRequest(http.MethodGet, "/agent-requests?status=approved&employee_id="+emp1ID.String(), nil)
	rec := httptest.NewRecorder()

	handler.ListAgentRequests(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify filtered results
	assert.Equal(t, float64(1), response["total"])
}

// TDD Lesson: Testing list with invalid employee_id (gracefully ignored)
func TestListAgentRequests_Success_WithInvalidEmployeeID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	requests := []db.AgentRequest{}

	// Expect list query without employee filter (invalid UUID ignored)
	mockDB.EXPECT().
		ListAgentRequests(gomock.Any(), db.ListAgentRequestsParams{
			Status:      nil,
			EmployeeID:  pgtype.UUID{Valid: false}, // Invalid UUID ignored
			QueryLimit:  100,
			QueryOffset: 0,
		}).
		Return(requests, nil)

	// Expect count query
	mockDB.EXPECT().
		CountAgentRequests(gomock.Any(), db.CountAgentRequestsParams{
			Status:     nil,
			EmployeeID: pgtype.UUID{Valid: false},
		}).
		Return(int64(0), nil)

	handler := handlers.NewAgentRequestsHandler(mockDB)

	// Create request with invalid employee_id
	req := httptest.NewRequest(http.MethodGet, "/agent-requests?employee_id=invalid-uuid", nil)
	rec := httptest.NewRecorder()

	handler.ListAgentRequests(rec, req)

	// Assert HTTP 200 OK (graceful handling)
	assert.Equal(t, http.StatusOK, rec.Code)
}

// TDD Lesson: Testing empty result set
func TestListAgentRequests_Success_EmptyResults(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Return empty slice
	requests := []db.AgentRequest{}

	mockDB.EXPECT().
		ListAgentRequests(gomock.Any(), gomock.Any()).
		Return(requests, nil)

	mockDB.EXPECT().
		CountAgentRequests(gomock.Any(), gomock.Any()).
		Return(int64(0), nil)

	handler := handlers.NewAgentRequestsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/agent-requests", nil)
	rec := httptest.NewRecorder()

	handler.ListAgentRequests(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify empty results
	assert.Equal(t, float64(0), response["total"])

	requestsArray, ok := response["requests"].([]interface{})
	require.True(t, ok)
	assert.Len(t, requestsArray, 0)
}

// TDD Lesson: Testing database error on list
func TestListAgentRequests_DatabaseError_OnList(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Simulate database error on list
	mockDB.EXPECT().
		ListAgentRequests(gomock.Any(), gomock.Any()).
		Return(nil, assert.AnError)

	handler := handlers.NewAgentRequestsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/agent-requests", nil)
	rec := httptest.NewRecorder()

	handler.ListAgentRequests(rec, req)

	// Assert HTTP 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	// Parse error response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Failed to list requests", response["error"])
}

// TDD Lesson: Testing fallback count on count error
func TestListAgentRequests_Success_CountErrorFallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	requests := []db.AgentRequest{
		{
			ID:          uuid.New(),
			EmployeeID:  uuid.New(),
			RequestType: "agent_access",
			RequestData: []byte(`{}`),
			Status:      "pending",
			Reason:      strPtr("Test"),
			CreatedAt:   pgtype.Timestamp{Valid: true},
		},
	}

	mockDB.EXPECT().
		ListAgentRequests(gomock.Any(), gomock.Any()).
		Return(requests, nil)

	// Count query fails
	mockDB.EXPECT().
		CountAgentRequests(gomock.Any(), gomock.Any()).
		Return(int64(0), assert.AnError)

	handler := handlers.NewAgentRequestsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/agent-requests", nil)
	rec := httptest.NewRecorder()

	handler.ListAgentRequests(rec, req)

	// Assert HTTP 200 OK (graceful fallback)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify fallback count (length of requests array)
	assert.Equal(t, float64(1), response["total"])
}
