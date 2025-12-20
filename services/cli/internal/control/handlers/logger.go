// Package handlers provides handler implementations for the Control Service pipeline.
package handlers

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/control"
)

// sensitiveHeaders are headers that should be redacted in logs.
var sensitiveHeaders = map[string]bool{
	"authorization":       true,
	"x-api-key":           true,
	"cookie":              true,
	"set-cookie":          true,
	"x-auth-token":        true,
	"x-access-token":      true,
	"proxy-authorization": true,
}

// Queue defines the interface for enqueueing log entries.
type Queue interface {
	Enqueue(entry control.LogEntry) error
}

// LoggerHandler captures request/response data and writes to a queue.
type LoggerHandler struct {
	queue Queue
}

// NewLoggerHandler creates a new logger handler.
func NewLoggerHandler(queue Queue) *LoggerHandler {
	return &LoggerHandler{
		queue: queue,
	}
}

// Name returns the handler name.
func (h *LoggerHandler) Name() string {
	return "logger"
}

// Priority returns the handler priority.
// Logger runs at mid-priority (50) - after policy checks, before analytics.
func (h *LoggerHandler) Priority() int {
	return 50
}

// HandleRequest logs an outgoing API request.
func (h *LoggerHandler) HandleRequest(ctx *control.HandlerContext, req *http.Request) control.Result {
	// Read and restore body
	var bodyStr string
	if req.Body != nil {
		bodyBytes, err := io.ReadAll(req.Body)
		if err == nil {
			bodyStr = string(bodyBytes)
			// Restore body for downstream handlers
			req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	entry := control.LogEntry{
		EmployeeID:    ctx.EmployeeID,
		OrgID:         ctx.OrgID,
		SessionID:     ctx.SessionID,
		AgentID:       ctx.AgentID,
		EventType:     "api_request",
		EventCategory: "proxy",
		Timestamp:     time.Now(),
		Payload: map[string]interface{}{
			"method":  req.Method,
			"url":     req.URL.String(),
			"host":    req.URL.Host,
			"headers": redactHeaders(req.Header),
		},
	}

	if bodyStr != "" {
		entry.Payload["body"] = bodyStr
	}

	// Enqueue is non-blocking (writes to disk)
	_ = h.queue.Enqueue(entry)

	return control.ContinueResult()
}

// HandleResponse logs an incoming API response.
func (h *LoggerHandler) HandleResponse(ctx *control.HandlerContext, res *http.Response) control.Result {
	// Read and restore body
	var bodyStr string
	if res.Body != nil {
		bodyBytes, err := io.ReadAll(res.Body)
		if err == nil {
			bodyStr = string(bodyBytes)
			// Restore body for downstream handlers
			res.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}
	}

	entry := control.LogEntry{
		EmployeeID:    ctx.EmployeeID,
		OrgID:         ctx.OrgID,
		SessionID:     ctx.SessionID,
		AgentID:       ctx.AgentID,
		EventType:     "api_response",
		EventCategory: "proxy",
		Timestamp:     time.Now(),
		Payload: map[string]interface{}{
			"status_code": res.StatusCode,
			"headers":     redactHeaders(res.Header),
		},
	}

	if bodyStr != "" {
		entry.Payload["body"] = bodyStr
	}

	// Include request URL for correlation
	if res.Request != nil {
		entry.Payload["url"] = res.Request.URL.String()
	}

	// Enqueue is non-blocking (writes to disk)
	_ = h.queue.Enqueue(entry)

	return control.ContinueResult()
}

// redactHeaders returns a copy of headers with sensitive values redacted.
func redactHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)

	for key, values := range headers {
		if len(values) == 0 {
			continue
		}

		// Check if header should be redacted (case-insensitive)
		if sensitiveHeaders[strings.ToLower(key)] {
			result[key] = "[REDACTED]"
		} else {
			result[key] = values[0] // Take first value
		}
	}

	return result
}
