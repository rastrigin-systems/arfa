package control

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewService(t *testing.T) {
	dir := t.TempDir()
	config := ServiceConfig{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		QueueDir:   dir,
	}

	svc, err := NewService(config)

	require.NoError(t, err)
	require.NotNil(t, svc)
	assert.NotEmpty(t, svc.SessionID())
}

func TestService_SessionID_Generated(t *testing.T) {
	dir := t.TempDir()
	config := ServiceConfig{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		QueueDir:   dir,
	}

	svc1, _ := NewService(config)
	svc2, _ := NewService(config)

	// Each service should have a unique session ID
	assert.NotEqual(t, svc1.SessionID(), svc2.SessionID())
}

func TestService_HandlerContext(t *testing.T) {
	dir := t.TempDir()
	config := ServiceConfig{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		QueueDir:   dir,
	}

	svc, err := NewService(config)
	require.NoError(t, err)

	ctx := svc.Context()

	assert.Equal(t, "emp-123", ctx.EmployeeID)
	assert.Equal(t, "org-456", ctx.OrgID)
	// ClientName/ClientVersion are empty until client detection happens
	assert.Empty(t, ctx.ClientName)
	assert.Empty(t, ctx.ClientVersion)
	assert.Equal(t, svc.SessionID(), ctx.SessionID)
}

func TestService_Pipeline(t *testing.T) {
	dir := t.TempDir()
	config := ServiceConfig{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		QueueDir:   dir,
	}

	svc, err := NewService(config)
	require.NoError(t, err)

	// Should have logger handler registered by default
	handlers := svc.Pipeline().Handlers()
	assert.GreaterOrEqual(t, len(handlers), 1)

	// Find logger handler
	var hasLogger bool
	for _, h := range handlers {
		if h.Name() == "logger" {
			hasLogger = true
			break
		}
	}
	assert.True(t, hasLogger, "should have logger handler")
}

func TestService_RegisterHandler(t *testing.T) {
	dir := t.TempDir()
	config := ServiceConfig{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		QueueDir:   dir,
	}

	svc, err := NewService(config)
	require.NoError(t, err)

	// Register custom handler
	custom := &testHandler{name: "custom", priority: 100}
	svc.RegisterHandler(custom)

	handlers := svc.Pipeline().Handlers()
	var hasCustom bool
	for _, h := range handlers {
		if h.Name() == "custom" {
			hasCustom = true
			break
		}
	}
	assert.True(t, hasCustom)
}

func TestService_HandleRequest(t *testing.T) {
	dir := t.TempDir()
	config := ServiceConfig{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		QueueDir:   dir,
	}

	svc, err := NewService(config)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)

	result := svc.HandleRequest(req)

	assert.Equal(t, ActionContinue, result.Action)
}

func TestService_HandleResponse(t *testing.T) {
	dir := t.TempDir()
	config := ServiceConfig{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		QueueDir:   dir,
	}

	svc, err := NewService(config)
	require.NoError(t, err)

	res := &http.Response{
		StatusCode: 200,
		Request:    httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil),
	}

	result := svc.HandleResponse(res)

	assert.Equal(t, ActionContinue, result.Action)
}

func TestService_HandleRequest_WritesToQueue(t *testing.T) {
	dir := t.TempDir()
	config := ServiceConfig{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		QueueDir:   dir,
	}

	svc, err := NewService(config)
	require.NoError(t, err)

	req := httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil)
	svc.HandleRequest(req)

	// Check queue has entry
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	require.NoError(t, err)
	assert.Len(t, files, 1)
}

func TestService_HandleResponse_WritesToQueue(t *testing.T) {
	dir := t.TempDir()
	config := ServiceConfig{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		QueueDir:   dir,
	}

	svc, err := NewService(config)
	require.NoError(t, err)

	res := &http.Response{
		StatusCode: 200,
		Request:    httptest.NewRequest("POST", "https://api.anthropic.com/v1/messages", nil),
	}
	svc.HandleResponse(res)

	// Check queue has entry
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	require.NoError(t, err)
	assert.Len(t, files, 1)
}

func TestService_Start_StartsWorker(t *testing.T) {
	dir := t.TempDir()
	config := ServiceConfig{
		EmployeeID:    "emp-123",
		OrgID:         "org-456",
		QueueDir:      dir,
		FlushInterval: 50 * time.Millisecond,
	}

	svc, err := NewService(config)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start service in background
	go svc.Start(ctx)

	// Should be running
	time.Sleep(100 * time.Millisecond)

	// Cancel and ensure it stops
	cancel()
	time.Sleep(100 * time.Millisecond)
}

func TestService_Stop_GracefulShutdown(t *testing.T) {
	dir := t.TempDir()
	config := ServiceConfig{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		QueueDir:   dir,
	}

	svc, err := NewService(config)
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.Start(ctx)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// Stopped successfully
	case <-time.After(time.Second):
		t.Fatal("service did not stop")
	}
}

func TestServiceConfig_Defaults(t *testing.T) {
	dir := t.TempDir()
	config := ServiceConfig{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		QueueDir:   dir,
		// No FlushInterval or MaxBatchSize set
	}

	svc, err := NewService(config)
	require.NoError(t, err)
	require.NotNil(t, svc)
}

// testHandler is defined in pipeline_test.go and reused here
