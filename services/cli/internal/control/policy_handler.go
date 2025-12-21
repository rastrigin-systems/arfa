// Package control provides the Control Service for intercepting and processing
// LLM API traffic through a pluggable handler pipeline.
package control

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
)

// PolicyHandler blocks tool calls based on a deny list.
// MVP: Hardcoded deny list, parses SSE stream, replaces blocked tools with error text.
type PolicyHandler struct {
	// denyList contains tool names that should be blocked.
	// For MVP, this is hardcoded. Later: loaded from ~/.ubik/policies.json
	denyList map[string]string // tool name -> reason
}

// NewPolicyHandler creates a new PolicyHandler with a hardcoded deny list for testing.
func NewPolicyHandler() *PolicyHandler {
	return &PolicyHandler{
		denyList: map[string]string{
			// Uncomment tools to block for testing:
			// "Bash":  "Shell commands are blocked by organization policy",
			// "Write": "File writes are blocked by organization policy",
		},
	}
}

// NewPolicyHandlerWithDenyList creates a PolicyHandler with a custom deny list.
func NewPolicyHandlerWithDenyList(denyList map[string]string) *PolicyHandler {
	return &PolicyHandler{denyList: denyList}
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
func (h *PolicyHandler) processSSEStream(data []byte) ([]byte, bool) {
	if len(h.denyList) == 0 {
		return data, false
	}

	var output bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer for large events

	var currentEvent string
	var currentData string
	var blockedIndices = make(map[int]string) // index -> reason
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
						if reason, blocked := h.denyList[strings.ToLower(toolName)]; blocked {
							// Block this tool - remember the index
							blockedIndices[block.Index] = reason
							wasModified = true

							// Write replacement text block instead
							h.writeBlockedEvent(&output, block.Index, toolName, reason)
							shouldWrite = false
						}
						// Also check original case
						if reason, blocked := h.denyList[toolName]; blocked {
							blockedIndices[block.Index] = reason
							wasModified = true
							h.writeBlockedEvent(&output, block.Index, toolName, reason)
							shouldWrite = false
						}
					}
				}

			case "content_block_delta":
				// Check if this delta belongs to a blocked block
				var delta struct {
					Index int `json:"index"`
				}
				if err := json.Unmarshal([]byte(currentData), &delta); err == nil {
					if _, blocked := blockedIndices[delta.Index]; blocked {
						shouldWrite = false // Skip deltas for blocked tools
					}
				}

			case "content_block_stop":
				var stop struct {
					Index int `json:"index"`
				}
				if err := json.Unmarshal([]byte(currentData), &stop); err == nil {
					if _, blocked := blockedIndices[stop.Index]; blocked {
						shouldWrite = false // Skip stop for blocked tools (we already wrote it)
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
