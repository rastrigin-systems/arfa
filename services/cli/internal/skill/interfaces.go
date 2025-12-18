package skill

import (
	"context"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/config"
)

// ============================================================================
// Dependency Interfaces
// ============================================================================
// These interfaces define what the skill package needs from other packages.
// This enables dependency injection and testability.
// ============================================================================

// APIClientInterface defines what skill needs from api.Client
type APIClientInterface interface {
	// ListSkills fetches all available skills from the catalog.
	ListSkills(ctx context.Context) (*api.ListSkillsResponse, error)
	// GetSkill fetches details for a specific skill by ID.
	GetSkill(ctx context.Context, skillID string) (*api.Skill, error)
	// ListEmployeeSkills fetches skills assigned to the authenticated employee.
	ListEmployeeSkills(ctx context.Context) (*api.ListEmployeeSkillsResponse, error)
	// GetEmployeeSkill fetches a specific skill assigned to the authenticated employee.
	GetEmployeeSkill(ctx context.Context, skillID string) (*api.EmployeeSkill, error)
}

// ConfigManagerInterface defines what skill needs from config.Manager
type ConfigManagerInterface interface {
	Load() (*config.Config, error)
	Save(cfg *config.Config) error
	GetConfigPath() string
}

// ============================================================================
// Service Interface
// ============================================================================

// ServiceInterface defines the contract for skill management operations.
// Implementations handle listing skills from catalog and local storage.
type ServiceInterface interface {
	// ListCatalogSkills fetches all available skills from the platform catalog.
	ListCatalogSkills(ctx context.Context) ([]api.Skill, error)

	// GetSkill fetches details for a specific skill from the catalog.
	GetSkill(ctx context.Context, skillID string) (*api.Skill, error)

	// GetSkillByName fetches a skill by name (searches catalog).
	GetSkillByName(ctx context.Context, name string) (*api.Skill, error)

	// ListEmployeeSkills fetches skills assigned to the authenticated employee.
	ListEmployeeSkills(ctx context.Context) ([]api.EmployeeSkill, error)

	// GetEmployeeSkill fetches a specific skill assigned to the employee.
	GetEmployeeSkill(ctx context.Context, skillID string) (*api.EmployeeSkill, error)

	// GetEmployeeSkillByName fetches an employee skill by name.
	GetEmployeeSkillByName(ctx context.Context, name string) (*api.EmployeeSkill, error)

	// GetLocalSkills returns locally installed skills from .claude/skills/.
	GetLocalSkills() ([]LocalSkillInfo, error)

	// GetLocalSkill returns details for a specific locally installed skill.
	GetLocalSkill(name string) (*LocalSkillInfo, error)
}

// Compile-time interface implementation check
var _ ServiceInterface = (*Service)(nil)
