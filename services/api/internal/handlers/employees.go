package handlers

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rastrigin-systems/arfa/generated/api"
	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/services/api/internal/auth"
)

// EmployeesHandler handles employee-related requests
type EmployeesHandler struct {
	db db.Querier
}

// NewEmployeesHandler creates a new employees handler
func NewEmployeesHandler(database db.Querier) *EmployeesHandler {
	return &EmployeesHandler{
		db: database,
	}
}

// ListEmployees handles GET /employees
// Returns paginated list of employees for the authenticated user's organization
func (h *EmployeesHandler) ListEmployees(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get org_id from context (set by JWT middleware)
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Organization ID not found in context")
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	// Status filter (optional)
	// Use pointer to string: nil = no filter, &value = filter by value
	var status *string
	if statusParam := query.Get("status"); statusParam != "" {
		status = &statusParam
	}

	// Team filter (optional)
	teamID := pgtype.UUID{Valid: false}
	if teamIDParam := query.Get("team_id"); teamIDParam != "" {
		if parsedTeamID, err := uuid.Parse(teamIDParam); err == nil {
			teamID = pgtype.UUID{Bytes: parsedTeamID, Valid: true}
		}
	}

	// Search filter (optional)
	var search *string
	if searchParam := query.Get("search"); searchParam != "" {
		search = &searchParam
	}

	// Pagination: limit (default 50, max 100)
	limit := 50
	if limitParam := query.Get("limit"); limitParam != "" {
		if parsedLimit, err := strconv.Atoi(limitParam); err == nil {
			if parsedLimit > 0 && parsedLimit <= 100 {
				limit = parsedLimit
			}
		}
	}

	// Pagination: offset (default 0)
	offset := 0
	if offsetParam := query.Get("offset"); offsetParam != "" {
		if parsedOffset, err := strconv.Atoi(offsetParam); err == nil {
			if parsedOffset >= 0 {
				offset = parsedOffset
			}
		}
	}

	// Query database for employees
	employees, err := h.db.ListEmployees(ctx, db.ListEmployeesParams{
		OrgID:       orgID,
		Status:      status,
		TeamID:      teamID,
		Search:      search,
		QueryLimit:  int32(limit),
		QueryOffset: int32(offset),
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch employees")
		return
	}

	// Query total count
	total, err := h.db.CountEmployees(ctx, db.CountEmployeesParams{
		OrgID:  orgID,
		Status: status,
		TeamID: teamID,
		Search: search,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to count employees")
		return
	}

	// Convert db.ListEmployeesRow to api.Employee
	apiEmployees := make([]api.Employee, len(employees))
	for i, emp := range employees {
		apiEmployees[i] = listEmployeesRowToAPI(emp)
	}

	// Build response
	response := api.ListEmployeesResponse{
		Employees: apiEmployees,
		Total:     total,
		Limit:     limit,
		Offset:    offset,
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetEmployee handles GET /employees/{id}
// Returns a single employee by ID with org isolation
func (h *EmployeesHandler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract employee_id from URL
	employeeIDStr := chi.URLParam(r, "employee_id")

	// Parse UUID
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid employee ID format")
		return
	}

	// Get org_id from context (set by JWT middleware)
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Organization ID not found in context")
		return
	}

	// Fetch employee from database
	employee, err := h.db.GetEmployee(ctx, employeeID)
	if err == pgx.ErrNoRows {
		writeError(w, http.StatusNotFound, "Employee not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch employee")
		return
	}

	// Verify org isolation - employee must belong to requesting org
	if employee.OrgID != orgID {
		// Return 404 (not 403) for security - don't reveal employee exists
		writeError(w, http.StatusNotFound, "Employee not found")
		return
	}

	// Convert to API type and return
	apiEmployee := dbGetEmployeeRowToAPI(employee)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiEmployee)
}

// CreateEmployee handles POST /employees
// Creates a new employee with auto-generated temporary password
func (h *EmployeesHandler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get org_id from context (set by JWT middleware)
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Organization ID not found in context")
		return
	}

	// Parse and validate request body
	var req api.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Email == "" {
		writeError(w, http.StatusUnprocessableEntity, "Email is required")
		return
	}
	if req.FullName == "" {
		writeError(w, http.StatusUnprocessableEntity, "Full name is required")
		return
	}
	if strings.TrimSpace(req.FullName) == "" {
		writeError(w, http.StatusUnprocessableEntity, "Full name cannot be empty")
		return
	}
	// Check if role_id is a valid (non-zero) UUID
	if uuid.UUID(req.RoleId) == uuid.Nil {
		writeError(w, http.StatusUnprocessableEntity, "Role ID is required")
		return
	}

	// Generate temporary password
	tempPassword, err := generateTempPassword(16)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to generate password")
		return
	}

	// Hash the password
	passwordHash, err := auth.HashPassword(tempPassword)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to hash password")
		return
	}

	// Convert team_id to pgtype.UUID
	var teamID pgtype.UUID
	if req.TeamId != nil {
		teamUUID := uuid.UUID(*req.TeamId)
		teamID = pgtype.UUID{
			Bytes: teamUUID,
			Valid: true,
		}
	}

	// Handle preferences
	preferences := []byte("{}")
	if req.Preferences != nil {
		prefsJSON, err := json.Marshal(req.Preferences)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid preferences format")
			return
		}
		preferences = prefsJSON
	}

	// Create employee in database
	employee, err := h.db.CreateEmployee(ctx, db.CreateEmployeeParams{
		OrgID:        orgID,
		TeamID:       teamID,
		RoleID:       uuid.UUID(req.RoleId),
		Email:        string(req.Email),
		FullName:     req.FullName,
		PasswordHash: passwordHash,
		Status:       "active", // New employees are active by default
		Preferences:  preferences,
	})

	if err != nil {
		// Check for unique constraint violation (duplicate email)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeError(w, http.StatusConflict, "Employee with this email already exists")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to create employee")
		return
	}

	// Convert to API type and return with temporary password
	apiEmployee := dbEmployeeToAPI(employee)
	response := api.CreateEmployeeResponse{
		Employee:          apiEmployee,
		TemporaryPassword: tempPassword,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// generateTempPassword generates a cryptographically secure random password
func generateTempPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	password := make([]byte, length)

	for i := range password {
		randomInt, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", err
		}
		password[i] = charset[randomInt.Int64()]
	}

	return string(password), nil
}

// UpdateEmployee handles PATCH /employees/{id}
// Updates employee with provided fields (all fields optional)
func (h *EmployeesHandler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract employee_id from URL
	employeeIDStr := chi.URLParam(r, "employee_id")

	// Parse UUID
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid employee ID format")
		return
	}

	// Get org_id from context (set by JWT middleware)
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Organization ID not found in context")
		return
	}

	// Fetch employee to verify org isolation
	existingEmployee, err := h.db.GetEmployee(ctx, employeeID)
	if err == pgx.ErrNoRows {
		writeError(w, http.StatusNotFound, "Employee not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch employee")
		return
	}

	// Verify org isolation
	if existingEmployee.OrgID != orgID {
		// Return 404 (not 403) for security - don't reveal employee exists
		writeError(w, http.StatusNotFound, "Employee not found")
		return
	}

	// Parse and validate request body
	var req api.UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Build update params, using existing values for fields not provided
	updateParams := db.UpdateEmployeeParams{
		ID:          employeeID,
		FullName:    existingEmployee.FullName,
		TeamID:      existingEmployee.TeamID,
		RoleID:      existingEmployee.RoleID,
		Status:      existingEmployee.Status,
		Preferences: existingEmployee.Preferences,
	}

	// Apply updates for provided fields
	if req.FullName != nil && *req.FullName != "" {
		updateParams.FullName = *req.FullName
	}

	if req.RoleId != nil {
		updateParams.RoleID = uuid.UUID(*req.RoleId)
	}

	if req.Status != nil {
		updateParams.Status = string(*req.Status)
	}

	// Handle team_id (can be set, removed, or left unchanged)
	if req.TeamId != nil {
		teamUUID := uuid.UUID(*req.TeamId)
		updateParams.TeamID = pgtype.UUID{
			Bytes: teamUUID,
			Valid: true,
		}
	}

	// Handle preferences
	if req.Preferences != nil {
		prefsJSON, err := json.Marshal(req.Preferences)
		if err != nil {
			writeError(w, http.StatusBadRequest, "Invalid preferences format")
			return
		}
		updateParams.Preferences = prefsJSON
	}

	// Update employee in database
	updatedEmployee, err := h.db.UpdateEmployee(ctx, updateParams)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to update employee")
		return
	}

	// Convert to API type and return
	apiEmployee := dbEmployeeToAPI(updatedEmployee)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiEmployee)
}

// DeleteEmployee handles DELETE /employees/{id}
// Soft deletes an employee (sets deleted_at)
func (h *EmployeesHandler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract employee_id from URL
	employeeIDStr := chi.URLParam(r, "employee_id")

	// Parse UUID
	employeeID, err := uuid.Parse(employeeIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid employee ID format")
		return
	}

	// Get org_id from context (set by JWT middleware)
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Organization ID not found in context")
		return
	}

	// Fetch employee to verify org isolation
	employee, err := h.db.GetEmployee(ctx, employeeID)
	if err == pgx.ErrNoRows {
		writeError(w, http.StatusNotFound, "Employee not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch employee")
		return
	}

	// Verify org isolation
	if employee.OrgID != orgID {
		// Return 404 (not 403) for security - don't reveal employee exists
		writeError(w, http.StatusNotFound, "Employee not found")
		return
	}

	// Hard delete employee
	if err := h.db.DeleteEmployee(ctx, employeeID); err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete employee")
		return
	}

	// Return 204 No Content
	w.WriteHeader(http.StatusNoContent)
}

// listEmployeesRowToAPI converts db.ListEmployeesRow to api.Employee
func listEmployeesRowToAPI(row db.ListEmployeesRow) api.Employee {
	apiEmp := api.Employee{
		Id:       &row.ID,
		OrgId:    row.OrgID,
		Email:    openapi_types.Email(row.Email),
		FullName: row.FullName,
		RoleId:   row.RoleID,
		Status:   api.EmployeeStatus(row.Status),
	}

	// Handle nullable fields
	if row.TeamID.Valid {
		teamUUID := row.TeamID.Bytes
		apiEmp.TeamId = (*openapi_types.UUID)(&teamUUID)
	}

	// Add team name if available (from JOIN)
	if row.TeamName != nil {
		apiEmp.TeamName = row.TeamName
	}

	if row.CreatedAt.Valid {
		apiEmp.CreatedAt = &row.CreatedAt.Time
	}

	if row.UpdatedAt.Valid {
		apiEmp.UpdatedAt = &row.UpdatedAt.Time
	}

	if row.LastLoginAt.Valid {
		apiEmp.LastLoginAt = &row.LastLoginAt.Time
	}

	// Handle preferences JSON
	if len(row.Preferences) > 0 && string(row.Preferences) != "null" {
		var prefs map[string]interface{}
		if err := json.Unmarshal(row.Preferences, &prefs); err == nil {
			apiEmp.Preferences = &prefs
		}
	}

	return apiEmp
}

// dbEmployeeToAPI converts db.Employee to api.Employee
func dbEmployeeToAPI(emp db.Employee) api.Employee {
	apiEmp := api.Employee{
		Id:       &emp.ID,
		OrgId:    emp.OrgID,
		Email:    openapi_types.Email(emp.Email),
		FullName: emp.FullName,
		RoleId:   emp.RoleID,
		Status:   api.EmployeeStatus(emp.Status),
	}

	// Handle nullable fields
	if emp.TeamID.Valid {
		teamUUID := emp.TeamID.Bytes
		apiEmp.TeamId = (*openapi_types.UUID)(&teamUUID)
	}

	if emp.CreatedAt.Valid {
		apiEmp.CreatedAt = &emp.CreatedAt.Time
	}

	if emp.UpdatedAt.Valid {
		apiEmp.UpdatedAt = &emp.UpdatedAt.Time
	}

	if emp.LastLoginAt.Valid {
		apiEmp.LastLoginAt = &emp.LastLoginAt.Time
	}

	// Handle preferences JSON
	if len(emp.Preferences) > 0 && string(emp.Preferences) != "null" {
		var prefs map[string]interface{}
		if err := json.Unmarshal(emp.Preferences, &prefs); err == nil {
			apiEmp.Preferences = &prefs
		}
	}

	return apiEmp
}

// dbGetEmployeeRowToAPI converts db.GetEmployeeRow to api.Employee (includes team_name)
func dbGetEmployeeRowToAPI(emp db.GetEmployeeRow) api.Employee {
	apiEmp := api.Employee{
		Id:       &emp.ID,
		OrgId:    emp.OrgID,
		Email:    openapi_types.Email(emp.Email),
		FullName: emp.FullName,
		RoleId:   emp.RoleID,
		Status:   api.EmployeeStatus(emp.Status),
	}

	// Handle nullable fields
	if emp.TeamID.Valid {
		teamUUID := emp.TeamID.Bytes
		apiEmp.TeamId = (*openapi_types.UUID)(&teamUUID)
	}

	// Handle team_name
	if emp.TeamName != nil {
		apiEmp.TeamName = emp.TeamName
	}

	if emp.CreatedAt.Valid {
		apiEmp.CreatedAt = &emp.CreatedAt.Time
	}

	if emp.UpdatedAt.Valid {
		apiEmp.UpdatedAt = &emp.UpdatedAt.Time
	}

	if emp.LastLoginAt.Valid {
		apiEmp.LastLoginAt = &emp.LastLoginAt.Time
	}

	// Handle preferences JSON
	if len(emp.Preferences) > 0 && string(emp.Preferences) != "null" {
		var prefs map[string]interface{}
		if err := json.Unmarshal(emp.Preferences, &prefs); err == nil {
			apiEmp.Preferences = &prefs
		}
	}

	return apiEmp
}
