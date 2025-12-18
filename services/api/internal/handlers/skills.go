package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"github.com/rastrigin-systems/ubik-enterprise/generated/api"
	"github.com/rastrigin-systems/ubik-enterprise/generated/db"
	"github.com/rastrigin-systems/ubik-enterprise/services/api/internal/middleware"
)

// SkillsHandler handles skill-related requests
type SkillsHandler struct {
	db db.Querier
}

// NewSkillsHandler creates a new skills handler
func NewSkillsHandler(database db.Querier) *SkillsHandler {
	return &SkillsHandler{
		db: database,
	}
}

// ============================================================================
// Skills Catalog Endpoints
// ============================================================================

// ListSkills handles GET /skills
// Returns list of all available skills from the catalog
func (h *SkillsHandler) ListSkills(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Query database for all active skills
	skills, err := h.db.ListSkills(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch skills")
		return
	}

	// Convert db.SkillCatalog to api.Skill
	apiSkills := make([]api.Skill, len(skills))
	for i, skill := range skills {
		apiSkills[i] = dbSkillToAPI(skill)
	}

	// Build response
	response := api.ListSkillsResponse{
		Skills: apiSkills,
		Total:  len(apiSkills),
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetSkill handles GET /skills/{skill_id}
// Returns a specific skill by ID
func (h *SkillsHandler) GetSkill(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Extract skill_id from URL path
	skillIDStr := chi.URLParam(r, "skill_id")
	skillID, err := uuid.Parse(skillIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid skill ID")
		return
	}

	// Query database
	skill, err := h.db.GetSkill(ctx, skillID)
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "Skill not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch skill")
		return
	}

	// Convert to API response
	apiSkill := dbSkillToAPI(skill)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiSkill)
}

// ============================================================================
// Employee Skills Endpoints
// ============================================================================

// ListEmployeeSkills handles GET /employees/me/skills
// Returns list of skills assigned to the authenticated employee
func (h *SkillsHandler) ListEmployeeSkills(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get employee ID from JWT context
	employeeID, err := middleware.GetEmployeeID(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get employee context")
		return
	}

	// Query database for employee's skills
	empSkills, err := h.db.ListEmployeeSkills(ctx, employeeID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to fetch employee skills")
		return
	}

	// Convert to API response
	apiSkills := make([]api.EmployeeSkill, len(empSkills))
	for i, skill := range empSkills {
		apiSkills[i] = dbEmployeeSkillToAPI(skill)
	}

	// Build response
	response := api.ListEmployeeSkillsResponse{
		Skills: apiSkills,
		Total:  len(apiSkills),
	}

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetEmployeeSkill handles GET /employees/me/skills/{skill_id}
// Returns a specific skill assigned to the authenticated employee
func (h *SkillsHandler) GetEmployeeSkill(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get employee ID from JWT context
	employeeID, err := middleware.GetEmployeeID(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "Failed to get employee context")
		return
	}

	// Extract skill_id from URL path
	skillIDStr := chi.URLParam(r, "skill_id")
	skillID, err := uuid.Parse(skillIDStr)
	if err != nil {
		writeError(w, http.StatusBadRequest, "Invalid skill ID")
		return
	}

	// Query database
	empSkill, err := h.db.GetEmployeeSkill(ctx, db.GetEmployeeSkillParams{
		EmployeeID: employeeID,
		SkillID:    skillID,
	})
	if err != nil {
		if err == sql.ErrNoRows {
			writeError(w, http.StatusNotFound, "Skill not found or not assigned to employee")
			return
		}
		writeError(w, http.StatusInternalServerError, "Failed to fetch employee skill")
		return
	}

	// Convert to API response
	apiSkill := dbGetEmployeeSkillRowToAPI(empSkill)

	// Write JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(apiSkill)
}

// ============================================================================
// Helper Functions - DB to API Converters
// ============================================================================

// dbSkillToAPI converts db.SkillCatalog to api.Skill
func dbSkillToAPI(skill db.SkillCatalog) api.Skill {
	apiSkill := api.Skill{
		Id:      (*openapi_types.UUID)(&skill.ID),
		Name:    skill.Name,
		Version: skill.Version,
	}

	// Handle nullable description and category
	if skill.Description != nil {
		apiSkill.Description = *skill.Description
	}
	if skill.Category != nil {
		apiSkill.Category = *skill.Category
	}
	if skill.IsActive != nil {
		apiSkill.IsActive = *skill.IsActive
	}

	// Handle created_at / updated_at
	if skill.CreatedAt.Valid {
		apiSkill.CreatedAt = &skill.CreatedAt.Time
	}
	if skill.UpdatedAt.Valid {
		apiSkill.UpdatedAt = &skill.UpdatedAt.Time
	}

	// Parse files JSON
	if len(skill.Files) > 0 {
		var files []map[string]interface{}
		if err := json.Unmarshal(skill.Files, &files); err == nil {
			filesArray := make([]struct {
				Content *string `json:"content,omitempty"`
				Path    *string `json:"path,omitempty"`
			}, len(files))
			for i, f := range files {
				if path, ok := f["path"].(string); ok {
					filesArray[i].Path = &path
				}
				if content, ok := f["content"].(string); ok {
					filesArray[i].Content = &content
				}
			}
			apiSkill.Files = filesArray
		}
	}

	// Parse dependencies JSON
	if len(skill.Dependencies) > 0 && string(skill.Dependencies) != "null" {
		var deps struct {
			McpServers *[]string `json:"mcp_servers,omitempty"`
			Skills     *[]string `json:"skills,omitempty"`
		}
		if err := json.Unmarshal(skill.Dependencies, &deps); err == nil {
			apiSkill.Dependencies = &deps
		}
	}

	return apiSkill
}

// dbEmployeeSkillToAPI converts db.ListEmployeeSkillsRow to api.EmployeeSkill
func dbEmployeeSkillToAPI(skill db.ListEmployeeSkillsRow) api.EmployeeSkill {
	apiSkill := api.EmployeeSkill{
		Id:      (*openapi_types.UUID)(&skill.ID),
		Name:    skill.Name,
		Version: skill.Version,
	}

	// Handle nullable fields
	if skill.Description != nil {
		apiSkill.Description = *skill.Description
	}
	if skill.Category != nil {
		apiSkill.Category = *skill.Category
	}
	if skill.IsActive != nil {
		apiSkill.IsActive = *skill.IsActive
	}
	if skill.IsEnabled != nil {
		apiSkill.IsEnabled = *skill.IsEnabled
	}
	if skill.InstalledAt.Valid {
		apiSkill.InstalledAt = &skill.InstalledAt.Time
	}

	// Parse files JSON
	if len(skill.Files) > 0 {
		var files []map[string]interface{}
		if err := json.Unmarshal(skill.Files, &files); err == nil {
			filesArray := make([]struct {
				Content *string `json:"content,omitempty"`
				Path    *string `json:"path,omitempty"`
			}, len(files))
			for i, f := range files {
				if path, ok := f["path"].(string); ok {
					filesArray[i].Path = &path
				}
				if content, ok := f["content"].(string); ok {
					filesArray[i].Content = &content
				}
			}
			apiSkill.Files = filesArray
		}
	}

	// Parse dependencies JSON
	if len(skill.Dependencies) > 0 && string(skill.Dependencies) != "null" {
		var deps struct {
			McpServers *[]string `json:"mcp_servers,omitempty"`
			Skills     *[]string `json:"skills,omitempty"`
		}
		if err := json.Unmarshal(skill.Dependencies, &deps); err == nil {
			apiSkill.Dependencies = &deps
		}
	}

	// Parse config JSON
	if len(skill.Config) > 0 && string(skill.Config) != "null" {
		var config map[string]interface{}
		if err := json.Unmarshal(skill.Config, &config); err == nil {
			apiSkill.Config = &config
		}
	}

	return apiSkill
}

// dbGetEmployeeSkillRowToAPI converts db.GetEmployeeSkillRow to api.EmployeeSkill
func dbGetEmployeeSkillRowToAPI(skill db.GetEmployeeSkillRow) api.EmployeeSkill {
	apiSkill := api.EmployeeSkill{
		Id:      (*openapi_types.UUID)(&skill.ID),
		Name:    skill.Name,
		Version: skill.Version,
	}

	// Handle nullable fields
	if skill.Description != nil {
		apiSkill.Description = *skill.Description
	}
	if skill.Category != nil {
		apiSkill.Category = *skill.Category
	}
	if skill.IsActive != nil {
		apiSkill.IsActive = *skill.IsActive
	}
	if skill.IsEnabled != nil {
		apiSkill.IsEnabled = *skill.IsEnabled
	}
	if skill.InstalledAt.Valid {
		apiSkill.InstalledAt = &skill.InstalledAt.Time
	}

	// Parse files JSON
	if len(skill.Files) > 0 {
		var files []map[string]interface{}
		if err := json.Unmarshal(skill.Files, &files); err == nil {
			filesArray := make([]struct {
				Content *string `json:"content,omitempty"`
				Path    *string `json:"path,omitempty"`
			}, len(files))
			for i, f := range files {
				if path, ok := f["path"].(string); ok {
					filesArray[i].Path = &path
				}
				if content, ok := f["content"].(string); ok {
					filesArray[i].Content = &content
				}
			}
			apiSkill.Files = filesArray
		}
	}

	// Parse dependencies JSON
	if len(skill.Dependencies) > 0 && string(skill.Dependencies) != "null" {
		var deps struct {
			McpServers *[]string `json:"mcp_servers,omitempty"`
			Skills     *[]string `json:"skills,omitempty"`
		}
		if err := json.Unmarshal(skill.Dependencies, &deps); err == nil {
			apiSkill.Dependencies = &deps
		}
	}

	// Parse config JSON
	if len(skill.Config) > 0 && string(skill.Config) != "null" {
		var config map[string]interface{}
		if err := json.Unmarshal(skill.Config, &config); err == nil {
			apiSkill.Config = &config
		}
	}

	return apiSkill
}
