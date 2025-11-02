package integration

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sergeirastrigin/ubik-enterprise/generated/db"
	"github.com/sergeirastrigin/ubik-enterprise/services/api/tests/testutil"
)

// ============================================================================
// Skill Catalog Integration Tests
// ============================================================================

func TestSkillCatalog_Integration_CRUD(t *testing.T) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)

	t.Run("CreateSkill_Success", func(t *testing.T) {
		filesJSON := json.RawMessage(`[{"path": "SKILL.md", "content": "# Test Skill"}]`)
		depsJSON := json.RawMessage(`{"mcp_servers": ["github"], "skills": []}`)

		skill, err := queries.CreateSkill(ctx, db.CreateSkillParams{
			Name:         "test-skill",
			Description:  testutil.ToNullString("Test skill for integration tests"),
			Category:     testutil.ToNullString("testing"),
			Version:      "1.0.0",
			Files:        filesJSON,
			Dependencies: depsJSON,
			IsActive:     testutil.ToNullBool(true),
		})

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, skill.ID)
		assert.Equal(t, "test-skill", skill.Name)
		assert.Equal(t, "1.0.0", skill.Version)
		assert.True(t, skill.IsActive.Bool)
	})

	t.Run("GetSkill_Success", func(t *testing.T) {
		// Use existing skill from seed data
		skills, err := queries.ListSkills(ctx)
		require.NoError(t, err)
		require.NotEmpty(t, skills, "Expected seed data to have skills")

		skill, err := queries.GetSkill(ctx, skills[0].ID)
		require.NoError(t, err)
		assert.Equal(t, skills[0].ID, skill.ID)
		assert.Equal(t, skills[0].Name, skill.Name)
	})

	t.Run("GetSkillByName_Success", func(t *testing.T) {
		skill, err := queries.GetSkillByName(ctx, "release-manager")
		require.NoError(t, err)
		assert.Equal(t, "release-manager", skill.Name)
		assert.NotEqual(t, uuid.Nil, skill.ID)
	})

	t.Run("ListSkills_Success", func(t *testing.T) {
		skills, err := queries.ListSkills(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(skills), 3, "Expected at least 3 skills from seed data")

		// All returned skills should be active
		for _, skill := range skills {
			assert.True(t, skill.IsActive.Bool, "ListSkills should only return active skills")
		}
	})

	t.Run("ListAllSkills_IncludesInactive", func(t *testing.T) {
		// Create an inactive skill
		filesJSON := json.RawMessage(`[{"path": "SKILL.md", "content": "# Inactive Skill"}]`)
		depsJSON := json.RawMessage(`{}`)

		inactive, err := queries.CreateSkill(ctx, db.CreateSkillParams{
			Name:         "inactive-skill",
			Description:  testutil.ToNullString("Inactive skill"),
			Category:     testutil.ToNullString("testing"),
			Version:      "1.0.0",
			Files:        filesJSON,
			Dependencies: depsJSON,
			IsActive:     testutil.ToNullBool(false),
		})
		require.NoError(t, err)

		// ListAllSkills should include inactive skills
		allSkills, err := queries.ListAllSkills(ctx)
		require.NoError(t, err)

		found := false
		for _, skill := range allSkills {
			if skill.ID == inactive.ID {
				found = true
				assert.False(t, skill.IsActive.Bool)
			}
		}
		assert.True(t, found, "ListAllSkills should include inactive skills")
	})

	t.Run("UpdateSkill_Success", func(t *testing.T) {
		// Create skill
		filesJSON := json.RawMessage(`[{"path": "SKILL.md", "content": "# Original"}]`)
		depsJSON := json.RawMessage(`{}`)

		skill, err := queries.CreateSkill(ctx, db.CreateSkillParams{
			Name:         "update-test-skill",
			Description:  testutil.ToNullString("Original description"),
			Category:     testutil.ToNullString("testing"),
			Version:      "1.0.0",
			Files:        filesJSON,
			Dependencies: depsJSON,
			IsActive:     testutil.ToNullBool(true),
		})
		require.NoError(t, err)

		// Update skill
		updatedFilesJSON := json.RawMessage(`[{"path": "SKILL.md", "content": "# Updated"}]`)
		updated, err := queries.UpdateSkill(ctx, db.UpdateSkillParams{
			ID:          skill.ID,
			Description: testutil.ToNullString("Updated description"),
			Version:     testutil.ToNullString("2.0.0"),
			Files:       updatedFilesJSON,
		})
		require.NoError(t, err)

		assert.Equal(t, skill.ID, updated.ID)
		assert.Equal(t, "Updated description", updated.Description.String)
		assert.Equal(t, "2.0.0", updated.Version)
	})

	t.Run("DeactivateSkill_Success", func(t *testing.T) {
		// Create active skill
		filesJSON := json.RawMessage(`[{"path": "SKILL.md", "content": "# Active"}]`)
		depsJSON := json.RawMessage(`{}`)

		skill, err := queries.CreateSkill(ctx, db.CreateSkillParams{
			Name:         "deactivate-test-skill",
			Description:  testutil.ToNullString("To be deactivated"),
			Category:     testutil.ToNullString("testing"),
			Version:      "1.0.0",
			Files:        filesJSON,
			Dependencies: depsJSON,
			IsActive:     testutil.ToNullBool(true),
		})
		require.NoError(t, err)

		// Deactivate
		err = queries.DeactivateSkill(ctx, skill.ID)
		require.NoError(t, err)

		// Verify deactivation
		deactivated, err := queries.GetSkill(ctx, skill.ID)
		require.NoError(t, err)
		assert.False(t, deactivated.IsActive.Bool)
	})
}

// ============================================================================
// Employee Skills Integration Tests
// ============================================================================

func TestEmployeeSkills_Integration_CRUD(t *testing.T) {
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

	t.Run("AssignSkillToEmployee_Success", func(t *testing.T) {
		// Get skill from seed data
		skill, err := queries.GetSkillByName(ctx, "github-task-manager")
		require.NoError(t, err)

		configJSON := json.RawMessage(`{"auto_assign": true}`)
		assignment, err := queries.AssignSkillToEmployee(ctx, db.AssignSkillToEmployeeParams{
			EmployeeID: employee.ID,
			SkillID:    skill.ID,
			IsEnabled:  testutil.ToNullBool(true),
			Config:     configJSON,
		})

		require.NoError(t, err)
		assert.NotEqual(t, uuid.Nil, assignment.ID)
		assert.Equal(t, employee.ID, assignment.EmployeeID)
		assert.Equal(t, skill.ID, assignment.SkillID)
		assert.True(t, assignment.IsEnabled.Bool)
	})

	t.Run("ListEmployeeSkills_Success", func(t *testing.T) {
		// Create test employee
		emp := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
			OrgID:    org.ID,
			RoleID:   role.ID,
			Email:    "emp-skills@example.com",
			FullName: "Skills Test User",
			Status:   "active",
		})

		// Assign multiple skills
		skills, err := queries.ListSkills(ctx)
		require.NoError(t, err)
		require.GreaterOrEqual(t, len(skills), 2)

		for i := 0; i < 2; i++ {
			configJSON := json.RawMessage(`{}`)
			_, err := queries.AssignSkillToEmployee(ctx, db.AssignSkillToEmployeeParams{
				EmployeeID: emp.ID,
				SkillID:    skills[i].ID,
				IsEnabled:  testutil.ToNullBool(true),
				Config:     configJSON,
			})
			require.NoError(t, err)
		}

		// List employee skills
		empSkills, err := queries.ListEmployeeSkills(ctx, emp.ID)
		require.NoError(t, err)
		assert.Equal(t, 2, len(empSkills))
	})

	t.Run("GetEmployeeSkill_Success", func(t *testing.T) {
		// Create test employee
		emp := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
			OrgID:    org.ID,
			RoleID:   role.ID,
			Email:    "get-skill@example.com",
			FullName: "Get Skill User",
			Status:   "active",
		})

		skill, err := queries.GetSkillByName(ctx, "release-manager")
		require.NoError(t, err)

		configJSON := json.RawMessage(`{"auto_release": true}`)
		_, err = queries.AssignSkillToEmployee(ctx, db.AssignSkillToEmployeeParams{
			EmployeeID: emp.ID,
			SkillID:    skill.ID,
			IsEnabled:  testutil.ToNullBool(true),
			Config:     configJSON,
		})
		require.NoError(t, err)

		// Get specific employee skill
		empSkill, err := queries.GetEmployeeSkill(ctx, db.GetEmployeeSkillParams{
			EmployeeID: emp.ID,
			SkillID:    skill.ID,
		})
		require.NoError(t, err)
		assert.Equal(t, skill.Name, empSkill.Name)
		assert.True(t, empSkill.IsEnabled.Bool)
	})

	t.Run("UpdateEmployeeSkill_Success", func(t *testing.T) {
		// Create test employee
		emp := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
			OrgID:    org.ID,
			RoleID:   role.ID,
			Email:    "update-skill@example.com",
			FullName: "Update Skill User",
			Status:   "active",
		})

		skill, err := queries.GetSkillByName(ctx, "code-reviewer")
		require.NoError(t, err)

		configJSON := json.RawMessage(`{"strict_mode": false}`)
		_, err = queries.AssignSkillToEmployee(ctx, db.AssignSkillToEmployeeParams{
			EmployeeID: emp.ID,
			SkillID:    skill.ID,
			IsEnabled:  testutil.ToNullBool(true),
			Config:     configJSON,
		})
		require.NoError(t, err)

		// Update employee skill
		newConfigJSON := json.RawMessage(`{"strict_mode": true}`)
		updated, err := queries.UpdateEmployeeSkill(ctx, db.UpdateEmployeeSkillParams{
			EmployeeID: emp.ID,
			SkillID:    skill.ID,
			IsEnabled:  testutil.ToNullBool(false),
			Config:     newConfigJSON,
		})
		require.NoError(t, err)
		assert.False(t, updated.IsEnabled.Bool)
	})

	t.Run("RemoveSkillFromEmployee_Success", func(t *testing.T) {
		// Create test employee
		emp := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
			OrgID:    org.ID,
			RoleID:   role.ID,
			Email:    "remove-skill@example.com",
			FullName: "Remove Skill User",
			Status:   "active",
		})

		skill, err := queries.GetSkillByName(ctx, "github-task-manager")
		require.NoError(t, err)

		configJSON := json.RawMessage(`{}`)
		_, err = queries.AssignSkillToEmployee(ctx, db.AssignSkillToEmployeeParams{
			EmployeeID: emp.ID,
			SkillID:    skill.ID,
			IsEnabled:  testutil.ToNullBool(true),
			Config:     configJSON,
		})
		require.NoError(t, err)

		// Remove skill
		err = queries.RemoveSkillFromEmployee(ctx, db.RemoveSkillFromEmployeeParams{
			EmployeeID: emp.ID,
			SkillID:    skill.ID,
		})
		require.NoError(t, err)

		// Verify removal
		empSkills, err := queries.ListEmployeeSkills(ctx, emp.ID)
		require.NoError(t, err)
		assert.Equal(t, 0, len(empSkills))
	})

	t.Run("CountEmployeeSkills_Success", func(t *testing.T) {
		// Create test employee
		emp := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
			OrgID:    org.ID,
			RoleID:   role.ID,
			Email:    "count-skills@example.com",
			FullName: "Count Skills User",
			Status:   "active",
		})

		skills, err := queries.ListSkills(ctx)
		require.NoError(t, err)

		// Assign 3 skills
		for i := 0; i < 3; i++ {
			configJSON := json.RawMessage(`{}`)
			_, err := queries.AssignSkillToEmployee(ctx, db.AssignSkillToEmployeeParams{
				EmployeeID: emp.ID,
				SkillID:    skills[i].ID,
				IsEnabled:  testutil.ToNullBool(true),
				Config:     configJSON,
			})
			require.NoError(t, err)
		}

		// Count skills
		count, err := queries.CountEmployeeSkills(ctx, emp.ID)
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)
	})

	t.Run("GetSkillUsageCount_Success", func(t *testing.T) {
		skill, err := queries.GetSkillByName(ctx, "release-manager")
		require.NoError(t, err)

		// Assign skill to multiple employees
		for i := 0; i < 3; i++ {
			emp := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
				OrgID:    org.ID,
				RoleID:   role.ID,
				Email:    testutil.RandomEmail(),
				FullName: "Usage Test User",
				Status:   "active",
			})

			configJSON := json.RawMessage(`{}`)
			_, err := queries.AssignSkillToEmployee(ctx, db.AssignSkillToEmployeeParams{
				EmployeeID: emp.ID,
				SkillID:    skill.ID,
				IsEnabled:  testutil.ToNullBool(true),
				Config:     configJSON,
			})
			require.NoError(t, err)
		}

		// Count usage
		count, err := queries.GetSkillUsageCount(ctx, skill.ID)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(3))
	})
}

// ============================================================================
// Multi-tenancy Tests
// ============================================================================

func TestEmployeeSkills_Integration_MultiTenancy(t *testing.T) {
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

	skill, err := queries.GetSkillByName(ctx, "github-task-manager")
	require.NoError(t, err)

	// Assign skill to both employees
	configJSON := json.RawMessage(`{}`)
	_, err = queries.AssignSkillToEmployee(ctx, db.AssignSkillToEmployeeParams{
		EmployeeID: emp1.ID,
		SkillID:    skill.ID,
		IsEnabled:  testutil.ToNullBool(true),
		Config:     configJSON,
	})
	require.NoError(t, err)

	_, err = queries.AssignSkillToEmployee(ctx, db.AssignSkillToEmployeeParams{
		EmployeeID: emp2.ID,
		SkillID:    skill.ID,
		IsEnabled:  testutil.ToNullBool(true),
		Config:     configJSON,
	})
	require.NoError(t, err)

	// Each employee should only see their own skills
	emp1Skills, err := queries.ListEmployeeSkills(ctx, emp1.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(emp1Skills))

	emp2Skills, err := queries.ListEmployeeSkills(ctx, emp2.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, len(emp2Skills))
}

// ============================================================================
// Helper function to execute test with proper context
// ============================================================================

func executeWithContext(t *testing.T, fn func(context.Context, *db.Queries)) {
	conn, queries := testutil.SetupTestDB(t)
	defer conn.Close(testutil.GetContext(t))
	ctx := testutil.GetContext(t)
	fn(ctx, queries)
}
