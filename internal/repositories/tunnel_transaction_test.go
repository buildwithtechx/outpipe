package repositories

import (
	"context"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
)

func TestTransitionWithDeliveriesRollsBackWhenOutboxInsertFails(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:tunnel-outbox-transaction?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&models.Tunnel{}, &models.WebhookSubscription{}, &models.WebhookDelivery{}); err != nil {
		t.Fatal(err)
	}

	repository, err := NewTunnelRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	tunnel := models.Tunnel{OrganizationID: "org-1", Name: "preview", Protocol: models.TunnelProtocolHTTP, Status: models.TunnelStatusCreated, TargetHost: "127.0.0.1", TargetPort: 3000, PublicHostname: "preview.outpipe.app"}
	if err := repository.Create(context.Background(), &tunnel); err != nil {
		t.Fatal(err)
	}

	subscription := models.WebhookSubscription{OrganizationID: tunnel.OrganizationID, Name: "events", URL: "https://example.test/events", Events: `["tunnel.connected"]`, Secret: "secret"}
	if err := db.Create(&subscription).Error; err != nil {
		t.Fatal(err)
	}

	existing := models.WebhookDelivery{SubscriptionID: subscription.ID, EventID: "event-1", EventType: "tunnel.connected", Payload: `{}`, Status: models.WebhookDeliveryPending, AvailableAt: time.Now().UTC()}
	if err := db.Create(&existing).Error; err != nil {
		t.Fatal(err)
	}

	err = repository.TransitionWithDeliveries(context.Background(), tunnel.ID, models.TunnelStatusActive, nil, []models.WebhookDelivery{{SubscriptionID: subscription.ID, EventID: existing.EventID, EventType: existing.EventType, Payload: `{}`, Status: models.WebhookDeliveryPending, AvailableAt: time.Now().UTC()}})
	if err == nil {
		t.Fatal("expected duplicate delivery to fail")
	}

	stored, err := repository.FindByID(context.Background(), tunnel.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Status != models.TunnelStatusCreated {
		t.Fatalf("tunnel status = %q, want %q after transaction rollback", stored.Status, models.TunnelStatusCreated)
	}
}
