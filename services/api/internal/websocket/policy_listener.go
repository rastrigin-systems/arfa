package websocket

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// PolicyChangeChannel is the PostgreSQL NOTIFY channel name
	PolicyChangeChannel = "policy_change"

	// Retry settings for connection failures
	listenerRetryDelay = 5 * time.Second
)

// PolicyListener listens for PostgreSQL NOTIFY events and forwards to PolicyHub
type PolicyListener struct {
	pool *pgxpool.Pool
	hub  *PolicyHub
	stop chan struct{}
}

// NewPolicyListener creates a new policy listener
func NewPolicyListener(pool *pgxpool.Pool, hub *PolicyHub) *PolicyListener {
	return &PolicyListener{
		pool: pool,
		hub:  hub,
		stop: make(chan struct{}),
	}
}

// Start begins listening for policy change notifications
func (l *PolicyListener) Start(ctx context.Context) {
	go l.listenLoop(ctx)
}

// Stop signals the listener to stop
func (l *PolicyListener) Stop() {
	close(l.stop)
}

// listenLoop continuously listens for notifications with automatic reconnection
func (l *PolicyListener) listenLoop(ctx context.Context) {
	for {
		select {
		case <-l.stop:
			return
		case <-ctx.Done():
			return
		default:
			if err := l.listen(ctx); err != nil {
				log.Printf("Policy listener error: %v, retrying in %v", err, listenerRetryDelay)
				select {
				case <-time.After(listenerRetryDelay):
				case <-l.stop:
					return
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// listen establishes a connection and listens for notifications
func (l *PolicyListener) listen(ctx context.Context) error {
	// Acquire a dedicated connection for LISTEN
	conn, err := l.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	// Subscribe to the channel
	_, err = conn.Exec(ctx, "LISTEN "+PolicyChangeChannel)
	if err != nil {
		return err
	}

	log.Printf("Policy listener started on channel: %s", PolicyChangeChannel)

	// Listen for notifications
	for {
		select {
		case <-l.stop:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Wait for notification with timeout
			notification, err := conn.Conn().WaitForNotification(ctx)
			if err != nil {
				return err
			}

			// Process notification
			l.processNotification(notification.Payload)
		}
	}
}

// processNotification handles a single notification payload
func (l *PolicyListener) processNotification(payload string) {
	// Parse the notification payload
	var raw struct {
		Action     string          `json:"action"`
		Policy     json.RawMessage `json:"policy,omitempty"`
		PolicyID   *string         `json:"policy_id,omitempty"`
		OrgID      string          `json:"org_id"`
		TeamID     *string         `json:"team_id,omitempty"`
		EmployeeID *string         `json:"employee_id,omitempty"`
	}

	if err := json.Unmarshal([]byte(payload), &raw); err != nil {
		log.Printf("Failed to parse policy notification: %v", err)
		return
	}

	// Build the notification
	notification := PolicyChangeNotification{
		Action: raw.Action,
	}

	// Parse org_id
	orgID, err := uuid.Parse(raw.OrgID)
	if err != nil {
		log.Printf("Invalid org_id in notification: %v", err)
		return
	}
	notification.OrgID = orgID

	// Parse optional team_id
	if raw.TeamID != nil && *raw.TeamID != "" {
		teamID, err := uuid.Parse(*raw.TeamID)
		if err == nil {
			notification.TeamID = &teamID
		}
	}

	// Parse optional employee_id
	if raw.EmployeeID != nil && *raw.EmployeeID != "" {
		employeeID, err := uuid.Parse(*raw.EmployeeID)
		if err == nil {
			notification.EmployeeID = &employeeID
		}
	}

	// Parse optional policy_id (for delete)
	if raw.PolicyID != nil && *raw.PolicyID != "" {
		policyID, err := uuid.Parse(*raw.PolicyID)
		if err == nil {
			notification.PolicyID = &policyID
		}
	}

	// Parse optional policy object (for create/update)
	if len(raw.Policy) > 0 && string(raw.Policy) != "null" {
		var policyData struct {
			ID         string          `json:"id"`
			OrgID      string          `json:"org_id"`
			TeamID     *string         `json:"team_id"`
			EmployeeID *string         `json:"employee_id"`
			ToolName   string          `json:"tool_name"`
			Action     string          `json:"action"`
			Reason     string          `json:"reason"`
			Conditions json.RawMessage `json:"conditions"`
			CreatedAt  *time.Time      `json:"created_at"`
			UpdatedAt  *time.Time      `json:"updated_at"`
		}

		if err := json.Unmarshal(raw.Policy, &policyData); err == nil {
			pd := PolicyData{
				ToolName: policyData.ToolName,
				Action:   policyData.Action,
				Reason:   policyData.Reason,
			}

			// Parse UUIDs
			if id, err := uuid.Parse(policyData.ID); err == nil {
				pd.ID = id
			}
			if orgID, err := uuid.Parse(policyData.OrgID); err == nil {
				pd.OrgID = orgID
			}
			if policyData.TeamID != nil && *policyData.TeamID != "" {
				if tid, err := uuid.Parse(*policyData.TeamID); err == nil {
					pd.TeamID = &tid
				}
			}
			if policyData.EmployeeID != nil && *policyData.EmployeeID != "" {
				if eid, err := uuid.Parse(*policyData.EmployeeID); err == nil {
					pd.EmployeeID = &eid
				}
			}

			// Parse conditions
			if len(policyData.Conditions) > 0 {
				var conditions map[string]interface{}
				if err := json.Unmarshal(policyData.Conditions, &conditions); err == nil {
					pd.Conditions = conditions
				}
			}

			// Determine scope
			if pd.EmployeeID != nil {
				pd.Scope = "employee"
			} else if pd.TeamID != nil {
				pd.Scope = "team"
			} else {
				pd.Scope = "organization"
			}

			if policyData.CreatedAt != nil {
				pd.CreatedAt = *policyData.CreatedAt
			}
			pd.UpdatedAt = policyData.UpdatedAt

			notification.Policy = &pd
		}
	}

	log.Printf("Policy notification: action=%s org=%s team=%v employee=%v",
		notification.Action, notification.OrgID,
		notification.TeamID, notification.EmployeeID)

	// Forward to hub
	l.hub.NotifyPolicyChange(notification)
}
