package service

import (
	"context"
	"log"
	"time"

	"github.com/rastrigin-systems/ubik-enterprise/generated/db"
)

// RetentionPolicy defines the retention policy for logs
type RetentionPolicy struct {
	RetentionDays int
	db            db.Querier
}

// NewRetentionPolicy creates a new retention policy
func NewRetentionPolicy(db db.Querier, retentionDays int) *RetentionPolicy {
	if retentionDays <= 0 {
		retentionDays = 30 // Default to 30 days
	}
	return &RetentionPolicy{
		RetentionDays: retentionDays,
		db:            db,
	}
}

// CleanupOldLogs deletes logs older than the retention period
func (rp *RetentionPolicy) CleanupOldLogs(ctx context.Context) error {
	cutoffDate := time.Now().Add(-time.Duration(rp.RetentionDays) * 24 * time.Hour)

	loggingSvc := NewLoggingService(rp.db)
	err := loggingSvc.DeleteOldLogs(ctx, cutoffDate)
	if err != nil {
		return err
	}

	log.Printf("Deleted activity logs older than %s (%d days)",
		cutoffDate.Format(time.RFC3339), rp.RetentionDays)
	return nil
}

// StartRetentionWorker starts a background worker that runs cleanup periodically
func (rp *RetentionPolicy) StartRetentionWorker(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run cleanup immediately on start
	if err := rp.CleanupOldLogs(ctx); err != nil {
		log.Printf("Error during initial cleanup: %v", err)
	}

	for {
		select {
		case <-ticker.C:
			if err := rp.CleanupOldLogs(ctx); err != nil {
				log.Printf("Error during scheduled cleanup: %v", err)
			}
		case <-ctx.Done():
			log.Println("Retention worker stopped")
			return
		}
	}
}
