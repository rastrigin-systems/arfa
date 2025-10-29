// Package types contains shared types used across the ubik-enterprise platform
package types

import (
	"time"

	"github.com/google/uuid"
)

// Common domain types that are shared between API and CLI

// OrganizationID represents a unique organization identifier
type OrganizationID uuid.UUID

// EmployeeID represents a unique employee identifier
type EmployeeID uuid.UUID

// AgentID represents a unique agent identifier
type AgentID uuid.UUID

// ConfigStatus represents the status of a configuration
type ConfigStatus string

const (
	ConfigStatusActive   ConfigStatus = "active"
	ConfigStatusInactive ConfigStatus = "inactive"
	ConfigStatusPending  ConfigStatus = "pending"
)

// Agent represents an AI agent in the catalog
type Agent struct {
	ID          uuid.UUID  `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Version     string     `json:"version"`
	IsActive    bool       `json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// AgentConfig represents an agent configuration instance
type AgentConfig struct {
	ID        uuid.UUID    `json:"id"`
	AgentID   uuid.UUID    `json:"agent_id"`
	OrgID     uuid.UUID    `json:"org_id"`
	Status    ConfigStatus `json:"status"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt *time.Time   `json:"updated_at,omitempty"`
}

// Organization represents a company/organization
type Organization struct {
	ID        uuid.UUID  `json:"id"`
	Name      string     `json:"name"`
	Slug      string     `json:"slug"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

// Employee represents a user/employee
type Employee struct {
	ID        uuid.UUID  `json:"id"`
	OrgID     uuid.UUID  `json:"org_id"`
	Email     string     `json:"email"`
	FirstName string     `json:"first_name"`
	LastName  string     `json:"last_name"`
	IsActive  bool       `json:"is_active"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}
