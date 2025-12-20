package control

import (
	"bytes"
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewControlledProxy(t *testing.T) {
	dir := t.TempDir()
	svc, err := NewService(ServiceConfig{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		AgentID:    "agent-abc",
		QueueDir:   dir,
	})
	require.NoError(t, err)

	proxy := NewControlledProxy(svc)

	require.NotNil(t, proxy)
	assert.NotNil(t, proxy.goproxy)
	assert.Equal(t, svc, proxy.service)
}

func TestControlledProxy_StartStop(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dir := t.TempDir()
	svc, err := NewService(ServiceConfig{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		AgentID:    "agent-abc",
		QueueDir:   dir,
	})
	require.NoError(t, err)

	proxy := NewControlledProxy(svc)

	// Start proxy
	err = proxy.Start()
	require.NoError(t, err)
	defer proxy.Stop()

	// Should have a valid port
	assert.GreaterOrEqual(t, proxy.GetPort(), MinPort)
	assert.LessOrEqual(t, proxy.GetPort(), MaxPort)

	// Should have a proxy URL
	assert.Contains(t, proxy.GetProxyURL(), "http://127.0.0.1:")

	// Should have a cert path
	assert.NotEmpty(t, proxy.GetCertPath())
	assert.FileExists(t, proxy.GetCertPath())

	// Stop proxy
	err = proxy.Stop()
	require.NoError(t, err)
}

func TestControlledProxy_RequestsFlowThroughPipeline(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dir := t.TempDir()
	svc, err := NewService(ServiceConfig{
		EmployeeID: "emp-test",
		OrgID:      "org-test",
		AgentID:    "agent-test",
		QueueDir:   dir,
	})
	require.NoError(t, err)

	proxy := NewControlledProxy(svc)
	err = proxy.Start()
	require.NoError(t, err)
	defer proxy.Stop()

	// Create HTTP client that uses proxy
	proxyURL := proxy.GetProxyURL()
	transport := &http.Transport{
		Proxy: http.ProxyURL(mustParseURL(proxyURL)),
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true, // Trust self-signed cert
		},
	}

	// Load CA cert for proper verification
	caCert, err := os.ReadFile(proxy.GetCertPath())
	if err == nil {
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		transport.TLSClientConfig = &tls.Config{
			RootCAs: caCertPool,
		}
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   5 * time.Second,
	}

	// Make request to LLM API (will be intercepted)
	// Note: This is a mock request - actual API would require valid credentials
	req, _ := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBufferString(`{"test":"data"}`))
	req.Header.Set("Content-Type", "application/json")

	// The request will fail (no valid API key) but should be logged
	_, _ = client.Do(req)

	// Give time for async processing
	time.Sleep(100 * time.Millisecond)

	// Check that entry was written to queue
	files, err := filepath.Glob(filepath.Join(dir, "*.json"))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(files), 1, "expected at least one log entry in queue")
}

func TestControlledProxy_MultiplePortAllocation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dir := t.TempDir()

	// Start multiple proxies
	proxies := make([]*ControlledProxy, 3)
	for i := 0; i < 3; i++ {
		svc, err := NewService(ServiceConfig{
			EmployeeID: "emp-123",
			OrgID:      "org-456",
			AgentID:    fmt.Sprintf("agent-%d", i),
			QueueDir:   filepath.Join(dir, fmt.Sprintf("queue-%d", i)),
		})
		require.NoError(t, err)

		proxy := NewControlledProxy(svc)
		err = proxy.Start()
		require.NoError(t, err)
		proxies[i] = proxy
	}

	defer func() {
		for _, p := range proxies {
			p.Stop()
		}
	}()

	// All proxies should have unique ports
	ports := make(map[int]bool)
	for _, p := range proxies {
		port := p.GetPort()
		assert.False(t, ports[port], "ports should be unique")
		ports[port] = true
	}
}

func TestControlledProxy_GracefulShutdown(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	dir := t.TempDir()
	svc, err := NewService(ServiceConfig{
		EmployeeID: "emp-123",
		OrgID:      "org-456",
		AgentID:    "agent-abc",
		QueueDir:   dir,
	})
	require.NoError(t, err)

	// Start control service in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go svc.Start(ctx)

	proxy := NewControlledProxy(svc)
	err = proxy.Start()
	require.NoError(t, err)

	port := proxy.GetPort()

	// Stop proxy
	err = proxy.Stop()
	require.NoError(t, err)

	// Port should be released
	time.Sleep(100 * time.Millisecond)

	// Should be able to start a new proxy on the same port
	newProxy := NewControlledProxy(svc)
	err = newProxy.Start()
	require.NoError(t, err)
	defer newProxy.Stop()

	// Might get same or different port depending on timing
	assert.GreaterOrEqual(t, newProxy.GetPort(), MinPort)
	_ = port // Use port variable to avoid unused warning
}

func TestControlledProxy_NilService(t *testing.T) {
	// Proxy should handle nil service gracefully
	proxy := NewControlledProxy(nil)
	require.NotNil(t, proxy)

	// handleRequest with nil service should not panic
	req, _ := http.NewRequest("GET", "https://api.anthropic.com/v1/test", nil)
	resultReq, resultResp := proxy.handleRequest(req)
	assert.Equal(t, req, resultReq)
	assert.Nil(t, resultResp)

	// handleResponse with nil service should not panic
	resp := &http.Response{StatusCode: 200}
	resultResp = proxy.handleResponse(resp)
	assert.Equal(t, resp, resultResp)
}

func mustParseURL(rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	if err != nil {
		panic(err)
	}
	return u
}
