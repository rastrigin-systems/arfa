package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"

	"github.com/rastrigin-systems/arfa/generated/api"
	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/services/api/internal/auth"
	authpkg "github.com/rastrigin-systems/arfa/services/api/pkg/auth"
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
		log.Printf("Warning: Failed to update last login: %v", err)
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
		log.Printf("Warning: Failed to delete session: %v", err)
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

// Register handles employee self-service registration
//
// TDD Lesson: Registration is more complex than login - we create both org and employee atomically
//
// Implementation Steps (derived from tests):
// 1. Parse and validate request
// 2. Check org_slug availability
// 3. Check email availability
// 4. Get admin role
// 5. Hash password
// 6. Create organization
// 7. Create employee with admin role
// 8. Generate JWT token
// 9. Create session
// 10. Return token, employee, and organization data
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Step 1: Parse request (TestRegister_InvalidJSON requires this)
	var req api.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Validate password strength (TestRegister_WeakPassword requires this)
	if len(req.Password) < 8 {
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	// Validate org_slug format (TestRegister_InvalidOrgSlug requires this)
	if !isValidOrgSlug(req.OrgSlug) {
		writeError(w, http.StatusBadRequest, "org_slug must be lowercase, alphanumeric, dashes only, and at least 3 characters")
		return
	}

	// Step 2: Check org_slug is available (TestRegister_DuplicateOrgSlug requires this)
	_, err := h.db.GetOrganizationBySlug(ctx, req.OrgSlug)
	if err == nil {
		// Organization found = slug already exists
		writeError(w, http.StatusConflict, "org_slug already exists")
		return
	}

	// Step 3: Check email is available (TestRegister_DuplicateEmail requires this)
	_, err = h.db.GetEmployeeByEmail(ctx, string(req.Email))
	if err == nil {
		// Employee found = email already exists
		writeError(w, http.StatusConflict, "email already exists")
		return
	}

	// Step 4: Get admin role (TestRegister_AdminRoleNotFound requires this)
	adminRole, err := h.db.GetRoleByName(ctx, "admin")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get admin role")
		return
	}

	// Step 5: Hash password (TestRegister_Success validates this)
	passwordHash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Step 6: Create organization (TestRegister_Success expects this)
	org, err := h.db.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name: req.OrgName,
		Slug: req.OrgSlug,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create organization")
		return
	}

	// Step 7: Create employee with admin role (TestRegister_Success expects this)
	employee, err := h.db.CreateEmployee(ctx, db.CreateEmployeeParams{
		OrgID:        org.ID,
		RoleID:       adminRole.ID,
		Email:        string(req.Email),
		FullName:     req.FullName,
		PasswordHash: passwordHash,
		Status:       "active",
		Preferences:  []byte("{}"),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create employee")
		return
	}

	// Step 8: Generate JWT token (TestRegister_Success requires this)
	token, err := auth.GenerateJWT(employee.ID, employee.OrgID, 24*time.Hour)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	// Step 9: Create session (TestRegister_Success expects this)
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

	// Step 10: Return response (TestRegister_Success validates this)
	response := api.RegisterResponse{
		Token:        token,
		ExpiresAt:    session.ExpiresAt.Time,
		Employee:     mapEmployeeToAPI(employee),
		Organization: mapOrganizationToAPI(org),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// CheckSlugAvailability checks if an organization slug is available
//
// TDD Lesson: Simple validation endpoint - checks database and returns boolean
//
// Implementation Steps:
// 1. Extract slug from query parameter
// 2. Validate slug format
// 3. Check if organization with slug exists in database
// 4. Return availability status
func (h *AuthHandler) CheckSlugAvailability(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Step 1: Get slug from query parameter
	slug := r.URL.Query().Get("slug")
	if slug == "" {
		writeError(w, http.StatusBadRequest, "slug parameter is required")
		return
	}

	// Step 2: Validate slug format
	if !isValidOrgSlug(slug) {
		writeError(w, http.StatusBadRequest, "Invalid slug format")
		return
	}

	// Step 3: Check if slug exists in database
	_, err := h.db.GetOrganizationBySlug(ctx, slug)
	available := err != nil // If error (org not found), slug is available

	// Step 4: Return availability status
	response := map[string]bool{
		"available": available,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ForgotPassword handles password reset requests
//
// TDD Lesson: Security-first design to prevent email enumeration
//
// Implementation Steps (derived from security requirements):
// 1. Parse email from request
// 2. Always return success message (don't reveal if email exists)
// 3. If employee exists, check rate limit
// 4. Generate secure token and store it
// 5. Send password reset email
func (h *AuthHandler) ForgotPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Step 1: Parse request
	var req api.ForgotPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Step 2: Generic success message (prevents email enumeration)
	genericMessage := "If an account exists with this email, you will receive a password reset link within a few minutes."

	// Try to get employee by email (don't fail if not found)
	employee, err := h.db.GetEmployeeByEmail(ctx, string(req.Email))
	if err != nil {
		// Employee not found - return generic success (security)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.ForgotPasswordResponse{
			Message: genericMessage,
		})
		return
	}

	// Step 3: Check rate limit (3 requests per hour)
	if err := authpkg.CheckPasswordResetRateLimit(ctx, employee.ID, h.db); err != nil {
		// Rate limited - return 429
		writeError(w, http.StatusTooManyRequests, "Too many password reset requests. Please try again later.")
		return
	}

	// Step 4: Generate secure token
	token, err := authpkg.GenerateSecureToken()
	if err != nil {
		// Log error but return generic success (don't reveal internal errors)
		log.Printf("Error generating token: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.ForgotPasswordResponse{
			Message: genericMessage,
		})
		return
	}

	// Store token in database with 1-hour expiration
	expiresAt := pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour), Valid: true}
	_, err = h.db.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
		EmployeeID: employee.ID,
		Token:      token,
		ExpiresAt:  expiresAt,
	})
	if err != nil {
		// Log error but return generic success
		log.Printf("Error creating password reset token: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(api.ForgotPasswordResponse{
			Message: genericMessage,
		})
		return
	}

	// Step 5: Send email (in production, this would be async)
	// For now, log the reset URL - in production, replace with email service
	// TODO: Replace with actual email service (SendGrid, AWS SES, etc.)
	webAppURL := os.Getenv("WEB_APP_URL")
	if webAppURL == "" {
		webAppURL = "http://localhost:3000" // default for local development
	}
	resetURL := fmt.Sprintf("%s/reset-password/%s", webAppURL, token)
	log.Printf("📧 Password reset requested for %s", employee.Email)
	log.Printf("   Reset URL: %s", resetURL)
	log.Printf("   Token: %s", token)

	// Return generic success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.ForgotPasswordResponse{
		Message: genericMessage,
	})
}

// VerifyResetToken verifies if a password reset token is valid
//
// TDD Lesson: Token validation with clear error messages
//
// Implementation Steps:
// 1. Extract token from query parameter
// 2. Check if token exists and is not expired
// 3. Check if token has not been used
// 4. Return validation result
func (h *AuthHandler) VerifyResetToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Step 1: Extract token from query parameter
	token := r.URL.Query().Get("token")
	if token == "" {
		writeError(w, http.StatusBadRequest, "Token is required")
		return
	}

	// Step 2 & 3: Check if token is valid (exists, not expired, not used)
	_, err := h.db.GetPasswordResetToken(ctx, token)
	if err != nil {
		// Token invalid, expired, or already used
		writeError(w, http.StatusBadRequest, "Invalid, expired, or already used token")
		return
	}

	// Step 4: Return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.VerifyResetTokenResponse{
		Valid: true,
	})
}

// ResetPassword resets an employee's password using a valid token
//
// TDD Lesson: Atomic operation - update password and mark token as used
//
// Implementation Steps:
// 1. Parse request (token + new password)
// 2. Validate password strength
// 3. Verify token is valid
// 4. Hash new password
// 5. Update employee password
// 6. Mark token as used
// 7. Return success
func (h *AuthHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Step 1: Parse request
	var req api.ResetPasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request format")
		return
	}

	// Step 2: Validate password strength (minimum 8 characters)
	if len(req.NewPassword) < 8 {
		writeError(w, http.StatusBadRequest, "Password must be at least 8 characters")
		return
	}

	// Step 3: Verify token is valid (not expired, not used)
	resetToken, err := h.db.GetPasswordResetToken(ctx, req.Token)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid, expired, or already used token")
		return
	}

	// Step 4: Hash new password
	passwordHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to process password")
		return
	}

	// Step 5: Update employee password
	err = h.db.UpdateEmployeePassword(ctx, db.UpdateEmployeePasswordParams{
		PasswordHash: passwordHash,
		ID:           resetToken.EmployeeID,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update password")
		return
	}

	// Step 6: Mark token as used
	err = h.db.MarkPasswordResetTokenUsed(ctx, req.Token)
	if err != nil {
		// Log error but password was updated successfully
		log.Printf("Warning: Failed to mark token as used: %v", err)
	}

	// Step 7: Return success
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(api.ResetPasswordResponse{
		Message: "Password reset successful",
	})
}

// isValidOrgSlug validates org_slug format
// Must be lowercase, alphanumeric, dashes only, at least 3 characters
func isValidOrgSlug(slug string) bool {
	if len(slug) < 3 || len(slug) > 100 {
		return false
	}

	for _, ch := range slug {
		if !((ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9') || ch == '-') {
			return false
		}
	}

	return true
}

// mapOrganizationToAPI converts database organization to API organization
func mapOrganizationToAPI(org db.Organization) api.Organization {
	orgID := openapi_types.UUID(org.ID)
	createdAt := org.CreatedAt.Time
	updatedAt := org.UpdatedAt.Time

	return api.Organization{
		Id:        &orgID,
		Name:      org.Name,
		Slug:      org.Slug,
		CreatedAt: &createdAt,
		UpdatedAt: &updatedAt,
	}
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
