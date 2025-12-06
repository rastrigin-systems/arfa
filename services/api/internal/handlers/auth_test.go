package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/auth"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/handlers"
)

// TestLogin_Success is our first test - it defines the EXPECTED behavior
//
// TDD Lesson: This test will FAIL initially because handlers.AuthHandler doesn't exist yet.
// That's GOOD! We want the test to fail first, then we'll make it pass.
func TestLogin_Success(t *testing.T) {
	// === ARRANGE === Setup test data and mocks
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Create test employee data
	employeeID := uuid.New()
	orgID := uuid.New()
	email := "alice@acme.com"
	password := "SecurePass123!"
	passwordHash, _ := auth.HashPassword(password)

	// Define what we EXPECT the database to return
	mockDB.EXPECT().
		GetEmployeeByEmail(gomock.Any(), email).
		Return(db.Employee{
			ID:           employeeID,
			OrgID:        orgID,
			Email:        email,
			PasswordHash: passwordHash,
			FullName:     "Alice Smith",
			Status:       "active",
		}, nil)

	// Expect session creation
	mockDB.EXPECT().
		CreateSession(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateSessionParams) (db.Session, error) {
			return db.Session{
				ID:         uuid.New(),
				EmployeeID: params.EmployeeID,
				TokenHash:  params.TokenHash,
				ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
				CreatedAt:  pgtype.Timestamp{Time: time.Now(), Valid: true},
			}, nil
		})

	// Expect last login update
	mockDB.EXPECT().
		UpdateEmployeeLastLogin(gomock.Any(), employeeID).
		Return(nil)

	// === ACT === Call the handler
	handler := handlers.NewAuthHandler(mockDB)

	requestBody := api.LoginRequest{
		Email:    openapi_types.Email(email),
		Password: password,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	// === ASSERT === Verify the results
	assert.Equal(t, http.StatusOK, rec.Code, "Should return 200 OK")

	var response api.LoginResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify token is returned
	assert.NotEmpty(t, response.Token, "Should return a JWT token")
	assert.Greater(t, len(response.Token), 50, "Token should be reasonably long")

	// Verify token expiration
	assert.True(t, response.ExpiresAt.After(time.Now()), "Token should not be expired")

	// Verify employee data
	require.NotNil(t, response.Employee.Id, "Employee ID should not be nil")
	assert.Equal(t, employeeID.String(), response.Employee.Id.String())
	assert.Equal(t, email, string(response.Employee.Email))
	assert.Equal(t, "Alice Smith", response.Employee.FullName)
}

// TDD Lesson: Let's add a test for INVALID password before implementing
// This is called "test-first thinking" - define ALL behaviors upfront
func TestLogin_InvalidPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	email := "alice@acme.com"
	correctPassword := "SecurePass123!"
	wrongPassword := "WrongPassword"
	passwordHash, _ := auth.HashPassword(correctPassword)

	// Expect database to return employee
	mockDB.EXPECT().
		GetEmployeeByEmail(gomock.Any(), email).
		Return(db.Employee{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: passwordHash,
			Status:       "active",
		}, nil)

	// NO session creation expected (password is wrong)
	// NO last login update expected

	handler := handlers.NewAuthHandler(mockDB)

	requestBody := api.LoginRequest{
		Email:    openapi_types.Email(email),
		Password: wrongPassword, // Wrong password!
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var errorResponse api.Error
	json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	assert.Contains(t, errorResponse.Error, "Invalid credentials")
}

// TDD Lesson: Test for user not found
func TestLogin_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	email := "nonexistent@example.com"

	// Expect database to return no employee (error)
	mockDB.EXPECT().
		GetEmployeeByEmail(gomock.Any(), email).
		Return(db.Employee{}, assert.AnError) // Simulate not found error

	handler := handlers.NewAuthHandler(mockDB)

	requestBody := api.LoginRequest{
		Email:    openapi_types.Email(email),
		Password: "SomePassword",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	// Should return 401 (don't reveal if user exists or not)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TDD Lesson: Test for invalid JSON
func TestLogin_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewAuthHandler(mockDB)

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TDD Lesson: Test for inactive user
func TestLogin_InactiveUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	email := "inactive@example.com"
	password := "SecurePass123!"
	passwordHash, _ := auth.HashPassword(password)

	mockDB.EXPECT().
		GetEmployeeByEmail(gomock.Any(), email).
		Return(db.Employee{
			ID:           uuid.New(),
			Email:        email,
			PasswordHash: passwordHash,
			Status:       "suspended", // User is suspended!
		}, nil)

	handler := handlers.NewAuthHandler(mockDB)

	requestBody := api.LoginRequest{
		Email:    openapi_types.Email(email),
		Password: password,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Login(rec, req)

	// Should return 403 Forbidden
	assert.Equal(t, http.StatusForbidden, rec.Code)

	var errorResponse api.Error
	json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	assert.Contains(t, errorResponse.Error, "Account is suspended")
}

// ============================================================================
// Logout Tests
// ============================================================================

// TDD Lesson: Testing logout - invalidate the session
func TestLogout_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Create a valid token
	employeeID := uuid.New()
	orgID := uuid.New()
	token, _ := auth.GenerateJWT(employeeID, orgID, 24*time.Hour)
	tokenHash := auth.HashToken(token)

	// Expect session deletion
	mockDB.EXPECT().
		DeleteSession(gomock.Any(), tokenHash).
		Return(nil)

	handler := handlers.NewAuthHandler(mockDB)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.Logout(rec, req)

	// Should return 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	json.Unmarshal(rec.Body.Bytes(), &response)
	assert.Equal(t, "Logged out successfully", response["message"])
}

// TDD Lesson: Test missing/invalid token
func TestLogout_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewAuthHandler(mockDB)

	// No database calls expected (token validation fails first)

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler.Logout(rec, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TDD Lesson: Test missing Authorization header
func TestLogout_MissingToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewAuthHandler(mockDB)

	// No database calls expected

	req := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	// No Authorization header
	rec := httptest.NewRecorder()

	handler.Logout(rec, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ============================================================================
// GetMe Tests
// ============================================================================

// TDD Lesson: GetMe returns the current employee from JWT token
func TestGetMe_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Create test data
	employeeID := uuid.New()
	orgID := uuid.New()
	roleID := uuid.New()
	email := "alice@acme.com"

	// Generate a valid token
	token, _ := auth.GenerateJWT(employeeID, orgID, 24*time.Hour)
	tokenHash := auth.HashToken(token)

	// Expect database to return employee data
	mockDB.EXPECT().
		GetSessionWithEmployee(gomock.Any(), tokenHash).
		Return(db.GetSessionWithEmployeeRow{
			ID:                uuid.New(),
			EmployeeID:        employeeID,
			TokenHash:         tokenHash,
			ExpiresAt:         pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
			CreatedAt:         pgtype.Timestamp{Time: time.Now(), Valid: true},
			OrgID:             orgID,
			TeamID:            pgtype.UUID{},
			RoleID:            roleID,
			Email:             email,
			FullName:          "Alice Smith",
			Status:            "active",
			Preferences:       []byte("{}"),
			LastLoginAt:       pgtype.Timestamp{},
			EmployeeCreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
			EmployeeUpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
		}, nil)

	handler := handlers.NewAuthHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.GetMe(rec, req)

	// Should return 200 OK
	assert.Equal(t, http.StatusOK, rec.Code)

	var response api.Employee
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err)

	assert.Equal(t, employeeID.String(), response.Id.String())
	assert.Equal(t, email, string(response.Email))
	assert.Equal(t, "Alice Smith", response.FullName)
	assert.Equal(t, "active", string(response.Status))
}

// TDD Lesson: Test invalid token
func TestGetMe_InvalidToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewAuthHandler(mockDB)

	// No database calls expected (token validation fails first)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	handler.GetMe(rec, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TDD Lesson: Test missing token
func TestGetMe_MissingToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewAuthHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	// No Authorization header
	rec := httptest.NewRecorder()

	handler.GetMe(rec, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TDD Lesson: Test expired token
func TestGetMe_ExpiredToken(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	employeeID := uuid.New()
	orgID := uuid.New()

	// Generate an expired token (1 nanosecond duration)
	token, _ := auth.GenerateJWT(employeeID, orgID, 1*time.Nanosecond)
	time.Sleep(2 * time.Millisecond) // Ensure it's expired

	handler := handlers.NewAuthHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.GetMe(rec, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// TDD Lesson: Test session not found (logged out)
func TestGetMe_SessionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	employeeID := uuid.New()
	orgID := uuid.New()
	token, _ := auth.GenerateJWT(employeeID, orgID, 24*time.Hour)
	tokenHash := auth.HashToken(token)

	// Expect database to return "not found" error
	mockDB.EXPECT().
		GetSessionWithEmployee(gomock.Any(), tokenHash).
		Return(db.GetSessionWithEmployeeRow{}, assert.AnError)

	handler := handlers.NewAuthHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	handler.GetMe(rec, req)

	// Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

// ============================================================================
// Register Tests
// ============================================================================

// TestRegister_Success is the first TDD test for registration
//
// TDD Lesson: Define EXPECTED behavior before implementation
func TestRegister_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Test data
	email := "alice@acme.com"
	password := "SecurePass123!"
	fullName := "Alice Smith"
	orgName := "ACME Corporation"
	orgSlug := "acme-corp"

	orgID := uuid.New()
	employeeID := uuid.New()
	roleID := uuid.New()
	passwordHash, _ := auth.HashPassword(password)

	// 1. Check org_slug is available (no existing org)
	mockDB.EXPECT().
		GetOrganizationBySlug(gomock.Any(), orgSlug).
		Return(db.Organization{}, assert.AnError) // Not found = available

	// 2. Check email is available (no existing employee)
	mockDB.EXPECT().
		GetEmployeeByEmail(gomock.Any(), email).
		Return(db.Employee{}, assert.AnError) // Not found = available

	// 3. Get admin role
	mockDB.EXPECT().
		GetRoleByName(gomock.Any(), "admin").
		Return(db.Role{
			ID:          roleID,
			Name:        "admin",
			Permissions: []byte(`["*"]`),
		}, nil)

	// 4. Create organization
	mockDB.EXPECT().
		CreateOrganization(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateOrganizationParams) (db.Organization, error) {
			return db.Organization{
				ID:        orgID,
				Name:      params.Name,
				Slug:      params.Slug,
				CreatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
				UpdatedAt: pgtype.Timestamp{Time: time.Now(), Valid: true},
			}, nil
		})

	// 5. Create employee
	mockDB.EXPECT().
		CreateEmployee(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateEmployeeParams) (db.Employee, error) {
			assert.Equal(t, orgID, params.OrgID)
			assert.Equal(t, roleID, params.RoleID)
			assert.Equal(t, email, params.Email)
			assert.Equal(t, fullName, params.FullName)
			// Verify password is hashed (not plain text)
			assert.NotEqual(t, password, params.PasswordHash)
			assert.True(t, len(params.PasswordHash) > 30) // bcrypt hash is long

			return db.Employee{
				ID:           employeeID,
				OrgID:        orgID,
				RoleID:       roleID,
				Email:        email,
				FullName:     fullName,
				PasswordHash: passwordHash,
				Status:       "active",
				CreatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
				UpdatedAt:    pgtype.Timestamp{Time: time.Now(), Valid: true},
			}, nil
		})

	// 6. Create session
	mockDB.EXPECT().
		CreateSession(gomock.Any(), gomock.Any()).
		DoAndReturn(func(ctx context.Context, params db.CreateSessionParams) (db.Session, error) {
			return db.Session{
				ID:         uuid.New(),
				EmployeeID: params.EmployeeID,
				TokenHash:  params.TokenHash,
				ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
				CreatedAt:  pgtype.Timestamp{Time: time.Now(), Valid: true},
			}, nil
		})

	// Create handler and request
	handler := handlers.NewAuthHandler(mockDB)

	requestBody := api.RegisterRequest{
		Email:    openapi_types.Email(email),
		Password: password,
		FullName: fullName,
		OrgName:  orgName,
		OrgSlug:  orgSlug,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	// Assert response
	assert.Equal(t, http.StatusCreated, rec.Code, "Should return 201 Created")

	var response api.RegisterResponse
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	require.NoError(t, err, "Response should be valid JSON")

	// Verify token
	assert.NotEmpty(t, response.Token, "Should return JWT token")
	assert.Greater(t, len(response.Token), 50, "Token should be reasonably long")
	assert.True(t, response.ExpiresAt.After(time.Now()), "Token should not be expired")

	// Verify employee data
	require.NotNil(t, response.Employee.Id)
	assert.Equal(t, employeeID.String(), response.Employee.Id.String())
	assert.Equal(t, email, string(response.Employee.Email))
	assert.Equal(t, fullName, response.Employee.FullName)

	// Verify organization data
	require.NotNil(t, response.Organization.Id)
	assert.Equal(t, orgID.String(), response.Organization.Id.String())
	assert.Equal(t, orgName, response.Organization.Name)
	assert.Equal(t, orgSlug, response.Organization.Slug)
}

// TestRegister_DuplicateOrgSlug tests org_slug uniqueness constraint
func TestRegister_DuplicateOrgSlug(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgSlug := "acme-corp"

	// Org slug already exists
	mockDB.EXPECT().
		GetOrganizationBySlug(gomock.Any(), orgSlug).
		Return(db.Organization{
			ID:   uuid.New(),
			Slug: orgSlug,
		}, nil) // Found = conflict

	handler := handlers.NewAuthHandler(mockDB)

	requestBody := api.RegisterRequest{
		Email:    openapi_types.Email("alice@acme.com"),
		Password: "SecurePass123!",
		FullName: "Alice Smith",
		OrgName:  "ACME Corporation",
		OrgSlug:  orgSlug,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	// Should return 409 Conflict
	assert.Equal(t, http.StatusConflict, rec.Code)

	var errorResponse api.Error
	json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	assert.Contains(t, errorResponse.Error, "org_slug")
	assert.Contains(t, errorResponse.Error, "already exists")
}

// TestRegister_DuplicateEmail tests email uniqueness constraint
func TestRegister_DuplicateEmail(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	email := "alice@acme.com"
	orgSlug := "acme-corp"

	// Org slug is available
	mockDB.EXPECT().
		GetOrganizationBySlug(gomock.Any(), orgSlug).
		Return(db.Organization{}, assert.AnError) // Not found = available

	// Email already exists
	mockDB.EXPECT().
		GetEmployeeByEmail(gomock.Any(), email).
		Return(db.Employee{
			ID:    uuid.New(),
			Email: email,
		}, nil) // Found = conflict

	handler := handlers.NewAuthHandler(mockDB)

	requestBody := api.RegisterRequest{
		Email:    openapi_types.Email(email),
		Password: "SecurePass123!",
		FullName: "Alice Smith",
		OrgName:  "ACME Corporation",
		OrgSlug:  orgSlug,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	// Should return 409 Conflict
	assert.Equal(t, http.StatusConflict, rec.Code)

	var errorResponse api.Error
	json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	assert.Contains(t, errorResponse.Error, "email")
	assert.Contains(t, errorResponse.Error, "already exists")
}

// TestRegister_InvalidJSON tests malformed request
func TestRegister_InvalidJSON(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewAuthHandler(mockDB)

	// Send invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rec.Code)
}

// TestRegister_WeakPassword tests password strength validation
func TestRegister_WeakPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewAuthHandler(mockDB)

	requestBody := api.RegisterRequest{
		Email:    openapi_types.Email("alice@acme.com"),
		Password: "weak", // Too short
		FullName: "Alice Smith",
		OrgName:  "ACME Corporation",
		OrgSlug:  "acme-corp",
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	// Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, rec.Code)

	var errorResponse api.Error
	json.Unmarshal(rec.Body.Bytes(), &errorResponse)
	assert.Contains(t, errorResponse.Error, "password")
}

// TestRegister_InvalidOrgSlug tests org_slug format validation
func TestRegister_InvalidOrgSlug(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)
	handler := handlers.NewAuthHandler(mockDB)

	testCases := []struct {
		name    string
		orgSlug string
	}{
		{"uppercase", "ACME-Corp"},
		{"spaces", "acme corp"},
		{"underscores", "acme_corp"},
		{"too short", "ab"},
		{"special chars", "acme@corp"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			requestBody := api.RegisterRequest{
				Email:    openapi_types.Email("alice@acme.com"),
				Password: "SecurePass123!",
				FullName: "Alice Smith",
				OrgName:  "ACME Corporation",
				OrgSlug:  tc.orgSlug,
			}
			bodyBytes, _ := json.Marshal(requestBody)

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler.Register(rec, req)

			// Should return 400 Bad Request
			assert.Equal(t, http.StatusBadRequest, rec.Code)

			var errorResponse api.Error
			json.Unmarshal(rec.Body.Bytes(), &errorResponse)
			assert.Contains(t, errorResponse.Error, "org_slug")
		})
	}
}

// TestRegister_AdminRoleNotFound tests missing admin role
func TestRegister_AdminRoleNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	orgSlug := "acme-corp"
	email := "alice@acme.com"

	// Org slug is available
	mockDB.EXPECT().
		GetOrganizationBySlug(gomock.Any(), orgSlug).
		Return(db.Organization{}, assert.AnError)

	// Email is available
	mockDB.EXPECT().
		GetEmployeeByEmail(gomock.Any(), email).
		Return(db.Employee{}, assert.AnError)

	// Admin role not found
	mockDB.EXPECT().
		GetRoleByName(gomock.Any(), "admin").
		Return(db.Role{}, assert.AnError)

	handler := handlers.NewAuthHandler(mockDB)

	requestBody := api.RegisterRequest{
		Email:    openapi_types.Email(email),
		Password: "SecurePass123!",
		FullName: "Alice Smith",
		OrgName:  "ACME Corporation",
		OrgSlug:  orgSlug,
	}
	bodyBytes, _ := json.Marshal(requestBody)

	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.Register(rec, req)

	// Should return 500 Internal Server Error
	assert.Equal(t, http.StatusInternalServerError, rec.Code)
}
