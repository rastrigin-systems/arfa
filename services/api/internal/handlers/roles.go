package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rastrigin-systems/arfa/generated/api"
	"github.com/rastrigin-systems/arfa/generated/db"
)

// RolesHandler handles role management requests
type RolesHandler struct {
	db db.Querier
}

// NewRolesHandler creates a new roles handler
func NewRolesHandler(database db.Querier) *RolesHandler {
	return &RolesHandler{
		db: database,
	}
}

// Request/Response types (since not in OpenAPI spec yet)
type CreateRoleRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type UpdateRoleRequest struct {
	Name        *string   `json:"name,omitempty"`
	Permissions *[]string `json:"permissions,omitempty"`
}

// ListRoles handles GET /roles
// Returns list of all roles
func (h *RolesHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Query database for all roles
	roles, err := h.db.ListRoles(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch roles")
		return
	}

	// Convert to API response with employee counts
	apiRoles := make([]api.Role, len(roles))
	for i, role := range roles {
		apiRole := dbRoleToAPI(role)

		// Get employee count for this role
		employeeCount, err := h.db.CountEmployeesByRole(ctx, role.ID)
		if err == nil {
			count := int(employeeCount)
			apiRole.EmployeeCount = &count
		}

		apiRoles[i] = apiRole
	}

	// Build response
	response := struct {
		Roles []api.Role `json:"roles"`
		Total int        `json:"total"`
	}{
		Roles: apiRoles,
		Total: len(apiRoles),
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// GetRole handles GET /roles/{role_id}
// Returns a single role by ID
func (h *RolesHandler) GetRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get role ID from URL
	roleIDStr := chi.URLParam(r, "role_id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid role ID")
		return
	}

	// Query database
	role, err := h.db.GetRole(ctx, roleID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Role not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch role")
		return
	}

	// Write JSON response
	apiRole := dbRoleToAPI(role)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiRole)
}

// CreateRole handles POST /roles
// Creates a new role (admin only)
func (h *RolesHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Parse request body
	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" {
		writeError(w, http.StatusUnprocessableEntity, "Name is required")
		return
	}

	// Convert permissions to JSON
	permissionsJSON, err := json.Marshal(req.Permissions)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to encode permissions")
		return
	}

	// Create role in database
	role, err := h.db.CreateRole(ctx, db.CreateRoleParams{
		Name:        req.Name,
		Permissions: permissionsJSON,
	})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to create role")
		return
	}

	// Write JSON response
	apiRole := dbRoleToAPI(role)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(apiRole)
}

// UpdateRole handles PATCH /roles/{role_id}
// Updates an existing role (admin only)
func (h *RolesHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get role ID from URL
	roleIDStr := chi.URLParam(r, "role_id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid role ID")
		return
	}

	// Parse request body
	var req UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Build update params - use empty values if not provided (COALESCE in SQL will keep existing)
	name := ""
	if req.Name != nil {
		name = *req.Name
	}

	var permissionsJSON []byte
	if req.Permissions != nil {
		var err error
		permissionsJSON, err = json.Marshal(*req.Permissions)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to encode permissions")
			return
		}
	}

	params := db.UpdateRoleParams{
		ID:          roleID,
		Name:        name,
		Permissions: permissionsJSON,
	}

	// Update role in database
	role, err := h.db.UpdateRole(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Role not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to update role")
		return
	}

	// Write JSON response
	apiRole := dbRoleToAPI(role)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(apiRole)
}

// DeleteRole handles DELETE /roles/{role_id}
// Deletes a role (admin only)
func (h *RolesHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get role ID from URL
	roleIDStr := chi.URLParam(r, "role_id")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid role ID")
		return
	}

	// Delete role from database
	err = h.db.DeleteRole(ctx, roleID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to delete role")
		return
	}

	// Write no content response
	w.WriteHeader(http.StatusNoContent)
}

// dbRoleToAPI converts a database role to an API role
func dbRoleToAPI(role db.Role) api.Role {
	// Parse permissions JSON to string array
	var permissions []string
	if len(role.Permissions) > 0 {
		_ = json.Unmarshal(role.Permissions, &permissions)
	}

	// Convert UUIDs
	roleIDUUID := openapi_types.UUID(role.ID)

	// Convert timestamps
	var createdAt *time.Time
	if role.CreatedAt.Valid {
		createdAt = &role.CreatedAt.Time
	}

	return api.Role{
		Id:          &roleIDUUID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: &permissions,
		CreatedAt:   createdAt,
	}
}
