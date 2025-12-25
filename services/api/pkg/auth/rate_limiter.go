package auth

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/rastrigin-systems/arfa/generated/db"
)

// ErrRateLimitExceeded is returned when the password reset rate limit is exceeded
var ErrRateLimitExceeded = errors.New("password reset rate limit exceeded: maximum 3 requests per hour")

// CheckPasswordResetRateLimit checks if an employee has exceeded the password reset rate limit.
// Rate limit: Maximum 3 password reset requests per employee per hour.
func CheckPasswordResetRateLimit(ctx context.Context, employeeID uuid.UUID, querier db.Querier) error {
	count, err := querier.CountRecentPasswordResetRequests(ctx, employeeID)
	if err != nil {
		return err
	}

	if count >= 3 {
		return ErrRateLimitExceeded
	}

	return nil
}
