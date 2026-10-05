package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/models"
)

func (s *UptimeService) Probe(ctx context.Context, monitor *models.UptimeMonitor) (models.UptimeCheck, error) {
	if monitor == nil {
		return models.UptimeCheck{}, fmt.Errorf("monitor is required")
	}
	timeout := time.Duration(monitor.TimeoutSeconds) * time.Second
	if timeout < time.Second || timeout > 60*time.Second {
		timeout = 10 * time.Second
	}
	probeCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	start := time.Now()
	statusCode, probeErr := s.probeTarget(probeCtx, monitor)
	success := probeErr == nil
	errorMsg := ""
	if probeErr != nil {
		errorMsg = probeErr.Error()
	}
	elapsed := time.Since(start)
	latency := elapsed.Milliseconds()
	checkID := make([]byte, 16)
	if _, err := rand.Read(checkID); err != nil {
		return models.UptimeCheck{}, fmt.Errorf("generate check id: %w", err)
	}

	if success && monitor.MaxLatencyMs > 0 && elapsed > time.Duration(monitor.MaxLatencyMs)*time.Millisecond {
		success = false
		errorMsg = "latency assertion failed"
	}
	now := time.Now().UTC()
	check := models.UptimeCheck{
		ID:           hex.EncodeToString(checkID),
		MonitorID:    monitor.ID,
		StatusCode:   statusCode,
		LatencyMs:    latency,
		Success:      success,
		ErrorMessage: errorMsg,
		CheckedAt:    now,
	}

	if err := s.repo.RecordCheck(ctx, &check); err != nil {
		return models.UptimeCheck{}, fmt.Errorf("persist uptime check: %w", err)
	}

	monitor.LastCheckAt = &now
	nextProbe := now.Add(time.Duration(monitor.IntervalSeconds) * time.Second)
	monitor.NextProbeAt = &nextProbe
	monitor.LatencyMs = latency
	if success {
		monitor.Status = models.MonitorStatusUp
	} else {
		monitor.Status = models.MonitorStatusDown
	}

	recent, err := s.repo.GetRecentChecks(ctx, monitor.ID, 30)
	historyErr := err
	if historyErr == nil && len(recent) > 0 {
		successCount := 0
		for _, c := range recent {
			if c.Success {
				successCount++
			}
		}
		monitor.UptimeRatio = float64(successCount) / float64(len(recent)) * 100.0
	}
	monitor.UpdatedAt = now
	if err := s.repo.UpdateMonitor(ctx, monitor); err != nil {
		return models.UptimeCheck{}, fmt.Errorf("update probed monitor: %w", err)
	}
	if historyErr != nil {
		return check, fmt.Errorf("get recent uptime checks: %w", historyErr)
	}

	return check, nil
}
