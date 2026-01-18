package control

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"time"
)

// ProxyInfo contains information about a running proxy instance.
type ProxyInfo struct {
	PID       int       `json:"pid"`
	Port      int       `json:"port"`
	StartedAt time.Time `json:"started_at"`
	CertPath  string    `json:"cert_path"`
}

// ProxyStatus represents the current status of the proxy.
type ProxyStatus struct {
	Running bool
	Info    *ProxyInfo
	Uptime  time.Duration
}

// PIDFile manages the proxy PID file for status detection.
type PIDFile struct {
	path string
}

// NewPIDFile creates a new PID file manager.
func NewPIDFile() (*PIDFile, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %w", err)
	}

	pidDir := filepath.Join(home, ".arfa")
	if err := os.MkdirAll(pidDir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	return &PIDFile{
		path: filepath.Join(pidDir, "proxy.pid"),
	}, nil
}

// Write writes the proxy info to the PID file.
func (p *PIDFile) Write(info ProxyInfo) error {
	data, err := json.MarshalIndent(info, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal proxy info: %w", err)
	}

	if err := os.WriteFile(p.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	return nil
}

// Read reads the proxy info from the PID file.
func (p *PIDFile) Read() (*ProxyInfo, error) {
	data, err := os.ReadFile(p.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read PID file: %w", err)
	}

	var info ProxyInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("failed to parse PID file: %w", err)
	}

	return &info, nil
}

// Remove removes the PID file.
func (p *PIDFile) Remove() error {
	if err := os.Remove(p.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove PID file: %w", err)
	}
	return nil
}

// GetStatus checks if the proxy is running and returns its status.
func (p *PIDFile) GetStatus() (*ProxyStatus, error) {
	info, err := p.Read()
	if err != nil {
		return nil, err
	}

	if info == nil {
		return &ProxyStatus{Running: false}, nil
	}

	// Check if process is still alive
	process, err := os.FindProcess(info.PID)
	if err != nil {
		// Process not found, clean up stale PID file
		_ = p.Remove()
		return &ProxyStatus{Running: false}, nil
	}

	// On Unix, FindProcess always succeeds, so we need to send signal 0
	// to check if process exists
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// Process doesn't exist, clean up stale PID file
		_ = p.Remove()
		return &ProxyStatus{Running: false}, nil
	}

	return &ProxyStatus{
		Running: true,
		Info:    info,
		Uptime:  time.Since(info.StartedAt),
	}, nil
}

// Path returns the path to the PID file.
func (p *PIDFile) Path() string {
	return p.path
}
