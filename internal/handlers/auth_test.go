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
	"github.com/sergeirastrigin/ubik-enterprise/internal/auth"
	"github.com/sergeirastrigin/ubik-enterprise/internal/handlers"
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
