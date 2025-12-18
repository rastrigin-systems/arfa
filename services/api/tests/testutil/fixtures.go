package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/require"

	"github.com/rastrigin-systems/ubik-enterprise/generated/db"
)

// CreateTestOrg creates a test organization with default values
func CreateTestOrg(t *testing.T, queries *db.Queries, ctx context.Context) db.Organization {
	org, err := queries.CreateOrganization(ctx, db.CreateOrganizationParams{
		Name: "Test Corporation",
		Slug: "test-corp-" + uuid.NewString()[:8],
	})
	require.NoError(t, err)
	return org
}

// CreateTestRole creates a test role (roles are global, not org-specific)
func CreateTestRole(t *testing.T, queries *db.Queries, ctx context.Context, name string) db.Role {
	role, err := queries.CreateRole(ctx, db.CreateRoleParams{
		Name:        name + "-" + uuid.NewString()[:4], // Make unique
		Permissions: []byte(`["read","write"]`),        // JSONB format
	})
	require.NoError(t, err)
	return role
}

// CreateTestEmployee creates a test employee with hashed password
func CreateTestEmployee(t *testing.T, queries *db.Queries, ctx context.Context, params TestEmployeeParams) db.Employee {
	if params.Email == "" {
		params.Email = "test-" + uuid.NewString()[:8] + "@example.com"
	}
	if params.PasswordHash == "" {
		// Default password hash for "password123"
		params.PasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"
	}
	if params.FullName == "" {
		params.FullName = "Test User"
	}
	if params.Status == "" {
		params.Status = "active"
	}

	createParams := db.CreateEmployeeParams{
		OrgID:        params.OrgID,
		RoleID:       params.RoleID,
		Email:        params.Email,
		FullName:     params.FullName,
		PasswordHash: params.PasswordHash,
		Status:       params.Status,
		Preferences:  []byte("{}"),
	}

	if params.TeamID != uuid.Nil {
		createParams.TeamID = pgtype.UUID{Bytes: params.TeamID, Valid: true}
	} else {
		createParams.TeamID = pgtype.UUID{Valid: false}
	}

	emp, err := queries.CreateEmployee(ctx, createParams)
	require.NoError(t, err)
	return emp
}

// TestEmployeeParams holds parameters for creating a test employee
type TestEmployeeParams struct {
	OrgID        uuid.UUID
	Email        string
	PasswordHash string
	FullName     string
	RoleID       uuid.UUID
	TeamID       uuid.UUID
	Status       string
}

// CreateTestSession creates a test session for an employee
func CreateTestSession(t *testing.T, queries *db.Queries, ctx context.Context, employeeID uuid.UUID, tokenHash string) db.Session {
	session, err := queries.CreateSession(ctx, CreateSessionParams(employeeID, tokenHash))
	require.NoError(t, err)
	return session
}

// CreateSessionParams creates session parameters with defaults
func CreateSessionParams(employeeID uuid.UUID, tokenHash string) db.CreateSessionParams {
	return db.CreateSessionParams{
		EmployeeID: employeeID,
		TokenHash:  tokenHash,
		IpAddress:  nil,
		UserAgent:  nil,
		ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(24 * time.Hour), Valid: true},
	}
}

// CreateTestTeam creates a test team in the organization
// Optional description can be passed as last parameter
func CreateTestTeam(t *testing.T, queries *db.Queries, ctx context.Context, orgID uuid.UUID, name string, description ...string) db.Team {
	params := db.CreateTeamParams{
		OrgID: orgID,
		Name:  name,
	}
	if len(description) > 0 && description[0] != "" {
		params.Description = &description[0]
	}
	team, err := queries.CreateTeam(ctx, params)
	require.NoError(t, err)
	return team
}

// GenerateUniqueName generates a unique name by appending a timestamp
func GenerateUniqueName(baseName string, timestamp int64) string {
	return baseName + "-" + uuid.NewString()[:8]
}

// CreateTeamAgentConfigParams creates parameters for team agent config creation
func CreateTeamAgentConfigParams(teamID, agentID uuid.UUID, configJSON []byte, isEnabled bool) db.CreateTeamAgentConfigParams {
	return db.CreateTeamAgentConfigParams{
		TeamID:         teamID,
		AgentID:        agentID,
		ConfigOverride: configJSON,
		IsEnabled:      isEnabled,
	}
}

// CreateUpdateOrgParams creates parameters for updating an organization
func CreateUpdateOrgParams(id uuid.UUID, name string, maxEmployees int32, maxAgentsPerEmployee int32, settings []byte) db.UpdateOrganizationParams {
	return db.UpdateOrganizationParams{
		ID:                   id,
		Name:                 name,
		MaxEmployees:         maxEmployees,
		MaxAgentsPerEmployee: maxAgentsPerEmployee,
		Settings:             settings,
	}
}

// RandomEmail generates a random email address for testing
func RandomEmail() string {
	return "test-" + uuid.NewString()[:8] + "@example.com"
}
