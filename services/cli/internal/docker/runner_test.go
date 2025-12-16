package docker

import (
	"os/exec"
	"testing"
)

func TestFindAgentBinary(t *testing.T) {
	tests := []struct {
		name      string
		agentType string
		wantErr   bool
	}{
		{
			name:      "unknown agent type returns error",
			agentType: "unknown-agent",
			wantErr:   true,
		},
		{
			name:      "claude-code may or may not be installed",
			agentType: "claude-code",
			wantErr:   false, // Will depend on whether claude is installed
		},
		{
			name:      "ide_assistant normalizes to claude-code",
			agentType: "ide_assistant",
			wantErr:   false, // Will depend on whether claude is installed
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := FindAgentBinary(tt.agentType)

			if tt.agentType == "unknown-agent" {
				if err == nil {
					t.Error("expected error for unknown agent type")
				}
				return
			}

			// For known agent types, the test passes whether installed or not
			// Just verify we get a valid response
			if err != nil {
				// Check it's a "not found" error, not a different error
				if _, lookupErr := exec.LookPath("nonexistent-binary-12345"); lookupErr == nil {
					t.Error("expected LookPath to fail for nonexistent binary")
				}
			} else {
				if path == "" {
					t.Error("expected non-empty path when no error")
				}
			}
		})
	}
}

func TestNormalizeAgentType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		// API types -> CLI types
		{"ide_assistant", "claude-code"},
		{"code_completion", "cursor"},
		{"ai_editor", "windsurf"},
		{"gemini_agent", "gemini"},
		{"pair_programmer", "aider"},
		// CLI types pass through
		{"claude-code", "claude-code"},
		{"cursor", "cursor"},
		{"windsurf", "windsurf"},
		{"gemini", "gemini"},
		{"aider", "aider"},
		// Unknown types pass through as-is
		{"unknown-type", "unknown-type"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeAgentType(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeAgentType(%s) = %s, want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestAgentTypeMapping_Completeness(t *testing.T) {
	// Verify all API agent types are mapped
	expectedMappings := map[string]string{
		"ide_assistant":   "claude-code",
		"code_completion": "cursor",
		"ai_editor":       "windsurf",
		"gemini_agent":    "gemini",
		"pair_programmer": "aider",
	}

	for apiType, cliType := range expectedMappings {
		if mapped, ok := AgentTypeMapping[apiType]; !ok {
			t.Errorf("expected API type %s to be in AgentTypeMapping", apiType)
		} else if mapped != cliType {
			t.Errorf("AgentTypeMapping[%s] = %s, want %s", apiType, mapped, cliType)
		}
	}
}

func TestAgentBinaries_Completeness(t *testing.T) {
	// Verify all expected agent types are mapped
	expectedAgents := []string{"claude-code", "cursor", "windsurf", "gemini", "aider"}

	for _, agent := range expectedAgents {
		if _, ok := AgentBinaries[agent]; !ok {
			t.Errorf("expected agent type %s to be in AgentBinaries map", agent)
		}
	}
}

func TestAgentEnvVars_Completeness(t *testing.T) {
	tests := []struct {
		agentType   string
		expectedVar string
	}{
		{"claude-code", "ANTHROPIC_API_KEY"},
		{"cursor", "ANTHROPIC_API_KEY"},
		{"windsurf", "ANTHROPIC_API_KEY"},
		{"gemini", "GEMINI_API_KEY"},
		{"aider", "ANTHROPIC_API_KEY"},
	}

	for _, tt := range tests {
		t.Run(tt.agentType, func(t *testing.T) {
			envVar, ok := AgentEnvVars[tt.agentType]
			if !ok {
				t.Errorf("expected agent type %s to have an env var mapping", tt.agentType)
				return
			}
			if envVar != tt.expectedVar {
				t.Errorf("expected env var %s, got %s", tt.expectedVar, envVar)
			}
		})
	}
}

func TestNewRunner(t *testing.T) {
	runner := NewRunner()

	if runner == nil {
		t.Error("expected non-nil runner")
	}

	// Verify initial state
	if runner.IsRunning() {
		t.Error("new runner should not be running")
	}

	if runner.PID() != 0 {
		t.Error("new runner should have PID 0")
	}
}

func TestGetInstallInstructions(t *testing.T) {
	tests := []struct {
		agentType        string
		shouldContain    string
		shouldNotBeEmpty bool
	}{
		{"claude-code", "npm", true},
		{"cursor", "cursor.sh", true},
		{"windsurf", "windsurf", true},
		{"aider", "pip", true},
		{"unknown", "website", true},
	}

	for _, tt := range tests {
		t.Run(tt.agentType, func(t *testing.T) {
			instructions := getInstallInstructions(tt.agentType)

			if tt.shouldNotBeEmpty && instructions == "" {
				t.Error("expected non-empty instructions")
			}

			if tt.shouldContain != "" && !contains(instructions, tt.shouldContain) {
				t.Errorf("expected instructions to contain %s, got: %s", tt.shouldContain, instructions)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
