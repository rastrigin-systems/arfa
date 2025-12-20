package control

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandlerContext_HasRequiredFields(t *testing.T) {
	ctx := &HandlerContext{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		SessionID:  "sess-789",
		AgentID:    "agent-abc",
	}

	assert.Equal(t, "emp-123", ctx.EmployeeID)
	assert.Equal(t, "org-456", ctx.OrgID)
	assert.Equal(t, "sess-789", ctx.SessionID)
	assert.Equal(t, "agent-abc", ctx.AgentID)
}

func TestHandlerContext_WithMetadata(t *testing.T) {
	ctx := &HandlerContext{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		SessionID:  "sess-789",
		AgentID:    "agent-abc",
		Metadata:   map[string]interface{}{"key": "value"},
	}

	assert.Equal(t, "value", ctx.Metadata["key"])
}

func TestResult_ContinueAction(t *testing.T) {
	result := Result{Action: ActionContinue}

	assert.Equal(t, ActionContinue, result.Action)
	assert.True(t, result.ShouldContinue())
	assert.False(t, result.ShouldBlock())
}

func TestResult_BlockAction(t *testing.T) {
	result := Result{
		Action: ActionBlock,
		Reason: "policy violation",
	}

	assert.Equal(t, ActionBlock, result.Action)
	assert.False(t, result.ShouldContinue())
	assert.True(t, result.ShouldBlock())
	assert.Equal(t, "policy violation", result.Reason)
}

func TestResult_WithError(t *testing.T) {
	result := Result{
		Action: ActionContinue,
		Error:  assert.AnError,
	}

	assert.Error(t, result.Error)
}

func TestResult_WithModifiedRequest(t *testing.T) {
	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-Modified", "true")

	result := Result{
		Action:          ActionContinue,
		ModifiedRequest: req,
	}

	assert.NotNil(t, result.ModifiedRequest)
	assert.Equal(t, "true", result.ModifiedRequest.Header.Get("X-Modified"))
}

// Mock handler for testing interface compliance
type mockHandler struct {
	name            string
	priority        int
	requestResult   Result
	responseResult  Result
	requestCalled   bool
	responseCalled  bool
}

func (m *mockHandler) Name() string {
	return m.name
}

func (m *mockHandler) Priority() int {
	return m.priority
}

func (m *mockHandler) HandleRequest(ctx *HandlerContext, req *http.Request) Result {
	m.requestCalled = true
	return m.requestResult
}

func (m *mockHandler) HandleResponse(ctx *HandlerContext, res *http.Response) Result {
	m.responseCalled = true
	return m.responseResult
}

func TestHandler_InterfaceCompliance(t *testing.T) {
	handler := &mockHandler{
		name:          "test-handler",
		priority:      100,
		requestResult: Result{Action: ActionContinue},
	}

	// Verify interface compliance
	var _ Handler = handler

	assert.Equal(t, "test-handler", handler.Name())
	assert.Equal(t, 100, handler.Priority())
}

func TestHandler_HandleRequest(t *testing.T) {
	handler := &mockHandler{
		name:          "test-handler",
		priority:      100,
		requestResult: Result{Action: ActionContinue},
	}

	ctx := &HandlerContext{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		SessionID:  "sess-789",
		AgentID:    "agent-abc",
	}
	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)

	result := handler.HandleRequest(ctx, req)

	assert.True(t, handler.requestCalled)
	assert.Equal(t, ActionContinue, result.Action)
}

func TestHandler_HandleResponse(t *testing.T) {
	handler := &mockHandler{
		name:           "test-handler",
		priority:       100,
		responseResult: Result{Action: ActionContinue},
	}

	ctx := &HandlerContext{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		SessionID:  "sess-789",
		AgentID:    "agent-abc",
	}
	res := &http.Response{
		StatusCode: 200,
		Request:    httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil),
	}

	result := handler.HandleResponse(ctx, res)

	assert.True(t, handler.responseCalled)
	assert.Equal(t, ActionContinue, result.Action)
}

func TestNewHandlerContext(t *testing.T) {
	ctx := NewHandlerContext("emp-123", "org-456", "sess-789", "agent-abc")

	require.NotNil(t, ctx)
	assert.Equal(t, "emp-123", ctx.EmployeeID)
	assert.Equal(t, "org-456", ctx.OrgID)
	assert.Equal(t, "sess-789", ctx.SessionID)
	assert.Equal(t, "agent-abc", ctx.AgentID)
	assert.NotNil(t, ctx.Metadata)
}

func TestContinueResult(t *testing.T) {
	result := ContinueResult()

	assert.Equal(t, ActionContinue, result.Action)
	assert.True(t, result.ShouldContinue())
}

func TestBlockResult(t *testing.T) {
	result := BlockResult("test reason")

	assert.Equal(t, ActionBlock, result.Action)
	assert.Equal(t, "test reason", result.Reason)
	assert.True(t, result.ShouldBlock())
}

func TestErrorResult(t *testing.T) {
	result := ErrorResult(assert.AnError)

	assert.Equal(t, ActionContinue, result.Action)
	assert.Error(t, result.Error)
}
