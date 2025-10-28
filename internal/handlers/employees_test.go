package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"github.com/sergeirastrigin/ubik-enterprise/internal/handlers"
)

// ============================================================================
// ListEmployees Tests
// ============================================================================

// TDD Lesson: Testing employee list endpoint with org isolation
func TestListEmployees_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	emp1ID := uuid.New()
	emp2ID := uuid.New()
	roleID := uuid.New()

	// Create test employees
	employees := []db.Employee{
		{
			ID:           emp1ID,
			OrgID:        orgID,
			Email:        "alice@example.com",
			FullName:     "Alice Smith",
			RoleID:       roleID,
			Status:       "active",
			TeamID:       pgtype.UUID{},
			PasswordHash: "hash1",
			Preferences:  []byte("{}"),
			CreatedAt:    pgtype.Timestamp{Valid: true},
			UpdatedAt:    pgtype.Timestamp{Valid: true},
		},
		{
			ID:           emp2ID,
			OrgID:        orgID,
			Email:        "bob@example.com",
			FullName:     "Bob Jones",
			RoleID:       roleID,
			Status:       "active",
			TeamID:       pgtype.UUID{},
			PasswordHash: "hash2",
			Preferences:  []byte("{}"),
			CreatedAt:    pgtype.Timestamp{Valid: true},
			UpdatedAt:    pgtype.Timestamp{Valid: true},
		},
	}

	// Expect database query with org_id, no filters, default pagination
	mockDB.EXPECT().
		ListEmployees(gomock.Any(), db.ListEmployeesParams{
			OrgID:       orgID,
			Status:      nil,                                        // No status filter (*string = nil)
			TeamID:      pgtype.UUID{Valid: false},                  // No team filter
			QueryLimit:  50,                                         // Default limit
			QueryOffset: 0,                                          // Default offset
		}).
		Return(employees, nil)

	// Expect count query
	mockDB.EXPECT().
		CountEmployees(gomock.Any(), db.CountEmployeesParams{
			OrgID:  orgID,
			Status: nil,
			TeamID: pgtype.UUID{Valid: false},
		}).
		Return(int64(2), nil)

	handler := handlers.NewEmployeesHandler(mockDB)

	// Create request with org_id in context (set by middleware)
	req := httptest.NewRequest(http.MethodGet, "/employees", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListEmployees(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListEmployeesResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify employees returned
	require.Len(t, response.Employees, 2)
	assert.Equal(t, "alice@example.com", string(response.Employees[0].Email))
	assert.Equal(t, "bob@example.com", string(response.Employees[1].Email))

	// Verify pagination info
	assert.Equal(t, int64(2), response.Total)
	assert.Equal(t, 50, response.Limit)
	assert.Equal(t, 0, response.Offset)
}

// TDD Lesson: Test filtering by status
func TestListEmployees_FilterByStatus(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	empID := uuid.New()
	roleID := uuid.New()

	// Only active employees
	employees := []db.Employee{
		{
			ID:           empID,
			OrgID:        orgID,
			Email:        "active@example.com",
			FullName:     "Active User",
			RoleID:       roleID,
			Status:       "active",
			TeamID:       pgtype.UUID{},
			PasswordHash: "hash",
			Preferences:  []byte("{}"),
			CreatedAt:    pgtype.Timestamp{Valid: true},
			UpdatedAt:    pgtype.Timestamp{Valid: true},
		},
	}

	// Status filter value
	activeStatus := "active"

	// Expect database query WITH status filter
	mockDB.EXPECT().
		ListEmployees(gomock.Any(), db.ListEmployeesParams{
			OrgID:       orgID,
			Status:      &activeStatus,                              // Status filter applied (*string)
			TeamID:      pgtype.UUID{Valid: false},
			QueryLimit:  50,
			QueryOffset: 0,
		}).
		Return(employees, nil)

	mockDB.EXPECT().
		CountEmployees(gomock.Any(), db.CountEmployeesParams{
			OrgID:  orgID,
			Status: &activeStatus,
			TeamID: pgtype.UUID{Valid: false},
		}).
		Return(int64(1), nil)

	handler := handlers.NewEmployeesHandler(mockDB)

	// Request with status query parameter
	req := httptest.NewRequest(http.MethodGet, "/employees?status=active", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListEmployees(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListEmployeesResponse
	json.Unmarshal(rec.Body.Bytes(), &response)

	require.Len(t, response.Employees, 1)
	assert.Equal(t, "active", string(response.Employees[0].Status))
}

// TDD Lesson: Test pagination with limit and offset
func TestListEmployees_Pagination(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()

	// Expect database query with custom limit and offset
	mockDB.EXPECT().
		ListEmployees(gomock.Any(), db.ListEmployeesParams{
			OrgID:       orgID,
			Status:      nil,
			TeamID:      pgtype.UUID{Valid: false},
			QueryLimit:  10,  // Custom limit
			QueryOffset: 20,  // Custom offset
		}).
		Return([]db.Employee{}, nil)

	mockDB.EXPECT().
		CountEmployees(gomock.Any(), gomock.Any()).
		Return(int64(100), nil)

	handler := handlers.NewEmployeesHandler(mockDB)

	// Request with pagination parameters
	req := httptest.NewRequest(http.MethodGet, "/employees?limit=10&offset=20", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListEmployees(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListEmployeesResponse
	json.Unmarshal(rec.Body.Bytes(), &response)

	assert.Equal(t, 10, response.Limit)
	assert.Equal(t, 20, response.Offset)
	assert.Equal(t, int64(100), response.Total)
}

// TDD Lesson: Test empty result (no employees found)
func TestListEmployees_EmptyResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()

	// Expect database query returning empty list
	mockDB.EXPECT().
		ListEmployees(gomock.Any(), gomock.Any()).
		Return([]db.Employee{}, nil)

	mockDB.EXPECT().
		CountEmployees(gomock.Any(), gomock.Any()).
		Return(int64(0), nil)

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/employees", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.ListEmployees(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListEmployeesResponse
	json.Unmarshal(rec.Body.Bytes(), &response)

	// Should return empty array, not null
	require.NotNil(t, response.Employees)
	assert.Len(t, response.Employees, 0)
	assert.Equal(t, int64(0), response.Total)
}

// ============================================================================
// GetEmployee Tests
// ============================================================================

// TDD Lesson: Testing employee retrieval by ID with org isolation
func TestGetEmployee_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	empID := uuid.New()
	roleID := uuid.New()

	// Employee to return
	employee := db.Employee{
		ID:           empID,
		OrgID:        orgID,
		Email:        "alice@example.com",
		FullName:     "Alice Smith",
		RoleID:       roleID,
		Status:       "active",
		TeamID:       pgtype.UUID{},
		PasswordHash: "hash1",
		Preferences:  []byte(`{"theme":"dark"}`),
		CreatedAt:    pgtype.Timestamp{Valid: true},
		UpdatedAt:    pgtype.Timestamp{Valid: true},
		LastLoginAt:  pgtype.Timestamp{Valid: true},
	}

	// Expect database query for single employee
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(employee, nil)

	handler := handlers.NewEmployeesHandler(mockDB)

	// Create request with employee_id in URL and org_id in context
	req := httptest.NewRequest(http.MethodGet, "/employees/"+empID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	// Use chi router to extract URL param
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", empID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.GetEmployee(rec, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Employee
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify employee data
	assert.Equal(t, empID, *response.Id)
	assert.Equal(t, "alice@example.com", string(response.Email))
	assert.Equal(t, "Alice Smith", response.FullName)
	assert.Equal(t, api.EmployeeStatusActive, response.Status)
}

// TDD Lesson: Test 404 when employee not found
func TestGetEmployee_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	empID := uuid.New()

	// Expect database query to return error (not found)
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(db.Employee{}, pgx.ErrNoRows)

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/employees/"+empID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", empID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.GetEmployee(rec, req)

	// Verify 404 response
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response api.Error
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "not found")
}

// TDD Lesson: Test org isolation - employee from different org returns 404
func TestGetEmployee_WrongOrg(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	otherOrgID := uuid.New() // Different org
	empID := uuid.New()
	roleID := uuid.New()

	// Employee belongs to different org
	employee := db.Employee{
		ID:           empID,
		OrgID:        otherOrgID, // Wrong org!
		Email:        "bob@other-org.com",
		FullName:     "Bob Other",
		RoleID:       roleID,
		Status:       "active",
		TeamID:       pgtype.UUID{},
		PasswordHash: "hash",
		Preferences:  []byte("{}"),
		CreatedAt:    pgtype.Timestamp{Valid: true},
		UpdatedAt:    pgtype.Timestamp{Valid: true},
	}

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(employee, nil)

	handler := handlers.NewEmployeesHandler(mockDB)

	// Request from orgID, but employee belongs to otherOrgID
	req := httptest.NewRequest(http.MethodGet, "/employees/"+empID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", empID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.GetEmployee(rec, req)

	// Should return 404 (not 403) for security - don't reveal employee exists
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// TDD Lesson: Test invalid UUID format
func TestGetEmployee_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/employees/invalid-uuid", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.GetEmployee(rec, req)

	// Should return 400 for invalid UUID
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
