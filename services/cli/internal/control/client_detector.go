package control

import (
	"regexp"
	"strings"
)

// ClientInfo contains detected AI client information from User-Agent.
type ClientInfo struct {
	Name    string // e.g., "claude-code", "cursor", "continue", "windsurf"
	Version string // e.g., "1.0.25", "0.43.0"
}

// Known AI client User-Agent patterns
var clientPatterns = []struct {
	pattern *regexp.Regexp
	name    string
}{
	// Claude Code: "claude-code/1.0.25" or "ClaudeCode/1.0.25"
	{regexp.MustCompile(`(?i)claude[-_]?code[/\s]+v?(\d+\.\d+(?:\.\d+)?)`), "claude-code"},

	// Claude CLI (part of Claude Code): "claude-cli/2.0.76 (external, cli)"
	// Maps to claude-code since it's the same product
	{regexp.MustCompile(`(?i)claude[-_]?cli[/\s]+v?(\d+\.\d+(?:\.\d+)?)`), "claude-code"},

	// Cursor: "Cursor/0.43.0" or "cursor/0.43.0"
	{regexp.MustCompile(`(?i)cursor[/\s]+v?(\d+\.\d+(?:\.\d+)?)`), "cursor"},

	// Continue: "Continue/1.0.0" or "continue-dev/1.0.0"
	{regexp.MustCompile(`(?i)continue(?:-dev)?[/\s]+v?(\d+\.\d+(?:\.\d+)?)`), "continue"},

	// Windsurf/Codeium: "Windsurf/1.0.0" or "Codeium/1.0.0"
	{regexp.MustCompile(`(?i)(?:windsurf|codeium)[/\s]+v?(\d+\.\d+(?:\.\d+)?)`), "windsurf"},

	// Aider: "aider/0.50.0"
	{regexp.MustCompile(`(?i)aider[/\s]+v?(\d+\.\d+(?:\.\d+)?)`), "aider"},

	// GitHub Copilot: "copilot/1.0.0" or "github-copilot/1.0.0"
	{regexp.MustCompile(`(?i)(?:github[-_]?)?copilot[/\s]+v?(\d+\.\d+(?:\.\d+)?)`), "copilot"},

	// OpenAI Codex CLI: "codex/1.0.0" or "openai-codex/1.0.0"
	{regexp.MustCompile(`(?i)(?:openai[-_]?)?codex[/\s]+v?(\d+\.\d+(?:\.\d+)?)`), "codex"},

	// Generic patterns for other clients
	{regexp.MustCompile(`(?i)vscode[/\s]+v?(\d+\.\d+(?:\.\d+)?)`), "vscode"},
	{regexp.MustCompile(`(?i)neovim[/\s]+v?(\d+\.\d+(?:\.\d+)?)`), "neovim"},
}

// DetectClient parses the User-Agent header and returns detected client info.
// Returns empty ClientInfo if no known client is detected.
func DetectClient(userAgent string) ClientInfo {
	if userAgent == "" {
		return ClientInfo{}
	}

	// Try each known pattern
	for _, cp := range clientPatterns {
		if matches := cp.pattern.FindStringSubmatch(userAgent); len(matches) >= 2 {
			return ClientInfo{
				Name:    cp.name,
				Version: matches[1],
			}
		}
	}

	// Try to detect from generic patterns
	// Look for product/version format at the start
	genericPattern := regexp.MustCompile(`^([a-zA-Z][\w-]*)[/\s]+v?(\d+\.\d+(?:\.\d+)?)`)
	if matches := genericPattern.FindStringSubmatch(userAgent); len(matches) >= 3 {
		name := strings.ToLower(matches[1])
		// Only return if it looks like an AI/coding tool
		if isLikelyAIClient(name) {
			return ClientInfo{
				Name:    name,
				Version: matches[2],
			}
		}
	}

	return ClientInfo{}
}

// isLikelyAIClient returns true if the name suggests an AI coding assistant.
func isLikelyAIClient(name string) bool {
	// Keywords that suggest AI/coding tools
	keywords := []string{
		"ai", "assistant", "code", "copilot", "llm",
		"claude", "gpt", "chat", "agent", "dev", "codex",
	}

	name = strings.ToLower(name)
	for _, kw := range keywords {
		if strings.Contains(name, kw) {
			return true
		}
	}
	return false
}

// DetectClientFromHeaders extracts User-Agent and detects the client.
// This is a convenience wrapper that handles header access.
func DetectClientFromHeaders(headers map[string][]string) ClientInfo {
	// Try different header name variations
	for _, key := range []string{"User-Agent", "user-agent", "USER-AGENT"} {
		if values, ok := headers[key]; ok && len(values) > 0 {
			return DetectClient(values[0])
		}
	}
	return ClientInfo{}
}
