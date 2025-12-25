package ui

import (
	"context"

	"github.com/rastrigin-systems/arfa/services/cli/internal/config"
)

// ConfigManagerInterface defines what the log streamer needs from config management
type ConfigManagerInterface interface {
	Load() (*config.Config, error)
	Save(cfg *config.Config) error
}

// LogStreamerInterface defines the contract for log streaming
type LogStreamerInterface interface {
	SetJSONOutput(enabled bool)
	SetVerbose(enabled bool)
	StreamLogs(ctx context.Context) error
}

// Compile-time interface checks
var _ LogStreamerInterface = (*LogStreamer)(nil)
