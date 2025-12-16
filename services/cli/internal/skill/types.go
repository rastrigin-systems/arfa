package skill

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
