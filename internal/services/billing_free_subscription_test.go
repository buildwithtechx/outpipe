package services

import (
	"context"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func TestProvisionFreeSubscriptionCreatesAnEntitledSubscription(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:billing-free-subscription?mode=memory&cache=shared"), &gorm.Config{})

	if err != nil {
		t.Fatal(err)
	}

	if err := db.AutoMigrate(&models.Plan{}, &models.Subscription{}); err != nil {
		t.Fatal(err)
	}

	repository, err := repositories.NewBillingRepository(db)

	if err != nil {
		t.Fatal(err)
	}

	service, err := NewBillingService(repository)

	if err != nil {
		t.Fatal(err)
	}

	const organizationID = "organization-free-plan"

	if err := service.ProvisionFreeSubscription(context.Background(), organizationID); err != nil {
		t.Fatal(err)
	}

	plan, subscription, err := service.Entitlements(context.Background(), organizationID)

	if err != nil {
		t.Fatal(err)
	}

	if plan.Key != "free" || subscription.PlanID != plan.ID || subscription.Status != models.SubscriptionStatusActive {
		t.Fatalf("expected active free entitlement, got plan %q with status %q", plan.Key, subscription.Status)
	}

	if err := service.ProvisionFreeSubscription(context.Background(), organizationID); err != nil {
		t.Fatal(err)
	}

	var count int64

	if err := db.Model(&models.Subscription{}).Where("organization_id = ?", organizationID).Count(&count).Error; err != nil {
		t.Fatal(err)
	}

	if count != 1 {
		t.Fatalf("expected one free subscription, got %d", count)
	}
}
