package policies

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func TestListCommand_EmptyPolicies(t *testing.T) {
	// Create mock server returning empty policies
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/employees/me/tool-policies" {
			resp := api.EmployeeToolPoliciesResponse{Policies: []api.ToolPolicy{}}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	// Create container with mock API client
	client := api.NewClient(server.URL)
	client.SetToken("test-token")
	c := container.NewTestContainer(container.WithMockAPIClient(client))
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
	reason := "Shell blocked"
	// Create mock server returning policies
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/employees/me/tool-policies" {
			resp := api.EmployeeToolPoliciesResponse{
				Policies: []api.ToolPolicy{
					{ToolName: "Bash", Action: api.ToolPolicyActionDeny, Reason: &reason, Scope: "organization"},
					{ToolName: "Write", Action: api.ToolPolicyActionAudit},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	client.SetToken("test-token")
	c := container.NewTestContainer(container.WithMockAPIClient(client))
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
	reason1 := "Shell blocked"
	reason2 := "Writes audited"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/employees/me/tool-policies" {
			resp := api.EmployeeToolPoliciesResponse{
				Policies: []api.ToolPolicy{
					{ToolName: "Bash", Action: api.ToolPolicyActionDeny, Reason: &reason1},
					{ToolName: "Write", Action: api.ToolPolicyActionAudit, Reason: &reason2},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	client.SetToken("test-token")
	c := container.NewTestContainer(container.WithMockAPIClient(client))
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
	reason := "Shell blocked"
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/employees/me/tool-policies" {
			resp := api.EmployeeToolPoliciesResponse{
				Policies: []api.ToolPolicy{
					{ToolName: "Bash", Action: api.ToolPolicyActionDeny, Reason: &reason},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
			return
		}
		http.NotFound(w, r)
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	client.SetToken("test-token")
	c := container.NewTestContainer(container.WithMockAPIClient(client))
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
}

func TestListCommand_APIError(t *testing.T) {
	// Create mock server that returns 401
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"Unauthorized"}`))
	}))
	defer server.Close()

	client := api.NewClient(server.URL)
	// No token set
	c := container.NewTestContainer(container.WithMockAPIClient(client))
	cmd := NewListCommand(c)

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	err := cmd.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to fetch policies")
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
