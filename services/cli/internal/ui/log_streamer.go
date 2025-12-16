package ui

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/api"
	"github.com/sergeirastrigin/ubik-enterprise/services/cli/internal/config"
)

// LogStreamer handles WebSocket connection and log streaming
type LogStreamer struct {
	apiClient     *api.Client
	configManager *config.Manager
	jsonOutput    bool // Whether to output full JSON
	verbose       bool // Whether to show full payloads
}

// NewLogStreamer creates a new LogStreamer
func NewLogStreamer(ac *api.Client, cm *config.Manager) *LogStreamer {
	return &LogStreamer{
		apiClient:     ac,
		configManager: cm,
		jsonOutput:    false,
		verbose:       false,
	}
}

// SetJSONOutput enables JSON output mode
func (ls *LogStreamer) SetJSONOutput(enabled bool) {
	ls.jsonOutput = enabled
}

// SetVerbose enables verbose output with full payloads
func (ls *LogStreamer) SetVerbose(enabled bool) {
	ls.verbose = enabled
}

// StreamLogs connects to the WebSocket endpoint and prints logs
func (ls *LogStreamer) StreamLogs(ctx context.Context) error {
	// Ensure authenticated
	cfg, err := ls.configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if cfg.Token == "" {
		return fmt.Errorf("not authenticated. Run 'ubik login' first")
	}

	// Construct WebSocket URL
	wsURL, err := url.Parse(cfg.PlatformURL)
	if err != nil {
		return fmt.Errorf("invalid platform URL: %w", err)
	}
	if wsURL.Scheme == "https" {
		wsURL.Scheme = "wss"
	} else {
		wsURL.Scheme = "ws"
	}
	wsURL.Path = "/api/v1/logs/stream"

	fmt.Printf("Connecting to WebSocket: %s...\n", wsURL.String())

	// Set up WebSocket dialer with authentication header
	headers := http.Header{}
	headers.Add("Authorization", "Bearer "+cfg.Token)
	dialer := websocket.DefaultDialer

	conn, _, err := dialer.DialContext(ctx, wsURL.String(), headers)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	defer conn.Close()

	fmt.Println("✓ Connected to log stream. Press Ctrl+C to exit.")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("Read error: %v", err)
				return
			}
			var logEntry StreamLogEntry
			if err := json.Unmarshal(message, &logEntry); err != nil {
				log.Printf("Failed to unmarshal log entry: %v", err)
				continue
			}

			// Format and print log entry
			if ls.jsonOutput {
				// Full JSON output
				jsonBytes, err := json.MarshalIndent(logEntry, "", "  ")
				if err == nil {
					fmt.Println(string(jsonBytes))
				}
			} else if ls.verbose && logEntry.Payload != nil && len(logEntry.Payload) > 0 {
				// Verbose mode with payload
				fmt.Printf("[%s] [%s:%s] %s\n",
					logEntry.Timestamp.Format("15:04:05"),
					logEntry.EventType,
					logEntry.EventCategory,
					logEntry.Content,
				)
				// Pretty print payload
				payloadJSON, err := json.MarshalIndent(logEntry.Payload, "  ", "  ")
				if err == nil {
					fmt.Printf("  Payload:\n  %s\n", string(payloadJSON))
				}
			} else {
				// Compact mode (original)
				fmt.Printf("[%s] [%s:%s] %s\n",
					logEntry.Timestamp.Format("15:04:05"),
					logEntry.EventType,
					logEntry.EventCategory,
					logEntry.Content,
				)
			}
		}
	}()

	select {
	case <-done:
	case <-interrupt:
		log.Println("Interrupt signal received. Closing connection...")
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Printf("Write close error: %v", err)
		}
		select {
		case <-done:
		case <-time.After(time.Second):
		}
	}

	return nil
}
