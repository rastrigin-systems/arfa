package handlers_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/handlers"
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
	employees := []db.ListEmployeesRow{
		{
			ID:           emp1ID,
			OrgID:        orgID,
			Email:        "alice@example.com",
			FullName:     "Alice Smith",
			RoleID:       roleID,
			Status:       "active",
			TeamID:       pgtype.UUID{},
			PasswordHash: "hash1",
			Preferences:  json.RawMessage("{}"),
			CreatedAt:    pgtype.Timestamp{Valid: true},
			UpdatedAt:    pgtype.Timestamp{Valid: true},
			TeamName:     nil, // No team assigned
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
			Preferences:  json.RawMessage("{}"),
			CreatedAt:    pgtype.Timestamp{Valid: true},
			UpdatedAt:    pgtype.Timestamp{Valid: true},
			TeamName:     nil, // No team assigned
		},
	}

	// Expect database query with org_id, no filters, default pagination
	mockDB.EXPECT().
		ListEmployees(gomock.Any(), db.ListEmployeesParams{
			OrgID:       orgID,
			Status:      nil,                       // No status filter (*string = nil)
			TeamID:      pgtype.UUID{Valid: false}, // No team filter
			QueryLimit:  50,                        // Default limit
			QueryOffset: 0,                         // Default offset
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
	employees := []db.ListEmployeesRow{
		{
			ID:           empID,
			OrgID:        orgID,
			Email:        "active@example.com",
			FullName:     "Active User",
			RoleID:       roleID,
			Status:       "active",
			TeamID:       pgtype.UUID{},
			PasswordHash: "hash",
			Preferences:  json.RawMessage("{}"),
			CreatedAt:    pgtype.Timestamp{Valid: true},
			UpdatedAt:    pgtype.Timestamp{Valid: true},
			TeamName:     nil,
		},
	}

	// Status filter value
	activeStatus := "active"

	// Expect database query WITH status filter
	mockDB.EXPECT().
		ListEmployees(gomock.Any(), db.ListEmployeesParams{
			OrgID:       orgID,
			Status:      &activeStatus, // Status filter applied (*string)
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
			QueryLimit:  10, // Custom limit
			QueryOffset: 20, // Custom offset
		}).
		Return([]db.ListEmployeesRow{}, nil)

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
		Return([]db.ListEmployeesRow{}, nil)

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
	employee := db.GetEmployeeRow{
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
		Return(db.GetEmployeeRow{}, pgx.ErrNoRows)

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
	employee := db.GetEmployeeRow{
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

// ============================================================================
// CreateEmployee Tests
// ============================================================================

// TDD Lesson: Testing employee creation with required fields only
func TestCreateEmployee_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	roleID := uuid.New()

	// Request payload with required fields only
	reqBody := `{
		"email": "newuser@example.com",
		"full_name": "New User",
		"role_id": "` + roleID.String() + `"
	}`

	// Expect CreateEmployee to be called
	// Note: password_hash will be generated, so we use gomock.Any()
	mockDB.EXPECT().
		CreateEmployee(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateEmployeeParams) (db.Employee, error) {
			// Verify params are correct
			assert.Equal(t, orgID, params.OrgID)
			assert.Equal(t, roleID, params.RoleID)
			assert.Equal(t, "newuser@example.com", params.Email)
			assert.Equal(t, "New User", params.FullName)
			assert.Equal(t, "active", params.Status)
			assert.False(t, params.TeamID.Valid) // No team
			assert.NotEmpty(t, params.PasswordHash)

			// Return created employee
			return db.Employee{
				ID:           uuid.New(),
				OrgID:        params.OrgID,
				Email:        params.Email,
				FullName:     params.FullName,
				RoleID:       params.RoleID,
				Status:       params.Status,
				TeamID:       params.TeamID,
				PasswordHash: params.PasswordHash,
				Preferences:  []byte("{}"),
				CreatedAt:    pgtype.Timestamp{Valid: true},
				UpdatedAt:    pgtype.Timestamp{Valid: true},
			}, nil
		})

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.CreateEmployee(rec, req)

	// Verify response
	assert.Equal(t, http.StatusCreated, rec.Code)

	var response api.CreateEmployeeResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify employee details
	assert.Equal(t, "newuser@example.com", string(response.Employee.Email))
	assert.Equal(t, "New User", response.Employee.FullName)
	assert.Equal(t, roleID, response.Employee.RoleId)
	assert.Equal(t, api.EmployeeStatusActive, response.Employee.Status)

	// Verify temporary password is returned
	assert.NotEmpty(t, response.TemporaryPassword)
	assert.GreaterOrEqual(t, len(response.TemporaryPassword), 16, "Temporary password should be at least 16 characters")
}

// TDD Lesson: Test creating employee with team_id
func TestCreateEmployee_WithTeam(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	roleID := uuid.New()
	teamID := uuid.New()

	// Request with team_id
	reqBody := `{
		"email": "teamuser@example.com",
		"full_name": "Team User",
		"role_id": "` + roleID.String() + `",
		"team_id": "` + teamID.String() + `"
	}`

	mockDB.EXPECT().
		CreateEmployee(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateEmployeeParams) (db.Employee, error) {
			// Verify team_id is set
			assert.True(t, params.TeamID.Valid)
			assert.Equal(t, teamID[:], params.TeamID.Bytes[:])

			return db.Employee{
				ID:           uuid.New(),
				OrgID:        params.OrgID,
				Email:        params.Email,
				FullName:     params.FullName,
				RoleID:       params.RoleID,
				Status:       params.Status,
				TeamID:       params.TeamID,
				PasswordHash: params.PasswordHash,
				Preferences:  []byte("{}"),
				CreatedAt:    pgtype.Timestamp{Valid: true},
				UpdatedAt:    pgtype.Timestamp{Valid: true},
			}, nil
		})

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.CreateEmployee(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)
}

// TDD Lesson: Test duplicate email returns 409 Conflict
func TestCreateEmployee_DuplicateEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	roleID := uuid.New()

	reqBody := `{
		"email": "existing@example.com",
		"full_name": "Duplicate User",
		"role_id": "` + roleID.String() + `"
	}`

	// Mock database returning unique constraint violation
	mockDB.EXPECT().
		CreateEmployee(gomock.Any(), gomock.Any()).
		Return(db.Employee{}, &pgconn.PgError{
			Code:           "23505", // unique_violation
			ConstraintName: "employees_email_key",
		})

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.CreateEmployee(rec, req)

	// Should return 409 Conflict
	assert.Equal(t, http.StatusConflict, rec.Code)

	var response api.Error
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "already exists")
}

// TDD Lesson: Test invalid JSON returns 400
func TestCreateEmployee_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()

	handler := handlers.NewEmployeesHandler(mockDB)

	// Invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.CreateEmployee(rec, req)

	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TDD Lesson: Test missing required fields returns 400
func TestCreateEmployee_MissingRequiredFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()

	testCases := []struct {
		name    string
		reqBody string
	}{
		{
			name:    "missing email",
			reqBody: `{"full_name": "User", "role_id": "` + uuid.New().String() + `"}`,
		},
		{
			name:    "missing full_name",
			reqBody: `{"email": "user@example.com", "role_id": "` + uuid.New().String() + `"}`,
		},
		{
			name:    "missing role_id",
			reqBody: `{"email": "user@example.com", "full_name": "User"}`,
		},
		{
			name:    "empty email",
			reqBody: `{"email": "", "full_name": "User", "role_id": "` + uuid.New().String() + `"}`,
		},
		{
			name:    "empty full_name",
			reqBody: `{"email": "user@example.com", "full_name": "", "role_id": "` + uuid.New().String() + `"}`,
		},
	}

	handler := handlers.NewEmployeesHandler(mockDB)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader(tc.reqBody))
			req.Header.Set("Content-Type", "application/json")
			req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
			rec := httptest.NewRecorder()

			handler.CreateEmployee(rec, req)

			// Should return 400 or 422
			assert.True(t, rec.Code == http.StatusBadRequest || rec.Code == http.StatusUnprocessableEntity,
				"Expected 400 or 422, got %d for %s", rec.Code, tc.name)
		})
	}
}

// TDD Lesson: Test invalid email format returns 400
func TestCreateEmployee_InvalidEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	roleID := uuid.New()

	reqBody := `{
		"email": "not-an-email",
		"full_name": "Invalid Email User",
		"role_id": "` + roleID.String() + `"
	}`

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))
	rec := httptest.NewRecorder()

	handler.CreateEmployee(rec, req)

	// Should return 400 or 422 for invalid email format
	assert.True(t, rec.Code == http.StatusBadRequest || rec.Code == http.StatusUnprocessableEntity)
}

// ============================================================================
// UpdateEmployee Tests
// ============================================================================

// TDD Lesson: Testing employee update with partial fields
func TestUpdateEmployee_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	empID := uuid.New()
	roleID := uuid.New()

	// Request to update full_name
	newName := "Updated Name"
	reqBody := `{"full_name": "` + newName + `"}`

	// First, expect GetEmployee to verify org isolation
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(db.GetEmployeeRow{
			ID:       empID,
			OrgID:    orgID,
			Email:    "user@example.com",
			FullName: "Old Name",
			RoleID:   roleID,
			Status:   "active",
		}, nil)

	// Then expect UpdateEmployee
	mockDB.EXPECT().
		UpdateEmployee(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.UpdateEmployeeParams) (db.Employee, error) {
			assert.Equal(t, empID, params.ID)
			assert.Equal(t, newName, params.FullName)

			return db.Employee{
				ID:          empID,
				OrgID:       orgID,
				Email:       "user@example.com",
				FullName:    newName,
				RoleID:      roleID,
				Status:      "active",
				Preferences: []byte("{}"),
				CreatedAt:   pgtype.Timestamp{Valid: true},
				UpdatedAt:   pgtype.Timestamp{Valid: true},
			}, nil
		})

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodPatch, "/employees/"+empID.String(), strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", empID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.UpdateEmployee(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Employee
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, newName, response.FullName)
}

// TDD Lesson: Test updating multiple fields including status
func TestUpdateEmployee_MultipleFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	empID := uuid.New()
	oldRoleID := uuid.New()
	newRoleID := uuid.New()
	teamID := uuid.New()

	reqBody := `{
		"full_name": "New Name",
		"role_id": "` + newRoleID.String() + `",
		"team_id": "` + teamID.String() + `",
		"status": "suspended"
	}`

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(db.GetEmployeeRow{
			ID:       empID,
			OrgID:    orgID,
			Email:    "user@example.com",
			FullName: "Old Name",
			RoleID:   oldRoleID,
			Status:   "active",
		}, nil)

	mockDB.EXPECT().
		UpdateEmployee(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.UpdateEmployeeParams) (db.Employee, error) {
			assert.Equal(t, empID, params.ID)
			assert.Equal(t, "New Name", params.FullName)
			assert.Equal(t, newRoleID, params.RoleID)
			assert.Equal(t, "suspended", params.Status)
			assert.True(t, params.TeamID.Valid)
			assert.Equal(t, teamID[:], params.TeamID.Bytes[:])

			return db.Employee{
				ID:          empID,
				OrgID:       orgID,
				Email:       "user@example.com",
				FullName:    "New Name",
				RoleID:      newRoleID,
				Status:      "suspended",
				TeamID:      params.TeamID,
				Preferences: []byte("{}"),
				CreatedAt:   pgtype.Timestamp{Valid: true},
				UpdatedAt:   pgtype.Timestamp{Valid: true},
			}, nil
		})

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodPatch, "/employees/"+empID.String(), strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", empID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.UpdateEmployee(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}

// TDD Lesson: Test employee not found returns 404
func TestUpdateEmployee_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	empID := uuid.New()

	reqBody := `{"full_name": "New Name"}`

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(db.GetEmployeeRow{}, pgx.ErrNoRows)

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodPatch, "/employees/"+empID.String(), strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", empID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.UpdateEmployee(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// TDD Lesson: Test org isolation - can't update employee from different org
func TestUpdateEmployee_WrongOrg(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	otherOrgID := uuid.New()
	empID := uuid.New()
	roleID := uuid.New()

	reqBody := `{"full_name": "New Name"}`

	// Employee belongs to different org
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(db.GetEmployeeRow{
			ID:       empID,
			OrgID:    otherOrgID, // Different org!
			Email:    "user@other.com",
			FullName: "Old Name",
			RoleID:   roleID,
			Status:   "active",
		}, nil)

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodPatch, "/employees/"+empID.String(), strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", empID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.UpdateEmployee(rec, req)

	// Should return 404 for security
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// TDD Lesson: Test invalid UUID format
func TestUpdateEmployee_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()

	reqBody := `{"full_name": "New Name"}`

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodPatch, "/employees/invalid-uuid", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.UpdateEmployee(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// ============================================================================
// DeleteEmployee Tests
// ============================================================================

// TDD Lesson: Testing soft delete (sets deleted_at)
func TestDeleteEmployee_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	empID := uuid.New()
	roleID := uuid.New()

	// First, verify org isolation
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(db.GetEmployeeRow{
			ID:       empID,
			OrgID:    orgID,
			Email:    "user@example.com",
			FullName: "User Name",
			RoleID:   roleID,
			Status:   "active",
		}, nil)

	// Then expect hard delete
	mockDB.EXPECT().
		DeleteEmployee(gomock.Any(), empID).
		Return(nil)

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodDelete, "/employees/"+empID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", empID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.DeleteEmployee(rec, req)

	// Should return 204 No Content
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

// TDD Lesson: Test deleting non-existent employee returns 404
func TestDeleteEmployee_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	empID := uuid.New()

	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(db.GetEmployeeRow{}, pgx.ErrNoRows)

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodDelete, "/employees/"+empID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", empID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.DeleteEmployee(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// TDD Lesson: Test org isolation for delete
func TestDeleteEmployee_WrongOrg(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()
	otherOrgID := uuid.New()
	empID := uuid.New()
	roleID := uuid.New()

	// Employee belongs to different org
	mockDB.EXPECT().
		GetEmployee(gomock.Any(), empID).
		Return(db.GetEmployeeRow{
			ID:       empID,
			OrgID:    otherOrgID, // Different org!
			Email:    "user@other.com",
			FullName: "User Name",
			RoleID:   roleID,
			Status:   "active",
		}, nil)

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodDelete, "/employees/"+empID.String(), nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", empID.String())
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.DeleteEmployee(rec, req)

	// Should return 404 for security
	assert.Equal(t, http.StatusNotFound, rec.Code)
}

// TDD Lesson: Test invalid UUID for delete
func TestDeleteEmployee_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgID := uuid.New()

	handler := handlers.NewEmployeesHandler(mockDB)

	req := httptest.NewRequest(http.MethodDelete, "/employees/invalid-uuid", nil)
	req = req.WithContext(handlers.SetOrgIDInContext(req.Context(), orgID))

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("employee_id", "invalid-uuid")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()

	handler.DeleteEmployee(rec, req)

	assert.Equal(t, http.StatusBadRequest, rec.Code)
}
