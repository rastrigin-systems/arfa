package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/handlers"
)

// ============================================================================
// GetEmployeeUsageStats Tests
// ============================================================================

// TDD Lesson: Testing employee usage stats retrieval
func TestGetEmployeeUsageStats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	empID := uuid.New()

	// Mock employee lookup (verify org ownership)
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(db.GetEmployeeRow{
			ID:       empID,
			OrgID:    orgID,
			FullName: "Alice Smith",
		}, nil)

	// Mock usage stats query
	mockDB.EXPECT().
		GetEmployeeUsageStats(gomock.Any(), gomock.Any()).
		Return(db.GetEmployeeUsageStatsRow{
			TotalRecords:  100,
			TotalApiCalls: 1500,
			TotalTokens:   50000,
			TotalCostUsd:  mustParseNumeric(t, "25.50"),
		}, nil)

	handler := handlers.NewUsageStatsHandler(mockDB)

	// Create request with employee_id URL param
	req := httptest.NewRequest(http.MethodGet, "/employees/"+empID.String()+"/usage", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	// Add URL params via chi context
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", empID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.GetEmployeeUsageStats(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, empID.String(), response["employee_id"])
	assert.Equal(t, "Alice Smith", response["employee_name"])
	assert.NotNil(t, response["period_start"])
	assert.NotNil(t, response["period_end"])
	assert.Equal(t, float64(1500), response["total_api_calls"])
	assert.Equal(t, float64(50000), response["total_tokens"])
	assert.NotNil(t, response["total_cost_usd"])
}

// TDD Lesson: Testing unauthorized access (no org_id)
func TestGetEmployeeUsageStats_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewUsageStatsHandler(mockDB)

	empID := uuid.New()

	// Create request WITHOUT org_id in context
	req := httptest.NewRequest(http.MethodGet, "/employees/"+empID.String()+"/usage", nil)

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", empID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.GetEmployeeUsageStats(rec, req)

	// Assert HTTP 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Unauthorized", response["error"])
}

// TDD Lesson: Testing invalid employee ID format
func TestGetEmployeeUsageStats_InvalidEmployeeID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	handler := handlers.NewUsageStatsHandler(mockDB)

	// Create request with invalid UUID
	req := httptest.NewRequest(http.MethodGet, "/employees/invalid-uuid/usage", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.GetEmployeeUsageStats(rec, req)

	// Assert HTTP 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TDD Lesson: Testing employee not found
func TestGetEmployeeUsageStats_EmployeeNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	empID := uuid.New()

	// Employee not found
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(db.GetEmployeeRow{}, assert.AnError)

	handler := handlers.NewUsageStatsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/employees/"+empID.String()+"/usage", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", empID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.GetEmployeeUsageStats(rec, req)

	// Assert HTTP 404 Not Found
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Employee not found", response["error"])
}

// TDD Lesson: Testing employee belongs to different org (forbidden)
func TestGetEmployeeUsageStats_WrongOrg(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	otherOrgID := uuid.New()
	empID := uuid.New()

	// Employee belongs to different org
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(db.GetEmployeeRow{
			ID:       empID,
			OrgID:    otherOrgID, // Different org!
			FullName: "Bob Jones",
		}, nil)

	handler := handlers.NewUsageStatsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/employees/"+empID.String()+"/usage", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", empID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.GetEmployeeUsageStats(rec, req)

	// Assert HTTP 403 Forbidden
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Forbidden", response["error"])
}

// TDD Lesson: Testing database error on stats query
func TestGetEmployeeUsageStats_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()
	empID := uuid.New()

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(db.GetEmployeeRow{
			ID:       empID,
			OrgID:    orgID,
			FullName: "Alice Smith",
		}, nil)

	// Stats query fails
	mockDB.EXPECT().
		GetEmployeeUsageStats(gomock.Any(), gomock.Any()).
		Return(db.GetEmployeeUsageStatsRow{}, assert.AnError)

	handler := handlers.NewUsageStatsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/employees/"+empID.String()+"/usage", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", empID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.GetEmployeeUsageStats(rec, req)

	// Assert HTTP 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Failed to get usage stats", response["error"])
}

// ============================================================================
// GetOrgUsageStats Tests
// ============================================================================

// TDD Lesson: Testing organization usage stats retrieval
func TestGetOrgUsageStats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()

	// Mock usage stats query for entire org
	mockDB.EXPECT().
		GetOrgUsageStats(gomock.Any(), gomock.Any()).
		Return(db.GetOrgUsageStatsRow{
			TotalRecords:  500,
			TotalApiCalls: 10000,
			TotalTokens:   250000,
			TotalCostUsd:  mustParseNumeric(t, "150.75"),
		}, nil)

	handler := handlers.NewUsageStatsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/usage/org", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.GetOrgUsageStats(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, orgID.String(), response["org_id"])
	assert.NotNil(t, response["period_start"])
	assert.NotNil(t, response["period_end"])
	assert.Equal(t, float64(10000), response["total_api_calls"])
	assert.Equal(t, float64(250000), response["total_tokens"])
	assert.NotNil(t, response["total_cost_usd"])
}

// TDD Lesson: Testing org stats unauthorized
func TestGetOrgUsageStats_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewUsageStatsHandler(mockDB)

	// Create request WITHOUT org_id
	req := httptest.NewRequest(http.MethodGet, "/usage/org", nil)
	rec := httptest.NewRecorder()

	handler.GetOrgUsageStats(rec, req)

	// Assert HTTP 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TDD Lesson: Testing org stats database error
func TestGetOrgUsageStats_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	orgID := uuid.New()

	// Stats query fails
	mockDB.EXPECT().
		GetOrgUsageStats(gomock.Any(), gomock.Any()).
		Return(db.GetOrgUsageStatsRow{}, assert.AnError)

	handler := handlers.NewUsageStatsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/usage/org", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.GetOrgUsageStats(rec, req)

	// Assert HTTP 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Failed to get usage stats", response["error"])
}

// ============================================================================
// GetCurrentEmployeeUsageStats Tests (authenticated employee)
// ============================================================================

// TDD Lesson: Testing current employee usage stats retrieval
func TestGetCurrentEmployeeUsageStats_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	empID := uuid.New()

	// Mock usage stats query
	mockDB.EXPECT().
		GetEmployeeUsageStats(gomock.Any(), gomock.Any()).
		Return(db.GetEmployeeUsageStatsRow{
			TotalRecords:  50,
			TotalApiCalls: 800,
			TotalTokens:   30000,
			TotalCostUsd:  mustParseNumeric(t, "12.50"),
		}, nil)

	handler := handlers.NewUsageStatsHandler(mockDB)

	// Mock employee_id in context (set by auth middleware)
	req := httptest.NewRequest(http.MethodGet, "/usage/me", nil)
	req = req.WithContext(handlers.SetEmployeeIDInContext(req.Context(), empID))
	rec := httptest.NewRecorder()

	handler.GetCurrentEmployeeUsageStats(rec, req)

	// Assert HTTP 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	// Parse response
	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	// Verify response structure
	assert.Equal(t, empID.String(), response["employee_id"])
	assert.NotNil(t, response["period_start"])
	assert.NotNil(t, response["period_end"])
	assert.Equal(t, float64(800), response["total_api_calls"])
	assert.Equal(t, float64(30000), response["total_tokens"])
}

// TDD Lesson: Testing current employee stats unauthorized
func TestGetCurrentEmployeeUsageStats_Unauthorized(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewUsageStatsHandler(mockDB)

	// Create request WITHOUT employee_id in context
	req := httptest.NewRequest(http.MethodGet, "/usage/me", nil)
	rec := httptest.NewRecorder()

	handler.GetCurrentEmployeeUsageStats(rec, req)

	// Assert HTTP 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TDD Lesson: Testing current employee stats database error
func TestGetCurrentEmployeeUsageStats_DatabaseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	empID := uuid.New()

	// Stats query fails
	mockDB.EXPECT().
		GetEmployeeUsageStats(gomock.Any(), gomock.Any()).
		Return(db.GetEmployeeUsageStatsRow{}, assert.AnError)

	handler := handlers.NewUsageStatsHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/usage/me", nil)
	req = req.WithContext(handlers.SetEmployeeIDInContext(req.Context(), empID))
	rec := httptest.NewRecorder()

	handler.GetCurrentEmployeeUsageStats(rec, req)

	// Assert HTTP 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, rec.Code)

	var response map[string]interface{}
	err := json.NewDecoder(rec.Body).Decode(&response)
	require.NoError(t, err)

	assert.Equal(t, "Failed to get usage stats", response["error"])
}

// ============================================================================
// Helper Functions
// ============================================================================

// Helper to parse numeric values for testing
func mustParseNumeric(t *testing.T, s string) pgtype.Numeric {
	// For testing purposes, create a simple numeric value
	// In real code, pgtype.Numeric handles arbitrary precision decimals
	var result pgtype.Numeric
	err := result.Scan(s)
	if err != nil {
		t.Logf("Warning: Failed to parse numeric value %s: %v", s, err)
		// Return a valid numeric with zero value
		result.Valid = true
	}
	return result
}
