package workers

import (
	"context"
	"fmt"
	"time"
)

const TelemetryReceiptRetentionDays = 120

type TelemetryReceiptStore interface {
	DeleteExpiredReceipts(ctx context.Context, cutoff time.Time) (int64, error)
}

type TelemetryReceiptJob struct {
	store TelemetryReceiptStore
	now   func() time.Time
}

func NewTelemetryReceiptJob(store TelemetryReceiptStore) (*TelemetryReceiptJob, error) {
	if store == nil {
		return nil, fmt.Errorf("telemetry receipt store is required")
	}
	return &TelemetryReceiptJob{store: store, now: time.Now}, nil
}

func (j *TelemetryReceiptJob) Name() string {
	return "telemetry-receipt-cleanup"
}

func (j *TelemetryReceiptJob) Run(ctx context.Context) error {
	cutoff := j.now().UTC().AddDate(0, 0, -TelemetryReceiptRetentionDays)
	if _, err := j.store.DeleteExpiredReceipts(ctx, cutoff); err != nil {
		return fmt.Errorf("clean up expired telemetry receipts: %w", err)
	}
	return nil
}
