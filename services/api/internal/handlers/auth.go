package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/internal/auth"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	db db.Querier
}

// NewAuthHandler creates a new authentication handler
//
// TDD Lesson: We create this constructor because our tests need it.
// We're writing "just enough" code to satisfy the test requirements.
func NewAuthHandler(database db.Querier) *AuthHandler {
	return &AuthHandler{
		db: database,
	}
}

// Login handles employee login requests
//
// TDD Lesson: This implementation is driven by our tests.
// Every line here exists because a test requires it.
//
// Implementation Steps (derived from tests):
// 1. Parse JSON request
// 2. Lookup employee by email
// 3. Verify password
// 4. Check employee status
// 5. Generate JWT token
// 6. Create session in database
// 7. Update last login time
// 8. Return token and employee data
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Step 1: Parse request (TestLogin_InvalidJSON requires this)
	var req api.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Step 2: Get employee by email (TestLogin_UserNotFound requires this)
	employee, err := h.db.GetEmployeeByEmail(ctx, string(req.Email))
	if err != nil {
		// Don't reveal whether user exists (security best practice)
		writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Step 3: Verify password (TestLogin_InvalidPassword requires this)
	if !auth.VerifyPassword(req.Password, employee.PasswordHash) {
		writeError(w, http.StatusUnauthorized, "Invalid credentials")
		return
	}

	// Step 4: Check employee status (TestLogin_InactiveUser requires this)
	if employee.Status != "active" {
		var message string
		switch employee.Status {
		case "suspended":
			message = "Account is suspended"
		case "inactive":
			message = "Account is inactive"
		default:
			message = "Account is not active"
		}
		writeError(w, http.StatusForbidden, message)
		return
	}

	// Step 5: Generate JWT token (TestLogin_Success requires this)
	token, err := auth.GenerateJWT(employee.ID, employee.OrgID, 24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Step 6: Create session (TestLogin_Success expects this)
	tokenHash := auth.HashToken(token)
	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()
	session, err := h.db.CreateSession(ctx, db.CreateSessionParams{
		EmployeeID: employee.ID,
		TokenHash:  tokenHash,
		IpAddress:  &ipAddress,
		UserAgent:  &userAgent,
		ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create session")
		return
	}

	// Step 7: Update last login (TestLogin_Success expects this)
	if err := h.db.UpdateEmployeeLastLogin(ctx, employee.ID); err != nil {
		// Log error but don't fail the request
		// (login succeeded even if we couldn't update last_login)
		fmt.Printf("Warning: Failed to update last login: %v\n", err)
	}

	// Step 8: Return response (TestLogin_Success validates this)
	response := api.LoginResponse{
		Token:     token,
		ExpiresAt: session.ExpiresAt.Time,
		Employee:  mapEmployeeToAPI(employee),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// Logout handles employee logout requests
//
// TDD Lesson: Logout is simpler than login - just invalidate the session
//
// Implementation Steps (derived from tests):
// 1. Extract token from Authorization header
// 2. Verify JWT token is valid
// 3. Hash the token
// 4. Delete session from database
// 5. Return success message
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Step 1: Extract token from Authorization header (TestLogout_MissingToken requires this)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeError(w, http.StatusUnauthorized, "Missing authorization header")
		return
	}

	// Extract "Bearer <token>" format
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		writeError(w, http.StatusUnauthorized, "Invalid authorization header format")
		return
	}
	token := authHeader[len(bearerPrefix):]

	// Step 2: Verify JWT token (TestLogout_InvalidToken requires this)
	_, err := auth.VerifyJWT(token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid token")
		return
	}

	// Step 3: Hash the token
	tokenHash := auth.HashToken(token)

	// Step 4: Delete session (TestLogout_Success expects this)
	if err := h.db.DeleteSession(ctx, tokenHash); err != nil {
		// Log error but don't expose details to client
		fmt.Printf("Warning: Failed to delete session: %v\n", err)
		writeError(w, http.StatusInternalServerError, "Failed to logout")
		return
	}

	// Step 5: Return success (TestLogout_Success validates this)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logged out successfully",
	})
}

// GetMe handles fetching the current logged-in employee's information
//
// TDD Lesson: GetMe verifies the token and returns employee data
//
// Implementation Steps (derived from tests):
// 1. Extract token from Authorization header
// 2. Verify JWT token is valid and not expired
// 3. Hash the token
// 4. Fetch session and employee from database
// 5. Return employee data
func (h *AuthHandler) GetMe(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Step 1: Extract token from Authorization header (TestGetMe_MissingToken requires this)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		writeError(w, http.StatusUnauthorized, "Missing authorization header")
		return
	}

	// Extract "Bearer <token>" format
	const bearerPrefix = "Bearer "
	if len(authHeader) < len(bearerPrefix) || authHeader[:len(bearerPrefix)] != bearerPrefix {
		writeError(w, http.StatusUnauthorized, "Invalid authorization header format")
		return
	}
	token := authHeader[len(bearerPrefix):]

	// Step 2: Verify JWT token (TestGetMe_InvalidToken and TestGetMe_ExpiredToken require this)
	_, err := auth.VerifyJWT(token)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Invalid or expired token")
		return
	}

	// Step 3: Hash the token
	tokenHash := auth.HashToken(token)

	// Step 4: Fetch session and employee (TestGetMe_Success expects this)
	// This also validates:
	// - Session exists and is not expired
	// - Employee exists and is not deleted
	// - Employee status is active
	sessionData, err := h.db.GetSessionWithEmployee(ctx, tokenHash)
	if err != nil {
		// Session not found (user logged out or session expired)
		writeError(w, http.StatusUnauthorized, "Session not found")
		return
	}

	// Step 5: Convert to API employee format and return
	employee := mapSessionDataToAPIEmployee(sessionData)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(employee)
}

// mapSessionDataToAPIEmployee converts GetSessionWithEmployeeRow to API Employee
//
// TDD Lesson: Separate conversion logic for cleaner code and easier testing
func mapSessionDataToAPIEmployee(data db.GetSessionWithEmployeeRow) api.Employee {
	// Convert UUIDs to OpenAPI UUID type
	empID := openapi_types.UUID(data.EmployeeID)
	orgID := openapi_types.UUID(data.OrgID)
	roleID := openapi_types.UUID(data.RoleID)

	// Convert email to OpenAPI Email type
	email := openapi_types.Email(data.Email)

	// Convert timestamps
	createdAt := data.EmployeeCreatedAt.Time
	updatedAt := data.EmployeeUpdatedAt.Time

	employee := api.Employee{
		Id:        &empID,
		OrgId:     orgID,
		Email:     email,
		FullName:  data.FullName,
		RoleId:    roleID,
		Status:    api.EmployeeStatus(data.Status),
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	// Handle nullable team_id
	if data.TeamID.Valid {
		teamID := openapi_types.UUID(data.TeamID.Bytes)
		employee.TeamId = &teamID
	}

	// Handle nullable last_login_at
	if data.LastLoginAt.Valid {
		employee.LastLoginAt = &data.LastLoginAt.Time
	}

	return employee
}

// mapEmployeeToAPI converts database employee to API employee
//
// TDD Lesson: We create helper functions as needed during implementation.
// This keeps the main logic clean and testable.
func mapEmployeeToAPI(emp db.Employee) api.Employee {
	// Convert UUIDs to OpenAPI UUID type
	empID := openapi_types.UUID(emp.ID)
	orgID := openapi_types.UUID(emp.OrgID)
	roleID := openapi_types.UUID(emp.RoleID)

	// Convert email to OpenAPI Email type
	email := openapi_types.Email(emp.Email)

	// Convert timestamps
	createdAt := emp.CreatedAt.Time
	updatedAt := emp.UpdatedAt.Time

	employee := api.Employee{
		Id:        &empID,
		OrgId:     orgID,
		Email:     email,
		FullName:  emp.FullName,
		RoleId:    roleID,
		Status:    api.EmployeeStatus(emp.Status),
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}

	// Handle nullable team_id
	if emp.TeamID.Valid {
		teamID := openapi_types.UUID(emp.TeamID.Bytes)
		employee.TeamId = &teamID
	}

	// Handle nullable last_login_at
	if emp.LastLoginAt.Valid {
		employee.LastLoginAt = &emp.LastLoginAt.Time
	}

	return employee
}

// writeError writes a JSON error response
//
// TDD Lesson: Centralized error handling makes code cleaner and more consistent.
func writeError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(api.Error{
		Error: message,
	})
}
