package service

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"
)

// getAppBaseURL returns the base URL for the web application.
// Can be overridden via ARFA_APP_URL environment variable.
func getAppBaseURL() string {
	if url := os.Getenv("ARFA_APP_URL"); url != "" {
		return strings.TrimSuffix(url, "/")
	}
	return "http://localhost:3000"
}

// EmailService defines the interface for sending emails
// This interface can be implemented by different providers (e.g., SendGrid, AWS SES, etc.)
type EmailService interface {
	// SendInvitation sends an invitation email to a new user
	SendInvitation(email InvitationEmail) error
}

// InvitationEmail contains all data needed to send an invitation email
type InvitationEmail struct {
	// Recipient is the email address of the person being invited
	Recipient string

	// InviterName is the name of the person who sent the invitation
	InviterName string

	// OrgName is the organization name
	OrgName string

	// Token is the unique invitation token
	Token string

	// ExpiresAt is when the invitation expires
	ExpiresAt time.Time
}

// Validate checks if the invitation email has all required fields
func (e *InvitationEmail) Validate() error {
	if e.Recipient == "" {
		return fmt.Errorf("recipient email is required")
	}

	if e.Token == "" {
		return fmt.Errorf("invitation token is required")
	}

	if time.Now().After(e.ExpiresAt) {
		return fmt.Errorf("invitation has expired")
	}

	return nil
}

// GenerateInvitationURL generates the invitation acceptance URL
func (e *InvitationEmail) GenerateInvitationURL(baseURL string) string {
	// Remove trailing slash if present
	baseURL = strings.TrimSuffix(baseURL, "/")
	return fmt.Sprintf("%s/accept-invitation?token=%s", baseURL, e.Token)
}

// MockEmailService is a mock implementation that logs emails instead of sending them
// This is useful for development and testing
type MockEmailService struct {
	mu         sync.RWMutex
	sentEmails []InvitationEmail
}

// NewMockEmailService creates a new mock email service
func NewMockEmailService() *MockEmailService {
	return &MockEmailService{
		sentEmails: make([]InvitationEmail, 0),
	}
}

// SendInvitation logs the invitation email to console
func (s *MockEmailService) SendInvitation(email InvitationEmail) error {
	// Validate email
	if err := email.Validate(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Add to history
	s.sentEmails = append(s.sentEmails, email)

	// Generate invitation URL
	invitationURL := email.GenerateInvitationURL(getAppBaseURL())

	// Log email details to console in a clear, readable format
	log.Printf("[EMAIL] Invitation Email Sent\n"+
		"  To: %s\n"+
		"  From: %s (%s)\n"+
		"  Subject: You're invited to join %s!\n"+
		"  Invitation Link: %s\n"+
		"  Token: %s\n"+
		"  Expires At: %s\n"+
		"  Sent At: %s",
		email.Recipient,
		email.InviterName,
		email.OrgName,
		email.OrgName,
		invitationURL,
		email.Token,
		email.ExpiresAt.Format(time.RFC3339),
		time.Now().Format(time.RFC3339),
	)

	return nil
}

// LastSentEmail returns the most recently sent email (for testing)
func (s *MockEmailService) LastSentEmail() *InvitationEmail {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if len(s.sentEmails) == 0 {
		return nil
	}

	return &s.sentEmails[len(s.sentEmails)-1]
}

// SentCount returns the number of emails sent (for testing)
func (s *MockEmailService) SentCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.sentEmails)
}

// Reset clears the sent email history (for testing)
func (s *MockEmailService) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.sentEmails = make([]InvitationEmail, 0)
}
