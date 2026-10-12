package workers_test

import (
	"context"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/workers"
)

type stubReceiptStore struct {
	cutoff  time.Time
	removed int64
	err     error
}

func (s *stubReceiptStore) DeleteExpiredReceipts(ctx context.Context, cutoff time.Time) (int64, error) {
	s.cutoff = cutoff
	return s.removed, s.err
}

func TestTelemetryReceiptJobExpiresAfterRetention(t *testing.T) {
	store := &stubReceiptStore{removed: 3}
	job, err := workers.NewTelemetryReceiptJob(store)
	if err != nil {
		t.Fatal(err)
	}
	if job.Name() == "" {
		t.Fatal("job name is required")
	}
	before := time.Now().UTC()
	if err := job.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	want := before.AddDate(0, 0, -workers.TelemetryReceiptRetentionDays)
	if store.cutoff.After(before.AddDate(0, 0, -workers.TelemetryReceiptRetentionDays).Add(time.Minute)) || store.cutoff.Before(want.Add(-time.Minute)) {
		t.Fatalf("expected ~120-day cutoff, got %s", store.cutoff.Format(time.RFC3339))
	}
}

func TestNewTelemetryReceiptJobRequiresStore(t *testing.T) {
	if _, err := workers.NewTelemetryReceiptJob(nil); err == nil {
		t.Fatal("expected error for missing store")
	}
}
