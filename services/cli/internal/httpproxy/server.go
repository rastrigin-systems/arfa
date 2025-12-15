package httpproxy

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/logging"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/logparser"
)

// sensitiveHeaderRegex matches headers that should be redacted in logs
var sensitiveHeaderRegex = regexp.MustCompile(`(?i)(auth|api-key|token|cookie)`)

// ProxyServer manages the embedded MITM proxy
type ProxyServer struct {
	proxy           *goproxy.ProxyHttpServer
	logger          logging.Logger
	certPath        string
	keyPath         string
	server          *http.Server
	anthropicParser *logparser.AnthropicParser
	port            int

	// Active session tracking (set via control socket)
	activeSessionID string
	activeAgentID   string
	sessionMu       sync.RWMutex
}

// NewProxyServer creates a new proxy server instance
func NewProxyServer(logger logging.Logger) *ProxyServer {
	return &ProxyServer{
		proxy:           goproxy.NewProxyHttpServer(),
		logger:          logger,
		anthropicParser: logparser.NewAnthropicParser(),
	}
}

// Start starts the proxy server on the specified port
func (s *ProxyServer) Start(port int) error {
	// 1. Setup CA certificates
	if err := s.setupCA(); err != nil {
		return fmt.Errorf("failed to setup CA: %w", err)
	}

	// 2. Configure proxy rules
	s.configureRules()

	// 3. Start server
	addr := fmt.Sprintf(":%d", port)
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.proxy,
	}

	s.port = port

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Proxy server error: %v\n", err)
		}
	}()

	fmt.Printf("✓ Proxy server started on %s\n", addr)
	return nil
}

// GetPort returns the port the proxy is running on
func (s *ProxyServer) GetPort() int {
	return s.port
}

// Stop stops the proxy server
func (s *ProxyServer) Stop(ctx context.Context) error {
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}

// GetCAPath returns the path to the CA certificate
func (s *ProxyServer) GetCAPath() string {
	return s.certPath
}

// SetSession sets the active session for logging API requests
func (s *ProxyServer) SetSession(sessionID, agentID string) {
	s.sessionMu.Lock()
	defer s.sessionMu.Unlock()
	s.activeSessionID = sessionID
	s.activeAgentID = agentID
}

// getActiveSession returns the currently active session info
func (s *ProxyServer) getActiveSession() (sessionID, agentID string) {
	s.sessionMu.RLock()
	defer s.sessionMu.RUnlock()
	return s.activeSessionID, s.activeAgentID
}

// setupCA ensures the CA certificate exists or generates it
func (s *ProxyServer) setupCA() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	certDir := filepath.Join(home, ".ubik", "certs")
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return err
	}

	s.certPath = filepath.Join(certDir, "ubik-ca.pem")
	s.keyPath = filepath.Join(certDir, "ubik-ca-key.pem")

	// Check if certs exist
	if _, err := os.Stat(s.certPath); err == nil {
		if _, err := os.Stat(s.keyPath); err == nil {
			// Load existing CA
			caCert, err := tls.LoadX509KeyPair(s.certPath, s.keyPath)
			if err != nil {
				return fmt.Errorf("failed to load existing CA: %w", err)
			}

			// Configure goproxy to use this CA
			s.proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
			goproxy.GoproxyCa = caCert
			goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&caCert)}
			goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&caCert)}
			goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&caCert)}
			goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&caCert)}

			return nil
		}
	}

	fmt.Println("Generating new CA certificate for MITM proxy...")

	// Generate new CA
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(2025),
		Subject: pkix.Name{
			Organization: []string{"Ubik Enterprise Proxy CA"},
			CommonName:   "ubik-proxy-ca",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0), // 10 years
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}

	// Save Cert
	certOut, err := os.Create(s.certPath)
	if err != nil {
		return err
	}
	pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes})
	certOut.Close()

	// Save Key
	keyOut, err := os.Create(s.keyPath)
	if err != nil {
		return err
	}
	pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey)})
	keyOut.Close()

	// Load the new CA into goproxy
	caCert, err := tls.LoadX509KeyPair(s.certPath, s.keyPath)
	if err != nil {
		return err
	}

	s.proxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	goproxy.GoproxyCa = caCert
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: goproxy.TLSConfigFromCA(&caCert)}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: goproxy.TLSConfigFromCA(&caCert)}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: goproxy.TLSConfigFromCA(&caCert)}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: goproxy.TLSConfigFromCA(&caCert)}

	return nil
}

// configureRules sets up interception rules for LLM providers
func (s *ProxyServer) configureRules() {
	// Targets: Anthropic, Google, OpenAI
	hostRegex := regexp.MustCompile(`(api.anthropic.com|generativelanguage.googleapis.com|api.openai.com)`)

	s.proxy.OnRequest(goproxy.ReqHostMatches(hostRegex)).DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			s.logRequest(r)
			return r, nil
		})

	s.proxy.OnResponse(goproxy.ReqHostMatches(hostRegex)).DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			s.logResponse(resp)
			return resp
		})
	// goproxy passes through all other requests by default
}

// logRequest captures and logs the request
func (s *ProxyServer) logRequest(r *http.Request) {
	if s.logger == nil {
		return
	}

	// Capture body
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body

	// Get session from control socket registration
	sessionID, agentID := s.getActiveSession()

	// Parse and classify based on provider
	if strings.Contains(r.URL.Host, "anthropic.com") && len(bodyBytes) > 0 {
		entries, err := s.anthropicParser.ParseRequest(bodyBytes)
		if err == nil {
			for _, entry := range entries {
				s.logger.LogClassified(entry)
			}
		}
	}

	// Also log raw request for debugging
	payload := map[string]interface{}{
		"method":  r.Method,
		"url":     r.URL.String(),
		"headers": redactHeaders(r.Header),
		"body":    string(bodyBytes),
	}

	// Add session tracking if available
	if sessionID != "" {
		payload["session_id"] = sessionID
	}
	if agentID != "" {
		payload["agent_id"] = agentID
	}

	s.logger.LogEvent("api_request", "proxy", fmt.Sprintf("%s %s", r.Method, r.URL.Host), payload)
}

// logResponse captures and logs the response
func (s *ProxyServer) logResponse(resp *http.Response) {
	if s.logger == nil {
		return
	}

	// Capture body (handle GZIP)
	bodyBytes, _ := io.ReadAll(resp.Body)
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // Restore body

	var decodedBody []byte
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(bytes.NewBuffer(bodyBytes))
		if err == nil {
			decodedBody, _ = io.ReadAll(reader)
			reader.Close()
		}
	} else {
		decodedBody = bodyBytes
	}

	// Get session from control socket registration
	sessionID, agentID := s.getActiveSession()

	// Parse and classify based on provider
	if resp.Request != nil && strings.Contains(resp.Request.URL.Host, "anthropic.com") && len(decodedBody) > 0 {
		entries, err := s.anthropicParser.ParseResponse(decodedBody)
		if err == nil {
			for _, entry := range entries {
				s.logger.LogClassified(entry)
			}
		}
	}

	// Log raw response
	payload := map[string]interface{}{
		"status":  resp.StatusCode,
		"headers": redactHeaders(resp.Header),
		"body":    string(decodedBody),
	}

	// Add session tracking if available
	if sessionID != "" {
		payload["session_id"] = sessionID
	}
	if agentID != "" {
		payload["agent_id"] = agentID
	}

	s.logger.LogEvent("api_response", "proxy", fmt.Sprintf("%d %s", resp.StatusCode, resp.Request.URL.Host), payload)
}

func redactHeaders(headers http.Header) map[string]string {
	redacted := make(map[string]string)
	for k, v := range headers {
		key := k // simplified
		// Redact sensitive headers
		if isSensitive(key) {
			redacted[key] = "[REDACTED]"
		} else {
			if len(v) > 0 {
				redacted[key] = v[0]
			}
		}
	}
	return redacted
}

func isSensitive(header string) bool {
	return sensitiveHeaderRegex.MatchString(header)
}
