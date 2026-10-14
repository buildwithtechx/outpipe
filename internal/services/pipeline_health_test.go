package services

import (
	"context"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func setupPipelineHealth(t *testing.T) (*PipelineHealthService, *repositories.GormWebhookRepository, *repositories.GormTelemetryExportRepository, *repositories.GormUptimeRepository) {
	t.Helper()
	db := reviewTestDatabase(t, &models.WebhookSubscription{}, &models.WebhookDelivery{}, &models.TelemetryExport{}, &models.UptimeStatusPage{}, &models.UptimeSubscriber{})
	webhooks, err := repositories.NewWebhookRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	exports := repositories.NewGormTelemetryExportRepository(db)
	uptime, err := repositories.NewUptimeRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	health, err := NewPipelineHealthService(webhooks, exports, uptime)
	if err != nil {
		t.Fatal(err)
	}
	return health, webhooks, exports, uptime
}

func TestPipelineHealthReportsHealthyWhenEmpty(t *testing.T) {
	health, _, _, _ := setupPipelineHealth(t)
	result, err := health.Health(context.Background(), "org-1")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Healthy || len(result.Queues) != 3 {
		t.Fatalf("health = %+v", result)
	}
	for _, queue := range result.Queues {
		if queue.Depth != 0 || queue.DeadLettered != 0 || !queue.Healthy {
			t.Fatalf("queue = %+v", queue)
		}
	}
}

func TestPipelineHealthFlagsDeadLettersAndLag(t *testing.T) {
	health, webhooks, exports, _ := setupPipelineHealth(t)
	ctx := context.Background()
	subscription := models.WebhookSubscription{OrganizationID: "org-1", Name: "hooks", URL: "https://example.com/hooks", Events: `["tunnel.connected"]`, Secret: "secret"}
	if err := webhooks.Create(ctx, &subscription); err != nil {
		t.Fatal(err)
	}
	old := time.Now().UTC().Add(-time.Hour)
	for i, status := range []models.WebhookDeliveryStatus{models.WebhookDeliveryPending, models.WebhookDeliveryFailed} {
		delivery := models.WebhookDelivery{SubscriptionID: subscription.ID, EventID: "event", EventType: "tunnel.connected", Payload: "{}", Status: status, AvailableAt: old}
		if i == 1 {
			delivery.EventID = "event-dead"
		}
		delivery.CreatedAt = old
		if i == 1 {
			delivery.Attempts = webhookMaxDeliveryAttempts
		}
		if err := webhooks.CreateDelivery(ctx, &delivery); err != nil {
			t.Fatal(err)
		}
	}
	failed := models.TelemetryExport{OrganizationID: "org-1", Signal: "captures", EventID: "event-1", Payload: "{}", EventTime: old, Status: models.TelemetryExportFailed, AvailableAt: old, Attempts: repositories.TelemetryExportMaxAttempts}
	if err := exports.Create(ctx, &failed); err != nil {
		t.Fatal(err)
	}
	result, err := health.Health(ctx, "org-1")
	if err != nil {
		t.Fatal(err)
	}
	if result.Healthy {
		t.Fatal("unhealthy pipelines reported healthy")
	}
	byQueue := make(map[string]QueueHealth)
	for _, queue := range result.Queues {
		byQueue[queue.Queue] = queue
	}
	webhook := byQueue["webhook-deliveries"]
	if webhook.Depth != 1 || webhook.DeadLettered != 1 || webhook.Healthy {
		t.Fatalf("webhook health = %+v", webhook)
	}
	if webhook.OldestAgeSecs == nil || *webhook.OldestAgeSecs < 3500 {
		t.Fatalf("webhook lag = %+v", webhook.OldestAgeSecs)
	}
	export := byQueue["telemetry-exports"]
	if export.Depth != 1 || export.DeadLettered != 1 || export.Healthy {
		t.Fatalf("export health = %+v", export)
	}
	notify := byQueue["subscriber-notifications"]
	if !notify.Healthy {
		t.Fatalf("notification health = %+v", notify)
	}
}

func TestPipelineDeadLettersAggregatesQueues(t *testing.T) {
	health, webhooks, exports, _ := setupPipelineHealth(t)
	ctx := context.Background()
	subscription := models.WebhookSubscription{OrganizationID: "org-1", Name: "hooks", URL: "https://example.com/hooks", Events: `["tunnel.connected"]`, Secret: "secret"}
	if err := webhooks.Create(ctx, &subscription); err != nil {
		t.Fatal(err)
	}
	delivery := models.WebhookDelivery{SubscriptionID: subscription.ID, EventID: "event", EventType: "tunnel.connected", Payload: "{}", Status: models.WebhookDeliveryFailed, AvailableAt: time.Now().UTC(), Attempts: webhookMaxDeliveryAttempts}
	if err := webhooks.CreateDelivery(ctx, &delivery); err != nil {
		t.Fatal(err)
	}
	job := models.TelemetryExport{OrganizationID: "org-1", Signal: "captures", EventID: "event-1", Payload: "{}", EventTime: time.Now().UTC(), Status: models.TelemetryExportFailed, AvailableAt: time.Now().UTC(), Attempts: repositories.TelemetryExportMaxAttempts, LastError: "analytics down"}
	if err := exports.Create(ctx, &job); err != nil {
		t.Fatal(err)
	}
	letters, err := health.DeadLetters(ctx, "org-1", 25)
	if err != nil {
		t.Fatal(err)
	}
	if len(letters) != 2 {
		t.Fatalf("dead letters = %+v", letters)
	}
	queues := map[string]bool{}
	for _, letter := range letters {
		queues[letter.Queue] = true
		if letter.LastError == "" && letter.Queue == "telemetry-exports" {
			t.Fatalf("export dead letter missing error: %+v", letter)
		}
	}
	if !queues["webhook-deliveries"] || !queues["telemetry-exports"] {
		t.Fatalf("queues = %v", queues)
	}
}
