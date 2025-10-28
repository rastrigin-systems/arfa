package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
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

	// Team filter (optional) - not currently used but supported by SQL
	// Use pgtype.UUID with Valid=false for no filter
	teamID := pgtype.UUID{Valid: false}
	// Future: add team_id query parameter support

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
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to count employees")
		return
	}

	// Convert db.Employee to api.Employee
	apiEmployees := make([]api.Employee, len(employees))
	for i, emp := range employees {
		apiEmployees[i] = dbEmployeeToAPI(emp)
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
	apiEmployee := dbEmployeeToAPI(employee)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiEmployee)
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
