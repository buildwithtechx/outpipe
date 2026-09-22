package repositories

import (
	"context"
	"errors"
	"sync"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

func TestApplyBillingEventReturnsDuplicateSentinel(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:billing-idempotency?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&models.BillingEvent{}); err != nil {
		t.Fatal(err)
	}

	repository, err := NewBillingRepository(db)
	if err != nil {
		t.Fatal(err)
	}

	first := &models.BillingEvent{
		Provider:        models.BillingProviderPolar,
		ProviderEventID: "event-duplicate",
		EventType:       "subscription.active",
		PayloadHash:     "hash-1",
	}
	if err := repository.ApplyBillingEvent(context.Background(), first, nil); err != nil {
		t.Fatal(err)
	}

	second := &models.BillingEvent{
		Provider:        models.BillingProviderPolar,
		ProviderEventID: "event-duplicate",
		EventType:       "subscription.active",
		PayloadHash:     "hash-2",
	}
	if err := repository.ApplyBillingEvent(context.Background(), second, nil); !errors.Is(err, ErrBillingEventDuplicate) {
		t.Fatalf("expected duplicate sentinel, got %v", err)
	}
}

func TestApplyBillingEventConcurrentlyRecordsOnlyOneEvent(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:billing-idempotency-concurrent?mode=memory&cache=shared&_busy_timeout=5000"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)

	if err := db.AutoMigrate(&models.BillingEvent{}); err != nil {
		t.Fatal(err)
	}

	repository, err := NewBillingRepository(db)
	if err != nil {
		t.Fatal(err)
	}

	const deliveries = 8
	errorsSeen := make(chan error, deliveries)
	var group sync.WaitGroup
	for range deliveries {
		group.Add(1)
		go func() {
			defer group.Done()
			event := &models.BillingEvent{
				Provider:        models.BillingProviderPolar,
				ProviderEventID: "event-concurrent",
				EventType:       "subscription.active",
				PayloadHash:     "same-payload",
			}
			errorsSeen <- repository.ApplyBillingEvent(context.Background(), event, nil)
		}()
	}
	group.Wait()
	close(errorsSeen)

	created := 0
	duplicates := 0
	for err := range errorsSeen {
		switch {
		case err == nil:
			created++
		case errors.Is(err, ErrBillingEventDuplicate):
			duplicates++
		default:
			t.Fatalf("unexpected concurrent delivery error: %v", err)
		}
	}
	if created != 1 || duplicates != deliveries-1 {
		t.Fatalf("expected one creation and %d duplicates, got %d creations and %d duplicates", deliveries-1, created, duplicates)
	}
}

func TestMarkBillingEventFailedPersistsRetryableFailure(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:billing-failure?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.BillingEvent{}); err != nil {
		t.Fatal(err)
	}
	repository, err := NewBillingRepository(db)
	if err != nil {
		t.Fatal(err)
	}

	reason := "provider subscription was not found"
	if err := repository.MarkBillingEventFailed(context.Background(), models.BillingProviderPolar, "event-failed", reason); err != nil {
		t.Fatal(err)
	}
	event, err := repository.FindBillingEvent(context.Background(), models.BillingProviderPolar, "event-failed")
	if err != nil {
		t.Fatal(err)
	}
	if event.FailureReason != reason || event.ProcessedAt != nil {
		t.Fatalf("expected retryable failed event, got failure=%q processed=%v", event.FailureReason, event.ProcessedAt)
	}
}
