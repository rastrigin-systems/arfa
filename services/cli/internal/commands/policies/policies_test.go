package policies

import (
	"bytes"
	"os"
	"testing"

	"github.com/rastrigin-systems/arfa/services/cli/internal/api"
	"github.com/rastrigin-systems/arfa/services/cli/internal/container"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPoliciesCommand(t *testing.T) {
	c := container.New()
	cmd := NewPoliciesCommand(c)

	assert.Equal(t, "policies", cmd.Use)
	assert.Equal(t, "Manage tool policies", cmd.Short)

	// Should have list subcommand
	listCmd, _, err := cmd.Find([]string{"list"})
	require.NoError(t, err)
	assert.Equal(t, "list", listCmd.Use)
}

func TestNewListCommand(t *testing.T) {
	c := container.New()
	cmd := NewListCommand(c)

	assert.Equal(t, "list", cmd.Use)
	assert.Equal(t, "List active tool policies", cmd.Short)

	// Should have --all and --json flags
	allFlag := cmd.Flags().Lookup("all")
	require.NotNil(t, allFlag)
	assert.Equal(t, "false", allFlag.DefValue)

	jsonFlag := cmd.Flags().Lookup("json")
	require.NotNil(t, jsonFlag)
	assert.Equal(t, "false", jsonFlag.DefValue)
}

func TestListCommand_NoPoliciesFile(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	c := container.New()
	cmd := NewListCommand(c)

	// Capture output
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "No tool policies found")
	assert.Contains(t, output, "arfa sync")
}

func TestListCommand_EmptyPolicies(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create empty policies file
	arfaDir := tempDir + "/.arfa"
	os.MkdirAll(arfaDir, 0700)
	os.WriteFile(arfaDir+"/policies.json", []byte(`{"policies":[],"version":1,"synced_at":"2024-01-15T10:00:00Z"}`), 0600)

	c := container.New()
	cmd := NewListCommand(c)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "No tool policies are configured")
}

func TestListCommand_WithPolicies(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create policies file with some policies
	arfaDir := tempDir + "/.arfa"
	os.MkdirAll(arfaDir, 0700)
	cacheContent := `{
		"policies": [
			{"tool_name": "Bash", "action": "deny", "reason": "Shell blocked", "scope": "organization"},
			{"tool_name": "Write", "action": "audit", "reason": "Audited"}
		],
		"version": 12345,
		"synced_at": "2024-01-15T10:00:00Z"
	}`
	os.WriteFile(arfaDir+"/policies.json", []byte(cacheContent), 0600)

	c := container.New()
	cmd := NewListCommand(c)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Should only show deny policies by default
	assert.Contains(t, output, "Bash")
	assert.Contains(t, output, "DENY")
	assert.Contains(t, output, "Shell blocked")
	// Write should not be shown (audit only)
	assert.NotContains(t, output, "Write")
}

func TestListCommand_WithAllFlag(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	// Create policies file with both deny and audit policies
	arfaDir := tempDir + "/.arfa"
	os.MkdirAll(arfaDir, 0700)
	cacheContent := `{
		"policies": [
			{"tool_name": "Bash", "action": "deny", "reason": "Shell blocked"},
			{"tool_name": "Write", "action": "audit", "reason": "Writes audited"}
		],
		"version": 12345,
		"synced_at": "2024-01-15T10:00:00Z"
	}`
	os.WriteFile(arfaDir+"/policies.json", []byte(cacheContent), 0600)

	c := container.New()
	cmd := NewListCommand(c)
	cmd.SetArgs([]string{"--all"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Should show both policies with --all
	assert.Contains(t, output, "Bash")
	assert.Contains(t, output, "Write")
	assert.Contains(t, output, "audit")
}

func TestListCommand_JSONOutput(t *testing.T) {
	tempDir := t.TempDir()
	originalHome := os.Getenv("HOME")
	os.Setenv("HOME", tempDir)
	defer os.Setenv("HOME", originalHome)

	arfaDir := tempDir + "/.arfa"
	os.MkdirAll(arfaDir, 0700)
	cacheContent := `{
		"policies": [
			{"tool_name": "Bash", "action": "deny", "reason": "Shell blocked"}
		],
		"version": 12345,
		"synced_at": "2024-01-15T10:00:00Z"
	}`
	os.WriteFile(arfaDir+"/policies.json", []byte(cacheContent), 0600)

	c := container.New()
	cmd := NewListCommand(c)
	cmd.SetArgs([]string{"--json"})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	require.NoError(t, err)

	output := buf.String()
	// Should be valid JSON
	assert.Contains(t, output, `"tool_name": "Bash"`)
	assert.Contains(t, output, `"action": "deny"`)
	assert.Contains(t, output, `"version": 12345`)
}

func TestFilterDenyPolicies(t *testing.T) {
	policies := []api.ToolPolicy{
		{ToolName: "Bash", Action: api.ToolPolicyActionDeny},
		{ToolName: "Write", Action: api.ToolPolicyActionAudit},
		{ToolName: "Read", Action: api.ToolPolicyActionDeny},
	}

	result := filterDenyPolicies(policies)

	assert.Len(t, result, 2)
	assert.Equal(t, "Bash", result[0].ToolName)
	assert.Equal(t, "Read", result[1].ToolName)
}
