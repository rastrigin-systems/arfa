package integration

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/rastrigin-systems/arfa/services/cli/internal/logging"
	"github.com/rastrigin-systems/arfa/services/cli/internal/proxy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestProxyWithLoggingIntegration tests the complete proxy + logging flow
func TestProxyWithLoggingIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create mock API client that captures logs
	mockAPI := &capturingAPIClient{
		logs: make([]logging.LogEntry, 0),
	}

	// Create logger
	loggerConfig := &logging.Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 100 * time.Millisecond,
		MaxRetries:    3,
		RetryBackoff:  50 * time.Millisecond,
	}

	logger, err := logging.NewLogger(loggerConfig, mockAPI)
	require.NoError(t, err)
	require.NotNil(t, logger)
	defer func() { _ = logger.Close() }()

	// Start session
	sessionID := logger.StartSession()
	clientName := "claude-code"
	clientVersion := "1.0.25"
	logger.SetClient(clientName, clientVersion)

	// Create proxy with logger
	p := proxy.New(logger)
	p.SetSession(sessionID.String(), clientName, clientVersion)

	err = p.Start()
	require.NoError(t, err)
	defer func() { _ = p.Stop() }()

	t.Logf("Proxy started on port %d", p.GetPort())

	// Create a test backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message": "hello from backend"}`))
	}))
	defer backend.Close()

	// Create HTTP client with proxy
	proxyURL, err := url.Parse(p.GetProxyURL())
	require.NoError(t, err)

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	// Make request through proxy
	resp, err := client.Get(backend.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// End session
	logger.EndSession()

	// Flush to ensure all logs are sent
	logger.Flush()

	// Wait for async operations
	time.Sleep(300 * time.Millisecond)

	// Verify session events were logged
	mockAPI.mu.Lock()
	logs := mockAPI.logs
	mockAPI.mu.Unlock()

	t.Logf("Captured %d log entries", len(logs))

	// Verify session_start and session_end events
	var foundStart, foundEnd bool
	for _, log := range logs {
		if log.EventType == "session_start" {
			foundStart = true
			assert.Equal(t, sessionID.String(), log.SessionID)
		}
		if log.EventType == "session_end" {
			foundEnd = true
			assert.Equal(t, sessionID.String(), log.SessionID)
		}
	}

	assert.True(t, foundStart, "Expected to find session_start event")
	assert.True(t, foundEnd, "Expected to find session_end event")
}

// TestProxyLoggingSessionPropagation tests that session ID is correctly propagated
func TestProxyLoggingSessionPropagation(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockAPI := &capturingAPIClient{
		logs: make([]logging.LogEntry, 0),
	}

	loggerConfig := &logging.Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 100 * time.Millisecond,
		MaxRetries:    3,
		RetryBackoff:  50 * time.Millisecond,
	}

	logger, err := logging.NewLogger(loggerConfig, mockAPI)
	require.NoError(t, err)
	defer func() { _ = logger.Close() }()

	// Set client info before starting session so session_start has it
	clientName := "claude-code"
	clientVersion := "1.0.25"
	logger.SetClient(clientName, clientVersion)

	sessionID := logger.StartSession()

	p := proxy.New(logger)
	p.SetSession(sessionID.String(), clientName, clientVersion)

	err = p.Start()
	require.NoError(t, err)
	defer func() { _ = p.Stop() }()

	// Log some events
	logger.LogEvent("custom_event_1", "test", "content 1", nil)
	logger.LogEvent("custom_event_2", "test", "content 2", map[string]interface{}{"key": "value"})
	logger.EndSession()

	logger.Flush()
	time.Sleep(200 * time.Millisecond)

	// Verify all events have correct session ID and agent ID
	mockAPI.mu.Lock()
	defer mockAPI.mu.Unlock()

	for _, log := range mockAPI.logs {
		assert.Equal(t, sessionID.String(), log.SessionID, "Log %s should have correct session ID", log.EventType)
		assert.Equal(t, clientName, log.ClientName, "Log %s should have correct client name", log.EventType)
	}
}

// TestMultipleProxiesWithDifferentSessions tests concurrent proxies with different sessions
func TestMultipleProxiesWithDifferentSessions(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockAPI := &capturingAPIClient{
		logs: make([]logging.LogEntry, 0),
	}

	loggerConfig := &logging.Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 100 * time.Millisecond,
		MaxRetries:    3,
		RetryBackoff:  50 * time.Millisecond,
	}

	logger, err := logging.NewLogger(loggerConfig, mockAPI)
	require.NoError(t, err)
	defer func() { _ = logger.Close() }()

	sessionID := logger.StartSession()
	logger.SetClient("multi-proxy-client", "1.0.0")

	// Start multiple proxies
	proxies := make([]*proxy.Proxy, 3)
	for i := 0; i < 3; i++ {
		p := proxy.New(logger)
		p.SetSession(sessionID.String(), "client-"+string(rune('A'+i)), "1.0."+string(rune('0'+i)))
		err := p.Start()
		require.NoError(t, err)
		proxies[i] = p
	}

	// Clean up all proxies
	defer func() {
		for _, p := range proxies {
			_ = p.Stop()
		}
	}()

	// Verify all proxies are on different ports
	ports := make(map[int]bool)
	for _, p := range proxies {
		port := p.GetPort()
		assert.False(t, ports[port], "Port %d should be unique", port)
		ports[port] = true
		t.Logf("Proxy on port %d", port)
	}

	// Create backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	// Make concurrent requests through different proxies
	var wg sync.WaitGroup
	for i, p := range proxies {
		wg.Add(1)
		go func(p *proxy.Proxy, id int) {
			defer wg.Done()

			proxyURL, _ := url.Parse(p.GetProxyURL())
			client := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				},
			}

			for j := 0; j < 5; j++ {
				resp, err := client.Get(backend.URL)
				if err == nil {
					_ = resp.Body.Close()
				}
			}
		}(p, i)
	}
	wg.Wait()

	logger.EndSession()
	logger.Flush()
	time.Sleep(200 * time.Millisecond)

	// Verify logs were captured
	mockAPI.mu.Lock()
	logCount := len(mockAPI.logs)
	mockAPI.mu.Unlock()

	assert.Greater(t, logCount, 0, "Should have captured logs")
}

// TestProxyLoggingDisabled tests that proxy works even when logging is disabled
func TestProxyLoggingDisabled(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create proxy without logger
	p := proxy.New(nil)
	p.SetSession("session-no-logger", "claude-code", "1.0.25")

	err := p.Start()
	require.NoError(t, err)
	defer func() { _ = p.Stop() }()

	// Create backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok": true}`))
	}))
	defer backend.Close()

	// Create HTTP client with proxy
	proxyURL, _ := url.Parse(p.GetProxyURL())
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	// Make request - should work without logging
	resp, err := client.Get(backend.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestProxyLoggingWithLargePayload tests handling of large request/response bodies
func TestProxyLoggingWithLargePayload(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockAPI := &capturingAPIClient{
		logs: make([]logging.LogEntry, 0),
	}

	loggerConfig := &logging.Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 100 * time.Millisecond,
		MaxRetries:    3,
		RetryBackoff:  50 * time.Millisecond,
	}

	logger, err := logging.NewLogger(loggerConfig, mockAPI)
	require.NoError(t, err)
	defer func() { _ = logger.Close() }()

	sessionID := logger.StartSession()
	logger.SetClient("large-payload-client", "1.0.25")

	p := proxy.New(logger)
	p.SetSession(sessionID.String(), "large-payload-client", "1.0.25")

	err = p.Start()
	require.NoError(t, err)
	defer func() { _ = p.Stop() }()

	// Create backend with large response
	largeResponse := make([]byte, 100*1024) // 100KB
	for i := range largeResponse {
		largeResponse[i] = byte('A' + (i % 26))
	}

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(largeResponse)
	}))
	defer backend.Close()

	proxyURL, _ := url.Parse(p.GetProxyURL())
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
	}

	resp, err := client.Get(backend.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	logger.EndSession()
	logger.Flush()
	time.Sleep(200 * time.Millisecond)
}

// TestProxyStopDuringRequest tests graceful proxy shutdown during active request
func TestProxyStopDuringRequest(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockAPI := &capturingAPIClient{
		logs: make([]logging.LogEntry, 0),
	}

	loggerConfig := &logging.Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 100 * time.Millisecond,
		MaxRetries:    3,
		RetryBackoff:  50 * time.Millisecond,
	}

	logger, err := logging.NewLogger(loggerConfig, mockAPI)
	require.NoError(t, err)
	defer func() { _ = logger.Close() }()

	sessionID := logger.StartSession()
	logger.SetClient("graceful-client", "1.0.25")

	p := proxy.New(logger)
	p.SetSession(sessionID.String(), "graceful-client", "1.0.25")

	err = p.Start()
	require.NoError(t, err)

	// Create slow backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	proxyURL, _ := url.Parse(p.GetProxyURL())
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(proxyURL),
		},
		Timeout: 5 * time.Second,
	}

	// Start request in background
	done := make(chan struct{})
	go func() {
		resp, _ := client.Get(backend.URL)
		if resp != nil {
			_ = resp.Body.Close()
		}
		close(done)
	}()

	// Stop proxy while request is in progress
	time.Sleep(50 * time.Millisecond)
	err = p.Stop()
	assert.NoError(t, err, "Proxy should stop gracefully")

	// Wait for request to complete (with timeout)
	select {
	case <-done:
		// Request completed or failed - both are acceptable
	case <-time.After(3 * time.Second):
		t.Fatal("Request did not complete within timeout")
	}

	logger.EndSession()
	logger.Flush()
}

// TestProxyLoggingEventTypes tests that different event types are correctly logged
func TestProxyLoggingEventTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockAPI := &capturingAPIClient{
		logs: make([]logging.LogEntry, 0),
	}

	loggerConfig := &logging.Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 100 * time.Millisecond,
		MaxRetries:    3,
		RetryBackoff:  50 * time.Millisecond,
	}

	logger, err := logging.NewLogger(loggerConfig, mockAPI)
	require.NoError(t, err)
	defer func() { _ = logger.Close() }()

	sessionID := logger.StartSession()
	logger.SetClient("event-types-client", "1.0.25")

	p := proxy.New(logger)
	p.SetSession(sessionID.String(), "event-types-client", "1.0.25")

	err = p.Start()
	require.NoError(t, err)
	defer func() { _ = p.Stop() }()

	// Log different event types
	logger.LogEvent("api_request", "proxy", "POST /v1/messages", map[string]interface{}{"method": "POST"})
	logger.LogEvent("api_response", "proxy", "200 OK", map[string]interface{}{"status": 200})
	logger.LogEvent("custom_event", "app", "custom content", nil)

	logger.EndSession()
	logger.Flush()
	time.Sleep(200 * time.Millisecond)

	// Verify all event types were logged
	mockAPI.mu.Lock()
	defer mockAPI.mu.Unlock()

	eventTypes := make(map[string]bool)
	for _, log := range mockAPI.logs {
		eventTypes[log.EventType] = true
	}

	assert.True(t, eventTypes["session_start"], "Should have session_start event")
	assert.True(t, eventTypes["session_end"], "Should have session_end event")
	assert.True(t, eventTypes["api_request"], "Should have api_request event")
	assert.True(t, eventTypes["api_response"], "Should have api_response event")
	assert.True(t, eventTypes["custom_event"], "Should have custom_event event")
}

// TestProxyRestartWithSameLogger tests restarting proxy with the same logger
func TestProxyRestartWithSameLogger(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	mockAPI := &capturingAPIClient{
		logs: make([]logging.LogEntry, 0),
	}

	loggerConfig := &logging.Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 100 * time.Millisecond,
		MaxRetries:    3,
		RetryBackoff:  50 * time.Millisecond,
	}

	logger, err := logging.NewLogger(loggerConfig, mockAPI)
	require.NoError(t, err)
	defer func() { _ = logger.Close() }()

	sessionID := logger.StartSession()
	logger.SetClient("restart-client", "1.0.25")

	// First proxy instance
	p1 := proxy.New(logger)
	p1.SetSession(sessionID.String(), "restart-client", "1.0.25")

	err = p1.Start()
	require.NoError(t, err)

	port1 := p1.GetPort()
	t.Logf("First proxy on port %d", port1)

	err = p1.Stop()
	require.NoError(t, err)

	// Second proxy instance with same logger
	p2 := proxy.New(logger)
	p2.SetSession(sessionID.String(), "restart-client", "1.0.25")

	err = p2.Start()
	require.NoError(t, err)
	defer func() { _ = p2.Stop() }()

	port2 := p2.GetPort()
	t.Logf("Second proxy on port %d", port2)

	// Should get the same port back (since first was stopped)
	assert.Equal(t, port1, port2, "Should reuse the same port after stop")

	logger.EndSession()
	logger.Flush()
	time.Sleep(200 * time.Millisecond)
}
