package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/sergeirastrigin/ubik-enterprise/generated/api"
	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
)

// OrganizationsHandler handles organization management requests
type OrganizationsHandler struct {
	db db.Querier
}

// NewOrganizationsHandler creates a new organizations handler
func NewOrganizationsHandler(database db.Querier) *OrganizationsHandler {
	return &OrganizationsHandler{
		db: database,
	}
}

// Request/Response types (since not in OpenAPI spec yet)
type UpdateOrganizationRequest struct {
	Name                 *string                 `json:"name,omitempty"`
	Settings             *map[string]interface{} `json:"settings,omitempty"`
	MaxEmployees         *int32                  `json:"max_employees,omitempty"`
	MaxAgentsPerEmployee *int32                  `json:"max_agents_per_employee,omitempty"`
}

// GetCurrentOrganization handles GET /organizations/current
// Returns the current organization details
func (h *OrganizationsHandler) GetCurrentOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get org ID from context (set by auth middleware)
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Query database
	org, err := h.db.GetOrganization(ctx, orgID)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Organization not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch organization")
		return
	}

	// Write JSON response
	apiOrg := dbOrganizationToAPI(org)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiOrg)
}

// UpdateCurrentOrganization handles PATCH /organizations/current
// Updates the current organization settings (admin only)
func (h *OrganizationsHandler) UpdateCurrentOrganization(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get org ID from context
	orgID, err := GetOrgID(ctx)
	if err != nil {
		writeError(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	// Parse request body
	var req UpdateOrganizationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Build update params - use empty/zero values if not provided (COALESCE in SQL will keep existing)
	name := ""
	if req.Name != nil {
		name = *req.Name
	}

	var settingsJSON []byte
	if req.Settings != nil {
		var err error
		settingsJSON, err = json.Marshal(*req.Settings)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "Failed to encode settings")
			return
		}
	}

	maxEmployees := int32(0)
	if req.MaxEmployees != nil {
		maxEmployees = *req.MaxEmployees
	}

	maxAgentsPerEmployee := int32(0)
	if req.MaxAgentsPerEmployee != nil {
		maxAgentsPerEmployee = *req.MaxAgentsPerEmployee
	}

	params := db.UpdateOrganizationParams{
		ID:                   orgID,
		Name:                 name,
		Settings:             settingsJSON,
		MaxEmployees:         maxEmployees,
		MaxAgentsPerEmployee: maxAgentsPerEmployee,
	}

	// Update organization in database
	org, err := h.db.UpdateOrganization(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			writeError(w, http.StatusNotFound, "Organization not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to update organization")
		return
	}

	// Write JSON response
	apiOrg := dbOrganizationToAPI(org)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiOrg)
}

// dbOrganizationToAPI converts a database organization to an API organization
func dbOrganizationToAPI(org db.Organization) api.Organization {
	// Parse settings JSON
	var settings map[string]interface{}
	if len(org.Settings) > 0 {
		json.Unmarshal(org.Settings, &settings)
	}

	// Convert UUIDs
	orgIDUUID := openapi_types.UUID(org.ID)

	// Convert timestamps
	var createdAt, updatedAt *time.Time
	if org.CreatedAt.Valid {
		createdAt = &org.CreatedAt.Time
	}
	if org.UpdatedAt.Valid {
		updatedAt = &org.UpdatedAt.Time
	}

	// Convert int32 to *int
	maxEmployees := int(org.MaxEmployees)
	maxAgentsPerEmployee := int(org.MaxAgentsPerEmployee)

	return api.Organization{
		Id:                   &orgIDUUID,
		Name:                 org.Name,
		Slug:                 org.Slug,
		Plan:                 api.OrganizationPlan(org.Plan),
		Settings:             &settings,
		MaxEmployees:         &maxEmployees,
		MaxAgentsPerEmployee: &maxAgentsPerEmployee,
		CreatedAt:            createdAt,
		UpdatedAt:            updatedAt,
	}
}
