package control

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testHandler is a configurable mock handler for testing
type testHandler struct {
	name          string
	priority      int
	onRequest     func(ctx *HandlerContext, req *http.Request) Result
	onResponse    func(ctx *HandlerContext, res *http.Response) Result
	requestCalls  int
	responseCalls int
}

func (h *testHandler) Name() string  { return h.name }
func (h *testHandler) Priority() int { return h.priority }

func (h *testHandler) HandleRequest(ctx *HandlerContext, req *http.Request) Result {
	h.requestCalls++
	if h.onRequest != nil {
		return h.onRequest(ctx, req)
	}
	return ContinueResult()
}

func (h *testHandler) HandleResponse(ctx *HandlerContext, res *http.Response) Result {
	h.responseCalls++
	if h.onResponse != nil {
		return h.onResponse(ctx, res)
	}
	return ContinueResult()
}

func TestNewPipeline(t *testing.T) {
	p := NewPipeline()

	require.NotNil(t, p)
	assert.Empty(t, p.Handlers())
}

func TestPipeline_Register(t *testing.T) {
	p := NewPipeline()
	h := &testHandler{name: "test", priority: 50}

	p.Register(h)

	handlers := p.Handlers()
	require.Len(t, handlers, 1)
	assert.Equal(t, "test", handlers[0].Name())
}

func TestPipeline_RegisterMultiple_SortedByPriority(t *testing.T) {
	p := NewPipeline()

	// Register in random order
	p.Register(&testHandler{name: "low", priority: 10})
	p.Register(&testHandler{name: "high", priority: 100})
	p.Register(&testHandler{name: "medium", priority: 50})

	handlers := p.Handlers()
	require.Len(t, handlers, 3)

	// Should be sorted high to low priority
	assert.Equal(t, "high", handlers[0].Name())
	assert.Equal(t, "medium", handlers[1].Name())
	assert.Equal(t, "low", handlers[2].Name())
}

func TestPipeline_ExecuteRequest_AllHandlersCalled(t *testing.T) {
	p := NewPipeline()

	h1 := &testHandler{name: "h1", priority: 100}
	h2 := &testHandler{name: "h2", priority: 50}
	h3 := &testHandler{name: "h3", priority: 10}

	p.Register(h1)
	p.Register(h2)
	p.Register(h3)

	ctx := NewHandlerContext("emp", "org", "sess")
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)

	result := p.ExecuteRequest(ctx, req)

	assert.Equal(t, ActionContinue, result.Action)
	assert.Equal(t, 1, h1.requestCalls)
	assert.Equal(t, 1, h2.requestCalls)
	assert.Equal(t, 1, h3.requestCalls)
}

func TestPipeline_ExecuteRequest_StopsOnBlock(t *testing.T) {
	p := NewPipeline()

	h1 := &testHandler{name: "h1", priority: 100}
	h2 := &testHandler{
		name:     "blocker",
		priority: 50,
		onRequest: func(ctx *HandlerContext, req *http.Request) Result {
			return BlockResult("blocked by policy")
		},
	}
	h3 := &testHandler{name: "h3", priority: 10}

	p.Register(h1)
	p.Register(h2)
	p.Register(h3)

	ctx := NewHandlerContext("emp", "org", "sess")
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)

	result := p.ExecuteRequest(ctx, req)

	assert.Equal(t, ActionBlock, result.Action)
	assert.Equal(t, "blocked by policy", result.Reason)
	assert.Equal(t, 1, h1.requestCalls)
	assert.Equal(t, 1, h2.requestCalls)
	assert.Equal(t, 0, h3.requestCalls) // Should not be called
}

func TestPipeline_ExecuteRequest_PropagatesModifiedRequest(t *testing.T) {
	p := NewPipeline()

	h1 := &testHandler{
		name:     "modifier",
		priority: 100,
		onRequest: func(ctx *HandlerContext, req *http.Request) Result {
			req.Header.Set("X-Modified-By", "h1")
			return Result{Action: ActionContinue, ModifiedRequest: req}
		},
	}

	var receivedHeader string
	h2 := &testHandler{
		name:     "receiver",
		priority: 50,
		onRequest: func(ctx *HandlerContext, req *http.Request) Result {
			receivedHeader = req.Header.Get("X-Modified-By")
			return ContinueResult()
		},
	}

	p.Register(h1)
	p.Register(h2)

	ctx := NewHandlerContext("emp", "org", "sess")
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)

	p.ExecuteRequest(ctx, req)

	assert.Equal(t, "h1", receivedHeader)
}

func TestPipeline_ExecuteResponse_AllHandlersCalled(t *testing.T) {
	p := NewPipeline()

	h1 := &testHandler{name: "h1", priority: 100}
	h2 := &testHandler{name: "h2", priority: 50}
	h3 := &testHandler{name: "h3", priority: 10}

	p.Register(h1)
	p.Register(h2)
	p.Register(h3)

	ctx := NewHandlerContext("emp", "org", "sess")
	res := &http.Response{StatusCode: 200}

	result := p.ExecuteResponse(ctx, res)

	assert.Equal(t, ActionContinue, result.Action)
	assert.Equal(t, 1, h1.responseCalls)
	assert.Equal(t, 1, h2.responseCalls)
	assert.Equal(t, 1, h3.responseCalls)
}

func TestPipeline_ExecuteResponse_StopsOnBlock(t *testing.T) {
	p := NewPipeline()

	h1 := &testHandler{name: "h1", priority: 100}
	h2 := &testHandler{
		name:     "blocker",
		priority: 50,
		onResponse: func(ctx *HandlerContext, res *http.Response) Result {
			return BlockResult("response blocked")
		},
	}
	h3 := &testHandler{name: "h3", priority: 10}

	p.Register(h1)
	p.Register(h2)
	p.Register(h3)

	ctx := NewHandlerContext("emp", "org", "sess")
	res := &http.Response{StatusCode: 200}

	result := p.ExecuteResponse(ctx, res)

	assert.Equal(t, ActionBlock, result.Action)
	assert.Equal(t, "response blocked", result.Reason)
	assert.Equal(t, 1, h1.responseCalls)
	assert.Equal(t, 1, h2.responseCalls)
	assert.Equal(t, 0, h3.responseCalls) // Should not be called
}

func TestPipeline_ExecuteRequest_EmptyPipeline(t *testing.T) {
	p := NewPipeline()

	ctx := NewHandlerContext("emp", "org", "sess")
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)

	result := p.ExecuteRequest(ctx, req)

	assert.Equal(t, ActionContinue, result.Action)
}

func TestPipeline_ExecuteResponse_EmptyPipeline(t *testing.T) {
	p := NewPipeline()

	ctx := NewHandlerContext("emp", "org", "sess")
	res := &http.Response{StatusCode: 200}

	result := p.ExecuteResponse(ctx, res)

	assert.Equal(t, ActionContinue, result.Action)
}

func TestPipeline_ExecuteRequest_ErrorContinues(t *testing.T) {
	p := NewPipeline()

	h1 := &testHandler{
		name:     "error-handler",
		priority: 100,
		onRequest: func(ctx *HandlerContext, req *http.Request) Result {
			return ErrorResult(assert.AnError)
		},
	}
	h2 := &testHandler{name: "h2", priority: 50}

	p.Register(h1)
	p.Register(h2)

	ctx := NewHandlerContext("emp", "org", "sess")
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)

	result := p.ExecuteRequest(ctx, req)

	// Error result should continue, not block
	assert.Equal(t, ActionContinue, result.Action)
	assert.Equal(t, 1, h1.requestCalls)
	assert.Equal(t, 1, h2.requestCalls) // Should still be called
}

func TestPipeline_HandlerOrder_ExecutedHighToLow(t *testing.T) {
	p := NewPipeline()

	var order []string

	p.Register(&testHandler{
		name:     "low",
		priority: 10,
		onRequest: func(ctx *HandlerContext, req *http.Request) Result {
			order = append(order, "low")
			return ContinueResult()
		},
	})
	p.Register(&testHandler{
		name:     "high",
		priority: 100,
		onRequest: func(ctx *HandlerContext, req *http.Request) Result {
			order = append(order, "high")
			return ContinueResult()
		},
	})
	p.Register(&testHandler{
		name:     "medium",
		priority: 50,
		onRequest: func(ctx *HandlerContext, req *http.Request) Result {
			order = append(order, "medium")
			return ContinueResult()
		},
	})

	ctx := NewHandlerContext("emp", "org", "sess")
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)

	p.ExecuteRequest(ctx, req)

	assert.Equal(t, []string{"high", "medium", "low"}, order)
}
