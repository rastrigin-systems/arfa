package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/internal/auth"
	"github.com/sergeirastrigin/ubik-enterprise/internal/handlers"
	"github.com/sergeirastrigin/ubik-enterprise/internal/middleware"
	"github.com/sergeirastrigin/ubik-enterprise/tests/testutil"
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
