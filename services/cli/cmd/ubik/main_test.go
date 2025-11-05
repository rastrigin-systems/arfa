package main

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/logging"
)

// TestLoggerInitialization tests that the logger is initialized correctly
func TestLoggerInitialization(t *testing.T) {
	tests := []struct {
		name        string
		envVar      string
		expectNil   bool
		description string
	}{
		{
			name:        "logger enabled by default",
			envVar:      "",
			expectNil:   false,
			description: "Logger should be created when no env var set",
		},
		{
			name:        "logger disabled via env var",
			envVar:      "1",
			expectNil:   true,
			description: "Logger should be nil when UBIK_NO_LOGGING=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable
			if tt.envVar != "" {
				os.Setenv("UBIK_NO_LOGGING", tt.envVar)
				defer os.Unsetenv("UBIK_NO_LOGGING")
			}

			// Create logger config
			loggerConfig := &logging.Config{
				Enabled:       true,
				BatchSize:     100,
				BatchInterval: 5 * time.Second,
				MaxRetries:    5,
				RetryBackoff:  1 * time.Second,
			}

			// Create mock API client
			apiClient := &mockAPIClient{}

			// Create logger
			logger, err := logging.NewLogger(loggerConfig, apiClient)

			// Validate
			if tt.expectNil {
				if logger != nil {
					t.Errorf("Expected logger to be nil, but got: %v", logger)
				}
				if err != nil {
					t.Errorf("Expected no error, but got: %v", err)
				}
			} else {
				if logger == nil {
					t.Errorf("Expected logger to be created, but got nil")
				}
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}

			// Clean up
			if logger != nil {
				logger.Close()
			}
		})
	}
}

// TestLoggerConfigDefaults tests that default values are set correctly
func TestLoggerConfigDefaults(t *testing.T) {
	config := &logging.Config{
		Enabled: true,
	}

	apiClient := &mockAPIClient{}
	logger, err := logging.NewLogger(config, apiClient)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}
	defer logger.Close()

	if logger == nil {
		t.Fatal("Expected logger to be created")
	}

	// The logger should apply defaults internally
	// We can't directly test the internal config, but we can verify
	// the logger was created successfully
}

// TestLoggerOptOut tests that UBIK_NO_LOGGING opt-out works
func TestLoggerOptOut(t *testing.T) {
	os.Setenv("UBIK_NO_LOGGING", "1")
	defer os.Unsetenv("UBIK_NO_LOGGING")

	config := &logging.Config{
		Enabled:       true,
		BatchSize:     100,
		BatchInterval: 5 * time.Second,
	}

	apiClient := &mockAPIClient{}
	logger, err := logging.NewLogger(config, apiClient)

	if err != nil {
		t.Errorf("Expected no error with opt-out, but got: %v", err)
	}
	if logger != nil {
		t.Errorf("Expected logger to be nil when opted out, but got: %v", logger)
	}
}

// TestLoggerConfigDisabled tests that disabled config prevents logger creation
func TestLoggerConfigDisabled(t *testing.T) {
	config := &logging.Config{
		Enabled: false,
	}

	apiClient := &mockAPIClient{}
	logger, err := logging.NewLogger(config, apiClient)

	if err != nil {
		t.Errorf("Expected no error with disabled config, but got: %v", err)
	}
	if logger != nil {
		t.Errorf("Expected logger to be nil when disabled, but got: %v", logger)
	}
}

// mockAPIClient is a simple mock for testing
type mockAPIClient struct{}

func (m *mockAPIClient) CreateLog(ctx context.Context, entry logging.LogEntry) error {
	return nil
}

func (m *mockAPIClient) CreateLogBatch(ctx context.Context, entries []logging.LogEntry) error {
	return nil
}
