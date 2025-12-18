package skill

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/config"
)

// Service handles skill management operations
type Service struct {
	client        APIClientInterface
	configManager ConfigManagerInterface
}

// NewService creates a new skill service with concrete types.
// This is the primary constructor for production use.
func NewService(client *api.Client, configManager *config.Manager) *Service {
	return &Service{
		client:        client,
		configManager: configManager,
	}
}

// NewServiceWithInterfaces creates a new skill service with interface types.
// This constructor enables dependency injection for testing with mocks.
func NewServiceWithInterfaces(client APIClientInterface, configManager ConfigManagerInterface) *Service {
	return &Service{
		client:        client,
		configManager: configManager,
	}
}

// ListCatalogSkills fetches all available skills from the platform catalog
func (s *Service) ListCatalogSkills(ctx context.Context) ([]api.Skill, error) {
	resp, err := s.client.ListSkills(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list catalog skills: %w", err)
	}
	return resp.Skills, nil
}

// GetSkill fetches details for a specific skill from the catalog
func (s *Service) GetSkill(ctx context.Context, skillID string) (*api.Skill, error) {
	skill, err := s.client.GetSkill(ctx, skillID)
	if err != nil {
		return nil, fmt.Errorf("failed to get skill: %w", err)
	}
	return skill, nil
}

// GetSkillByName fetches a skill by name (searches catalog)
func (s *Service) GetSkillByName(ctx context.Context, name string) (*api.Skill, error) {
	skills, err := s.ListCatalogSkills(ctx)
	if err != nil {
		return nil, err
	}

	for _, skill := range skills {
		if skill.Name == name {
			return &skill, nil
		}
	}

	return nil, fmt.Errorf("skill '%s' not found in catalog", name)
}

// ListEmployeeSkills fetches skills assigned to the authenticated employee
func (s *Service) ListEmployeeSkills(ctx context.Context) ([]api.EmployeeSkill, error) {
	resp, err := s.client.ListEmployeeSkills(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list employee skills: %w", err)
	}
	return resp.Skills, nil
}

// GetEmployeeSkill fetches a specific skill assigned to the employee
func (s *Service) GetEmployeeSkill(ctx context.Context, skillID string) (*api.EmployeeSkill, error) {
	skill, err := s.client.GetEmployeeSkill(ctx, skillID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee skill: %w", err)
	}
	return skill, nil
}

// GetEmployeeSkillByName fetches an employee skill by name
func (s *Service) GetEmployeeSkillByName(ctx context.Context, name string) (*api.EmployeeSkill, error) {
	skills, err := s.ListEmployeeSkills(ctx)
	if err != nil {
		return nil, err
	}

	for _, skill := range skills {
		if skill.Name == name {
			return &skill, nil
		}
	}

	return nil, fmt.Errorf("skill '%s' not found in your assigned skills", name)
}

// GetLocalSkills returns locally installed skills from .claude/skills/
func (s *Service) GetLocalSkills() ([]LocalSkillInfo, error) {
	// Get current working directory (or could be configurable)
	cwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get current directory: %w", err)
	}

	skillsDir := filepath.Join(cwd, ".claude", "skills")

	// Check if skills directory exists
	if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
		return []LocalSkillInfo{}, nil
	}

	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read skills directory: %w", err)
	}

	var localSkills []LocalSkillInfo
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		skillName := entry.Name()
		skillPath := filepath.Join(skillsDir, skillName)

		// Check for SKILL.md file
		skillMdPath := filepath.Join(skillPath, "SKILL.md")
		if _, err := os.Stat(skillMdPath); os.IsNotExist(err) {
			// Skip if no SKILL.md found
			continue
		}

		// List all files in skill directory
		var files []string
		skillFiles, err := os.ReadDir(skillPath)
		if err != nil {
			continue
		}
		for _, f := range skillFiles {
			if !f.IsDir() {
				files = append(files, f.Name())
			}
		}

		// Try to read metadata.json if it exists
		metadataPath := filepath.Join(skillPath, "metadata.json")
		var metadata struct {
			Version     string `json:"version,omitempty"`
			Description string `json:"description,omitempty"`
			Category    string `json:"category,omitempty"`
			IsEnabled   bool   `json:"is_enabled,omitempty"`
			InstalledAt string `json:"installed_at,omitempty"`
		}

		if metadataBytes, err := os.ReadFile(metadataPath); err == nil {
			json.Unmarshal(metadataBytes, &metadata)
		}

		localSkills = append(localSkills, LocalSkillInfo{
			Name:        skillName,
			Version:     metadata.Version,
			Description: metadata.Description,
			Category:    metadata.Category,
			Files:       files,
			IsEnabled:   metadata.IsEnabled,
			InstalledAt: metadata.InstalledAt,
		})
	}

	return localSkills, nil
}

// GetLocalSkill returns details for a specific locally installed skill
func (s *Service) GetLocalSkill(name string) (*LocalSkillInfo, error) {
	skills, err := s.GetLocalSkills()
	if err != nil {
		return nil, err
	}

	for _, skill := range skills {
		if skill.Name == name {
			return &skill, nil
		}
	}

	return nil, fmt.Errorf("skill '%s' not found locally", name)
}
