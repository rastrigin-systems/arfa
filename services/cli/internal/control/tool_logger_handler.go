package control

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
)

// ToolCallLoggerHandler extracts tool_use events from SSE streams and logs them.
// This provides structured visibility into what tools Claude is invoking.
type ToolCallLoggerHandler struct {
	queue LoggerQueue
}

// NewToolCallLoggerHandler creates a new tool call logger handler.
func NewToolCallLoggerHandler(queue LoggerQueue) *ToolCallLoggerHandler {
	return &ToolCallLoggerHandler{
		queue: queue,
	}
}

// Name returns the handler name.
func (h *ToolCallLoggerHandler) Name() string {
	return "ToolCallLogger"
}

// Priority returns 40 (after PolicyHandler at 110, after LoggerHandler at 50).
// Runs after raw logging to ensure logs are captured even if parsing fails.
func (h *ToolCallLoggerHandler) Priority() int {
	return 40
}

// HandleRequest is a no-op for tool call logging (we log from responses).
func (h *ToolCallLoggerHandler) HandleRequest(ctx *HandlerContext, req *http.Request) Result {
	return ContinueResult()
}

// HandleResponse parses SSE stream and logs tool_use events.
func (h *ToolCallLoggerHandler) HandleResponse(ctx *HandlerContext, res *http.Response) Result {
	if res == nil || res.Body == nil {
		return ContinueResult()
	}

	// Only process SSE streams
	contentType := res.Header.Get("Content-Type")
	if !strings.Contains(contentType, "text/event-stream") {
		return ContinueResult()
	}

	// Read entire body
	bodyBytes, err := io.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		// Restore empty body and continue
		res.Body = io.NopCloser(bytes.NewReader([]byte{}))
		return ContinueResult()
	}

	// Parse and log tool calls
	h.extractAndLogToolCalls(ctx, bodyBytes)

	// Restore body for downstream handlers
	res.Body = io.NopCloser(bytes.NewReader(bodyBytes))
	res.ContentLength = int64(len(bodyBytes))

	return ContinueResult()
}

// toolCallEntry represents a parsed tool call for logging.
type toolCallEntry struct {
	ToolName    string
	ToolID      string
	ToolInput   map[string]interface{}
	Blocked     bool
	BlockReason string
}

// pendingToolCall tracks tool_use blocks during SSE parsing.
type pendingToolCall struct {
	toolName  string
	toolID    string
	inputJSON strings.Builder
}

// extractAndLogToolCalls parses SSE events and logs each tool_use block.
func (h *ToolCallLoggerHandler) extractAndLogToolCalls(ctx *HandlerContext, data []byte) {
	if h.queue == nil {
		return
	}

	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1MB buffer

	var currentEvent string
	var currentData string
	pendingCalls := make(map[int]*pendingToolCall) // index -> pending call

	for scanner.Scan() {
		line := scanner.Text()

		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
			currentData = ""
		} else if strings.HasPrefix(line, "data: ") {
			currentData = strings.TrimPrefix(line, "data: ")
		} else if line == "" && currentEvent != "" {
			// End of SSE event - process it
			h.processSSEEvent(currentEvent, currentData, pendingCalls, ctx)
			currentEvent = ""
			currentData = ""
		}
	}
}

// processSSEEvent handles a single SSE event during parsing.
func (h *ToolCallLoggerHandler) processSSEEvent(event, data string, pendingCalls map[int]*pendingToolCall, ctx *HandlerContext) {
	// Clean up data - Anthropic SSE sometimes has trailing whitespace and extra braces
	data = cleanSSEData(data)

	switch event {
	case "content_block_start":
		var block struct {
			Index        int `json:"index"`
			ContentBlock struct {
				Type string `json:"type"`
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"content_block"`
		}
		if err := json.Unmarshal([]byte(data), &block); err == nil {
			if block.ContentBlock.Type == "tool_use" {
				pendingCalls[block.Index] = &pendingToolCall{
					toolName: block.ContentBlock.Name,
					toolID:   block.ContentBlock.ID,
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
		if err := json.Unmarshal([]byte(data), &delta); err == nil {
			if pending, ok := pendingCalls[delta.Index]; ok {
				if delta.Delta.Type == "input_json_delta" {
					pending.inputJSON.WriteString(delta.Delta.PartialJSON)
				}
			}
		}

	case "content_block_stop":
		var stop struct {
			Index int `json:"index"`
		}
		if err := json.Unmarshal([]byte(data), &stop); err == nil {
			if pending, ok := pendingCalls[stop.Index]; ok {
				// Tool call complete - log it
				h.logToolCall(ctx, pending)
				delete(pendingCalls, stop.Index)
			}
		}
	}
}

// logToolCall creates and enqueues a log entry for a tool call.
func (h *ToolCallLoggerHandler) logToolCall(ctx *HandlerContext, call *pendingToolCall) {
	// Parse accumulated JSON input
	var toolInput map[string]interface{}
	inputStr := call.inputJSON.String()
	if inputStr != "" {
		if err := json.Unmarshal([]byte(inputStr), &toolInput); err != nil {
			// If parsing fails, store raw string
			toolInput = map[string]interface{}{"_raw": inputStr}
		}
	}

	entry := LogEntry{
		EmployeeID:    ctx.EmployeeID,
		OrgID:         ctx.OrgID,
		SessionID:     ctx.SessionID,
		ClientName:    ctx.ClientName,
		ClientVersion: ctx.ClientVersion,
		EventType:     "tool_call",
		EventCategory: "classified",
		Timestamp:     time.Now(),
		Payload: map[string]interface{}{
			"tool_name":  call.toolName,
			"tool_id":    call.toolID,
			"tool_input": toolInput,
			"blocked":    false,
		},
	}

	_ = h.queue.Enqueue(entry)
}

// LogBlockedToolCall logs a tool call that was blocked by policy.
// Called by PolicyHandler when it blocks a tool.
func (h *ToolCallLoggerHandler) LogBlockedToolCall(ctx *HandlerContext, toolName, toolID, reason string, toolInput map[string]interface{}) {
	if h.queue == nil {
		return
	}

	entry := LogEntry{
		EmployeeID:    ctx.EmployeeID,
		OrgID:         ctx.OrgID,
		SessionID:     ctx.SessionID,
		ClientName:    ctx.ClientName,
		ClientVersion: ctx.ClientVersion,
		EventType:     "tool_call",
		EventCategory: "classified",
		Timestamp:     time.Now(),
		Payload: map[string]interface{}{
			"tool_name":    toolName,
			"tool_id":      toolID,
			"tool_input":   toolInput,
			"blocked":      true,
			"block_reason": reason,
		},
	}

	_ = h.queue.Enqueue(entry)
}

// cleanSSEData removes trailing garbage from SSE data lines.
// Anthropic's SSE stream sometimes includes trailing whitespace and extra braces.
func cleanSSEData(data string) string {
	if len(data) == 0 {
		return data
	}

	// Find valid JSON by balancing braces, accounting for strings
	openBraces := 0
	lastValidPos := -1
	inString := false
	escaped := false

	for i, c := range data {
		if escaped {
			escaped = false
			continue
		}

		if c == '\\' && inString {
			escaped = true
			continue
		}

		if c == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		switch c {
		case '{':
			openBraces++
		case '}':
			openBraces--
			if openBraces == 0 {
				lastValidPos = i + 1
			}
		}
	}

	if lastValidPos > 0 && lastValidPos < len(data) {
		return data[:lastValidPos]
	}

	return data
}
