package services_test

import (
	"context"
	"errors"
	"testing"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func setupQuotaObservability(t *testing.T) (*services.ObservabilityService, repositories.BillingRepository) {
	t.Helper()
	db, cleanup := setupTestObservabilityDB(t)
	t.Cleanup(cleanup)
	if err := db.AutoMigrate(&models.Plan{}, &models.Subscription{}, &models.TelemetryMetric{}, &models.OrganizationMonthlyUsage{}, &models.OrganizationTelemetryReceipt{}); err != nil {
		t.Fatalf("migrate quota tables: %v", err)
	}
	billing, err := repositories.NewBillingRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	quota, err := repositories.NewTelemetryQuotaRepository(db)
	if err != nil {
		t.Fatal(err)
	}
	svc := services.NewObservabilityService(repositories.NewGormObservabilityRepository(db))
	svc.SetQuota(quota, billing)
	plan := models.Plan{Key: "tiny", Name: "Tiny", Currency: "USD", MaxTunnels: 1, MaxDomains: 0, MaxMembers: 1, TelemetryBytes: 200, RetentionDays: 3, Features: `{}`, Active: true}
	if err := billing.CreatePlan(context.Background(), &plan); err != nil {
		t.Fatal(err)
	}
	stored, err := billing.FindPlan(context.Background(), "tiny")
	if err != nil {
		t.Fatal(err)
	}
	subscription := models.Subscription{OrganizationID: "org-quota", PlanID: stored.ID, Provider: models.BillingProviderInternal, ProviderSubID: "free:org-quota", Status: models.SubscriptionStatusActive, BillingInterval: models.BillingIntervalMonth}
	if err := billing.CreateSubscription(context.Background(), &subscription); err != nil {
		t.Fatal(err)
	}
	return svc, billing
}

func quotaMetricBatch() []byte {
	return []byte(`{"resourceMetrics":[{"scopeMetrics":[{"metrics":[{"name":"requests","sum":{"dataPoints":[{"timeUnixNano":"1000000000","asInt":"7"}]}}]}]}]}`)
}

func TestIngestMetricsEnforcesMonthlyQuota(t *testing.T) {
	svc, _ := setupQuotaObservability(t)
	ctx := context.Background()
	if _, err := svc.IngestMetrics(ctx, "org-quota", quotaMetricBatch(), "application/json"); err != nil {
		t.Fatalf("first batch rejected: %v", err)
	}
	var exceeded *services.TelemetryQuotaExceededError
	if _, err := svc.IngestMetrics(ctx, "org-quota", quotaMetricBatch(), "application/json"); err == nil {
		t.Log("identical retry treated as duplicate")
	} else if !errors.As(err, &exceeded) {
		t.Fatalf("unexpected retry error: %v", err)
	}
	other := []byte(`{"resourceMetrics":[{"scopeMetrics":[{"metrics":[{"name":"other-metric-with-a-different-name","sum":{"dataPoints":[{"timeUnixNano":"1000000000","asInt":"9"}]}}]}]}]}`)
	if _, err := svc.IngestMetrics(ctx, "org-quota", other, "application/json"); !errors.As(err, &exceeded) {
		t.Fatalf("expected quota exceeded for new bytes, got %v", err)
	}
}
