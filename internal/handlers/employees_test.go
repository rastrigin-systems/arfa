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
