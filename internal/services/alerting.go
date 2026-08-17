package services

import (
	"context"
	"fmt"

	"outpipe.dev/outpipe/internal/infra/telemetry"
)

type AlertType string

const (
	AlertFailedWebhook          AlertType = "failed_webhook"
	AlertRepeatedPaymentFailure AlertType = "repeated_payment_failure"
	AlertStalePresenceGrowth    AlertType = "stale_presence_growth"
	AlertQuotaInconsistency     AlertType = "quota_inconsistency"
)

type AlertService struct {
	reporter telemetry.Reporter
}

func NewAlertService(reporter telemetry.Reporter) (*AlertService, error) {

	if reporter == nil {
		reporter = telemetry.NopReporter{}
	}

	return &AlertService{reporter: reporter}, nil
}

func (s *AlertService) AlertFailedWebhook(ctx context.Context, provider, eventID, reason string) error {

	if err := s.reporter.Report(ctx, telemetry.Event{
		Name: string(AlertFailedWebhook),
		Properties: map[string]any{
			"provider": provider,
			"event_id": eventID,
			"reason":   reason,
		},
	}); err != nil {
		return fmt.Errorf("report webhook failure alert: %w", err)
	}

	return nil
}

func (s *AlertService) AlertRepeatedPaymentFailure(ctx context.Context, organizationID, subscriptionID string, failureCount int) error {

	if err := s.reporter.Report(ctx, telemetry.Event{
		Name: string(AlertRepeatedPaymentFailure),
		Properties: map[string]any{
			"organization_id": organizationID,
			"subscription_id": subscriptionID,
			"failure_count":   failureCount,
		},
	}); err != nil {
		return fmt.Errorf("report payment failure alert: %w", err)
	}

	return nil
}

func (s *AlertService) AlertStalePresenceGrowth(ctx context.Context, organizationID string, staleCount int) error {

	if err := s.reporter.Report(ctx, telemetry.Event{
		Name: string(AlertStalePresenceGrowth),
		Properties: map[string]any{
			"organization_id": organizationID,
			"stale_count":     staleCount,
		},
	}); err != nil {
		return fmt.Errorf("report stale presence alert: %w", err)
	}

	return nil
}

func (s *AlertService) AlertQuotaInconsistency(ctx context.Context, organizationID string, used, allowed int64) error {

	if err := s.reporter.Report(ctx, telemetry.Event{
		Name: string(AlertQuotaInconsistency),
		Properties: map[string]any{
			"organization_id": organizationID,
			"used":            used,
			"allowed":         allowed,
		},
	}); err != nil {
		return fmt.Errorf("report quota inconsistency alert: %w", err)
	}

	return nil
}
