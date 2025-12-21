// Package control provides the Control Service for intercepting and processing
// LLM API traffic through a pluggable handler pipeline.
package control

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/api"
)

// PolicyHandler blocks tool calls based on policies loaded from cache.
// Policies are synced via `ubik sync` and stored in ~/.ubik/policies.json.
type PolicyHandler struct {
	// denyList contains tool names that should be blocked unconditionally.
	// Built from policies loaded from ~/.ubik/policies.json
	denyList map[string]string // tool name -> reason

	// globPatterns contains tool name patterns that end with %
	// These match tools that start with the pattern prefix
	globPatterns map[string]string // pattern prefix -> reason

	// conditionalPolicies contains policies with conditions that need parameter evaluation.
	// Key is tool name (lowercase), value is list of policies with conditions.
	conditionalPolicies map[string][]conditionalPolicy
}

// conditionalPolicy represents a policy with conditions to evaluate against tool input.
type conditionalPolicy struct {
	ToolName   string
	Reason     string
	Conditions map[string]interface{}
}

// policyCacheFile represents the cached policies structure.
type policyCacheFile struct {
	Policies []api.ToolPolicy `json:"policies"`
	Version  int              `json:"version"`
	SyncedAt string           `json:"synced_at"`
}

// NewPolicyHandler creates a new PolicyHandler that loads policies from cache.
// If cache is unavailable or empty, no tools are blocked.
func NewPolicyHandler() *PolicyHandler {
	h := &PolicyHandler{
		denyList:            make(map[string]string),
		globPatterns:        make(map[string]string),
		conditionalPolicies: make(map[string][]conditionalPolicy),
	}
	h.loadFromCache()
	return h
}

// NewPolicyHandlerWithDenyList creates a PolicyHandler with a custom deny list.
// Used for testing.
func NewPolicyHandlerWithDenyList(denyList map[string]string) *PolicyHandler {
	return &PolicyHandler{
		denyList:            denyList,
		globPatterns:        make(map[string]string),
		conditionalPolicies: make(map[string][]conditionalPolicy),
	}
}

// NewPolicyHandlerWithPolicies creates a PolicyHandler with a list of policies.
// Used for testing with full policy objects.
func NewPolicyHandlerWithPolicies(policies []api.ToolPolicy) *PolicyHandler {
	h := &PolicyHandler{
		denyList:            make(map[string]string),
		globPatterns:        make(map[string]string),
		conditionalPolicies: make(map[string][]conditionalPolicy),
	}
	h.buildDenyList(policies)
	return h
}

// loadFromCache loads policies from ~/.ubik/policies.json
func (h *PolicyHandler) loadFromCache() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return // Silently fail - no policies blocked
	}

	policiesPath := filepath.Join(homeDir, ".ubik", "policies.json")
	data, err := os.ReadFile(policiesPath)
	if err != nil {
		return // File doesn't exist or can't be read - no policies blocked
	}

	var cache policyCacheFile
	if err := json.Unmarshal(data, &cache); err != nil {
		return // Invalid JSON - no policies blocked
	}

	h.buildDenyList(cache.Policies)
}

// buildDenyList converts policies to the internal deny list format.
// Only policies with action="deny" are added.
// Policies with conditions are stored separately for parameter evaluation.
func (h *PolicyHandler) buildDenyList(policies []api.ToolPolicy) {
	for _, policy := range policies {
		if policy.Action != api.ToolPolicyActionDeny {
			continue // Skip audit-only policies
		}

		reason := "Tool blocked by organization policy"
		if policy.Reason != nil && *policy.Reason != "" {
			reason = *policy.Reason
		}

		toolName := policy.ToolName

		// Handle policies with conditions - these need parameter evaluation
		if len(policy.Conditions) > 0 {
			cp := conditionalPolicy{
				ToolName:   toolName,
				Reason:     reason,
				Conditions: policy.Conditions,
			}
			// Store by lowercase tool name for case-insensitive matching
			key := strings.ToLower(toolName)
			h.conditionalPolicies[key] = append(h.conditionalPolicies[key], cp)
			continue
		}

		// Handle glob patterns (e.g., "mcp__gcloud__%")
		if strings.HasSuffix(toolName, "%") {
			prefix := strings.TrimSuffix(toolName, "%")
			h.globPatterns[prefix] = reason
		} else {
			// Exact match - store in both original case and lowercase
			h.denyList[toolName] = reason
			h.denyList[strings.ToLower(toolName)] = reason
		}
	}
}

// isBlocked checks if a tool should be blocked, returning the reason if so.
func (h *PolicyHandler) isBlocked(toolName string) (string, bool) {
	// Check exact match first (case-sensitive)
	if reason, ok := h.denyList[toolName]; ok {
		return reason, true
	}

	// Check lowercase match
	if reason, ok := h.denyList[strings.ToLower(toolName)]; ok {
		return reason, true
	}

	// Check glob patterns
	for prefix, reason := range h.globPatterns {
		if strings.HasPrefix(toolName, prefix) || strings.HasPrefix(strings.ToLower(toolName), strings.ToLower(prefix)) {
			return reason, true
		}
	}

	return "", false
}

// pendingBlock tracks tool_use blocks that need condition evaluation.
// We buffer events until we have enough input to evaluate conditions.
type pendingBlock struct {
	index           int
	toolName        string
	startEvent      string // The original content_block_start event
	startData       string // The original data for content_block_start
	deltaEvents     []string
	inputJSON       strings.Builder // Accumulated JSON input from deltas
}

// hasConditionalPolicies checks if a tool has policies with conditions.
func (h *PolicyHandler) hasConditionalPolicies(toolName string) bool {
	key := strings.ToLower(toolName)
	policies, exists := h.conditionalPolicies[key]
	return exists && len(policies) > 0
}

// evaluateConditions checks if tool input matches any conditional policy.
// Returns (reason, blocked) - if blocked is true, the tool should be denied.
func (h *PolicyHandler) evaluateConditions(toolName string, input string) (string, bool) {
	key := strings.ToLower(toolName)
	policies, exists := h.conditionalPolicies[key]
	if !exists {
		return "", false
	}

	// Parse the input JSON to extract parameter values
	var inputMap map[string]interface{}
	if err := json.Unmarshal([]byte(input), &inputMap); err != nil {
		// If we can't parse, don't block (fail open)
		return "", false
	}

	for _, policy := range policies {
		if h.matchesConditions(inputMap, policy.Conditions) {
			return policy.Reason, true
		}
	}

	return "", false
}

// matchesConditions checks if the input matches all conditions in a policy.
// Conditions use regex patterns that must match parameter values.
func (h *PolicyHandler) matchesConditions(input map[string]interface{}, conditions map[string]interface{}) bool {
	for paramName, condition := range conditions {
		// Get the parameter value from input
		inputValue, exists := input[paramName]
		if !exists {
			return false // Parameter doesn't exist, condition not met
		}

		// Convert input value to string for pattern matching
		var inputStr string
		switch v := inputValue.(type) {
		case string:
			inputStr = v
		case float64:
			inputStr = fmt.Sprintf("%v", v)
		case bool:
			inputStr = fmt.Sprintf("%v", v)
		default:
			// For complex types (arrays, objects), marshal to JSON
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return false
			}
			inputStr = string(jsonBytes)
		}

		// Check condition type
		switch c := condition.(type) {
		case string:
			// Direct string - treat as regex pattern
			if !h.matchesPattern(inputStr, c) {
				return false
			}
		case map[string]interface{}:
			// Object with operator
			if pattern, ok := c["matches"].(string); ok {
				if !h.matchesPattern(inputStr, pattern) {
					return false
				}
			} else if pattern, ok := c["contains"].(string); ok {
				if !strings.Contains(inputStr, pattern) {
					return false
				}
			} else if pattern, ok := c["equals"].(string); ok {
				if inputStr != pattern {
					return false
				}
			}
		}
	}
	return true
}

// matchesPattern checks if a string matches a regex pattern.
func (h *PolicyHandler) matchesPattern(s, pattern string) bool {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false // Invalid regex, don't match
	}
	return re.MatchString(s)
}

// Name returns the handler name.
func (h *PolicyHandler) Name() string {
	return "PolicyHandler"
}

// Priority returns 110 (higher than LoggerHandler at 50, runs first).
func (h *PolicyHandler) Priority() int {
	return 110
}

// HandleRequest is a no-op for policy handler (we block in response).
func (h *PolicyHandler) HandleRequest(ctx *HandlerContext, req *http.Request) Result {
	return ContinueResult()
}

// HandleResponse parses SSE stream and blocks denied tools.
func (h *PolicyHandler) HandleResponse(ctx *HandlerContext, res *http.Response) Result {
	if res == nil || res.Body == nil {
		return ContinueResult()
	}

	// Only process SSE streams from Anthropic
	contentType := res.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		return ContinueResult()
	}

	// Read entire body
	bodyBytes, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return ErrorResult(err)
	}

	// Parse and potentially modify the SSE stream
	modified, wasModified := h.processSSEStream(bodyBytes)

	// Always restore body (we consumed it by reading)
	// Return ModifiedResponse so pipeline passes restored body to next handler
	if wasModified {
		res.Body = io.NopCloser(bytes.NewReader(modified))
		res.ContentLength = int64(len(modified))
	} else {
		res.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		res.ContentLength = int64(len(bodyBytes))
	}

	return Result{
		Action:           ActionContinue,
		ModifiedResponse: res,
	}
}

// SSE event types we care about
type sseContentBlockStart struct {
	Type         string `json:"type"`
	Index        int    `json:"index"`
	ContentBlock struct {
		Type  string `json:"type"`
		ID    string `json:"id"`
		Name  string `json:"name"`
		Input any    `json:"input"`
	} `json:"content_block"`
}

// processSSEStream parses SSE events and replaces blocked tool_use blocks with error text.
// For tools with conditional policies, we buffer events until we have the full input.
func (h *PolicyHandler) processSSEStream(data []byte) ([]byte, bool) {
	if len(h.denyList) == 0 && len(h.globPatterns) == 0 && len(h.conditionalPolicies) == 0 {
		return data, false
	}

	var output bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer for large events

	var currentEvent string
	var currentData string
	var blockedIndices = make(map[int]string)    // index -> reason (unconditionally blocked)
	var pendingBlocks = make(map[int]*pendingBlock) // index -> pending block (needs condition evaluation)
	wasModified := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
			currentData = ""
		} else if strings.HasPrefix(line, "data: ") {
			currentData = strings.TrimPrefix(line, "data: ")
		} else if line == "" && currentEvent != "" {
			// End of SSE event - process it
			shouldWrite := true

			switch currentEvent {
			case "content_block_start":
				var block sseContentBlockStart
				if err := json.Unmarshal([]byte(currentData), &block); err == nil {
					if block.ContentBlock.Type == "tool_use" {
						toolName := block.ContentBlock.Name

						// Check unconditional block first
						if reason, blocked := h.isBlocked(toolName); blocked {
							blockedIndices[block.Index] = reason
							wasModified = true
							h.writeBlockedEvent(&output, block.Index, toolName, reason)
							shouldWrite = false
						} else if h.hasConditionalPolicies(toolName) {
							// Tool has conditional policies - buffer for evaluation
							pendingBlocks[block.Index] = &pendingBlock{
								index:      block.Index,
								toolName:   toolName,
								startEvent: currentEvent,
								startData:  currentData,
							}
							shouldWrite = false // Buffer, don't write yet
						}
					}
				}

			case "content_block_delta":
				var delta struct {
					Index int `json:"index"`
					Delta struct {
						Type        string `json:"type"`
						PartialJSON string `json:"partial_json"`
					} `json:"delta"`
				}
				if err := json.Unmarshal([]byte(currentData), &delta); err == nil {
					if _, blocked := blockedIndices[delta.Index]; blocked {
						shouldWrite = false // Skip deltas for blocked tools
					} else if pending, ok := pendingBlocks[delta.Index]; ok {
						// Buffer this delta for later evaluation
						pending.deltaEvents = append(pending.deltaEvents,
							"event: "+currentEvent+"\ndata: "+currentData+"\n\n")
						// Accumulate input JSON from input_json_delta events
						if delta.Delta.Type == "input_json_delta" {
							pending.inputJSON.WriteString(delta.Delta.PartialJSON)
						}
						shouldWrite = false
					}
				}

			case "content_block_stop":
				var stop struct {
					Index int `json:"index"`
				}
				if err := json.Unmarshal([]byte(currentData), &stop); err == nil {
					if _, blocked := blockedIndices[stop.Index]; blocked {
						shouldWrite = false // Skip stop for blocked tools (we already wrote it)
					} else if pending, ok := pendingBlocks[stop.Index]; ok {
						// Evaluate conditions against accumulated input
						input := pending.inputJSON.String()
						if reason, blocked := h.evaluateConditions(pending.toolName, input); blocked {
							// Condition matched - block this tool
							wasModified = true
							h.writeBlockedEvent(&output, pending.index, pending.toolName, reason)
						} else {
							// No conditions matched - flush buffered events
							output.WriteString("event: ")
							output.WriteString(pending.startEvent)
							output.WriteString("\n")
							output.WriteString("data: ")
							output.WriteString(pending.startData)
							output.WriteString("\n\n")
							for _, event := range pending.deltaEvents {
								output.WriteString(event)
							}
							// Write the stop event
							output.WriteString("event: ")
							output.WriteString(currentEvent)
							output.WriteString("\n")
							output.WriteString("data: ")
							output.WriteString(currentData)
							output.WriteString("\n\n")
						}
						delete(pendingBlocks, stop.Index)
						shouldWrite = false
					}
				}
			}

			if shouldWrite {
				output.WriteString("event: ")
				output.WriteString(currentEvent)
				output.WriteString("\n")
				output.WriteString("data: ")
				output.WriteString(currentData)
				output.WriteString("\n\n")
			}

			currentEvent = ""
			currentData = ""
		}
	}

	return output.Bytes(), wasModified
}

// writeBlockedEvent writes SSE events for a blocked tool (text block with error message).
func (h *PolicyHandler) writeBlockedEvent(w *bytes.Buffer, index int, toolName, reason string) {
	errorText := h.formatBlockError(toolName, reason)

	// Write content_block_start for text
	startData := map[string]any{
		"type":  "content_block_start",
		"index": index,
		"content_block": map[string]any{
			"type": "text",
			"text": "",
		},
	}
	startJSON, _ := json.Marshal(startData)
	w.WriteString("event: content_block_start\n")
	w.WriteString("data: ")
	w.Write(startJSON)
	w.WriteString("\n\n")

	// Write content_block_delta with error text
	deltaData := map[string]any{
		"type":  "content_block_delta",
		"index": index,
		"delta": map[string]any{
			"type": "text_delta",
			"text": errorText,
		},
	}
	deltaJSON, _ := json.Marshal(deltaData)
	w.WriteString("event: content_block_delta\n")
	w.WriteString("data: ")
	w.Write(deltaJSON)
	w.WriteString("\n\n")

	// Write content_block_stop
	stopData := map[string]any{
		"type":  "content_block_stop",
		"index": index,
	}
	stopJSON, _ := json.Marshal(stopData)
	w.WriteString("event: content_block_stop\n")
	w.WriteString("data: ")
	w.Write(stopJSON)
	w.WriteString("\n\n")
}

// formatBlockError creates the user-friendly error message.
func (h *PolicyHandler) formatBlockError(toolName, reason string) string {
	return "\n\n[TOOL BLOCKED BY ORGANIZATION POLICY]\n\n" +
		"Tool: " + toolName + "\n" +
		"Reason: " + reason + "\n\n" +
		"This restriction is set by your company administrator.\n" +
		"To see all tool restrictions, run: ubik policies list\n\n"
}
