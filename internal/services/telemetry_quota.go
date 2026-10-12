package services

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type TelemetryQuotaExceededError struct {
	ResetsAt time.Time
}

func (e *TelemetryQuotaExceededError) Error() string {
	return fmt.Sprintf("monthly telemetry allowance exceeded; resets at %s", e.ResetsAt.UTC().Format(time.RFC3339))
}

func (s *ObservabilityService) SetQuota(quota repositories.TelemetryQuotaRepository, billing repositories.BillingRepository) {
	s.quota = quota
	s.quotaBilling = billing
}

func (s *ObservabilityService) admitTelemetry(ctx context.Context, orgID, signal string, raw []byte) (bool, error) {
	if s.quota == nil {
		return false, nil
	}
	if orgID == "" {
		return false, fmt.Errorf("organization is required")
	}
	month, _, resetsAt := models.TelemetryUsagePeriod(time.Now().UTC())
	limit := models.DefaultTelemetryBytesMonthly
	if s.quotaBilling != nil {
		if subscription, err := s.quotaBilling.FindSubscription(ctx, orgID); err == nil {
			if plan, err := s.quotaBilling.FindPlanByID(ctx, subscription.PlanID); err == nil {
				limit = plan.TelemetryBytes
				if limit == 0 {
					limit = models.DefaultTelemetryBytesMonthly
				}
			}
		} else if !errors.Is(err, repositories.ErrNotFound) {
			return false, fmt.Errorf("resolve telemetry quota plan: %w", err)
		}
	}
	duplicate, err := s.quota.AdmitBatch(ctx, orgID, month, telemetryReceiptKey(signal, raw), int64(len(raw)), limit)
	if err != nil {
		if errors.Is(err, repositories.ErrTelemetryQuotaExceeded) {
			return false, &TelemetryQuotaExceededError{ResetsAt: resetsAt}
		}
		return false, err
	}
	return duplicate, nil
}

func (s *ObservabilityService) refundTelemetry(ctx context.Context, orgID, signal string, raw []byte) {
	if s.quota == nil || orgID == "" {
		return
	}
	month, _, _ := models.TelemetryUsagePeriod(time.Now().UTC())
	_ = s.quota.RefundBatch(ctx, orgID, month, telemetryReceiptKey(signal, raw), int64(len(raw)))
}

func telemetryReceiptKey(signal string, raw []byte) string {
	sum := sha256.Sum256(append(append([]byte(signal), 0), raw...))
	return hex.EncodeToString(sum[:])
}
