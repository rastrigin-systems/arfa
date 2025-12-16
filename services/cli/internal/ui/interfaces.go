package ui

import (
	"context"

	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/api"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/config"
)

// ConfigManagerInterface defines what the picker needs from config management
type ConfigManagerInterface interface {
	Load() (*config.Config, error)
	Save(cfg *config.Config) error
}

// AgentPickerInterface defines the contract for agent selection
type AgentPickerInterface interface {
	SelectAgent(agents []api.AgentConfig, saveAsDefault bool, forceInteractive bool) (*api.AgentConfig, error)
	ConfirmSaveDefault() bool
	GetDefaultAgent() string
	ClearDefault() error
}

// LogStreamerInterface defines the contract for log streaming
type LogStreamerInterface interface {
	SetJSONOutput(enabled bool)
	SetVerbose(enabled bool)
	StreamLogs(ctx context.Context) error
}

// Compile-time interface checks
var _ AgentPickerInterface = (*AgentPicker)(nil)
var _ LogStreamerInterface = (*LogStreamer)(nil)
