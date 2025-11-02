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

		dockerImage := "test-provider/server:latest"
		mcp, err := queries.CreateMCPServer(ctx, db.CreateMCPServerParams{
			Name:              "test-mcp-server",
			Provider:          "test-provider",
			Version:           "1.0.0",
			Description:       "Test MCP server for integration tests",
			ConnectionSchema:  connSchemaJSON,
			Capabilities:      capabilitiesJSON,
			RequiresCredentials: false,
			IsApproved:        true,
			DockerImage:       &dockerImage,
			ConfigTemplate:    configTemplateJSON,
			RequiredEnvVars:   requiredEnvVarsJSON,
		})

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, mcp.ID)
		assert.Equal(t, "test-mcp-server", mcp.Name)
		assert.Equal(t, "test-provider", mcp.Provider)
		assert.Equal(t, "1.0.0", mcp.Version)
		assert.True(t, mcp.IsApproved)
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
			assert.True(t, mcp.IsApproved, "ListMCPServers should only return approved MCPs")
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
			IsApproved:        false,
		})
		require.NoError(t, err)

		// ListAllMCPServers should include unapproved
		allMCPs, err := queries.ListAllMCPServers(ctx)
		require.NoError(t, err)

		found := false
		for _, mcp := range allMCPs {
			if mcp.ID == unapproved.ID {
				found = true
				assert.False(t, mcp.IsApproved)
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
			IsApproved:        true,
		})
		require.NoError(t, err)

		// Update MCP
		updatedDesc := "Updated description"
		updatedVer := "2.0.0"
		updatedImg := "test/updated:latest"
		updated, err := queries.UpdateMCPServer(ctx, db.UpdateMCPServerParams{
			ID:          mcp.ID,
			Description: &updatedDesc,
			Version:     &updatedVer,
			DockerImage: &updatedImg,
		})
		require.NoError(t, err)

		assert.Equal(t, mcp.ID, updated.ID)
		assert.Equal(t, "Updated description", updated.Description)
		assert.Equal(t, "2.0.0", updated.Version)
		assert.Equal(t, "test/updated:latest", *updated.DockerImage)
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
			IsApproved:        false,
		})
		require.NoError(t, err)

		// Approve
		err = queries.ApproveMCPServer(ctx, mcp.ID)
		require.NoError(t, err)

		// Verify approval
		approved, err := queries.GetMCPServer(ctx, mcp.ID)
		require.NoError(t, err)
		assert.True(t, approved.IsApproved)
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
			IsApproved:        true,
		})
		require.NoError(t, err)

		// Disapprove
		err = queries.DisapproveMCPServer(ctx, mcp.ID)
		require.NoError(t, err)

		// Verify disapproval
		disapproved, err := queries.GetMCPServer(ctx, mcp.ID)
		require.NoError(t, err)
		assert.False(t, disapproved.IsApproved)
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
			IsEnabled:        &[]bool{true}[0],
		})

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, config.ID)
		assert.Equal(t, employee.ID, config.EmployeeID)
		assert.Equal(t, mcp.ID, config.McpCatalogID)
		assert.True(t, *config.IsEnabled)
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
				IsEnabled:        &[]bool{true}[0],
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
			IsEnabled:        &[]bool{true}[0],
		})
		require.NoError(t, err)

		// Get specific employee MCP config
		empMCP, err := queries.GetEmployeeMCPConfig(ctx, db.GetEmployeeMCPConfigParams{
			EmployeeID:   emp.ID,
			McpCatalogID: mcp.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, mcp.Name, empMCP.Name)
		assert.True(t, *empMCP.IsEnabled)
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
			IsEnabled:        &[]bool{true}[0],
		})
		require.NoError(t, err)

		// Update employee MCP config
		newConfigJSON := json.RawMessage(`{"connection": "new"}`)
		updated, err := queries.UpdateEmployeeMCPConfig(ctx, db.UpdateEmployeeMCPConfigParams{
			EmployeeID:       emp.ID,
			McpCatalogID:     mcp.ID,
			IsEnabled:        &[]bool{false}[0],
			ConnectionConfig: newConfigJSON,
		})
		require.NoError(t, err)
		assert.False(t, *updated.IsEnabled)
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
			IsEnabled:        &[]bool{true}[0],
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
				IsEnabled:        &[]bool{true}[0],
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
				IsEnabled:        &[]bool{true}[0],
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
		IsEnabled:        &[]bool{true}[0],
	})
	require.NoError(t, err)

	_, err = queries.CreateEmployeeMCPConfig(ctx, db.CreateEmployeeMCPConfigParams{
		EmployeeID:       emp2.ID,
		McpCatalogID:     mcp.ID,
		ConnectionConfig: configJSON,
		IsEnabled:        &[]bool{true}[0],
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

// ============================================================================
// HTTP API Integration Tests
// ============================================================================

func TestMCPAPI_ListMCPServers(t *testing.T) {
	ts := testutil.NewTestServer(t)
	defer ts.Close()

	// Create test org, role, and employee
	org := testutil.CreateTestOrg(t, ts.Queries, ts.Ctx)
	role := testutil.CreateTestRole(t, ts.Queries, ts.Ctx, "developer")
	employee := testutil.CreateTestEmployee(t, ts.Queries, ts.Ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "api-user@example.com",
		FullName: "API Test User",
		Status:   "active",
	})

	// Login to get token
	token := testutil.LoginEmployee(t, ts, employee.Email, "testpassword123")

	t.Run("Success", func(t *testing.T) {
		resp := testutil.DoAuthenticatedRequest(t, ts, "GET", "/mcp-servers", nil, token)
		assert.Equal(t, 200, resp.StatusCode)

		var response struct {
			Servers []map[string]interface{} `json:"servers"`
			Total   int                      `json:"total"`
		}
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		// Should return only approved MCPs from seed data
		assert.Greater(t, response.Total, 0)
		assert.Len(t, response.Servers, response.Total)

		// Verify all returned MCPs are approved
		for _, server := range response.Servers {
			isApproved, ok := server["is_approved"].(bool)
			assert.True(t, ok, "is_approved should be a boolean")
			assert.True(t, isApproved, "All MCPs should be approved")
		}
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		resp := testutil.DoAuthenticatedRequest(t, ts, "GET", "/mcp-servers", nil, "")
		assert.Equal(t, 401, resp.StatusCode)
	})

	t.Run("Unauthorized_InvalidToken", func(t *testing.T) {
		resp := testutil.DoAuthenticatedRequest(t, ts, "GET", "/mcp-servers", nil, "invalid-token")
		assert.Equal(t, 401, resp.StatusCode)
	})
}

func TestMCPAPI_GetMCPServer(t *testing.T) {
	ts := testutil.NewTestServer(t)
	defer ts.Close()

	// Create test org, role, and employee
	org := testutil.CreateTestOrg(t, ts.Queries, ts.Ctx)
	role := testutil.CreateTestRole(t, ts.Queries, ts.Ctx, "developer")
	employee := testutil.CreateTestEmployee(t, ts.Queries, ts.Ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "api-user2@example.com",
		FullName: "API Test User 2",
		Status:   "active",
	})

	// Login to get token
	token := testutil.LoginEmployee(t, ts, employee.Email, "testpassword123")

	// Get an MCP from seed data
	mcp, err := ts.Queries.GetMCPServerByName(ts.Ctx, "GitHub")
	require.NoError(t, err)

	t.Run("Success", func(t *testing.T) {
		url := "/mcp-servers/" + mcp.ID.String()
		resp := testutil.DoAuthenticatedRequest(t, ts, "GET", url, nil, token)
		assert.Equal(t, 200, resp.StatusCode)

		var response map[string]interface{}
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, mcp.ID.String(), response["id"])
		assert.Equal(t, mcp.Name, response["name"])
		assert.Equal(t, mcp.Provider, response["provider"])
	})

	t.Run("NotFound", func(t *testing.T) {
		invalidID := uuid.New().String()
		url := "/mcp-servers/" + invalidID
		resp := testutil.DoAuthenticatedRequest(t, ts, "GET", url, nil, token)
		assert.Equal(t, 404, resp.StatusCode)
	})

	t.Run("BadRequest_InvalidUUID", func(t *testing.T) {
		resp := testutil.DoAuthenticatedRequest(t, ts, "GET", "/mcp-servers/not-a-uuid", nil, token)
		// Chi router will return 404 for invalid UUID patterns
		assert.True(t, resp.StatusCode == 400 || resp.StatusCode == 404)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		url := "/mcp-servers/" + mcp.ID.String()
		resp := testutil.DoAuthenticatedRequest(t, ts, "GET", url, nil, "")
		assert.Equal(t, 401, resp.StatusCode)
	})
}

func TestMCPAPI_ListEmployeeMCPServers(t *testing.T) {
	ts := testutil.NewTestServer(t)
	defer ts.Close()

	// Create test org, role, and employee
	org := testutil.CreateTestOrg(t, ts.Queries, ts.Ctx)
	role := testutil.CreateTestRole(t, ts.Queries, ts.Ctx, "developer")
	employee := testutil.CreateTestEmployee(t, ts.Queries, ts.Ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "emp-api@example.com",
		FullName: "Employee API User",
		Status:   "active",
	})

	// Login to get token
	token := testutil.LoginEmployee(t, ts, employee.Email, "testpassword123")

	// Assign some MCPs to the employee
	mcps, err := ts.Queries.ListMCPServers(ts.Ctx)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(mcps), 2)

	for i := 0; i < 2; i++ {
		configJSON := json.RawMessage(`{"key": "value"}`)
		_, err := ts.Queries.CreateEmployeeMCPConfig(ts.Ctx, db.CreateEmployeeMCPConfigParams{
			EmployeeID:       employee.ID,
			McpCatalogID:     mcps[i].ID,
			ConnectionConfig: configJSON,
			IsEnabled:        &[]bool{true}[0],
		})
		require.NoError(t, err)
	}

	t.Run("Success", func(t *testing.T) {
		resp := testutil.DoAuthenticatedRequest(t, ts, "GET", "/employees/me/mcp-servers", nil, token)
		assert.Equal(t, 200, resp.StatusCode)

		var response struct {
			Servers []map[string]interface{} `json:"servers"`
			Total   int                      `json:"total"`
		}
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		// Should return 2 MCPs we assigned
		assert.Equal(t, 2, response.Total)
		assert.Len(t, response.Servers, 2)

		// Verify fields are present
		for _, server := range response.Servers {
			assert.NotNil(t, server["id"])
			assert.NotNil(t, server["name"])
			assert.NotNil(t, server["provider"])
			assert.NotNil(t, server["is_enabled"])
			assert.NotNil(t, server["configured_at"])
		}
	})

	t.Run("Success_NoMCPs", func(t *testing.T) {
		// Create employee with no MCPs
		empNoMCPs := testutil.CreateTestEmployee(t, ts.Queries, ts.Ctx, testutil.TestEmployeeParams{
			OrgID:    org.ID,
			RoleID:   role.ID,
			Email:    "no-mcps@example.com",
			FullName: "No MCPs User",
			Status:   "active",
		})

		tokenNoMCPs := testutil.LoginEmployee(t, ts, empNoMCPs.Email, "testpassword123")

		resp := testutil.DoAuthenticatedRequest(t, ts, "GET", "/employees/me/mcp-servers", nil, tokenNoMCPs)
		assert.Equal(t, 200, resp.StatusCode)

		var response struct {
			Servers []map[string]interface{} `json:"servers"`
			Total   int                      `json:"total"`
		}
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, 0, response.Total)
		assert.Empty(t, response.Servers)
	})

	t.Run("Unauthorized_NoToken", func(t *testing.T) {
		resp := testutil.DoAuthenticatedRequest(t, ts, "GET", "/employees/me/mcp-servers", nil, "")
		assert.Equal(t, 401, resp.StatusCode)
	})
}

// ============================================================================
// Multi-tenancy API Tests
// ============================================================================

func TestMCPAPI_MultiTenancy(t *testing.T) {
	ts := testutil.NewTestServer(t)
	defer ts.Close()

	// Create two organizations
	org1 := testutil.CreateTestOrg(t, ts.Queries, ts.Ctx)
	org2 := testutil.CreateTestOrg(t, ts.Queries, ts.Ctx)

	role := testutil.CreateTestRole(t, ts.Queries, ts.Ctx, "developer")

	// Create employees in each org
	emp1 := testutil.CreateTestEmployee(t, ts.Queries, ts.Ctx, testutil.TestEmployeeParams{
		OrgID:    org1.ID,
		RoleID:   role.ID,
		Email:    "emp1-mt@org1.com",
		FullName: "Employee 1 MT",
		Status:   "active",
	})

	emp2 := testutil.CreateTestEmployee(t, ts.Queries, ts.Ctx, testutil.TestEmployeeParams{
		OrgID:    org2.ID,
		RoleID:   role.ID,
		Email:    "emp2-mt@org2.com",
		FullName: "Employee 2 MT",
		Status:   "active",
	})

	// Login both employees
	token1 := testutil.LoginEmployee(t, ts, emp1.Email, "testpassword123")
	token2 := testutil.LoginEmployee(t, ts, emp2.Email, "testpassword123")

	// Get MCP from seed data
	mcp, err := ts.Queries.GetMCPServerByName(ts.Ctx, "GitHub")
	require.NoError(t, err)

	// Assign MCP to emp1 only
	configJSON := json.RawMessage(`{"token": "emp1-token"}`)
	_, err = ts.Queries.CreateEmployeeMCPConfig(ts.Ctx, db.CreateEmployeeMCPConfigParams{
		EmployeeID:       emp1.ID,
		McpCatalogID:     mcp.ID,
		ConnectionConfig: configJSON,
		IsEnabled:        &[]bool{true}[0],
	})
	require.NoError(t, err)

	t.Run("Employee1_SeesMCPs", func(t *testing.T) {
		resp := testutil.DoAuthenticatedRequest(t, ts, "GET", "/employees/me/mcp-servers", nil, token1)
		assert.Equal(t, 200, resp.StatusCode)

		var response struct {
			Servers []map[string]interface{} `json:"servers"`
			Total   int                      `json:"total"`
		}
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, 1, response.Total)
	})

	t.Run("Employee2_DoesNotSeeMCPs", func(t *testing.T) {
		resp := testutil.DoAuthenticatedRequest(t, ts, "GET", "/employees/me/mcp-servers", nil, token2)
		assert.Equal(t, 200, resp.StatusCode)

		var response struct {
			Servers []map[string]interface{} `json:"servers"`
			Total   int                      `json:"total"`
		}
		err := json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		// emp2 should not see emp1's MCPs
		assert.Equal(t, 0, response.Total)
	})
}
