package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/rastrigin-systems/arfa/generated/api"
	"github.com/rastrigin-systems/arfa/services/api/internal/auth"
	"github.com/rastrigin-systems/arfa/services/api/internal/handlers"
	"github.com/rastrigin-systems/arfa/services/api/internal/middleware"
	"github.com/rastrigin-systems/arfa/services/api/tests/testutil"
)

// ============================================================================
// List Employees Integration Tests
// ============================================================================

// TDD Lesson: Integration test with real database
// Tests complete flow: org isolation, pagination, and filtering
func TestListEmployees_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)

	// Create test role
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create multiple employees in the same org
	employee1 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "alice@example.com",
		FullName: "Alice Smith",
		Status:   "active",
	})

	employee2 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "bob@example.com",
		FullName: "Bob Jones",
		Status:   "active",
	})

	employee3 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "charlie@example.com",
		FullName: "Charlie Brown",
		Status:   "suspended",
	})

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee1.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee1.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewEmployeesHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees", handler.ListEmployees)

	// Test 1: List all employees (no filter)
	req := httptest.NewRequest(http.MethodGet, "/employees", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListEmployeesResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should return all 3 employees
	assert.Equal(t, int64(3), response.Total)
	require.Len(t, response.Employees, 3)
	assert.Equal(t, 50, response.Limit)
	assert.Equal(t, 0, response.Offset)

	// Verify employee data
	emails := []string{
		string(response.Employees[0].Email),
		string(response.Employees[1].Email),
		string(response.Employees[2].Email),
	}
	assert.Contains(t, emails, employee1.Email)
	assert.Contains(t, emails, employee2.Email)
	assert.Contains(t, emails, employee3.Email)
}

// TDD Lesson: Test org isolation - employees from different orgs should not be visible
func TestListEmployees_Integration_OrgIsolation(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create two organizations
	org1 := testutil.CreateTestOrg(t, queries, ctx)
	org2 := testutil.CreateTestOrg(t, queries, ctx)

	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create employee in org1
	emp1 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org1.ID,
		RoleID:   role.ID,
		Email:    "alice@org1.com",
		FullName: "Alice Org1",
		Status:   "active",
	})

	// Create employee in org2
	testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org2.ID,
		RoleID:   role.ID,
		Email:    "bob@org2.com",
		FullName: "Bob Org2",
		Status:   "active",
	})

	// Authenticate as employee from org1
	token, _ := auth.GenerateJWT(emp1.ID, org1.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(emp1.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewEmployeesHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees", handler.ListEmployees)

	// List employees
	req := httptest.NewRequest(http.MethodGet, "/employees", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListEmployeesResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should only return 1 employee from org1
	assert.Equal(t, int64(1), response.Total)
	require.Len(t, response.Employees, 1)
	assert.Equal(t, "alice@org1.com", string(response.Employees[0].Email))
	assert.Equal(t, "Alice Org1", response.Employees[0].FullName)
}

// TDD Lesson: Test status filtering
func TestListEmployees_Integration_FilterByStatus(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create employees with different statuses
	activeEmp := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "active@example.com",
		FullName: "Active User",
		Status:   "active",
	})

	testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "suspended@example.com",
		FullName: "Suspended User",
		Status:   "suspended",
	})

	testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "inactive@example.com",
		FullName: "Inactive User",
		Status:   "inactive",
	})

	// Create session
	token, _ := auth.GenerateJWT(activeEmp.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(activeEmp.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler
	handler := handlers.NewEmployeesHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees", handler.ListEmployees)

	// Test filtering by status=active
	req := httptest.NewRequest(http.MethodGet, "/employees?status=active", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListEmployeesResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should only return active employees
	assert.Equal(t, int64(1), response.Total)
	require.Len(t, response.Employees, 1)
	assert.Equal(t, "active@example.com", string(response.Employees[0].Email))
	assert.Equal(t, api.EmployeeStatusActive, response.Employees[0].Status)
}

// TDD Lesson: Test pagination with limit and offset
func TestListEmployees_Integration_Pagination(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create 5 employees
	emp1 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "emp1@example.com",
		FullName: "Employee 1",
		Status:   "active",
	})

	for i := 2; i <= 5; i++ {
		testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
			OrgID:    org.ID,
			RoleID:   role.ID,
			Email:    "emp" + string(rune('0'+i)) + "@example.com",
			FullName: "Employee " + string(rune('0'+i)),
			Status:   "active",
		})
	}

	// Create session
	token, _ := auth.GenerateJWT(emp1.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(emp1.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler
	handler := handlers.NewEmployeesHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees", handler.ListEmployees)

	// Test pagination: limit=2, offset=0 (first page)
	req := httptest.NewRequest(http.MethodGet, "/employees?limit=2&offset=0", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.ListEmployeesResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Should return 2 employees, total 5
	assert.Equal(t, int64(5), response.Total)
	require.Len(t, response.Employees, 2)
	assert.Equal(t, 2, response.Limit)
	assert.Equal(t, 0, response.Offset)

	// Test second page: limit=2, offset=2
	req2 := httptest.NewRequest(http.MethodGet, "/employees?limit=2&offset=2", nil)
	req2.Header.Set("Authorization", "Bearer "+token)
	rec2 := httptest.NewRecorder()

	router.ServeHTTP(rec2, req2)

	assert.Equal(t, http.StatusOK, rec2.Code)

	var response2 api.ListEmployeesResponse
	err = json.Unmarshal(rec2.Body.Bytes(), &response2)
	require.NoError(t, err)

	// Should return 2 different employees
	assert.Equal(t, int64(5), response2.Total)
	require.Len(t, response2.Employees, 2)
	assert.Equal(t, 2, response2.Limit)
	assert.Equal(t, 2, response2.Offset)

	// Verify no overlap between pages
	page1Emails := []string{
		string(response.Employees[0].Email),
		string(response.Employees[1].Email),
	}
	page2Emails := []string{
		string(response2.Employees[0].Email),
		string(response2.Employees[1].Email),
	}

	for _, email := range page2Emails {
		assert.NotContains(t, page1Emails, email, "Pages should not overlap")
	}
}

// ============================================================================
// Get Employee Integration Tests
// ============================================================================

// TDD Lesson: Integration test for GET /employees/{id} with org isolation
func TestGetEmployee_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization and role
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create employee
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "alice@example.com",
		FullName: "Alice Smith",
		Status:   "active",
	})

	// Create session for authentication
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewEmployeesHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees/{employee_id}", handler.GetEmployee)

	// Request employee by ID
	req := httptest.NewRequest(http.MethodGet, "/employees/"+employee.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Employee
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify employee data
	assert.Equal(t, employee.ID, *response.Id)
	assert.Equal(t, "alice@example.com", string(response.Email))
	assert.Equal(t, "Alice Smith", response.FullName)
	assert.Equal(t, api.EmployeeStatusActive, response.Status)
}

// TDD Lesson: Test org isolation - cannot fetch employee from different org
func TestGetEmployee_Integration_OrgIsolation(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create two organizations
	org1 := testutil.CreateTestOrg(t, queries, ctx)
	org2 := testutil.CreateTestOrg(t, queries, ctx)

	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create employee in org1
	emp1 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org1.ID,
		RoleID:   role.ID,
		Email:    "alice@org1.com",
		FullName: "Alice Org1",
		Status:   "active",
	})

	// Create employee in org2
	emp2 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org2.ID,
		RoleID:   role.ID,
		Email:    "bob@org2.com",
		FullName: "Bob Org2",
		Status:   "active",
	})

	// Authenticate as employee from org1
	token, _ := auth.GenerateJWT(emp1.ID, org1.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(emp1.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewEmployeesHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees/{employee_id}", handler.GetEmployee)

	// Try to fetch employee from org2
	req := httptest.NewRequest(http.MethodGet, "/employees/"+emp2.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Should return 404 (not 403) for security - don't reveal employee exists
	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response api.Error
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "not found")
}

// TDD Lesson: Test 404 when employee doesn't exist
func TestGetEmployee_Integration_NotFound(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create employee for authentication
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "alice@example.com",
		FullName: "Alice Smith",
		Status:   "active",
	})

	// Create session
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler
	handler := handlers.NewEmployeesHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Get("/employees/{employee_id}", handler.GetEmployee)

	// Try to fetch non-existent employee
	nonExistentID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/employees/"+nonExistentID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusNotFound, rec.Code)

	var response api.Error
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Contains(t, response.Error, "not found")
}

// ============================================================================
// Create Employee Integration Tests
// ============================================================================

// TDD Lesson: Integration test for POST /employees
func TestCreateEmployee_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test organization and role
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create admin employee for authentication
	admin := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Create session for authentication
	token, _ := auth.GenerateJWT(admin.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(admin.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler with middleware
	handler := handlers.NewEmployeesHandler(queries)

	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Post("/employees", handler.CreateEmployee)

	// Request payload
	reqBody := `{
		"email": "newuser@example.com",
		"full_name": "New Employee",
		"role_id": "` + role.ID.String() + `"
	}`

	// Create employee
	req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	// TDD Fix: Unmarshal into correct response type (CreateEmployeeResponse, not Employee)
	var response api.CreateEmployeeResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify temporary password is returned
	assert.NotEmpty(t, response.TemporaryPassword)
	assert.Greater(t, len(response.TemporaryPassword), 10) // Should be at least 10 chars

	// Verify employee data
	assert.NotNil(t, response.Employee.Id)
	assert.Equal(t, "newuser@example.com", string(response.Employee.Email))
	assert.Equal(t, "New Employee", response.Employee.FullName)
	assert.Equal(t, role.ID, response.Employee.RoleId)
	assert.Equal(t, api.EmployeeStatusActive, response.Employee.Status)
	assert.Equal(t, org.ID, response.Employee.OrgId)

	// Verify employee was created in database
	createdEmployee, err := queries.GetEmployee(ctx, *response.Employee.Id)
	require.NoError(t, err)
	assert.Equal(t, "newuser@example.com", createdEmployee.Email)
	assert.Equal(t, org.ID, createdEmployee.OrgID)
}

// TDD Lesson: Test creating employee with team_id
func TestCreateEmployee_Integration_WithTeam(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create a team
	team := testutil.CreateTestTeam(t, queries, ctx, org.ID, "Engineering")

	// Create admin for authentication
	admin := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Create session
	token, _ := auth.GenerateJWT(admin.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(admin.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler
	handler := handlers.NewEmployeesHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Post("/employees", handler.CreateEmployee)

	// Request with team_id
	reqBody := `{
		"email": "teamuser@example.com",
		"full_name": "Team Member",
		"role_id": "` + role.ID.String() + `",
		"team_id": "` + team.ID.String() + `"
	}`

	req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	// TDD Fix: Unmarshal into correct response type (CreateEmployeeResponse, not Employee)
	var response api.CreateEmployeeResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify temporary password is returned
	assert.NotEmpty(t, response.TemporaryPassword)

	// Verify team_id is set
	assert.NotNil(t, response.Employee.TeamId)
	assert.Equal(t, team.ID, uuid.UUID(*response.Employee.TeamId))

	// Verify in database
	createdEmployee, err := queries.GetEmployee(ctx, *response.Employee.Id)
	require.NoError(t, err)
	assert.True(t, createdEmployee.TeamID.Valid)
	assert.Equal(t, team.ID[:], createdEmployee.TeamID.Bytes[:])
}

// TDD Lesson: Verify CreateEmployeeResponse structure (Issue #3 fix)
// This test documents the fix for the JSON unmarshaling issue where
// CreateEmployee returns CreateEmployeeResponse (with Employee + TemporaryPassword)
// not just Employee directly
func TestCreateEmployee_Integration_ResponseStructure(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create admin for authentication
	admin := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin User",
		Status:   "active",
	})

	// Create session
	token, _ := auth.GenerateJWT(admin.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(admin.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler
	handler := handlers.NewEmployeesHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Post("/employees", handler.CreateEmployee)

	// Create employee
	reqBody := `{
		"email": "newuser@example.com",
		"full_name": "New User",
		"role_id": "` + role.ID.String() + `"
	}`

	req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	// ✅ CRITICAL: Response must be CreateEmployeeResponse, not Employee directly
	var response api.CreateEmployeeResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err, "Response should unmarshal to CreateEmployeeResponse")

	// Verify response structure
	assert.NotEmpty(t, response.TemporaryPassword, "TemporaryPassword should be returned")
	assert.Greater(t, len(response.TemporaryPassword), 10, "TemporaryPassword should be strong")
	assert.NotNil(t, response.Employee.Id, "Employee.Id should be set")
	assert.NotEqual(t, uuid.Nil, response.Employee.RoleId, "Employee.RoleId should be set")
	assert.NotEmpty(t, response.Employee.Status, "Employee.Status should be set")
	assert.Equal(t, api.EmployeeStatusActive, response.Employee.Status)

	// Verify attempting to unmarshal directly into Employee would fail silently
	// (fields would be zero values, causing the bug in Issue #3)
	var wrongResponse api.Employee
	err = json.Unmarshal(rec.Body.Bytes(), &wrongResponse)
	require.NoError(t, err, "JSON unmarshals without error, but into wrong type")

	// This is the bug - fields are zero values when unmarshaling wrong type
	assert.Empty(t, wrongResponse.Email, "Email is empty when unmarshaling to wrong type")
	assert.Empty(t, wrongResponse.Status, "Status is empty when unmarshaling to wrong type")
	assert.Equal(t, uuid.Nil, wrongResponse.RoleId, "RoleId is nil UUID when unmarshaling to wrong type")
}

// TDD Lesson: Test duplicate email returns 409 Conflict
func TestCreateEmployee_Integration_DuplicateEmail(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create existing employee
	existing := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "existing@example.com",
		FullName: "Existing User",
		Status:   "active",
	})

	// Create session
	token, _ := auth.GenerateJWT(existing.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(existing.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler
	handler := handlers.NewEmployeesHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Post("/employees", handler.CreateEmployee)

	// Try to create employee with duplicate email
	reqBody := `{
		"email": "existing@example.com",
		"full_name": "Duplicate User",
		"role_id": "` + role.ID.String() + `"
	}`

	req := httptest.NewRequest(http.MethodPost, "/employees", strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Should return 409 Conflict
	assert.Equal(t, http.StatusConflict, rec.Code)

	var errorResponse api.Error
	err = json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	require.NoError(t, err)
	assert.Contains(t, errorResponse.Error, "already exists")
}

// ============================================================================
// Update Employee Integration Tests
// ============================================================================

// TDD Lesson: Integration test for PATCH /employees/{id}
func TestUpdateEmployee_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create employee to update
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "user@example.com",
		FullName: "Old Name",
		Status:   "active",
	})

	// Create session
	token, _ := auth.GenerateJWT(employee.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(employee.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler
	handler := handlers.NewEmployeesHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Patch("/employees/{employee_id}", handler.UpdateEmployee)

	// Update employee's name and status
	reqBody := `{
		"full_name": "New Name",
		"status": "suspended"
	}`

	req := httptest.NewRequest(http.MethodPatch, "/employees/"+employee.ID.String(), strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Employee
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify updated fields
	assert.Equal(t, "New Name", response.FullName)
	assert.Equal(t, api.EmployeeStatusSuspended, response.Status)
	assert.Equal(t, employee.Email, string(response.Email)) // Email unchanged

	// Verify in database
	updated, err := queries.GetEmployee(ctx, employee.ID)
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.FullName)
	assert.Equal(t, "suspended", updated.Status)
}

// TDD Lesson: Test org isolation for update
func TestUpdateEmployee_Integration_OrgIsolation(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create two organizations
	org1 := testutil.CreateTestOrg(t, queries, ctx)
	org2 := testutil.CreateTestOrg(t, queries, ctx)

	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create employee in org1
	emp1 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org1.ID,
		RoleID:   role.ID,
		Email:    "user1@org1.com",
		FullName: "User 1",
		Status:   "active",
	})

	// Create employee in org2
	emp2 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org2.ID,
		RoleID:   role.ID,
		Email:    "user2@org2.com",
		FullName: "User 2",
		Status:   "active",
	})

	// Authenticate as emp1 (org1)
	token, _ := auth.GenerateJWT(emp1.ID, org1.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(emp1.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler
	handler := handlers.NewEmployeesHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Patch("/employees/{employee_id}", handler.UpdateEmployee)

	// Try to update emp2 (different org)
	reqBody := `{"full_name": "Hacked Name"}`

	req := httptest.NewRequest(http.MethodPatch, "/employees/"+emp2.ID.String(), strings.NewReader(reqBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Should return 404 for security
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// Verify emp2 was NOT updated
	unchanged, err := queries.GetEmployee(ctx, emp2.ID)
	require.NoError(t, err)
	assert.Equal(t, "User 2", unchanged.FullName) // Still original name
}

// ============================================================================
// Delete Employee Integration Tests
// ============================================================================

// TDD Lesson: Integration test for DELETE /employees/{id}
func TestDeleteEmployee_Integration_Success(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create employee to delete
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "todelete@example.com",
		FullName: "To Delete",
		Status:   "active",
	})

	// Create another employee for auth
	admin := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "admin@example.com",
		FullName: "Admin",
		Status:   "active",
	})

	// Create session
	token, _ := auth.GenerateJWT(admin.ID, org.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(admin.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler
	handler := handlers.NewEmployeesHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Delete("/employees/{employee_id}", handler.DeleteEmployee)

	// Delete employee
	req := httptest.NewRequest(http.MethodDelete, "/employees/"+employee.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Should return 204 No Content
	assert.Equal(t, http.StatusNoContent, rec.Code)

	// Verify employee is soft-deleted (GetEmployee should return error)
	_, err = queries.GetEmployee(ctx, employee.ID)
	assert.Error(t, err) // Should not find deleted employee
}

// TDD Lesson: Test org isolation for delete
func TestDeleteEmployee_Integration_OrgIsolation(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create two organizations
	org1 := testutil.CreateTestOrg(t, queries, ctx)
	org2 := testutil.CreateTestOrg(t, queries, ctx)

	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create employee in org1
	emp1 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org1.ID,
		RoleID:   role.ID,
		Email:    "user1@org1.com",
		FullName: "User 1",
		Status:   "active",
	})

	// Create employee in org2
	emp2 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org2.ID,
		RoleID:   role.ID,
		Email:    "user2@org2.com",
		FullName: "User 2",
		Status:   "active",
	})

	// Authenticate as emp1 (org1)
	token, _ := auth.GenerateJWT(emp1.ID, org1.ID, 24*time.Hour)
	tokenHash := auth.HashToken(token)
	_, err := queries.CreateSession(ctx, testutil.CreateSessionParams(emp1.ID, tokenHash))
	require.NoError(t, err)

	// Setup handler
	handler := handlers.NewEmployeesHandler(queries)
	router := chi.NewRouter()
	router.Use(middleware.JWTAuth(queries))
	router.Delete("/employees/{employee_id}", handler.DeleteEmployee)

	// Try to delete emp2 (different org)
	req := httptest.NewRequest(http.MethodDelete, "/employees/"+emp2.ID.String(), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Should return 404 for security
	assert.Equal(t, http.StatusNotFound, rec.Code)

	// Verify emp2 was NOT deleted
	notDeleted, err := queries.GetEmployee(ctx, emp2.ID)
	require.NoError(t, err)
	assert.Equal(t, "User 2", notDeleted.FullName) // Still exists
}
