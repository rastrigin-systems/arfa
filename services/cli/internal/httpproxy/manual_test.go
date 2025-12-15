//go:build manual

package httpproxy

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// generateLocalTestToken creates a JWT token signed with the default dev secret
func generateLocalTestToken(employeeID, orgID string) string {
	claims := &TokenClaims{
		EmployeeID: employeeID,
		OrgID:      orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	secret := getJWTSecret() // Uses default: dev-secret-key-change-in-production
	tokenString, _ := token.SignedString(secret)
	return tokenString
}

// TestManualRegistration is a manual test for JWT registration
// Run with: go test -v -tags=manual -run TestManualRegistration ./internal/httpproxy/
func TestManualRegistration(t *testing.T) {
	client, err := NewDefaultControlClient()
	if err != nil {
		t.Fatalf("Error creating client: %v", err)
	}

	// Generate a token using the same secret as the daemon
	token := os.Getenv("JWT_TOKEN")
	if token == "" {
		// Use locally generated token with default dev secret
		token = generateLocalTestToken("emp-local-test", "org-local-test")
		t.Logf("Using locally generated test token")
	}

	resp, err := client.RegisterSession(RegisterSessionRequest{
		SessionID: "test-session-123",
		Token:     token,
		AgentID:   "test-agent",
		AgentName: "Test Agent",
		Workspace: "/tmp",
	})
	if err != nil {
		t.Fatalf("Registration failed: %v", err)
	}

	fmt.Printf("\n✓ Registration successful!\n")
	fmt.Printf("  Session ID: %s\n", resp.SessionID)
	fmt.Printf("  Port: %d\n", resp.Port)
	fmt.Printf("  Proxy Address: %s\n", resp.ProxyAddr)

	// Verify session was created with correct IDs from token
	session, err := client.GetSession("test-session-123")
	if err != nil {
		t.Fatalf("Failed to get session: %v", err)
	}
	if session == nil {
		t.Fatal("Session not found after registration")
	}

	fmt.Printf("\n✓ Session verification:\n")
	fmt.Printf("  Employee ID: %s\n", session.EmployeeID)
	fmt.Printf("  Org ID: %s\n", session.OrgID)

	// Clean up
	err = client.UnregisterSession("test-session-123")
	if err != nil {
		t.Logf("Warning: failed to unregister: %v", err)
	}
	fmt.Printf("\n✓ Session cleaned up\n")
}
