// Package proxy provides a simplified in-process MITM proxy for logging LLM API requests.
//
// This is a simplified replacement for the complex httpproxy package. Key differences:
// - Runs in-process (no daemon)
// - Session ID passed directly (no session manager)
// - Simple port allocation (no Unix socket control API)
// - Just logging (no policy engine)
package proxy

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
	"net"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/elazarl/goproxy"
	"github.com/rastrigin-systems/ubik-enterprise/pkg/types"
	"github.com/rastrigin-systems/ubik-enterprise/services/cli/internal/logparser"
)

const (
	// MinPort is the starting port for proxy allocation
	MinPort = 8082
	// MaxPort is the ending port for proxy allocation (supports 10 concurrent instances)
	MaxPort = 8091
)

// sensitiveHeaderRegex matches headers that should be redacted in logs
var sensitiveHeaderRegex = regexp.MustCompile(`(?i)(auth|api-key|token|cookie|x-api-key)`)

// llmHostRegex matches LLM provider hosts to intercept
var llmHostRegex = regexp.MustCompile(`(api\.anthropic\.com|generativelanguage\.googleapis\.com|api\.openai\.com)`)

// Logger defines the interface for logging proxy events.
// This is a minimal interface - only what the proxy needs.
type Logger interface {
	// LogEvent logs a proxy event
	LogEvent(eventType, category, content string, metadata map[string]interface{})
	// LogClassified logs a parsed/classified log entry
	LogClassified(entry types.ClassifiedLogEntry)
}

// Proxy provides in-process HTTPS interception for LLM API logging.
type Proxy struct {
	goproxy   *goproxy.ProxyHttpServer
	server    *http.Server
	logger    Logger
	parser    *logparser.AnthropicParser
	port      int
	certPath  string
	keyPath   string
	sessionID string
	agentID   string
	mu        sync.RWMutex // Protects sessionID and agentID
}

// New creates a new proxy instance.
func New(logger Logger) *Proxy {
	return &Proxy{
		goproxy: goproxy.NewProxyHttpServer(),
		logger:  logger,
		parser:  logparser.NewAnthropicParser(),
	}
}

// SetSession sets the session and agent ID for log entries.
func (p *Proxy) SetSession(sessionID, agentID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.sessionID = sessionID
	p.agentID = agentID
}

// Start starts the proxy on an available port in the range [MinPort, MaxPort].
// Returns an error if no port is available.
func (p *Proxy) Start() error {
	// Setup CA certificates first
	if err := p.setupCA(); err != nil {
		return fmt.Errorf("failed to setup CA: %w", err)
	}

	// Configure interception rules
	p.configureRules()

	// Try to find an available port
	for port := MinPort; port <= MaxPort; port++ {
		if err := p.tryStart(port); err == nil {
			p.port = port
			return nil
		}
	}

	return fmt.Errorf("no available port in range %d-%d", MinPort, MaxPort)
}

// tryStart attempts to start the proxy on the specified port.
func (p *Proxy) tryStart(port int) error {
	addr := fmt.Sprintf("127.0.0.1:%d", port)

	// Check if port is available
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	listener.Close()

	// Start server
	p.server = &http.Server{
		Addr:    addr,
		Handler: p.goproxy,
	}

	errCh := make(chan error, 1)
	go func() {
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	// Give server a moment to start
	time.Sleep(50 * time.Millisecond)

	// Check for immediate errors
	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

// Stop stops the proxy server.
func (p *Proxy) Stop() error {
	if p.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.server.Shutdown(ctx)
}

// GetPort returns the port the proxy is running on.
func (p *Proxy) GetPort() int {
	return p.port
}

// GetProxyURL returns the proxy URL for use in HTTP_PROXY env var.
func (p *Proxy) GetProxyURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", p.port)
}

// GetCertPath returns the path to the CA certificate.
func (p *Proxy) GetCertPath() string {
	return p.certPath
}

// setupCA ensures the CA certificate exists or generates it.
func (p *Proxy) setupCA() error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}

	certDir := filepath.Join(home, ".ubik", "certs")
	if err := os.MkdirAll(certDir, 0700); err != nil {
		return err
	}

	p.certPath = filepath.Join(certDir, "ubik-ca.pem")
	p.keyPath = filepath.Join(certDir, "ubik-ca-key.pem")

	// Check if certs exist
	if _, err := os.Stat(p.certPath); err == nil {
		if _, err := os.Stat(p.keyPath); err == nil {
			return p.loadCA()
		}
	}

	// Generate new CA
	return p.generateCA()
}

// loadCA loads existing CA certificate.
func (p *Proxy) loadCA() error {
	caCert, err := tls.LoadX509KeyPair(p.certPath, p.keyPath)
	if err != nil {
		return fmt.Errorf("failed to load CA: %w", err)
	}

	p.configureGoproxyCA(&caCert)
	return nil
}

// generateCA generates a new CA certificate.
func (p *Proxy) generateCA() error {
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"Ubik Enterprise Proxy CA"},
			CommonName:   "ubik-proxy-ca",
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
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

	// Save certificate
	certOut, err := os.Create(p.certPath)
	if err != nil {
		return err
	}
	if err := pem.Encode(certOut, &pem.Block{Type: "CERTIFICATE", Bytes: caBytes}); err != nil {
		certOut.Close()
		return err
	}
	certOut.Close()

	// Save key
	keyOut, err := os.OpenFile(p.keyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	if err := pem.Encode(keyOut, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey)}); err != nil {
		keyOut.Close()
		return err
	}
	keyOut.Close()

	// Load the new CA
	caCert, err := tls.LoadX509KeyPair(p.certPath, p.keyPath)
	if err != nil {
		return err
	}

	p.configureGoproxyCA(&caCert)
	return nil
}

// configureGoproxyCA configures goproxy to use the CA certificate.
func (p *Proxy) configureGoproxyCA(caCert *tls.Certificate) {
	p.goproxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	goproxy.GoproxyCa = *caCert
	tlsConfig := goproxy.TLSConfigFromCA(caCert)
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: tlsConfig}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: tlsConfig}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: tlsConfig}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: tlsConfig}
}

// configureRules sets up interception rules for LLM providers.
func (p *Proxy) configureRules() {
	// Intercept LLM API requests
	p.goproxy.OnRequest(goproxy.ReqHostMatches(llmHostRegex)).DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			p.logRequest(r)
			return r, nil
		})

	// Intercept LLM API responses
	p.goproxy.OnResponse(goproxy.ReqHostMatches(llmHostRegex)).DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			p.logResponse(resp)
			return resp
		})
}

// logRequest logs an intercepted request.
func (p *Proxy) logRequest(r *http.Request) {
	if p.logger == nil {
		return
	}

	// Read and restore body
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		return
	}
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Get session info under read lock
	p.mu.RLock()
	sessionID := p.sessionID
	agentID := p.agentID
	p.mu.RUnlock()

	// Parse and log classified entries for Anthropic
	if strings.Contains(r.URL.Host, "anthropic.com") && len(bodyBytes) > 0 {
		entries, err := p.parser.ParseRequest(bodyBytes)
		if err == nil {
			for _, entry := range entries {
				entry.SessionID = sessionID
				entry.AgentID = agentID
				p.logger.LogClassified(entry)
			}
		}
	}

	// Log raw request
	payload := map[string]interface{}{
		"method":     r.Method,
		"url":        r.URL.String(),
		"headers":    redactHeaders(r.Header),
		"body":       string(bodyBytes),
		"session_id": sessionID,
		"agent_id":   agentID,
	}

	p.logger.LogEvent("api_request", "proxy", fmt.Sprintf("%s %s", r.Method, r.URL.Host), payload)
}

// logResponse logs an intercepted response.
func (p *Proxy) logResponse(resp *http.Response) {
	if p.logger == nil || resp == nil {
		return
	}

	// Read and restore body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// Decompress if gzipped
	decodedBody := bodyBytes
	if resp.Header.Get("Content-Encoding") == "gzip" {
		reader, err := gzip.NewReader(bytes.NewBuffer(bodyBytes))
		if err == nil {
			decodedBody, _ = io.ReadAll(reader)
			reader.Close()
		}
	}

	// Get session info under read lock
	p.mu.RLock()
	sessionID := p.sessionID
	agentID := p.agentID
	p.mu.RUnlock()

	// Parse and log classified entries for Anthropic
	if resp.Request != nil && strings.Contains(resp.Request.URL.Host, "anthropic.com") && len(decodedBody) > 0 {
		entries, err := p.parser.ParseResponse(decodedBody)
		if err == nil {
			for _, entry := range entries {
				entry.SessionID = sessionID
				entry.AgentID = agentID
				p.logger.LogClassified(entry)
			}
		}
	}

	// Log raw response
	url := ""
	if resp.Request != nil {
		url = resp.Request.URL.Host
	}

	payload := map[string]interface{}{
		"status":     resp.StatusCode,
		"headers":    redactHeaders(resp.Header),
		"body":       string(decodedBody),
		"session_id": sessionID,
		"agent_id":   agentID,
	}

	p.logger.LogEvent("api_response", "proxy", fmt.Sprintf("%d %s", resp.StatusCode, url), payload)
}

// redactHeaders redacts sensitive headers.
func redactHeaders(headers http.Header) map[string]string {
	redacted := make(map[string]string)
	for k, v := range headers {
		if sensitiveHeaderRegex.MatchString(k) {
			redacted[k] = "[REDACTED]"
		} else if len(v) > 0 {
			redacted[k] = v[0]
		}
	}
	return redacted
}
