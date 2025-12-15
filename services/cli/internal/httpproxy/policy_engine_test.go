package httpproxy

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPolicyEngine(t *testing.T) {
	pe := NewPolicyEngine()
	assert.NotNil(t, pe)
	assert.False(t, pe.IsPlatformHealthy())
}

func TestPolicyEngine_FailClosed_PlatformUnreachable(t *testing.T) {
	pe := NewPolicyEngine()
	// Platform is not healthy by default (fail-closed)

	session := &Session{
		ID:        "test-session",
		Port:      8100,
		AgentName: "Claude Code",
	}

	decision := pe.EvaluateRequest(session, []byte(`{"messages": [{"role": "user", "content": "Hello"}]}`))

	assert.Equal(t, ActionBlock, decision.Action)
	assert.Contains(t, decision.Reason, "Platform unreachable")
}

func TestPolicyEngine_AllowWhenHealthy(t *testing.T) {
	pe := NewPolicyEngine()
	pe.SetPlatformHealthy(true)

	session := &Session{
		ID:        "test-session",
		Port:      8100,
		AgentName: "Claude Code",
	}

	decision := pe.EvaluateRequest(session, []byte(`{"messages": [{"role": "user", "content": "Hello"}]}`))

	assert.Equal(t, ActionAllow, decision.Action)
}

func TestPolicyEngine_BlockPII_CreditCard(t *testing.T) {
	pe := NewPolicyEngine()
	pe.SetPlatformHealthy(true)
	pe.EnablePIIDetection(true)

	session := &Session{
		ID:        "test-session",
		Port:      8100,
		AgentName: "Claude Code",
	}

	// Request with credit card number
	body := []byte(`{"messages": [{"role": "user", "content": "My card is 4532015112830366"}]}`)
	decision := pe.EvaluateRequest(session, body)

	assert.Equal(t, ActionBlock, decision.Action)
	assert.Contains(t, decision.Reason, "PII detected")
}

func TestPolicyEngine_BlockPII_SSN(t *testing.T) {
	pe := NewPolicyEngine()
	pe.SetPlatformHealthy(true)
	pe.EnablePIIDetection(true)

	session := &Session{
		ID:        "test-session",
		Port:      8100,
		AgentName: "Claude Code",
	}

	// Request with SSN
	body := []byte(`{"messages": [{"role": "user", "content": "My SSN is 123-45-6789"}]}`)
	decision := pe.EvaluateRequest(session, body)

	assert.Equal(t, ActionBlock, decision.Action)
	assert.Contains(t, decision.Reason, "PII detected")
}

func TestPolicyEngine_BlockPII_APIKey(t *testing.T) {
	pe := NewPolicyEngine()
	pe.SetPlatformHealthy(true)
	pe.EnablePIIDetection(true)

	session := &Session{
		ID:        "test-session",
		Port:      8100,
		AgentName: "Claude Code",
	}

	// Request with API key
	body := []byte(`{"messages": [{"role": "user", "content": "Use this key: sk-ant-api03-abcdefghijklmnopqrstuvwxyz"}]}`)
	decision := pe.EvaluateRequest(session, body)

	assert.Equal(t, ActionBlock, decision.Action)
	assert.Contains(t, decision.Reason, "PII detected")
}

func TestPolicyEngine_AllowWithoutPII(t *testing.T) {
	pe := NewPolicyEngine()
	pe.SetPlatformHealthy(true)
	pe.EnablePIIDetection(true)

	session := &Session{
		ID:        "test-session",
		Port:      8100,
		AgentName: "Claude Code",
	}

	// Clean request
	body := []byte(`{"messages": [{"role": "user", "content": "Write a function to sort an array"}]}`)
	decision := pe.EvaluateRequest(session, body)

	assert.Equal(t, ActionAllow, decision.Action)
}

func TestPolicyEngine_BlockDeniedTool(t *testing.T) {
	pe := NewPolicyEngine()
	pe.SetPlatformHealthy(true)
	pe.SetToolDenyList([]string{"execute_command", "run_shell"})

	session := &Session{
		ID:        "test-session",
		Port:      8100,
		AgentName: "Claude Code",
	}

	// Request with denied tool
	body := []byte(`{"tools": [{"name": "execute_command", "description": "Run a command"}]}`)
	decision := pe.EvaluateRequest(session, body)

	assert.Equal(t, ActionBlock, decision.Action)
	assert.Contains(t, decision.Reason, "tool not allowed")
}

func TestPolicyEngine_AllowApprovedTool(t *testing.T) {
	pe := NewPolicyEngine()
	pe.SetPlatformHealthy(true)
	pe.SetToolAllowList([]string{"read_file", "write_file"})

	session := &Session{
		ID:        "test-session",
		Port:      8100,
		AgentName: "Claude Code",
	}

	// Request with allowed tool
	body := []byte(`{"tools": [{"name": "read_file", "description": "Read a file"}]}`)
	decision := pe.EvaluateRequest(session, body)

	assert.Equal(t, ActionAllow, decision.Action)
}

func TestPolicyEngine_BlockUnapprovedToolWhenAllowListSet(t *testing.T) {
	pe := NewPolicyEngine()
	pe.SetPlatformHealthy(true)
	pe.SetToolAllowList([]string{"read_file"})

	session := &Session{
		ID:        "test-session",
		Port:      8100,
		AgentName: "Claude Code",
	}

	// Request with tool not in allow list
	body := []byte(`{"tools": [{"name": "execute_shell", "description": "Run shell command"}]}`)
	decision := pe.EvaluateRequest(session, body)

	assert.Equal(t, ActionBlock, decision.Action)
	assert.Contains(t, decision.Reason, "tool not allowed")
}

func TestPolicyEngine_DenyListTakesPrecedence(t *testing.T) {
	pe := NewPolicyEngine()
	pe.SetPlatformHealthy(true)
	pe.SetToolAllowList([]string{"execute_command"}) // Even if allowed
	pe.SetToolDenyList([]string{"execute_command"})  // Deny takes precedence

	session := &Session{
		ID:        "test-session",
		Port:      8100,
		AgentName: "Claude Code",
	}

	body := []byte(`{"tools": [{"name": "execute_command", "description": "Run a command"}]}`)
	decision := pe.EvaluateRequest(session, body)

	assert.Equal(t, ActionBlock, decision.Action)
}

func TestPolicyEngine_SyncPolicies(t *testing.T) {
	pe := NewPolicyEngine()

	// Create a mock platform client
	mockClient := &mockPlatformClient{
		policies: &PolicySet{
			Version: "v1",
			PII: PIIPolicyConfig{
				Enabled:  true,
				Patterns: []string{`\d{3}-\d{2}-\d{4}`},
			},
			Tools: ToolPolicyConfig{
				DenyList: []string{"dangerous_tool"},
			},
		},
	}
	pe.SetPlatformClient(mockClient)

	// Sync policies
	err := pe.SyncPolicies(context.Background())
	require.NoError(t, err)

	// Should now be healthy and have policies
	assert.True(t, pe.IsPlatformHealthy())
}

func TestPolicyEngine_SyncPoliciesFailure(t *testing.T) {
	pe := NewPolicyEngine()
	pe.SetPlatformHealthy(true) // Start healthy

	// Create a failing mock client
	mockClient := &mockPlatformClient{
		err: assert.AnError,
	}
	pe.SetPlatformClient(mockClient)

	// Sync should fail
	err := pe.SyncPolicies(context.Background())
	assert.Error(t, err)

	// Should become unhealthy after failed sync
	assert.False(t, pe.IsPlatformHealthy())
}

func TestPolicyEngine_BackgroundSync(t *testing.T) {
	pe := NewPolicyEngine()
	pe.SetSyncInterval(50 * time.Millisecond)

	callCount := 0
	mockClient := &mockPlatformClient{
		policies: &PolicySet{Version: "v1"},
		onFetch: func() {
			callCount++
		},
	}
	pe.SetPlatformClient(mockClient)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start background sync
	err := pe.Start(ctx)
	require.NoError(t, err)

	// Wait for a few syncs
	time.Sleep(200 * time.Millisecond)
	cancel()

	// Should have synced multiple times
	assert.GreaterOrEqual(t, callCount, 2)
}

func TestPolicyEngine_LastSyncTime(t *testing.T) {
	pe := NewPolicyEngine()

	mockClient := &mockPlatformClient{
		policies: &PolicySet{Version: "v1"},
	}
	pe.SetPlatformClient(mockClient)

	before := time.Now()
	err := pe.SyncPolicies(context.Background())
	require.NoError(t, err)
	after := time.Now()

	lastSync := pe.LastSyncTime()
	assert.True(t, lastSync.After(before) || lastSync.Equal(before))
	assert.True(t, lastSync.Before(after) || lastSync.Equal(after))
}

// mockPlatformClient is a test double for the platform client
type mockPlatformClient struct {
	policies *PolicySet
	err      error
	onFetch  func()
}

func (m *mockPlatformClient) FetchPolicies(ctx context.Context) (*PolicySet, error) {
	if m.onFetch != nil {
		m.onFetch()
	}
	if m.err != nil {
		return nil, m.err
	}
	return m.policies, nil
}
