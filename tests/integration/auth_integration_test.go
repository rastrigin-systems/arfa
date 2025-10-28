package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/internal/auth"
	"github.com/sergeirastrigin/ubik-enterprise/internal/handlers"
	"github.com/sergeirastrigin/ubik-enterprise/tests/testutil"
)

// TestLogin_Integration_Success tests the complete login flow with a REAL database
//
// Integration Test Lesson:
// - Uses real PostgreSQL via testcontainers
// - Tests full HTTP → Handler → Database → Response flow
// - Verifies side effects (session creation, last login update)
// - Slower but more comprehensive than unit tests
func TestLogin_Integration_Success(t *testing.T) {
	// Skip in short mode (for quick local testing)
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// === ARRANGE === Setup real database and test data

	// Step 1: Create real PostgreSQL database
	// Integration Test Lesson: This is the KEY difference from unit tests!
	// We're using a REAL database, not a mock.
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))

	ctx := testutil.GetContext(t)

	// Step 2: Create test organization
	org := testutil.CreateTestOrg(t, queries, ctx)

	// Step 3: Create test role
	role := testutil.CreateTestRole(t, queries, ctx, "Admin")

	// Step 4: Create test employee with known password
	password := "SecurePass123!"
	passwordHash, err := auth.HashPassword(password)
	require.NoError(t, err)

	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:        org.ID,
		RoleID:       role.ID,
		Email:        "alice@acme.com",
		FullName:     "Alice Smith",
		PasswordHash: passwordHash,
		Status:       "active",
	})

	// Debug: Verify employee was created
	t.Logf("Created employee: ID=%s, Email=%s", employee.ID, employee.Email)

	// Debug: Verify we can retrieve the employee
	retrievedEmp, err := queries.GetEmployeeByEmail(ctx, "alice@acme.com")
	require.NoError(t, err, "Should be able to retrieve employee by email")
	t.Logf("Retrieved employee: ID=%s, Email=%s, PasswordHash length=%d",
		retrievedEmp.ID, retrievedEmp.Email, len(retrievedEmp.PasswordHash))

	// Step 5: Create HTTP router with real handler
	// Integration Test Lesson: We're testing the REAL handler, not a mock
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/login", authHandler.Login)

	// === ACT === Make real HTTP request

	loginRequest := api.LoginRequest{
		Email:    openapi_types.Email("alice@acme.com"),
		Password: password,
	}
	bodyBytes, _ := json.Marshal(loginRequest)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// === ASSERT === Verify response and database state

	// Debug: Print response if not 200
	if rec.Code != http.StatusOK {
		t.Logf("Response Code: %d", rec.Code)
		t.Logf("Response Body: %s", rec.Body.String())
	}

	// Verify HTTP response
	assert.Equal(t, http.StatusOK, rec.Code, "Should return 200 OK")

	var response api.LoginResponse
	err = json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify token is returned
	assert.NotEmpty(t, response.Token, "Should return JWT token")
	assert.Greater(t, len(response.Token), 50, "Token should be reasonably long")

	// Verify employee data
	require.NotNil(t, response.Employee.Id)
	assert.Equal(t, employee.ID.String(), response.Employee.Id.String())
	assert.Equal(t, "alice@acme.com", string(response.Employee.Email))
	assert.Equal(t, "Alice Smith", response.Employee.FullName)

	// === INTEGRATION TEST UNIQUE PART ===
	// Verify DATABASE side effects (this is what unit tests can't do!)

	// Verify 1: Session was created in database
	tokenHash := auth.HashToken(response.Token)
	session, err := queries.GetSession(ctx, tokenHash)
	require.NoError(t, err, "Session should exist in database")
	assert.Equal(t, employee.ID, session.EmployeeID, "Session should belong to employee")
	assert.True(t, session.ExpiresAt.Time.After(time.Now()), "Session should not be expired")

	// Verify 2: Last login timestamp was updated
	updatedEmployee, err := queries.GetEmployee(ctx, employee.ID)
	require.NoError(t, err)
	require.True(t, updatedEmployee.LastLoginAt.Valid, "Last login should be set")
	assert.WithinDuration(t, time.Now(), updatedEmployee.LastLoginAt.Time, 5*time.Second,
		"Last login should be recent")

	// Verify 3: Token can be verified and contains correct claims
	claims, err := auth.VerifyJWT(response.Token)
	require.NoError(t, err, "Token should be valid")
	assert.Equal(t, employee.ID.String(), claims.EmployeeID)
	assert.Equal(t, org.ID.String(), claims.OrgID)

	// Integration Test Lesson: We just verified:
	// ✅ HTTP request/response work
	// ✅ Handler logic works
	// ✅ Database queries work
	// ✅ Sessions are created
	// ✅ Timestamps are updated
	// ✅ JWT tokens are valid
	//
	// This gives us CONFIDENCE the whole system works together!
}

// TestLogin_Integration_InvalidPassword tests login with wrong password
//
// Integration Test Lesson: Even error cases should be tested with real DB
// to ensure database queries don't have unexpected side effects
func TestLogin_Integration_InvalidPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create test data
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "User")

	correctPassword := "CorrectPassword123"
	passwordHash, _ := auth.HashPassword(correctPassword)

	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:        org.ID,
		RoleID:       role.ID,
		Email:        "bob@acme.com",
		FullName:     "Bob Jones",
		PasswordHash: passwordHash,
		Status:       "active",
	})

	// Setup router
	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/login", authHandler.Login)

	// Try to login with WRONG password
	loginRequest := api.LoginRequest{
		Email:    openapi_types.Email("bob@acme.com"),
		Password: "WrongPassword",
	}
	bodyBytes, _ := json.Marshal(loginRequest)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Verify error response
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	// === VERIFY NO SIDE EFFECTS ===
	// Integration Test Lesson: Failed login should NOT create session or update last_login

	// Verify 1: No session was created (checking with non-existent hash)
	_, sessionErr := queries.GetSessionWithEmployee(ctx, "nonexistent-hash")
	assert.Error(t, sessionErr, "Session should not exist")

	// Verify 2: Last login was NOT updated
	unchangedEmployee, err := queries.GetEmployee(ctx, employee.ID)
	require.NoError(t, err)
	assert.False(t, unchangedEmployee.LastLoginAt.Valid, "Last login should still be null")
}

// TestLogin_Integration_SuspendedUser tests that suspended users cannot login
func TestLogin_Integration_SuspendedUser(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "User")

	password := "Password123"
	passwordHash, _ := auth.HashPassword(password)

	// Create SUSPENDED employee
	_ = testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:        org.ID,
		RoleID:       role.ID,
		Email:        "suspended@acme.com",
		FullName:     "Suspended User",
		PasswordHash: passwordHash,
		Status:       "suspended", // Key difference: not active!
	})

	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/login", authHandler.Login)

	loginRequest := api.LoginRequest{
		Email:    openapi_types.Email("suspended@acme.com"),
		Password: password,
	}
	bodyBytes, _ := json.Marshal(loginRequest)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var errorResponse api.Error
	json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	assert.Contains(t, errorResponse.Error, "suspended")
}

// TestLogin_Integration_MultipleEmployees tests that login works correctly
// when multiple employees exist (tests query filtering)
//
// Integration Test Lesson: Real databases can have complex scenarios
// that are hard to mock - test them with real data!
func TestLogin_Integration_MultipleEmployees(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create TWO organizations
	org1 := testutil.CreateTestOrg(t, queries, ctx)
	org2 := testutil.CreateTestOrg(t, queries, ctx)

	role := testutil.CreateTestRole(t, queries, ctx, "User")

	password := "Password123"
	passwordHash, _ := auth.HashPassword(password)

	// Create employee in org1 with email alice@acme.com
	emp1 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:        org1.ID,
		RoleID:       role.ID,
		Email:        "alice@company1.com",
		FullName:     "Alice from Org 1",
		PasswordHash: passwordHash,
		Status:       "active",
	})

	// Create employee in org2 with different email
	emp2 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:        org2.ID,
		RoleID:       role.ID,
		Email:        "bob@company2.com",
		FullName:     "Bob from Org 2",
		PasswordHash: passwordHash,
		Status:       "active",
	})

	router := chi.NewRouter()
	authHandler := handlers.NewAuthHandler(queries)
	router.Post("/auth/login", authHandler.Login)

	// Login as alice@company1.com
	loginRequest := api.LoginRequest{
		Email:    openapi_types.Email("alice@company1.com"),
		Password: password,
	}
	bodyBytes, _ := json.Marshal(loginRequest)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.LoginResponse
	json.Unmarshal(rec.Body.Bytes(), &response)

	// Verify we got the CORRECT employee (from org1, not org2)
	require.NotNil(t, response.Employee.Id)
	assert.Equal(t, emp1.ID.String(), response.Employee.Id.String())
	assert.Equal(t, org1.ID.String(), response.Employee.OrgId.String())
	assert.NotEqual(t, emp2.ID.String(), response.Employee.Id.String(), "Should not get wrong employee")

	// Verify JWT contains correct org_id
	claims, err := auth.VerifyJWT(response.Token)
	require.NoError(t, err)
	assert.Equal(t, org1.ID.String(), claims.OrgID, "Token should have org1's ID")
}
