package cli

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
)

// LogStreamer handles WebSocket connection and log streaming
type LogStreamer struct {
	platformClient *PlatformClient
	configManager  *ConfigManager
}

// NewLogStreamer creates a new LogStreamer
func NewLogStreamer(pc *PlatformClient, cm *ConfigManager) *LogStreamer {
	return &LogStreamer{
		platformClient: pc,
		configManager:  cm,
	}
}

// StreamLogEntry represents a log entry received from the WebSocket
type StreamLogEntry struct {
	SessionID     string                 `json:"session_id,omitempty"`
	AgentID       string                 `json:"agent_id,omitempty"`
	EventType     string                 `json:"event_type"`
	EventCategory string                 `json:"event_category"`
	Content       string                 `json:"content,omitempty"`
	Payload       map[string]interface{} `json:"payload,omitempty"`
	Timestamp     time.Time              `json:"timestamp"`
}

// StreamLogs connects to the WebSocket endpoint and prints logs
func (ls *LogStreamer) StreamLogs(ctx context.Context) error {
	// Ensure authenticated
	config, err := ls.configManager.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	if config.Token == "" {
		return fmt.Errorf("not authenticated. Run 'ubik login' first")
	}

	// Construct WebSocket URL
	wsURL, err := url.Parse(config.PlatformURL)
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
	headers.Add("Authorization", "Bearer "+config.Token)
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
			fmt.Printf("[%s] [%s:%s] %s\n",
				logEntry.Timestamp.Format("15:04:05"),
				logEntry.EventType,
				logEntry.EventCategory,
				logEntry.Content,
			)
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
