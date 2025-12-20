package control

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
	"time"

	"github.com/elazarl/goproxy"
)

const (
	// MinPort is the starting port for proxy allocation
	MinPort = 8082
	// MaxPort is the ending port for proxy allocation (supports 10 concurrent instances)
	MaxPort = 8091
)

// llmHostRegex matches LLM provider hosts to intercept
var llmHostRegex = regexp.MustCompile(`(api\.anthropic\.com|generativelanguage\.googleapis\.com|api\.openai\.com)`)

// ControlledProxy provides in-process HTTPS interception integrated with the Control Service.
// All intercepted requests/responses flow through the Control Service pipeline.
type ControlledProxy struct {
	goproxy  *goproxy.ProxyHttpServer
	server   *http.Server
	service  *Service
	port     int
	certPath string
	keyPath  string
}

// NewControlledProxy creates a new proxy integrated with a Control Service.
func NewControlledProxy(service *Service) *ControlledProxy {
	return &ControlledProxy{
		goproxy: goproxy.NewProxyHttpServer(),
		service: service,
	}
}

// Start starts the proxy on an available port in the range [MinPort, MaxPort].
func (p *ControlledProxy) Start() error {
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
func (p *ControlledProxy) tryStart(port int) error {
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
func (p *ControlledProxy) Stop() error {
	if p.server == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return p.server.Shutdown(ctx)
}

// GetPort returns the port the proxy is running on.
func (p *ControlledProxy) GetPort() int {
	return p.port
}

// GetProxyURL returns the proxy URL for use in HTTP_PROXY env var.
func (p *ControlledProxy) GetProxyURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", p.port)
}

// GetCertPath returns the path to the CA certificate.
func (p *ControlledProxy) GetCertPath() string {
	return p.certPath
}

// setupCA ensures the CA certificate exists or generates it.
func (p *ControlledProxy) setupCA() error {
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
func (p *ControlledProxy) loadCA() error {
	caCert, err := tls.LoadX509KeyPair(p.certPath, p.keyPath)
	if err != nil {
		return fmt.Errorf("failed to load CA: %w", err)
	}

	p.configureGoproxyCA(&caCert)
	return nil
}

// generateCA generates a new CA certificate.
func (p *ControlledProxy) generateCA() error {
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
func (p *ControlledProxy) configureGoproxyCA(caCert *tls.Certificate) {
	p.goproxy.OnRequest().HandleConnect(goproxy.AlwaysMitm)
	goproxy.GoproxyCa = *caCert
	tlsConfig := goproxy.TLSConfigFromCA(caCert)
	goproxy.OkConnect = &goproxy.ConnectAction{Action: goproxy.ConnectAccept, TLSConfig: tlsConfig}
	goproxy.MitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectMitm, TLSConfig: tlsConfig}
	goproxy.HTTPMitmConnect = &goproxy.ConnectAction{Action: goproxy.ConnectHTTPMitm, TLSConfig: tlsConfig}
	goproxy.RejectConnect = &goproxy.ConnectAction{Action: goproxy.ConnectReject, TLSConfig: tlsConfig}
}

// configureRules sets up interception rules for LLM providers.
func (p *ControlledProxy) configureRules() {
	// Intercept LLM API requests
	p.goproxy.OnRequest(goproxy.ReqHostMatches(llmHostRegex)).DoFunc(
		func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
			return p.handleRequest(r)
		})

	// Intercept LLM API responses
	p.goproxy.OnResponse(goproxy.ReqHostMatches(llmHostRegex)).DoFunc(
		func(resp *http.Response, ctx *goproxy.ProxyCtx) *http.Response {
			return p.handleResponse(resp)
		})
}

// handleRequest processes an intercepted request through the Control Service pipeline.
func (p *ControlledProxy) handleRequest(r *http.Request) (*http.Request, *http.Response) {
	if p.service == nil {
		return r, nil
	}

	result := p.service.HandleRequest(r)

	// If blocked, return an error response
	if result.Action == ActionBlock {
		return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusForbidden, result.Reason)
	}

	// Use modified request if provided
	if result.ModifiedRequest != nil {
		return result.ModifiedRequest, nil
	}

	return r, nil
}

// handleResponse processes an intercepted response through the Control Service pipeline.
func (p *ControlledProxy) handleResponse(resp *http.Response) *http.Response {
	if p.service == nil || resp == nil {
		return resp
	}

	// Decompress gzip responses before passing to pipeline
	if resp.Header.Get("Content-Encoding") == "gzip" && resp.Body != nil {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err == nil {
			reader, err := gzip.NewReader(bytes.NewBuffer(bodyBytes))
			if err == nil {
				decodedBody, err := io.ReadAll(reader)
				reader.Close()
				if err == nil {
					// Replace body with decompressed content
					resp.Body = io.NopCloser(bytes.NewBuffer(decodedBody))
					resp.Header.Del("Content-Encoding")
					resp.ContentLength = int64(len(decodedBody))
				} else {
					// Restore original body if decompression fails
					resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
				}
			} else {
				// Restore original body if gzip reader fails
				resp.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			}
		}
	}

	result := p.service.HandleResponse(resp)

	// If blocked, we can't really block responses (already in flight)
	// Just log and continue
	if result.Action == ActionBlock {
		// Response blocking is a future enhancement
		// For now, just log the blocked response
	}

	// Use modified response if provided
	if result.ModifiedResponse != nil {
		return result.ModifiedResponse
	}

	return resp
}
