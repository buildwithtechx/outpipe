package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"outpipe.dev/outpipe/internal/models"
)

func (s *AlertBridgeService) ProcessDueAlerts(ctx context.Context, limit int) error {
	alerts, err := s.alerts.ClaimDueAlerts(ctx, s.now().UTC(), limit)
	if err != nil {
		return err
	}
	for i := range alerts {
		if err := s.evaluate(ctx, &alerts[i]); err != nil {
			return fmt.Errorf("evaluate alert %s: %w", alerts[i].ID, err)
		}
	}
	return nil
}

func (s *AlertBridgeService) evaluate(ctx context.Context, alert *models.ObservabilityAlert) error {
	var monitorIDs []string
	if err := json.Unmarshal([]byte(alert.MonitorIDs), &monitorIDs); err != nil {
		return fmt.Errorf("decode alert monitors: %w", err)
	}
	now := s.now().UTC()
	failed, recovered := false, true
	seen := 0
	for _, monitorID := range monitorIDs {
		monitor, err := s.uptime.GetMonitor(ctx, monitorID)
		if err != nil {
			continue
		}
		if monitor.Status == models.MonitorStatusPaused {
			continue
		}
		window := alert.ConsecutiveFailures
		if alert.ConsecutiveRecovered > window {
			window = alert.ConsecutiveRecovered
		}
		checks, err := s.uptime.GetRecentChecks(ctx, monitorID, window)
		if err != nil {
			return fmt.Errorf("read monitor checks: %w", err)
		}
		if len(checks) == 0 {
			recovered = false
			continue
		}
		seen++
		consecutiveFailures, consecutiveSuccesses := 0, 0
		for _, check := range checks {
			if check.Success {
				break
			}
			consecutiveFailures++
		}
		for _, check := range checks {
			if !check.Success {
				break
			}
			consecutiveSuccesses++
		}
		if consecutiveFailures >= alert.ConsecutiveFailures {
			failed = true
		}
		if consecutiveSuccesses < alert.ConsecutiveRecovered {
			recovered = false
		}
	}
	alert.LastEvaluatedAt = &now
	alert.NextEvaluationAt = now.Add(time.Duration(alert.EvalIntervalSeconds) * time.Second)
	if seen == 0 {
		return s.alerts.Update(ctx, alert)
	}
	switch alert.State {
	case models.AlertStateOK:
		if failed {
			return s.fire(ctx, alert, now)
		}
		alert.FailureStreak = 0
		alert.RecoveryStreak++
	case models.AlertStateFiring:
		if recovered {
			return s.resolve(ctx, alert, now)
		}
		alert.RecoveryStreak = 0
		alert.FailureStreak++
	}
	return s.alerts.Update(ctx, alert)
}

func (s *AlertBridgeService) fire(ctx context.Context, alert *models.ObservabilityAlert, now time.Time) error {
	title := alert.StatusPageTitle
	if title == "" {
		title = alert.Name
	}
	severity := models.IncidentSeverityMinor
	if alert.StatusPageImpact == models.AlertImpactOutage {
		severity = models.IncidentSeverityMajor
	}
	var monitorIDs []string
	componentsJSON := alert.StatusPageComponentIDs
	if componentsJSON == "" {
		componentsJSON = alert.MonitorIDs
	}
	_ = json.Unmarshal([]byte(componentsJSON), &monitorIDs)
	incident, err := s.incidents.CreateIncident(ctx, alert.OrganizationID, CreateIncidentInput{
		Title:        title,
		Severity:     severity,
		Message:      fmt.Sprintf("Alert %q is firing", alert.Name),
		ComponentIDs: monitorIDs,
	})
	if err != nil {
		return fmt.Errorf("publish firing incident: %w", err)
	}
	alert.State = models.AlertStateFiring
	alert.FailureStreak = 0
	alert.RecoveryStreak = 0
	alert.LastStateChangedAt = &now
	if err := s.alerts.Update(ctx, alert); err != nil {
		return err
	}
	return s.alerts.OpenIncident(ctx, &models.AlertIncident{
		OrganizationID: alert.OrganizationID,
		AlertID:        alert.ID,
		IncidentID:     incident.ID,
		StartedAt:      now,
	})
}

func (s *AlertBridgeService) resolve(ctx context.Context, alert *models.ObservabilityAlert, now time.Time) error {
	link, err := s.alerts.GetOpenIncident(ctx, alert.ID)
	if err != nil {
		return fmt.Errorf("find open alert incident: %w", err)
	}
	if err := s.incidents.AddIncidentUpdate(ctx, link.IncidentID, models.IncidentStatusResolved, fmt.Sprintf("Alert %q recovered", alert.Name)); err != nil {
		return fmt.Errorf("annotate recovered incident: %w", err)
	}
	alert.State = models.AlertStateOK
	alert.FailureStreak = 0
	alert.RecoveryStreak = 0
	alert.LastStateChangedAt = &now
	if err := s.alerts.Update(ctx, alert); err != nil {
		return err
	}
	return s.alerts.ResolveIncident(ctx, alert.ID, now)
}
