package email

import (
	"strings"
	"testing"
)

// TestMockEmailService_SendPasswordResetEmail tests the mock email service
func TestMockEmailService_SendPasswordResetEmail(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		service := NewMockEmailService()

		err := service.SendPasswordResetEmail("alice@acme.com", "test-token-123")
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}

		// Verify the email was recorded
		if len(service.SentEmails) != 1 {
			t.Fatalf("Expected 1 sent email, got %d", len(service.SentEmails))
		}

		sentEmail := service.SentEmails[0]
		if sentEmail.To != "alice@acme.com" {
			t.Errorf("Expected email to 'alice@acme.com', got '%s'", sentEmail.To)
		}

		if sentEmail.Token != "test-token-123" {
			t.Errorf("Expected token 'test-token-123', got '%s'", sentEmail.Token)
		}
	})

	t.Run("MultipleEmails", func(t *testing.T) {
		service := NewMockEmailService()

		// Send multiple emails
		_ = service.SendPasswordResetEmail("user1@example.com", "token1")
		_ = service.SendPasswordResetEmail("user2@example.com", "token2")
		_ = service.SendPasswordResetEmail("user3@example.com", "token3")

		if len(service.SentEmails) != 3 {
			t.Fatalf("Expected 3 sent emails, got %d", len(service.SentEmails))
		}

		// Verify order is preserved
		if service.SentEmails[0].To != "user1@example.com" {
			t.Errorf("Expected first email to 'user1@example.com'")
		}
		if service.SentEmails[1].To != "user2@example.com" {
			t.Errorf("Expected second email to 'user2@example.com'")
		}
		if service.SentEmails[2].To != "user3@example.com" {
			t.Errorf("Expected third email to 'user3@example.com'")
		}
	})

	t.Run("InvalidEmail", func(t *testing.T) {
		service := NewMockEmailService()

		err := service.SendPasswordResetEmail("", "token")
		if err == nil {
			t.Fatal("Expected error for empty email, got nil")
		}

		if !strings.Contains(err.Error(), "email") {
			t.Errorf("Expected error message to contain 'email', got: %v", err)
		}
	})

	t.Run("InvalidToken", func(t *testing.T) {
		service := NewMockEmailService()

		err := service.SendPasswordResetEmail("test@example.com", "")
		if err == nil {
			t.Fatal("Expected error for empty token, got nil")
		}

		if !strings.Contains(err.Error(), "token") {
			t.Errorf("Expected error message to contain 'token', got: %v", err)
		}
	})
}

// TestMockEmailService_Reset tests the Reset method
func TestMockEmailService_Reset(t *testing.T) {
	service := NewMockEmailService()

	// Send some emails
	_ = service.SendPasswordResetEmail("user1@example.com", "token1")
	_ = service.SendPasswordResetEmail("user2@example.com", "token2")

	if len(service.SentEmails) != 2 {
		t.Fatalf("Expected 2 sent emails before reset, got %d", len(service.SentEmails))
	}

	// Reset the service
	service.Reset()

	// Verify sent emails are cleared
	if len(service.SentEmails) != 0 {
		t.Errorf("Expected 0 sent emails after reset, got %d", len(service.SentEmails))
	}
}
