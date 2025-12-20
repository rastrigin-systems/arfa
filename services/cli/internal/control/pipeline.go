package control

import (
	"net/http"
	"sort"
	"sync"
)

// Pipeline orchestrates the execution of handlers in priority order.
type Pipeline struct {
	mu       sync.RWMutex
	handlers []Handler
}

// NewPipeline creates a new empty pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{
		handlers: make([]Handler, 0),
	}
}

// Register adds a handler to the pipeline.
// Handlers are automatically sorted by priority (highest first).
func (p *Pipeline) Register(h Handler) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.handlers = append(p.handlers, h)
	p.sortHandlers()
}

// Handlers returns a copy of the registered handlers in priority order.
func (p *Pipeline) Handlers() []Handler {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make([]Handler, len(p.handlers))
	copy(result, p.handlers)
	return result
}

// sortHandlers sorts handlers by priority (highest first).
// Must be called with lock held.
func (p *Pipeline) sortHandlers() {
	sort.Slice(p.handlers, func(i, j int) bool {
		return p.handlers[i].Priority() > p.handlers[j].Priority()
	})
}

// ExecuteRequest runs all handlers for an outgoing request.
// Handlers are executed in priority order (highest first).
// Execution stops if any handler returns ActionBlock.
// If a handler returns a ModifiedRequest, subsequent handlers receive it.
func (p *Pipeline) ExecuteRequest(ctx *HandlerContext, req *http.Request) Result {
	p.mu.RLock()
	handlers := make([]Handler, len(p.handlers))
	copy(handlers, p.handlers)
	p.mu.RUnlock()

	currentReq := req
	var lastResult Result

	for _, h := range handlers {
		result := h.HandleRequest(ctx, currentReq)

		if result.ShouldBlock() {
			return result
		}

		// Use modified request for next handler if provided
		if result.ModifiedRequest != nil {
			currentReq = result.ModifiedRequest
		}

		lastResult = result
	}

	// If no handlers, return continue
	if len(handlers) == 0 {
		return ContinueResult()
	}

	// Return the last result (with potentially modified request)
	if currentReq != req {
		lastResult.ModifiedRequest = currentReq
	}

	return lastResult
}

// ExecuteResponse runs all handlers for an incoming response.
// Handlers are executed in priority order (highest first).
// Execution stops if any handler returns ActionBlock.
// If a handler returns a ModifiedResponse, subsequent handlers receive it.
func (p *Pipeline) ExecuteResponse(ctx *HandlerContext, res *http.Response) Result {
	p.mu.RLock()
	handlers := make([]Handler, len(p.handlers))
	copy(handlers, p.handlers)
	p.mu.RUnlock()

	currentRes := res
	var lastResult Result

	for _, h := range handlers {
		result := h.HandleResponse(ctx, currentRes)

		if result.ShouldBlock() {
			return result
		}

		// Use modified response for next handler if provided
		if result.ModifiedResponse != nil {
			currentRes = result.ModifiedResponse
		}

		lastResult = result
	}

	// If no handlers, return continue
	if len(handlers) == 0 {
		return ContinueResult()
	}

	// Return the last result (with potentially modified response)
	if currentRes != res {
		lastResult.ModifiedResponse = currentRes
	}

	return lastResult
}
