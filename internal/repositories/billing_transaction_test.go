package repositories

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

func TestBillingEventAndSubscriptionCommitTogether(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:billing-transaction?mode=memory&cache=shared"), &gorm.Config{})

	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&models.Subscription{}, &models.BillingEvent{}); err != nil {
		t.Fatal(err)
	}

	subscription := models.Subscription{OrganizationID: "org-1", PlanID: "plan-1", Provider: models.BillingProviderPolar, ProviderSubID: "sub-1", Status: models.SubscriptionStatusPastDue, BillingInterval: "month"}

	if err := db.Create(&subscription).Error; err != nil {
		t.Fatal(err)
	}

	repository, err := NewBillingRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	subscription.Status = models.SubscriptionStatusActive
	event := &models.BillingEvent{Provider: models.BillingProviderPolar, ProviderEventID: "event-1", EventType: "subscription.active", PayloadHash: "hash"}

	if err := repository.ApplyBillingEvent(context.Background(), event, &subscription); err != nil {
		t.Fatal(err)
	}

	stored, err := repository.FindSubscriptionByProvider(context.Background(), models.BillingProviderPolar, "sub-1")

	if err != nil {
		t.Fatal(err)
	}

	if stored.Status != models.SubscriptionStatusActive {
		t.Fatalf("expected active subscription, got %s", stored.Status)
	}

	storedEvent, err := repository.FindBillingEvent(context.Background(), models.BillingProviderPolar, "event-1")

	if err != nil {
		t.Fatal(err)
	}

	if storedEvent.ProcessedAt == nil {
		t.Fatal("expected billing event to be marked processed")
	}
}
