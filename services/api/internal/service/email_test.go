package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMockEmailService_SendInvitation(t *testing.T) {
	tests := []struct {
		name        string
		recipient   string
		inviterName string
		orgName     string
		token       string
		expiresAt   time.Time
		wantErr     bool
	}{
		{
			name:        "sends invitation email successfully",
			recipient:   "newuser@example.com",
			inviterName: "John Doe",
			orgName:     "Acme Corp",
			token:       "abc123def456",
			expiresAt:   time.Now().Add(7 * 24 * time.Hour),
			wantErr:     false,
		},
		{
			name:        "handles empty recipient",
			recipient:   "",
			inviterName: "John Doe",
			orgName:     "Acme Corp",
			token:       "abc123def456",
			expiresAt:   time.Now().Add(7 * 24 * time.Hour),
			wantErr:     true,
		},
		{
			name:        "handles empty token",
			recipient:   "newuser@example.com",
			inviterName: "John Doe",
			orgName:     "Acme Corp",
			token:       "",
			expiresAt:   time.Now().Add(7 * 24 * time.Hour),
			wantErr:     true,
		},
		{
			name:        "handles expired invitation",
			recipient:   "newuser@example.com",
			inviterName: "John Doe",
			orgName:     "Acme Corp",
			token:       "abc123def456",
			expiresAt:   time.Now().Add(-24 * time.Hour), // Already expired
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock email service
			service := NewMockEmailService()

			// Send invitation email
			err := service.SendInvitation(InvitationEmail{
				Recipient:   tt.recipient,
				InviterName: tt.inviterName,
				OrgName:     tt.orgName,
				Token:       tt.token,
				ExpiresAt:   tt.expiresAt,
			})

			// Verify error expectation
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestInvitationEmail_Validate(t *testing.T) {
	tests := []struct {
		name    string
		email   InvitationEmail
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid invitation email",
			email: InvitationEmail{
				Recipient:   "user@example.com",
				InviterName: "John Doe",
				OrgName:     "Acme Corp",
				Token:       "abc123def456",
				ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
			},
			wantErr: false,
		},
		{
			name: "empty recipient",
			email: InvitationEmail{
				InviterName: "John Doe",
				OrgName:     "Acme Corp",
				Token:       "abc123def456",
				ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
			},
			wantErr: true,
			errMsg:  "recipient email is required",
		},
		{
			name: "empty token",
			email: InvitationEmail{
				Recipient:   "user@example.com",
				InviterName: "John Doe",
				OrgName:     "Acme Corp",
				ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
			},
			wantErr: true,
			errMsg:  "invitation token is required",
		},
		{
			name: "expired invitation",
			email: InvitationEmail{
				Recipient:   "user@example.com",
				InviterName: "John Doe",
				OrgName:     "Acme Corp",
				Token:       "abc123def456",
				ExpiresAt:   time.Now().Add(-24 * time.Hour),
			},
			wantErr: true,
			errMsg:  "invitation has expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.email.Validate()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMockEmailService_LogFormat(t *testing.T) {
	t.Run("logs contain all required information", func(t *testing.T) {
		service := NewMockEmailService()

		expiresAt := time.Now().Add(7 * 24 * time.Hour)
		email := InvitationEmail{
			Recipient:   "newuser@example.com",
			InviterName: "John Doe",
			OrgName:     "Acme Corp",
			Token:       "abc123def456",
			ExpiresAt:   expiresAt,
		}

		err := service.SendInvitation(email)
		require.NoError(t, err)

		// Verify the last sent email matches
		lastEmail := service.LastSentEmail()
		require.NotNil(t, lastEmail)
		assert.Equal(t, email.Recipient, lastEmail.Recipient)
		assert.Equal(t, email.InviterName, lastEmail.InviterName)
		assert.Equal(t, email.OrgName, lastEmail.OrgName)
		assert.Equal(t, email.Token, lastEmail.Token)
		assert.Equal(t, email.ExpiresAt, lastEmail.ExpiresAt)
	})
}

func TestMockEmailService_MultipleEmails(t *testing.T) {
	t.Run("tracks multiple sent emails", func(t *testing.T) {
		service := NewMockEmailService()

		// Send first email
		email1 := InvitationEmail{
			Recipient:   "user1@example.com",
			InviterName: "John Doe",
			OrgName:     "Acme Corp",
			Token:       "token1",
			ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		}
		err := service.SendInvitation(email1)
		require.NoError(t, err)

		// Send second email
		email2 := InvitationEmail{
			Recipient:   "user2@example.com",
			InviterName: "Jane Smith",
			OrgName:     "Beta Inc",
			Token:       "token2",
			ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		}
		err = service.SendInvitation(email2)
		require.NoError(t, err)

		// Verify count
		assert.Equal(t, 2, service.SentCount())

		// Verify last email is the second one
		lastEmail := service.LastSentEmail()
		require.NotNil(t, lastEmail)
		assert.Equal(t, "user2@example.com", lastEmail.Recipient)
		assert.Equal(t, "token2", lastEmail.Token)
	})
}

func TestMockEmailService_Reset(t *testing.T) {
	t.Run("resets sent email history", func(t *testing.T) {
		service := NewMockEmailService()

		// Send an email
		email := InvitationEmail{
			Recipient:   "user@example.com",
			InviterName: "John Doe",
			OrgName:     "Acme Corp",
			Token:       "abc123",
			ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
		}
		err := service.SendInvitation(email)
		require.NoError(t, err)

		// Verify email was sent
		assert.Equal(t, 1, service.SentCount())
		assert.NotNil(t, service.LastSentEmail())

		// Reset
		service.Reset()

		// Verify history is cleared
		assert.Equal(t, 0, service.SentCount())
		assert.Nil(t, service.LastSentEmail())
	})
}

func TestInvitationEmail_GenerateInvitationURL(t *testing.T) {
	tests := []struct {
		name     string
		baseURL  string
		token    string
		expected string
	}{
		{
			name:     "generates URL with trailing slash",
			baseURL:  "https://app.arfa.com/",
			token:    "abc123",
			expected: "https://app.arfa.com/accept-invitation?token=abc123",
		},
		{
			name:     "generates URL without trailing slash",
			baseURL:  "https://app.arfa.com",
			token:    "abc123",
			expected: "https://app.arfa.com/accept-invitation?token=abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email := InvitationEmail{
				Recipient:   "user@example.com",
				InviterName: "John Doe",
				OrgName:     "Acme Corp",
				Token:       tt.token,
				ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
			}

			url := email.GenerateInvitationURL(tt.baseURL)
			assert.Equal(t, tt.expected, url)
		})
	}
}
