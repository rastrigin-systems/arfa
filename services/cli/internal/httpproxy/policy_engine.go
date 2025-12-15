package httpproxy

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sync"
	"time"
)

// Policy action constants
const (
	ActionAllow = "allow"
	ActionBlock = "block"
)

// PolicyDecision represents the result of a policy evaluation
type PolicyDecision struct {
	Action   string `json:"action"`
	Reason   string `json:"reason,omitempty"`
	PolicyID string `json:"policy_id,omitempty"`
}

// PolicySet contains all policies loaded from the platform
type PolicySet struct {
	Version string           `json:"version"`
	PII     PIIPolicyConfig  `json:"pii"`
	Tools   ToolPolicyConfig `json:"tools"`
}

// PIIPolicyConfig configures PII detection
type PIIPolicyConfig struct {
	Enabled  bool     `json:"enabled"`
	Patterns []string `json:"patterns,omitempty"`
}

// ToolPolicyConfig configures tool filtering
type ToolPolicyConfig struct {
	AllowList []string `json:"allow_list,omitempty"`
	DenyList  []string `json:"deny_list,omitempty"`
}

// PlatformClient interface for fetching policies from the platform
type PlatformClient interface {
	FetchPolicies(ctx context.Context) (*PolicySet, error)
}

// PolicyEngine evaluates requests against security policies
type PolicyEngine struct {
	mu              sync.RWMutex
	policies        *PolicySet
	platformHealthy bool
	lastSync        time.Time
	platformClient  PlatformClient
	syncInterval    time.Duration

	// PII detection
	piiEnabled  bool
	piiPatterns []*regexp.Regexp

	// Tool filtering
	toolAllowList map[string]bool
	toolDenyList  map[string]bool
}

// NewPolicyEngine creates a new policy engine
func NewPolicyEngine() *PolicyEngine {
	pe := &PolicyEngine{
		platformHealthy: false, // Fail-closed by default
		syncInterval:    30 * time.Second,
		piiPatterns:     make([]*regexp.Regexp, 0),
		toolAllowList:   make(map[string]bool),
		toolDenyList:    make(map[string]bool),
	}

	// Load default PII patterns
	pe.loadDefaultPIIPatterns()

	return pe
}

// loadDefaultPIIPatterns loads built-in PII detection patterns
func (pe *PolicyEngine) loadDefaultPIIPatterns() {
	defaultPatterns := []string{
		// Credit card numbers (Visa, MasterCard, Amex, etc.)
		`\b4[0-9]{12}(?:[0-9]{3})?\b`,     // Visa
		`\b5[1-5][0-9]{14}\b`,             // MasterCard
		`\b3[47][0-9]{13}\b`,              // American Express
		`\b6(?:011|5[0-9]{2})[0-9]{12}\b`, // Discover

		// SSN
		`\b\d{3}-\d{2}-\d{4}\b`,

		// API keys (common patterns)
		`\bsk-[A-Za-z0-9]{20,}\b`,                         // OpenAI/Anthropic style
		`\bsk-ant-api[A-Za-z0-9\-]{20,}\b`,                // Anthropic
		`\b(api[_-]?key|apikey)[=:]\s*[A-Za-z0-9]{16,}\b`, // Generic API key
		`\bghp_[A-Za-z0-9]{36,}\b`,                        // GitHub PAT
		`\bglpat-[A-Za-z0-9\-]{20,}\b`,                    // GitLab PAT
	}

	for _, pattern := range defaultPatterns {
		if re, err := regexp.Compile(pattern); err == nil {
			pe.piiPatterns = append(pe.piiPatterns, re)
		}
	}
}

// IsPlatformHealthy returns whether the platform is reachable
func (pe *PolicyEngine) IsPlatformHealthy() bool {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.platformHealthy
}

// SetPlatformHealthy sets the platform health status
func (pe *PolicyEngine) SetPlatformHealthy(healthy bool) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.platformHealthy = healthy
}

// EnablePIIDetection enables or disables PII detection
func (pe *PolicyEngine) EnablePIIDetection(enabled bool) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.piiEnabled = enabled
}

// SetToolAllowList sets the list of allowed tools
func (pe *PolicyEngine) SetToolAllowList(tools []string) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.toolAllowList = make(map[string]bool)
	for _, tool := range tools {
		pe.toolAllowList[tool] = true
	}
}

// SetToolDenyList sets the list of denied tools
func (pe *PolicyEngine) SetToolDenyList(tools []string) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.toolDenyList = make(map[string]bool)
	for _, tool := range tools {
		pe.toolDenyList[tool] = true
	}
}

// SetPlatformClient sets the platform client for policy sync
func (pe *PolicyEngine) SetPlatformClient(client PlatformClient) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.platformClient = client
}

// SetSyncInterval sets the policy sync interval
func (pe *PolicyEngine) SetSyncInterval(interval time.Duration) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.syncInterval = interval
}

// LastSyncTime returns the time of the last successful policy sync
func (pe *PolicyEngine) LastSyncTime() time.Time {
	pe.mu.RLock()
	defer pe.mu.RUnlock()
	return pe.lastSync
}

// EvaluateRequest evaluates a request against security policies
func (pe *PolicyEngine) EvaluateRequest(session *Session, body []byte) *PolicyDecision {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	// Fail-closed: block if platform is unreachable
	if !pe.platformHealthy {
		return &PolicyDecision{
			Action: ActionBlock,
			Reason: "Platform unreachable - security policy requires connection",
		}
	}

	// Check for PII in request body
	if pe.piiEnabled {
		if decision := pe.checkPII(body); decision != nil {
			return decision
		}
	}

	// Check tool usage
	if decision := pe.checkTools(body); decision != nil {
		return decision
	}

	return &PolicyDecision{
		Action: ActionAllow,
	}
}

// checkPII scans the request body for PII
func (pe *PolicyEngine) checkPII(body []byte) *PolicyDecision {
	content := string(body)

	for _, pattern := range pe.piiPatterns {
		if pattern.MatchString(content) {
			return &PolicyDecision{
				Action: ActionBlock,
				Reason: "PII detected in request",
			}
		}
	}

	return nil
}

// checkTools validates tool usage against allow/deny lists
func (pe *PolicyEngine) checkTools(body []byte) *PolicyDecision {
	// Parse tools from request
	tools := extractToolNames(body)

	for _, tool := range tools {
		// Deny list takes precedence
		if pe.toolDenyList[tool] {
			return &PolicyDecision{
				Action: ActionBlock,
				Reason: fmt.Sprintf("tool not allowed: %s is in deny list", tool),
			}
		}

		// If allow list is set, tool must be in it
		if len(pe.toolAllowList) > 0 && !pe.toolAllowList[tool] {
			return &PolicyDecision{
				Action: ActionBlock,
				Reason: fmt.Sprintf("tool not allowed: %s is not in allow list", tool),
			}
		}
	}

	return nil
}

// extractToolNames extracts tool names from an API request body
func extractToolNames(body []byte) []string {
	var tools []string

	// Try to parse as JSON with tools array
	var request struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}

	if err := json.Unmarshal(body, &request); err == nil {
		for _, tool := range request.Tools {
			if tool.Name != "" {
				tools = append(tools, tool.Name)
			}
		}
	}

	return tools
}

// SyncPolicies fetches policies from the platform
func (pe *PolicyEngine) SyncPolicies(ctx context.Context) error {
	pe.mu.RLock()
	client := pe.platformClient
	pe.mu.RUnlock()

	if client == nil {
		return fmt.Errorf("no platform client configured")
	}

	policies, err := client.FetchPolicies(ctx)
	if err != nil {
		pe.mu.Lock()
		pe.platformHealthy = false
		pe.mu.Unlock()
		return fmt.Errorf("failed to fetch policies: %w", err)
	}

	pe.mu.Lock()
	defer pe.mu.Unlock()

	pe.policies = policies
	pe.platformHealthy = true
	pe.lastSync = time.Now()

	// Apply fetched policies
	pe.applyPolicies(policies)

	return nil
}

// applyPolicies updates the engine with the fetched policies
func (pe *PolicyEngine) applyPolicies(policies *PolicySet) {
	if policies == nil {
		return
	}

	// Apply PII config
	pe.piiEnabled = policies.PII.Enabled
	if len(policies.PII.Patterns) > 0 {
		pe.piiPatterns = make([]*regexp.Regexp, 0)
		for _, pattern := range policies.PII.Patterns {
			if re, err := regexp.Compile(pattern); err == nil {
				pe.piiPatterns = append(pe.piiPatterns, re)
			}
		}
	}

	// Apply tool config
	pe.toolAllowList = make(map[string]bool)
	for _, tool := range policies.Tools.AllowList {
		pe.toolAllowList[tool] = true
	}

	pe.toolDenyList = make(map[string]bool)
	for _, tool := range policies.Tools.DenyList {
		pe.toolDenyList[tool] = true
	}
}

// Start begins background policy synchronization
func (pe *PolicyEngine) Start(ctx context.Context) error {
	// Initial sync
	if err := pe.SyncPolicies(ctx); err != nil {
		// Log but don't fail - will retry
		fmt.Printf("Initial policy sync failed: %v\n", err)
	}

	// Background sync goroutine
	go func() {
		pe.mu.RLock()
		interval := pe.syncInterval
		pe.mu.RUnlock()

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := pe.SyncPolicies(ctx); err != nil {
					fmt.Printf("Policy sync failed: %v\n", err)
				}
			}
		}
	}()

	return nil
}

// Stop stops the policy engine (cleanup if needed)
func (pe *PolicyEngine) Stop() {
	// Currently no cleanup needed
	// Context cancellation handles stopping background sync
}
