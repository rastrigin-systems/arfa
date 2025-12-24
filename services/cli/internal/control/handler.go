// Package control provides the Control Service for intercepting and processing
// LLM API traffic through a pluggable handler pipeline.
package control

import (
	"net/http"
)

// Action represents what should happen after a handler processes a request/response.
type Action int

const (
	// ActionContinue indicates the pipeline should continue to the next handler.
	ActionContinue Action = iota
	// ActionBlock indicates the request/response should be blocked.
	ActionBlock
)

// Result is returned by handlers after processing a request or response.
type Result struct {
	// Action indicates whether to continue or block.
	Action Action

	// Reason provides context when blocking (for logging/debugging).
	Reason string

	// Error captures any error that occurred during handling.
	// A non-nil error with ActionContinue means "log error but continue".
	Error error

	// ModifiedRequest allows handlers to modify the request before forwarding.
	// If nil, the original request is used.
	ModifiedRequest *http.Request

	// ModifiedResponse allows handlers to modify the response before returning.
	// If nil, the original response is used.
	ModifiedResponse *http.Response
}

// ShouldContinue returns true if the pipeline should continue.
func (r Result) ShouldContinue() bool {
	return r.Action == ActionContinue
}

// ShouldBlock returns true if the request/response should be blocked.
func (r Result) ShouldBlock() bool {
	return r.Action == ActionBlock
}

// ContinueResult creates a Result that continues the pipeline.
func ContinueResult() Result {
	return Result{Action: ActionContinue}
}

// BlockResult creates a Result that blocks with a reason.
func BlockResult(reason string) Result {
	return Result{Action: ActionBlock, Reason: reason}
}

// ErrorResult creates a Result that continues but records an error.
func ErrorResult(err error) Result {
	return Result{Action: ActionContinue, Error: err}
}

// HandlerContext provides context to handlers about the current request/response.
// This includes ownership fields for log attribution.
type HandlerContext struct {
	// EmployeeID identifies the employee making the request.
	EmployeeID string

	// OrgID identifies the organization for multi-tenancy.
	OrgID string

	// SessionID groups related requests within a CLI session.
	SessionID string

	// ClientName identifies the AI client (e.g., "claude-code", "cursor", "continue").
	// Detected from User-Agent headers.
	ClientName string

	// ClientVersion is the version of the AI client (e.g., "1.0.25").
	// Detected from User-Agent headers.
	ClientVersion string

	// Metadata allows handlers to pass data to downstream handlers.
	Metadata map[string]interface{}
}

// NewHandlerContext creates a new HandlerContext with the required ownership fields.
func NewHandlerContext(employeeID, orgID, sessionID string) *HandlerContext {
	return &HandlerContext{
		EmployeeID: employeeID,
		OrgID:      orgID,
		SessionID:  sessionID,
		Metadata:   make(map[string]interface{}),
	}
}

// SetClient updates the client detection fields from a ClientInfo.
func (ctx *HandlerContext) SetClient(info ClientInfo) {
	ctx.ClientName = info.Name
	ctx.ClientVersion = info.Version
}

// Handler defines the interface for processing requests and responses.
// Handlers are called in priority order (higher priority first).
type Handler interface {
	// Name returns a unique identifier for this handler.
	Name() string

	// Priority returns the handler's priority (higher = earlier in pipeline).
	// Suggested ranges:
	//   100+ : Policy/blocking handlers (run first)
	//   50-99: Logging/capture handlers
	//   1-49 : Analytics/metrics handlers (run last)
	Priority() int

	// HandleRequest processes an outgoing request.
	// Called before the request is forwarded to the LLM API.
	HandleRequest(ctx *HandlerContext, req *http.Request) Result

	// HandleResponse processes an incoming response.
	// Called before the response is returned to the agent.
	HandleResponse(ctx *HandlerContext, res *http.Response) Result
}
