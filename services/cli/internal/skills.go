package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/api"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/config"
)

// SkillsService handles skill management operations
type SkillsService struct {
	client        *api.Client
	configManager *config.Manager
}

// NewSkillsService creates a new skills service
func NewSkillsService(client *api.Client, configManager *config.Manager) *SkillsService {
	return &SkillsService{
		client:        client,
		configManager: configManager,
	}
}

// LocalSkillInfo represents locally installed skill information
type LocalSkillInfo struct {
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description,omitempty"`
	Category    string   `json:"category,omitempty"`
	Files       []string `json:"files"`
	IsEnabled   bool     `json:"is_enabled"`
	InstalledAt string   `json:"installed_at,omitempty"`
}

// ListCatalogSkills fetches all available skills from the platform catalog
func (ss *SkillsService) ListCatalogSkills(ctx context.Context) ([]api.Skill, error) {
	resp, err := ss.client.ListSkills(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list catalog skills: %w", err)
	}
	return resp.Skills, nil
}

// GetSkill fetches details for a specific skill from the catalog
func (ss *SkillsService) GetSkill(ctx context.Context, skillID string) (*api.Skill, error) {
	skill, err := ss.client.GetSkill(ctx, skillID)
	if err != nil {
		return nil, fmt.Errorf("failed to get skill: %w", err)
	}
	return skill, nil
}

// GetSkillByName fetches a skill by name (searches catalog)
func (ss *SkillsService) GetSkillByName(ctx context.Context, name string) (*api.Skill, error) {
	skills, err := ss.ListCatalogSkills(ctx)
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
func (ss *SkillsService) ListEmployeeSkills(ctx context.Context) ([]api.EmployeeSkill, error) {
	resp, err := ss.client.ListEmployeeSkills(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list employee skills: %w", err)
	}
	return resp.Skills, nil
}

// GetEmployeeSkill fetches a specific skill assigned to the employee
func (ss *SkillsService) GetEmployeeSkill(ctx context.Context, skillID string) (*api.EmployeeSkill, error) {
	skill, err := ss.client.GetEmployeeSkill(ctx, skillID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee skill: %w", err)
	}
	return skill, nil
}

// GetEmployeeSkillByName fetches an employee skill by name
func (ss *SkillsService) GetEmployeeSkillByName(ctx context.Context, name string) (*api.EmployeeSkill, error) {
	skills, err := ss.ListEmployeeSkills(ctx)
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
func (ss *SkillsService) GetLocalSkills() ([]LocalSkillInfo, error) {
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
func (ss *SkillsService) GetLocalSkill(name string) (*LocalSkillInfo, error) {
	skills, err := ss.GetLocalSkills()
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
