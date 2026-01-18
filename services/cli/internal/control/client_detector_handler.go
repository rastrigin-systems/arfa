package control

import (
	"net/http"
)

// ClientDetectorHandler detects the AI client from User-Agent headers.
// This handler should run first (highest priority) to populate client info
// for all downstream handlers.
type ClientDetectorHandler struct{}

// NewClientDetectorHandler creates a new client detector handler.
func NewClientDetectorHandler() *ClientDetectorHandler {
	return &ClientDetectorHandler{}
}

// Name returns the handler name.
func (h *ClientDetectorHandler) Name() string {
	return "client_detector"
}

// Priority returns the handler priority.
// This should be the highest priority to run before all other handlers.
func (h *ClientDetectorHandler) Priority() int {
	return 200 // Higher than PolicyHandler (110)
}

// HandleRequest detects the client from User-Agent and sets context fields.
func (h *ClientDetectorHandler) HandleRequest(ctx *HandlerContext, req *http.Request) Result {
	// Detect client from User-Agent header
	userAgent := req.Header.Get("User-Agent")
	if userAgent != "" {
		clientInfo := DetectClient(userAgent)
		// Only update context if we detected a known client
		// This prevents overwriting with empty values from unrecognized User-Agents
		if clientInfo.Name != "" {
			ctx.SetClient(clientInfo)
		}
	}

	return ContinueResult()
}

// HandleResponse is a no-op for this handler.
func (h *ClientDetectorHandler) HandleResponse(ctx *HandlerContext, res *http.Response) Result {
	return ContinueResult()
}
