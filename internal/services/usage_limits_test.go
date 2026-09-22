package services

import (
	"context"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/models"
)

type usageLimitRepository struct {
	seenLimit int
}

func (r *usageLimitRepository) CreateEvent(context.Context, *models.UsageEvent) error { return nil }
func (r *usageLimitRepository) UpsertSnapshot(context.Context, *models.UsageSnapshot) error {
	return nil
}
func (r *usageLimitRepository) FindSnapshot(context.Context, string, time.Time) (models.UsageSnapshot, error) {
	return models.UsageSnapshot{}, nil
}
func (r *usageLimitRepository) ListEvents(_ context.Context, _ string, _, _ time.Time, limit int, _ time.Time, _ string) ([]models.UsageEvent, error) {
	r.seenLimit = limit
	return nil, nil
}
func (r *usageLimitRepository) ListRequestEvents(context.Context, string, time.Time, time.Time, int) ([]models.UsageEvent, error) {
	return nil, nil
}
func (r *usageLimitRepository) AggregatePeriod(context.Context, string, time.Time, time.Time) (models.UsageSnapshot, error) {
	return models.UsageSnapshot{}, nil
}
func (r *usageLimitRepository) DeleteBefore(context.Context, string, time.Time) (int64, error) {
	return 0, nil
}

func TestListEventsRejectsOversizedLimitAndPreservesEmptyPages(t *testing.T) {
	repository := &usageLimitRepository{}
	service, err := NewUsageService(repository)
	if err != nil {
		t.Fatal(err)
	}
	from := time.Now().UTC().Add(-time.Hour)
	to := time.Now().UTC()

	if _, err := service.ListEvents(context.Background(), "org-1", from, to, 1001, ""); err == nil {
		t.Fatal("oversized usage limit succeeded; want an error")
	}
	page, err := service.ListEvents(context.Background(), "org-1", from, to, 100, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Events) != 0 || page.NextCursor != "" || repository.seenLimit != 101 {
		t.Fatalf("empty bounded page = %#v, repository limit = %d", page, repository.seenLimit)
	}
}
