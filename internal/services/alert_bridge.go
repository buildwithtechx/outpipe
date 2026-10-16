package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type AlertBridgeService struct {
	alerts    repositories.AlertRepository
	uptime    repositories.UptimeRepository
	incidents *UptimeService
	now       func() time.Time
}

func NewAlertBridgeService(alerts repositories.AlertRepository, uptime repositories.UptimeRepository, incidents *UptimeService) (*AlertBridgeService, error) {
	if alerts == nil || uptime == nil || incidents == nil {
		return nil, fmt.Errorf("alert, uptime, and incident services are required")
	}
	return &AlertBridgeService{alerts: alerts, uptime: uptime, incidents: incidents, now: time.Now}, nil
}

type CreateAlertInput struct {
	Name                 string   `json:"name"`
	MonitorIDs           []string `json:"monitorIds"`
	ConsecutiveFailures  int      `json:"consecutiveFailures"`
	ConsecutiveRecovered int      `json:"consecutiveRecoveries"`
	EvalIntervalSeconds  int      `json:"evalIntervalSeconds"`
	StatusPageTitle      string   `json:"statusPageTitle"`
	StatusPageImpact     string   `json:"statusPageImpact"`
}

func (s *AlertBridgeService) Create(ctx context.Context, orgID string, input CreateAlertInput) (models.ObservabilityAlert, error) {
	monitorIDs, err := s.resolveMonitors(ctx, orgID, input.MonitorIDs)
	if err != nil {
		return models.ObservabilityAlert{}, err
	}
	failures := input.ConsecutiveFailures
	if failures <= 0 {
		failures = 3
	}
	recovered := input.ConsecutiveRecovered
	if recovered <= 0 {
		recovered = 2
	}
	interval := input.EvalIntervalSeconds
	if interval <= 0 {
		interval = 60
	}
	impact := models.AlertImpact(strings.ToLower(strings.TrimSpace(input.StatusPageImpact)))
	if impact == "" {
		impact = models.AlertImpactDegraded
	}
	alert := models.ObservabilityAlert{
		OrganizationID:       orgID,
		Name:                 strings.TrimSpace(input.Name),
		MonitorIDs:           monitorIDs,
		ConsecutiveFailures:  failures,
		ConsecutiveRecovered: recovered,
		EvalIntervalSeconds:  interval,
		State:                models.AlertStateOK,
		Enabled:              true,
		StatusPageTitle:      strings.TrimSpace(input.StatusPageTitle),
		StatusPageImpact:     impact,
		NextEvaluationAt:     s.now().UTC(),
	}
	if alert.Name == "" {
		return models.ObservabilityAlert{}, fmt.Errorf("alert name is required")
	}
	if failures < 1 || failures > 10 || recovered < 1 || recovered > 10 || interval < 30 || interval > 3600 {
		return models.ObservabilityAlert{}, fmt.Errorf("alert thresholds are out of range")
	}
	if impact != models.AlertImpactDegraded && impact != models.AlertImpactOutage {
		return models.ObservabilityAlert{}, fmt.Errorf("alert impact must be degraded or outage")
	}
	if err := s.alerts.Create(ctx, &alert); err != nil {
		return models.ObservabilityAlert{}, err
	}
	return alert, nil
}

func (s *AlertBridgeService) resolveMonitors(ctx context.Context, orgID string, monitorIDs []string) (string, error) {
	if len(monitorIDs) == 0 || len(monitorIDs) > 100 {
		return "", fmt.Errorf("alert requires between 1 and 100 monitors")
	}
	monitors, err := s.uptime.ListOrgMonitors(ctx, orgID)
	if err != nil {
		return "", fmt.Errorf("list alert monitors: %w", err)
	}
	owned := map[string]bool{}
	for _, monitor := range monitors {
		owned[monitor.ID] = true
	}
	seen := map[string]bool{}
	unique := make([]string, 0, len(monitorIDs))
	for _, id := range monitorIDs {
		id = strings.TrimSpace(id)
		if id == "" || seen[id] {
			continue
		}
		if !owned[id] {
			return "", fmt.Errorf("monitor %q is not available", id)
		}
		seen[id] = true
		unique = append(unique, id)
	}
	if len(unique) == 0 {
		return "", fmt.Errorf("alert requires between 1 and 100 monitors")
	}
	encoded, err := json.Marshal(unique)
	if err != nil {
		return "", fmt.Errorf("encode alert monitors: %w", err)
	}
	return string(encoded), nil
}

func (s *AlertBridgeService) List(ctx context.Context, orgID string) ([]models.ObservabilityAlert, error) {
	return s.alerts.FindByOrganization(ctx, orgID)
}

func (s *AlertBridgeService) Delete(ctx context.Context, orgID, id string) error {
	alert, err := s.alerts.FindByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return err
	}
	if alert.OrganizationID != orgID {
		return repositories.ErrNotFound
	}
	return s.alerts.Delete(ctx, alert.ID)
}

func (s *AlertBridgeService) SetEnabled(ctx context.Context, orgID, id string, enabled bool) (models.ObservabilityAlert, error) {
	alert, err := s.alerts.FindByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return models.ObservabilityAlert{}, err
	}
	if alert.OrganizationID != orgID {
		return models.ObservabilityAlert{}, repositories.ErrNotFound
	}
	alert.Enabled = enabled
	alert.NextEvaluationAt = s.now().UTC()
	if err := s.alerts.Update(ctx, &alert); err != nil {
		return models.ObservabilityAlert{}, err
	}
	return alert, nil
}

func (s *AlertBridgeService) SetStatusPageLink(ctx context.Context, orgID, id, title, impact string, componentIDs []string) (models.ObservabilityAlert, error) {
	alert, err := s.alerts.FindByID(ctx, strings.TrimSpace(id))
	if err != nil {
		return models.ObservabilityAlert{}, err
	}
	if alert.OrganizationID != orgID {
		return models.ObservabilityAlert{}, repositories.ErrNotFound
	}
	parsed := models.AlertImpact(strings.ToLower(strings.TrimSpace(impact)))
	if parsed != "" && parsed != models.AlertImpactDegraded && parsed != models.AlertImpactOutage {
		return models.ObservabilityAlert{}, fmt.Errorf("alert impact must be degraded or outage")
	}
	if len(componentIDs) > 0 {
		encoded, err := s.resolveMonitors(ctx, orgID, componentIDs)
		if err != nil {
			return models.ObservabilityAlert{}, err
		}
		alert.StatusPageComponentIDs = encoded
	} else {
		alert.StatusPageComponentIDs = ""
	}
	if strings.TrimSpace(title) != "" {
		alert.StatusPageTitle = strings.TrimSpace(title)
	}
	if parsed != "" {
		alert.StatusPageImpact = parsed
	}
	if err := s.alerts.Update(ctx, &alert); err != nil {
		return models.ObservabilityAlert{}, err
	}
	return alert, nil
}
