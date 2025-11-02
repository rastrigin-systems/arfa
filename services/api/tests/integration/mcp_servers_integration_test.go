package integration

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/tests/testutil"
)

// ============================================================================
// MCP Catalog Integration Tests
// ============================================================================

func TestMCPCatalog_Integration_CRUD(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	t.Run("CreateMCPServer_Success", func(t *testing.T) {
		connSchemaJSON := json.RawMessage(`{"url": "http://example.com"}`)
		capabilitiesJSON := json.RawMessage(`["read", "write"]`)
		configTemplateJSON := json.RawMessage(`{"token": "${API_TOKEN}"}`)
		requiredEnvVarsJSON := json.RawMessage(`["API_TOKEN"]`)

		mcp, err := queries.CreateMCPServer(ctx, db.CreateMCPServerParams{
			Name:              "test-mcp-server",
			Provider:          "test-provider",
			Version:           "1.0.0",
			Description:       "Test MCP server for integration tests",
			ConnectionSchema:  connSchemaJSON,
			Capabilities:      capabilitiesJSON,
			RequiresCredentials: false,
			IsApproved:        testutil.ToNullBool(true),
			DockerImage:       testutil.ToNullString("test-provider/server:latest"),
			ConfigTemplate:    configTemplateJSON,
			RequiredEnvVars:   requiredEnvVarsJSON,
		})

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, mcp.ID)
		assert.Equal(t, "test-mcp-server", mcp.Name)
		assert.Equal(t, "test-provider", mcp.Provider)
		assert.Equal(t, "1.0.0", mcp.Version)
		assert.True(t, mcp.IsApproved.Bool)
	})

	t.Run("GetMCPServer_Success", func(t *testing.T) {
		// Use existing MCP from seed data
		mcps, err := queries.ListMCPServers(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, mcps, "Expected seed data to have MCP servers")

		mcp, err := queries.GetMCPServer(ctx, mcps[0].ID)
		require.NoError(t, err)
		assert.Equal(t, mcps[0].ID, mcp.ID)
		assert.Equal(t, mcps[0].Name, mcp.Name)
	})

	t.Run("GetMCPServerByName_Success", func(t *testing.T) {
		mcp, err := queries.GetMCPServerByName(ctx, "GitHub")
		require.NoError(t, err)
		assert.Equal(t, "GitHub", mcp.Name)
		assert.NotEqual(t, uuid.Nil, mcp.ID)
	})

	t.Run("ListMCPServers_OnlyApproved", func(t *testing.T) {
		mcps, err := queries.ListMCPServers(ctx)
		require.NoError(t, err)

		// All returned MCPs should be approved
		for _, mcp := range mcps {
			assert.True(t, mcp.IsApproved.Bool, "ListMCPServers should only return approved MCPs")
		}
	})

	t.Run("ListAllMCPServers_IncludesUnapproved", func(t *testing.T) {
		// Create unapproved MCP
		connSchemaJSON := json.RawMessage(`{}`)
		capabilitiesJSON := json.RawMessage(`[]`)

		unapproved, err := queries.CreateMCPServer(ctx, db.CreateMCPServerParams{
			Name:              "unapproved-mcp",
			Provider:          "test",
			Version:           "1.0.0",
			Description:       "Unapproved MCP",
			ConnectionSchema:  connSchemaJSON,
			Capabilities:      capabilitiesJSON,
			RequiresCredentials: false,
			IsApproved:        testutil.ToNullBool(false),
		})
		require.NoError(t, err)

		// ListAllMCPServers should include unapproved
		allMCPs, err := queries.ListAllMCPServers(ctx)
		require.NoError(t, err)

		found := false
		for _, mcp := range allMCPs {
			if mcp.ID == unapproved.ID {
				found = true
				assert.False(t, mcp.IsApproved.Bool)
			}
		}
		assert.True(t, found, "ListAllMCPServers should include unapproved MCPs")
	})

	t.Run("UpdateMCPServer_Success", func(t *testing.T) {
		// Create MCP
		connSchemaJSON := json.RawMessage(`{}`)
		capabilitiesJSON := json.RawMessage(`[]`)

		mcp, err := queries.CreateMCPServer(ctx, db.CreateMCPServerParams{
			Name:              "update-test-mcp",
			Provider:          "test-provider",
			Version:           "1.0.0",
			Description:       "Original description",
			ConnectionSchema:  connSchemaJSON,
			Capabilities:      capabilitiesJSON,
			RequiresCredentials: false,
			IsApproved:        testutil.ToNullBool(true),
		})
		require.NoError(t, err)

		// Update MCP
		updated, err := queries.UpdateMCPServer(ctx, db.UpdateMCPServerParams{
			ID:          mcp.ID,
			Description: testutil.ToNullString("Updated description"),
			Version:     testutil.ToNullString("2.0.0"),
			DockerImage: testutil.ToNullString("test/updated:latest"),
		})
		require.NoError(t, err)

		assert.Equal(t, mcp.ID, updated.ID)
		assert.Equal(t, "Updated description", updated.Description)
		assert.Equal(t, "2.0.0", updated.Version)
		assert.Equal(t, "test/updated:latest", updated.DockerImage.String)
	})

	t.Run("ApproveMCPServer_Success", func(t *testing.T) {
		// Create unapproved MCP
		connSchemaJSON := json.RawMessage(`{}`)
		capabilitiesJSON := json.RawMessage(`[]`)

		mcp, err := queries.CreateMCPServer(ctx, db.CreateMCPServerParams{
			Name:              "approve-test-mcp",
			Provider:          "test",
			Version:           "1.0.0",
			Description:       "To be approved",
			ConnectionSchema:  connSchemaJSON,
			Capabilities:      capabilitiesJSON,
			RequiresCredentials: false,
			IsApproved:        testutil.ToNullBool(false),
		})
		require.NoError(t, err)

		// Approve
		err = queries.ApproveMCPServer(ctx, mcp.ID)
		require.NoError(t, err)

		// Verify approval
		approved, err := queries.GetMCPServer(ctx, mcp.ID)
		require.NoError(t, err)
		assert.True(t, approved.IsApproved.Bool)
	})

	t.Run("DisapproveMCPServer_Success", func(t *testing.T) {
		// Create approved MCP
		connSchemaJSON := json.RawMessage(`{}`)
		capabilitiesJSON := json.RawMessage(`[]`)

		mcp, err := queries.CreateMCPServer(ctx, db.CreateMCPServerParams{
			Name:              "disapprove-test-mcp",
			Provider:          "test",
			Version:           "1.0.0",
			Description:       "To be disapproved",
			ConnectionSchema:  connSchemaJSON,
			Capabilities:      capabilitiesJSON,
			RequiresCredentials: false,
			IsApproved:        testutil.ToNullBool(true),
		})
		require.NoError(t, err)

		// Disapprove
		err = queries.DisapproveMCPServer(ctx, mcp.ID)
		require.NoError(t, err)

		// Verify disapproval
		disapproved, err := queries.GetMCPServer(ctx, mcp.ID)
		require.NoError(t, err)
		assert.False(t, disapproved.IsApproved.Bool)
	})
}

// ============================================================================
// Employee MCP Configuration Integration Tests
// ============================================================================

func TestEmployeeMCPConfigs_Integration_CRUD(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Setup test data
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "developer")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "dev@example.com",
		FullName: "Developer User",
		Status:   "active",
	})

	t.Run("CreateEmployeeMCPConfig_Success", func(t *testing.T) {
		// Get MCP from seed data
		mcp, err := queries.GetMCPServerByName(ctx, "GitHub")
		require.NoError(t, err)

		configJSON := json.RawMessage(`{"token": "ghp_test123"}`)
		config, err := queries.CreateEmployeeMCPConfig(ctx, db.CreateEmployeeMCPConfigParams{
			EmployeeID:       employee.ID,
			McpCatalogID:     mcp.ID,
			ConnectionConfig: configJSON,
			IsEnabled:        testutil.ToNullBool(true),
		})

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, config.ID)
		assert.Equal(t, employee.ID, config.EmployeeID)
		assert.Equal(t, mcp.ID, config.McpCatalogID)
		assert.True(t, config.IsEnabled.Bool)
	})

	t.Run("ListEmployeeMCPConfigs_Success", func(t *testing.T) {
		// Create test employee
		emp := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
			OrgID:    org.ID,
			RoleID:   role.ID,
			Email:    "emp-mcps@example.com",
			FullName: "MCP Test User",
			Status:   "active",
		})

		// Assign multiple MCPs
		mcps, err := queries.ListMCPServers(ctx)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(mcps), 2)

		for i := 0; i < 2; i++ {
			configJSON := json.RawMessage(`{}`)
			_, err := queries.CreateEmployeeMCPConfig(ctx, db.CreateEmployeeMCPConfigParams{
				EmployeeID:       emp.ID,
				McpCatalogID:     mcps[i].ID,
				ConnectionConfig: configJSON,
				IsEnabled:        testutil.ToNullBool(true),
			})
			require.NoError(t, err)
		}

		// List employee MCP configs
		empMCPs, err := queries.ListEmployeeMCPConfigs(ctx, emp.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, len(empMCPs))
	})

	t.Run("GetEmployeeMCPConfig_Success", func(t *testing.T) {
		// Create test employee
		emp := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
			OrgID:    org.ID,
			RoleID:   role.ID,
			Email:    "get-mcp@example.com",
			FullName: "Get MCP User",
			Status:   "active",
		})

		mcp, err := queries.GetMCPServerByName(ctx, "Filesystem")
		require.NoError(t, err)

		configJSON := json.RawMessage(`{"paths": ["/home/user"]}`)
		_, err = queries.CreateEmployeeMCPConfig(ctx, db.CreateEmployeeMCPConfigParams{
			EmployeeID:       emp.ID,
			McpCatalogID:     mcp.ID,
			ConnectionConfig: configJSON,
			IsEnabled:        testutil.ToNullBool(true),
		})
		require.NoError(t, err)

		// Get specific employee MCP config
		empMCP, err := queries.GetEmployeeMCPConfig(ctx, db.GetEmployeeMCPConfigParams{
			EmployeeID:   emp.ID,
			McpCatalogID: mcp.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, mcp.Name, empMCP.Name)
		assert.True(t, empMCP.IsEnabled.Bool)
	})

	t.Run("UpdateEmployeeMCPConfig_Success", func(t *testing.T) {
		// Create test employee
		emp := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
			OrgID:    org.ID,
			RoleID:   role.ID,
			Email:    "update-mcp@example.com",
			FullName: "Update MCP User",
			Status:   "active",
		})

		mcp, err := queries.GetMCPServerByName(ctx, "PostgreSQL")
		require.NoError(t, err)

		configJSON := json.RawMessage(`{"connection": "old"}`)
		_, err = queries.CreateEmployeeMCPConfig(ctx, db.CreateEmployeeMCPConfigParams{
			EmployeeID:       emp.ID,
			McpCatalogID:     mcp.ID,
			ConnectionConfig: configJSON,
			IsEnabled:        testutil.ToNullBool(true),
		})
		require.NoError(t, err)

		// Update employee MCP config
		newConfigJSON := json.RawMessage(`{"connection": "new"}`)
		updated, err := queries.UpdateEmployeeMCPConfig(ctx, db.UpdateEmployeeMCPConfigParams{
			EmployeeID:       emp.ID,
			McpCatalogID:     mcp.ID,
			IsEnabled:        testutil.ToNullBool(false),
			ConnectionConfig: newConfigJSON,
		})
		require.NoError(t, err)
		assert.False(t, updated.IsEnabled.Bool)
	})

	t.Run("DeleteEmployeeMCPConfig_Success", func(t *testing.T) {
		// Create test employee
		emp := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
			OrgID:    org.ID,
			RoleID:   role.ID,
			Email:    "delete-mcp@example.com",
			FullName: "Delete MCP User",
			Status:   "active",
		})

		mcp, err := queries.GetMCPServerByName(ctx, "Slack")
		require.NoError(t, err)

		configJSON := json.RawMessage(`{}`)
		_, err = queries.CreateEmployeeMCPConfig(ctx, db.CreateEmployeeMCPConfigParams{
			EmployeeID:       emp.ID,
			McpCatalogID:     mcp.ID,
			ConnectionConfig: configJSON,
			IsEnabled:        testutil.ToNullBool(true),
		})
		require.NoError(t, err)

		// Delete config
		err = queries.DeleteEmployeeMCPConfig(ctx, db.DeleteEmployeeMCPConfigParams{
			EmployeeID:   emp.ID,
			McpCatalogID: mcp.ID,
		})
		require.NoError(t, err)

		// Verify deletion
		empMCPs, err := queries.ListEmployeeMCPConfigs(ctx, emp.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(empMCPs))
	})

	t.Run("CountEmployeeMCPConfigs_Success", func(t *testing.T) {
		// Create test employee
		emp := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
			OrgID:    org.ID,
			RoleID:   role.ID,
			Email:    "count-mcps@example.com",
			FullName: "Count MCPs User",
			Status:   "active",
		})

		mcps, err := queries.ListMCPServers(ctx)
		require.NoError(t, err)

		// Assign 3 MCPs
		for i := 0; i < 3; i++ {
			configJSON := json.RawMessage(`{}`)
			_, err := queries.CreateEmployeeMCPConfig(ctx, db.CreateEmployeeMCPConfigParams{
				EmployeeID:       emp.ID,
				McpCatalogID:     mcps[i].ID,
				ConnectionConfig: configJSON,
				IsEnabled:        testutil.ToNullBool(true),
			})
			require.NoError(t, err)
		}

		// Count MCPs
		count, err := queries.CountEmployeeMCPConfigs(ctx, emp.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})

	t.Run("GetMCPUsageCount_Success", func(t *testing.T) {
		mcp, err := queries.GetMCPServerByName(ctx, "GitHub")
		require.NoError(t, err)

		// Assign MCP to multiple employees
		for i := 0; i < 3; i++ {
			emp := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
				OrgID:    org.ID,
				RoleID:   role.ID,
				Email:    testutil.RandomEmail(),
				FullName: "Usage Test User",
				Status:   "active",
			})

			configJSON := json.RawMessage(`{}`)
			_, err := queries.CreateEmployeeMCPConfig(ctx, db.CreateEmployeeMCPConfigParams{
				EmployeeID:       emp.ID,
				McpCatalogID:     mcp.ID,
				ConnectionConfig: configJSON,
				IsEnabled:        testutil.ToNullBool(true),
			})
			require.NoError(t, err)
		}

		// Count usage
		count, err := queries.GetMCPUsageCount(ctx, mcp.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(3))
	})
}

// ============================================================================
// Multi-tenancy Tests
// ============================================================================

func TestEmployeeMCPConfigs_Integration_MultiTenancy(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	// Create two organizations
	org1 := testutil.CreateTestOrg(t, queries, ctx)
	org2 := testutil.CreateTestOrg(t, queries, ctx)

	role := testutil.CreateTestRole(t, queries, ctx, "developer")

	// Create employees in each org
	emp1 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org1.ID,
		RoleID:   role.ID,
		Email:    "emp1@org1.com",
		FullName: "Employee 1",
		Status:   "active",
	})

	emp2 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org2.ID,
		RoleID:   role.ID,
		Email:    "emp2@org2.com",
		FullName: "Employee 2",
		Status:   "active",
	})

	mcp, err := queries.GetMCPServerByName(ctx, "GitHub")
	require.NoError(t, err)

	// Assign MCP to both employees
	configJSON := json.RawMessage(`{}`)
	_, err = queries.CreateEmployeeMCPConfig(ctx, db.CreateEmployeeMCPConfigParams{
		EmployeeID:       emp1.ID,
		McpCatalogID:     mcp.ID,
		ConnectionConfig: configJSON,
		IsEnabled:        testutil.ToNullBool(true),
	})
	require.NoError(t, err)

	_, err = queries.CreateEmployeeMCPConfig(ctx, db.CreateEmployeeMCPConfigParams{
		EmployeeID:       emp2.ID,
		McpCatalogID:     mcp.ID,
		ConnectionConfig: configJSON,
		IsEnabled:        testutil.ToNullBool(true),
	})
	require.NoError(t, err)

	// Each employee should only see their own MCPs
	emp1MCPs, err := queries.ListEmployeeMCPConfigs(ctx, emp1.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(emp1MCPs))

	emp2MCPs, err := queries.ListEmployeeMCPConfigs(ctx, emp2.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(emp2MCPs))
}
