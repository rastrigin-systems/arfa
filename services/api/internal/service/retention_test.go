package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/sergeirastrigin/ubik-enterprise/generated/mocks"
)

func TestRetentionPolicy_CleanupOldLogs(t *testing.T) {
	tests := []struct {
		name          string
		retentionDays int
		mockSetup     func(*mocks.MockQuerier)
		wantErr       bool
	}{
		{
			name:          "cleanup with 30 day retention",
			retentionDays: 30,
			mockSetup: func(m *mocks.MockQuerier) {
				m.EXPECT().
					DeleteOldLogs(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:          "cleanup with 90 day retention",
			retentionDays: 90,
			mockSetup: func(m *mocks.MockQuerier) {
				m.EXPECT().
					DeleteOldLogs(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
		{
			name:          "default retention when zero",
			retentionDays: 0, // Should default to 30 days
			mockSetup: func(m *mocks.MockQuerier) {
				m.EXPECT().
					DeleteOldLogs(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockDB := mocks.NewMockQuerier(ctrl)
			tt.mockSetup(mockDB)

			rp := NewRetentionPolicy(mockDB, tt.retentionDays)

			err := rp.CleanupOldLogs(context.Background())
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewRetentionPolicy_DefaultsTo30Days(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Test with zero
	rp := NewRetentionPolicy(mockDB, 0)
	assert.Equal(t, 30, rp.RetentionDays, "should default to 30 days when 0")

	// Test with negative
	rp = NewRetentionPolicy(mockDB, -10)
	assert.Equal(t, 30, rp.RetentionDays, "should default to 30 days when negative")

	// Test with positive
	rp = NewRetentionPolicy(mockDB, 60)
	assert.Equal(t, 60, rp.RetentionDays, "should use provided value when positive")
}

func TestRetentionPolicy_StartRetentionWorker(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockDB := mocks.NewMockQuerier(ctrl)

	// Expect cleanup to be called at least once (on start)
	mockDB.EXPECT().
		DeleteOldLogs(gomock.Any(), gomock.Any()).
		Return(nil).
		MinTimes(1)

	rp := NewRetentionPolicy(mockDB, 30)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Run worker in background
	done := make(chan bool)
	go func() {
		rp.StartRetentionWorker(ctx, 50*time.Millisecond)
		done <- true
	}()

	// Wait for context to cancel or timeout
	select {
	case <-done:
		// Worker stopped as expected
	case <-time.After(200 * time.Millisecond):
		t.Fatal("worker did not stop in time")
	}
}
