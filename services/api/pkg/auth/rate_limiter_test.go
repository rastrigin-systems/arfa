package auth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/rastrigin-systems/arfa/generated/db"
	"github.com/rastrigin-systems/arfa/services/api/tests/testutil"
)

func TestCheckPasswordResetRateLimit(t *testing.T) {
	// Setup test database
	conn, queries := testutil.SetupTestDB(t)
	ctx := context.Background()

	// Create test employee
	org := testutil.CreateTestOrg(t, queries, ctx)
	role := testutil.CreateTestRole(t, queries, ctx, "Member")
	employee := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
		OrgID:    org.ID,
		RoleID:   role.ID,
		Email:    "ratelimit@example.com",
		FullName: "Test User",
	})

	// Helper to clean up password reset tokens
	cleanupTokens := func(employeeID uuid.UUID) {
		_, err := conn.Exec(ctx, "DELETE FROM password_reset_tokens WHERE employee_id = $1", employeeID)
		if err != nil {
			t.Fatalf("Failed to cleanup tokens: %v", err)
		}
	}

	t.Run("allows first request (no previous requests)", func(t *testing.T) {
		cleanupTokens(employee.ID)
		err := CheckPasswordResetRateLimit(ctx, employee.ID, queries)
		if err != nil {
			t.Fatalf("Expected no error for first request, got %v", err)
		}
	})

	t.Run("allows up to 3 requests within 1 hour", func(t *testing.T) {
		cleanupTokens(employee.ID)

		// Create 3 tokens (simulating 3 requests)
		for i := 0; i < 3; i++ {
			token, _ := GenerateSecureToken()
			_, err := queries.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
				EmployeeID: employee.ID,
				Token:      token,
				ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour), Valid: true},
			})
			if err != nil {
				t.Fatalf("Failed to create token %d: %v", i, err)
			}
		}

		// 4th request should fail
		err := CheckPasswordResetRateLimit(ctx, employee.ID, queries)
		if err == nil {
			t.Fatal("Expected rate limit error after 3 requests, got nil")
		}
		if err != ErrRateLimitExceeded {
			t.Fatalf("Expected ErrRateLimitExceeded, got %v", err)
		}
	})

	t.Run("allows request after cleanup (simulating time passing)", func(t *testing.T) {
		cleanupTokens(employee.ID)

		// Simulate: no tokens in last hour = rate limit check passes
		err := CheckPasswordResetRateLimit(ctx, employee.ID, queries)
		if err != nil {
			t.Fatalf("Expected no error after cleanup, got %v", err)
		}
	})

	t.Run("rate limit is per employee (different employees have separate limits)", func(t *testing.T) {
		cleanupTokens(employee.ID)

		// Create another employee
		employee2 := testutil.CreateTestEmployee(t, queries, ctx, testutil.TestEmployeeParams{
			OrgID:    org.ID,
			RoleID:   role.ID,
			Email:    "ratelimit2@example.com",
			FullName: "Test User 2",
		})
		defer cleanupTokens(employee2.ID)

		// Max out employee1's rate limit
		for i := 0; i < 3; i++ {
			token, _ := GenerateSecureToken()
			queries.CreatePasswordResetToken(ctx, db.CreatePasswordResetTokenParams{
				EmployeeID: employee.ID,
				Token:      token,
				ExpiresAt:  pgtype.Timestamp{Time: time.Now().Add(1 * time.Hour), Valid: true},
			})
		}

		// Employee1 should be rate limited
		err1 := CheckPasswordResetRateLimit(ctx, employee.ID, queries)
		if err1 != ErrRateLimitExceeded {
			t.Fatalf("Expected ErrRateLimitExceeded for employee1, got %v", err1)
		}

		// Employee2 should NOT be rate limited
		err2 := CheckPasswordResetRateLimit(ctx, employee2.ID, queries)
		if err2 != nil {
			t.Fatalf("Expected no error for employee2, got %v", err2)
		}
	})

	t.Run("allows request for non-existent employee (count = 0)", func(t *testing.T) {
		fakeEmployeeID := uuid.New()
		err := CheckPasswordResetRateLimit(ctx, fakeEmployeeID, queries)
		// Should allow (count = 0 for non-existent employee)
		if err != nil {
			t.Fatalf("Expected no error for non-existent employee, got %v", err)
		}
	})
}
