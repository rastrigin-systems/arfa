# Email Service Documentation

## Overview

The email service provides a clean interface for sending invitation emails. It includes a mock implementation for development/testing and is designed to be easily swappable with real email providers (e.g., SendGrid, AWS SES).

## Architecture

### Interface

```go
type EmailService interface {
    SendInvitation(email InvitationEmail) error
}
```

### Mock Implementation

The `MockEmailService` logs emails to console instead of sending them. This is useful for:
- Local development
- Testing
- Environments without email provider configuration

### Log Format

```
[EMAIL] Invitation Email Sent
  To: user@example.com
  From: John Doe (Acme Corp)
  Subject: You're invited to join Acme Corp!
  Invitation Link: https://app.arfa.com/accept-invitation?token=abc123...
  Token: abc123...
  Expires At: 2025-11-14T16:09:17Z
  Sent At: 2025-11-07T16:09:17Z
```

## Usage

### In Handlers

The `InvitationHandler` accepts an `EmailService` via dependency injection:

```go
// Create mock email service
emailService := service.NewMockEmailService()

// Create handler with email service
handler := handlers.NewInvitationHandler(database, emailService)
```

### In Tests

Tests use the mock email service with inspection methods:

```go
emailService := service.NewMockEmailService()
handler := handlers.NewInvitationHandler(mockDB, emailService)

// After sending...
assert.Equal(t, 1, emailService.SentCount())
lastEmail := emailService.LastSentEmail()
assert.Equal(t, "user@example.com", lastEmail.Recipient)
```

## Integration Points

### Create Invitation Endpoint

When an invitation is created via `POST /invitations`:

1. Invitation is saved to database
2. Email info is fetched (inviter name, org name)
3. Email is sent via `emailService.SendInvitation()`
4. **Email failure does NOT fail the request** - invitation is already created
5. Errors are logged but request returns 201 Created

This approach ensures:
- Invitations are always saved successfully
- Email delivery is best-effort
- Users get immediate feedback
- Failed emails can be retried later (future enhancement)

## Configuration

### Environment Variables (Future)

```bash
# Email provider (mock, sendgrid, ses)
EMAIL_PROVIDER=mock

# Base URL for invitation links
BASE_URL=https://app.arfa.com

# SendGrid API key (when using sendgrid provider)
SENDGRID_API_KEY=SG.xxx...

# AWS SES configuration (when using ses provider)
AWS_SES_REGION=us-east-1
AWS_SES_FROM_EMAIL=noreply@arfa.com
```

## Future Enhancements

### Real Email Providers

To add SendGrid support:

```go
type SendGridEmailService struct {
    apiKey  string
    baseURL string
}

func (s *SendGridEmailService) SendInvitation(email InvitationEmail) error {
    // Implement SendGrid API call
    // Use email templates
    // Handle errors and retries
}
```

### Email Templates

Add HTML email templates with:
- Company branding
- Responsive design
- Clear call-to-action buttons
- Legal footer (privacy, terms)

### Email Queue

For production reliability:
- Queue failed emails for retry
- Use message broker (RabbitMQ, Redis)
- Implement exponential backoff
- Track delivery status

### Email Analytics

Track email metrics:
- Emails sent
- Delivery rate
- Open rate (if using tracking pixels)
- Click-through rate
- Bounce rate

## Testing

### Unit Tests

```bash
# Test email service
go test -v ./services/api/internal/service -run TestMockEmailService

# Test invitation handler with email
go test -v ./services/api/internal/handlers -run TestCreateInvitation_Success
```

### Integration Tests

Integration tests use the mock email service to verify:
- Emails are sent when invitations are created
- Email contains correct recipient and token
- Email failures don't break invitation creation

## Security Considerations

### Token Security

- Tokens are 256-bit cryptographically secure random strings
- Tokens expire after 7 days
- Tokens can only be used once
- Failed login attempts don't reveal token validity

### Email Content

- Never include passwords in emails
- Only include single-use invitation tokens
- Use HTTPS for invitation links
- Set appropriate email headers (SPF, DKIM, DMARC)

### Rate Limiting

- Maximum 20 invitations per organization per day
- Prevents email spam abuse
- Tracked at database level
- Future: Add per-IP rate limiting

## Troubleshooting

### Emails Not Appearing in Logs

Check:
1. Handler is using correct email service instance
2. GetInvitationEmailInfo query returns data
3. Database has employee and organization records
4. Console output is not being filtered

### Email Validation Errors

Common errors:
- `recipient email is required` - Check invitation email field
- `invitation token is required` - Check token generation
- `invitation has expired` - Check expires_at timestamp

### Production Email Delivery Issues

When implementing real email provider:
1. Verify API credentials
2. Check email provider logs
3. Monitor bounce rates
4. Implement retry logic
5. Add delivery webhooks

## Related Files

- `/services/api/internal/service/email.go` - Email service implementation
- `/services/api/internal/service/email_test.go` - Email service tests
- `/services/api/internal/handlers/invitations.go` - Handler integration
- `/sqlc/queries/invitations.sql` - Database queries including email info

## See Also

- [Testing Guide](../development/testing.md)
- [Contributing](../development/contributing.md)
