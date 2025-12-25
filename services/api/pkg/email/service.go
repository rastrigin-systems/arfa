package email

import (
	"errors"
	"fmt"
	"log"
	"sync"
)

// EmailService defines the interface for sending emails
type EmailService interface {
	SendPasswordResetEmail(email, token string) error
}

// SentEmail represents an email that was sent (for testing)
type SentEmail struct {
	To    string
	Token string
}

// MockEmailService is a mock implementation of EmailService for testing
// It logs emails to console and records them for verification in tests
type MockEmailService struct {
	SentEmails []SentEmail
	mu         sync.Mutex
}

// NewMockEmailService creates a new mock email service
func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		SentEmails: make([]SentEmail, 0),
	}
}

// SendPasswordResetEmail sends a password reset email (mock implementation)
// In production, this would send an actual email via SendGrid, AWS SES, etc.
func (s *MockEmailService) SendPasswordResetEmail(email, token string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate inputs
	if email == "" {
		return errors.New("email address cannot be empty")
	}
	if token == "" {
		return errors.New("reset token cannot be empty")
	}

	// Log to console (useful for development)
	resetURL := fmt.Sprintf("https://app.arfa.cloud/reset-password/%s", token)
	log.Printf("📧 Mock Email Service: Password Reset Email\n")
	log.Printf("   To: %s\n", email)
	log.Printf("   Reset URL: %s\n", resetURL)
	log.Printf("   Token: %s\n", token)

	// Record sent email (for test verification)
	s.SentEmails = append(s.SentEmails, SentEmail{
		To:    email,
		Token: token,
	})

	return nil
}

// Reset clears the sent emails list (useful for tests)
func (s *MockEmailService) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.SentEmails = make([]SentEmail, 0)
}
