package proxy

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/rastrigin-systems/arfa/services/cli/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLogger implements Logger interface for testing with thread-safety
type mockLogger struct {
	mu         sync.Mutex
	events     []logEvent
	classified []types.ClassifiedLogEntry
}

type logEvent struct {
	eventType string
	category  string
	content   string
	metadata  map[string]interface{}
}

func (m *mockLogger) LogEvent(eventType, category, content string, metadata map[string]interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, logEvent{eventType, category, content, metadata})
}

func (m *mockLogger) LogClassified(entry types.ClassifiedLogEntry) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.classified = append(m.classified, entry)
}

func TestNew(t *testing.T) {
	logger := &mockLogger{}
	p := New(logger)

	assert.NotNil(t, p)
	assert.NotNil(t, p.goproxy)
	assert.NotNil(t, p.parser)
	assert.Equal(t, logger, p.logger)
}

func TestSetSession(t *testing.T) {
	p := New(nil)
	p.SetSession("session-123", "claude-code", "1.0.25")

	assert.Equal(t, "session-123", p.sessionID)
	assert.Equal(t, "claude-code", p.clientName)
	assert.Equal(t, "1.0.25", p.clientVersion)
}

func TestStartStop(t *testing.T) {
	logger := &mockLogger{}
	p := New(logger)

	err := p.Start()
	require.NoError(t, err)
	assert.GreaterOrEqual(t, p.GetPort(), MinPort)
	assert.LessOrEqual(t, p.GetPort(), MaxPort)
	assert.NotEmpty(t, p.GetProxyURL())
	assert.NotEmpty(t, p.GetCertPath())

	err = p.Stop()
	assert.NoError(t, err)
}

func TestGetProxyURL(t *testing.T) {
	p := &Proxy{port: 8082}
	assert.Equal(t, "http://127.0.0.1:8082", p.GetProxyURL())
}

func TestMultipleInstances(t *testing.T) {
	// Test that multiple proxies can start on different ports
	logger := &mockLogger{}

	p1 := New(logger)
	err := p1.Start()
	require.NoError(t, err)
	defer func() { _ = p1.Stop() }()

	p2 := New(logger)
	err = p2.Start()
	require.NoError(t, err)
	defer func() { _ = p2.Stop() }()

	// They should be on different ports
	assert.NotEqual(t, p1.GetPort(), p2.GetPort())
}

func TestRedactHeaders(t *testing.T) {
	headers := http.Header{
		"Content-Type":    {"application/json"},
		"Authorization":   {"Bearer secret-token"},
		"X-Api-Key":       {"secret-key"},
		"X-Custom-Header": {"visible-value"},
	}

	redacted := redactHeaders(headers)

	assert.Equal(t, "application/json", redacted["Content-Type"])
	assert.Equal(t, "[REDACTED]", redacted["Authorization"])
	assert.Equal(t, "[REDACTED]", redacted["X-Api-Key"])
	assert.Equal(t, "visible-value", redacted["X-Custom-Header"])
}

func TestLLMHostRegex(t *testing.T) {
	tests := []struct {
		host    string
		matches bool
	}{
		{"api.anthropic.com", true},
		{"api.openai.com", true},
		{"generativelanguage.googleapis.com", true},
		{"example.com", false},
		{"github.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.host, func(t *testing.T) {
			assert.Equal(t, tt.matches, llmHostRegex.MatchString(tt.host))
		})
	}
}

func TestProxyIntegration(t *testing.T) {
	// Create a test backend that mimics an LLM API
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"response": "hello"}`))
	}))
	defer backend.Close()

	logger := &mockLogger{}
	p := New(logger)
	p.SetSession("test-session", "claude-code", "1.0.25")

	err := p.Start()
	require.NoError(t, err)
	defer func() { _ = p.Stop() }()

	// Make a request through the proxy to the test backend
	// Note: This tests the proxy machinery but not HTTPS interception
	// (which would require trusting the CA cert)
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(mustParseURL(t, p.GetProxyURL())),
		},
	}

	resp, err := client.Get(backend.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func mustParseURL(t *testing.T, rawURL string) *url.URL {
	u, err := url.Parse(rawURL)
	require.NoError(t, err)
	return u
}

// TestNewWithNilLogger tests that proxy can be created without a logger
func TestNewWithNilLogger(t *testing.T) {
	p := New(nil)
	assert.NotNil(t, p)
	assert.Nil(t, p.logger)
	assert.NotNil(t, p.goproxy)
	assert.NotNil(t, p.parser)
}

// TestStopWithoutStart tests stopping a proxy that was never started
func TestStopWithoutStart(t *testing.T) {
	p := New(nil)
	err := p.Stop()
	assert.NoError(t, err) // Should not error
}

// TestDoubleStop tests stopping a proxy twice
func TestDoubleStop(t *testing.T) {
	logger := &mockLogger{}
	p := New(logger)

	err := p.Start()
	require.NoError(t, err)

	err = p.Stop()
	assert.NoError(t, err)

	err = p.Stop()
	assert.NoError(t, err) // Second stop should also succeed
}

// TestCAGenerationAndLoading tests CA certificate generation and loading
func TestCAGenerationAndLoading(t *testing.T) {
	logger := &mockLogger{}
	p := New(logger)

	err := p.Start()
	require.NoError(t, err)
	defer func() { _ = p.Stop() }()

	certPath := p.GetCertPath()
	assert.NotEmpty(t, certPath)
	assert.FileExists(t, certPath)

	// Verify the cert file contains PEM data
	certData, err := os.ReadFile(certPath)
	require.NoError(t, err)
	assert.Contains(t, string(certData), "-----BEGIN CERTIFICATE-----")
	assert.Contains(t, string(certData), "-----END CERTIFICATE-----")

	// Verify key file also exists
	keyPath := filepath.Join(filepath.Dir(certPath), "arfa-ca-key.pem")
	assert.FileExists(t, keyPath)

	keyData, err := os.ReadFile(keyPath)
	require.NoError(t, err)
	assert.Contains(t, string(keyData), "-----BEGIN RSA PRIVATE KEY-----")
}

// TestCAReuse tests that existing CA is reused on subsequent starts
func TestCAReuse(t *testing.T) {
	logger := &mockLogger{}

	// Start first proxy
	p1 := New(logger)
	err := p1.Start()
	require.NoError(t, err)

	certPath1 := p1.GetCertPath()
	certData1, err := os.ReadFile(certPath1)
	require.NoError(t, err)

	_ = p1.Stop()

	// Start second proxy - should reuse the same CA
	p2 := New(logger)
	err = p2.Start()
	require.NoError(t, err)
	defer func() { _ = p2.Stop() }()

	certPath2 := p2.GetCertPath()
	certData2, err := os.ReadFile(certPath2)
	require.NoError(t, err)

	// Should be the same certificate
	assert.Equal(t, certPath1, certPath2)
	assert.Equal(t, certData1, certData2)
}

// TestPortRange tests that proxies use the expected port range
func TestPortRange(t *testing.T) {
	var proxies []*Proxy
	var ports []int

	// Start multiple proxies
	for i := 0; i < 5; i++ {
		p := New(&mockLogger{})
		err := p.Start()
		require.NoError(t, err)
		proxies = append(proxies, p)
		ports = append(ports, p.GetPort())
	}

	// Clean up
	defer func() {
		for _, p := range proxies {
			_ = p.Stop()
		}
	}()

	// All ports should be within range
	for _, port := range ports {
		assert.GreaterOrEqual(t, port, MinPort)
		assert.LessOrEqual(t, port, MaxPort)
	}

	// All ports should be unique
	portSet := make(map[int]bool)
	for _, port := range ports {
		assert.False(t, portSet[port], "Port %d was reused", port)
		portSet[port] = true
	}
}

// TestConcurrentProxyAccess tests thread-safety of proxy operations
func TestConcurrentProxyAccess(t *testing.T) {
	logger := &mockLogger{}
	p := New(logger)
	p.SetSession("test-session", "claude-code", "1.0.25")

	err := p.Start()
	require.NoError(t, err)
	defer func() { _ = p.Stop() }()

	// Concurrent SetSession calls
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			p.SetSession("session-"+string(rune('a'+id)), "client-"+string(rune('a'+id)), "1.0."+string(rune('0'+id)))
		}(i)
	}
	wg.Wait()

	// Concurrent GetPort and GetProxyURL calls
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = p.GetPort()
			_ = p.GetProxyURL()
			_ = p.GetCertPath()
		}()
	}
	wg.Wait()
}

// TestRedactHeadersComprehensive tests header redaction with various patterns
func TestRedactHeadersComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		headers  http.Header
		expected map[string]string
	}{
		{
			name: "authorization variants",
			headers: http.Header{
				"Authorization":       {"Bearer token"},
				"Proxy-Authorization": {"Basic creds"},
				"X-Authorization":     {"custom"},
			},
			expected: map[string]string{
				"Authorization":       "[REDACTED]",
				"Proxy-Authorization": "[REDACTED]",
				"X-Authorization":     "[REDACTED]",
			},
		},
		{
			name: "api key variants",
			headers: http.Header{
				"X-Api-Key":         {"key1"},
				"Api-Key":           {"key2"},
				"X-API-KEY":         {"key3"},
				"Anthropic-Api-Key": {"key4"},
			},
			expected: map[string]string{
				"X-Api-Key":         "[REDACTED]",
				"Api-Key":           "[REDACTED]",
				"X-API-KEY":         "[REDACTED]",
				"Anthropic-Api-Key": "[REDACTED]",
			},
		},
		{
			name: "token variants",
			headers: http.Header{
				"X-Auth-Token":  {"token1"},
				"Access-Token":  {"token2"},
				"Refresh-Token": {"token3"},
			},
			expected: map[string]string{
				"X-Auth-Token":  "[REDACTED]",
				"Access-Token":  "[REDACTED]",
				"Refresh-Token": "[REDACTED]",
			},
		},
		{
			name: "cookie variants",
			headers: http.Header{
				"Cookie":     {"session=abc123"},
				"Set-Cookie": {"session=abc123; HttpOnly"},
			},
			expected: map[string]string{
				"Cookie":     "[REDACTED]",
				"Set-Cookie": "[REDACTED]",
			},
		},
		{
			name: "safe headers",
			headers: http.Header{
				"Content-Type":   {"application/json"},
				"Accept":         {"*/*"},
				"User-Agent":     {"test-agent/1.0"},
				"Content-Length": {"100"},
				"X-Request-Id":   {"req-123"},
			},
			expected: map[string]string{
				"Content-Type":   "application/json",
				"Accept":         "*/*",
				"User-Agent":     "test-agent/1.0",
				"Content-Length": "100",
				"X-Request-Id":   "req-123",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redacted := redactHeaders(tt.headers)
			for key, expectedVal := range tt.expected {
				assert.Equal(t, expectedVal, redacted[key], "Header %s", key)
			}
		})
	}
}

// TestLLMHostRegexComprehensive tests LLM host matching with various hosts
func TestLLMHostRegexComprehensive(t *testing.T) {
	tests := []struct {
		host    string
		matches bool
		desc    string
	}{
		// Should match
		{"api.anthropic.com", true, "Anthropic API"},
		{"api.openai.com", true, "OpenAI API"},
		{"generativelanguage.googleapis.com", true, "Google Gemini API"},

		// Should NOT match
		{"example.com", false, "Generic domain"},
		{"github.com", false, "GitHub"},
		{"anthropic.com", false, "Anthropic main site (not API)"},
		{"openai.com", false, "OpenAI main site (not API)"},
		{"googleapis.com", false, "Google APIs (not generative)"},
		{"api.github.com", false, "GitHub API"},
		{"api.example.com", false, "Generic API"},
		{"api-anthropic.com", false, "Similar but different domain"},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			result := llmHostRegex.MatchString(tt.host)
			assert.Equal(t, tt.matches, result, "Host: %s", tt.host)
		})
	}
}

// TestProxyWithSessionMetadata tests that session metadata is included in logs
func TestProxyWithSessionMetadata(t *testing.T) {
	logger := &mockLogger{}
	p := New(logger)

	sessionID := "test-session-12345"
	clientName := "claude-code"
	clientVersion := "1.0.25"
	p.SetSession(sessionID, clientName, clientVersion)

	err := p.Start()
	require.NoError(t, err)
	defer func() { _ = p.Stop() }()

	// Verify session is set
	assert.Equal(t, sessionID, p.sessionID)
	assert.Equal(t, clientName, p.clientName)
	assert.Equal(t, clientVersion, p.clientVersion)
}

// TestProxyHTTPRequest tests HTTP (non-HTTPS) request proxying
func TestProxyHTTPRequest(t *testing.T) {
	// Create a test backend
	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "ok"}`))
	}))
	defer backend.Close()

	logger := &mockLogger{}
	p := New(logger)
	p.SetSession("session-1", "claude-code", "1.0.25")

	err := p.Start()
	require.NoError(t, err)
	defer func() { _ = p.Stop() }()

	// Create client with proxy
	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(mustParseURL(t, p.GetProxyURL())),
		},
	}

	// Make request through proxy
	resp, err := client.Get(backend.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestProxyMethods tests different HTTP methods
func TestProxyMethods(t *testing.T) {
	methods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}

	for _, method := range methods {
		t.Run(method, func(t *testing.T) {
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, method, r.Method)
				w.WriteHeader(http.StatusOK)
			}))
			defer backend.Close()

			logger := &mockLogger{}
			p := New(logger)

			err := p.Start()
			require.NoError(t, err)
			defer func() { _ = p.Stop() }()

			client := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(mustParseURL(t, p.GetProxyURL())),
				},
			}

			req, err := http.NewRequest(method, backend.URL, nil)
			require.NoError(t, err)

			resp, err := client.Do(req)
			require.NoError(t, err)
			_ = resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
		})
	}
}

// TestProxyErrorHandling tests proxy behavior with backend errors
func TestProxyErrorHandling(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"BadRequest", http.StatusBadRequest},
		{"Unauthorized", http.StatusUnauthorized},
		{"NotFound", http.StatusNotFound},
		{"InternalServerError", http.StatusInternalServerError},
		{"ServiceUnavailable", http.StatusServiceUnavailable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer backend.Close()

			logger := &mockLogger{}
			p := New(logger)

			err := p.Start()
			require.NoError(t, err)
			defer func() { _ = p.Stop() }()

			client := &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(mustParseURL(t, p.GetProxyURL())),
				},
			}

			resp, err := client.Get(backend.URL)
			require.NoError(t, err)
			_ = resp.Body.Close()

			// Proxy should pass through the error status code
			assert.Equal(t, tt.statusCode, resp.StatusCode)
		})
	}
}
