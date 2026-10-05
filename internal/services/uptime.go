package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type UptimeService struct {
	repo       repositories.UptimeRepository
	httpClient *http.Client
}

func NewUptimeService(repo repositories.UptimeRepository) (*UptimeService, error) {
	if repo == nil {
		return nil, fmt.Errorf("uptime repository is required")
	}
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("stopped after 5 redirects")
			}
			return nil
		},
	}
	return &UptimeService{repo: repo, httpClient: client}, nil
}

type CreateMonitorInput struct {
	Name            string `json:"name"`
	URL             string `json:"url"`
	Protocol        string `json:"protocol"`
	Method          string `json:"method"`
	IntervalSeconds int    `json:"intervalSeconds"`
	TimeoutSeconds  int    `json:"timeoutSeconds"`
}

func (s *UptimeService) CreateMonitor(ctx context.Context, orgID string, input CreateMonitorInput) (models.UptimeMonitor, error) {
	if strings.TrimSpace(input.Name) == "" {
		return models.UptimeMonitor{}, fmt.Errorf("monitor name is required")
	}
	if strings.TrimSpace(input.URL) == "" {
		return models.UptimeMonitor{}, fmt.Errorf("target URL or host is required")
	}
	if input.Protocol == "" {
		input.Protocol = "https"
	}
	if input.Method == "" {
		input.Method = "GET"
	}
	if input.IntervalSeconds <= 0 {
		input.IntervalSeconds = 60
	}
	if input.TimeoutSeconds <= 0 {
		input.TimeoutSeconds = 10
	}

	randomID := make([]byte, 16)
	if _, err := rand.Read(randomID); err != nil {
		return models.UptimeMonitor{}, fmt.Errorf("generate monitor id: %w", err)
	}

	now := time.Now().UTC()
	monitor := models.UptimeMonitor{
		ID:              hex.EncodeToString(randomID),
		OrganizationID:  orgID,
		Name:            strings.TrimSpace(input.Name),
		URL:             strings.TrimSpace(input.URL),
		Protocol:        strings.ToLower(input.Protocol),
		Method:          strings.ToUpper(input.Method),
		IntervalSeconds: input.IntervalSeconds,
		TimeoutSeconds:  input.TimeoutSeconds,
		Status:          models.MonitorStatusUp,
		LatencyMs:       0,
		UptimeRatio:     100.0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.repo.CreateMonitor(ctx, &monitor); err != nil {
		return models.UptimeMonitor{}, err
	}
	return monitor, nil
}

func (s *UptimeService) ListOrgMonitors(ctx context.Context, orgID string) ([]models.UptimeMonitor, error) {
	return s.repo.ListOrgMonitors(ctx, orgID)
}

func (s *UptimeService) GetMonitor(ctx context.Context, id string) (*models.UptimeMonitor, error) {
	return s.repo.GetMonitor(ctx, id)
}

func (s *UptimeService) DeleteMonitor(ctx context.Context, id string) error {
	return s.repo.DeleteMonitor(ctx, id)
}

func (s *UptimeService) GetStatusPageByOrg(ctx context.Context, orgID string) (*models.UptimeStatusPage, error) {
	return s.repo.GetStatusPageByOrg(ctx, orgID)
}

type CreateIncidentInput struct {
	Title    string                  `json:"title"`
	Severity models.IncidentSeverity `json:"severity"`
	Message  string                  `json:"message"`
}

func (s *UptimeService) CreateIncident(ctx context.Context, orgID string, input CreateIncidentInput) (models.UptimeIncident, error) {
	if strings.TrimSpace(input.Title) == "" {
		return models.UptimeIncident{}, fmt.Errorf("incident title is required")
	}
	if input.Severity == "" {
		input.Severity = models.IncidentSeverityMinor
	}

	randomID := make([]byte, 16)
	if _, err := rand.Read(randomID); err != nil {
		return models.UptimeIncident{}, fmt.Errorf("generate incident id: %w", err)
	}

	now := time.Now().UTC()
	incident := models.UptimeIncident{
		ID:             hex.EncodeToString(randomID),
		OrganizationID: orgID,
		Title:          strings.TrimSpace(input.Title),
		Status:         models.IncidentStatusInvestigating,
		Severity:       input.Severity,
		StartedAt:      now,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.repo.CreateIncident(ctx, &incident); err != nil {
		return models.UptimeIncident{}, err
	}

	if strings.TrimSpace(input.Message) != "" {
		updateID := make([]byte, 16)
		_, _ = rand.Read(updateID)
		initialUpdate := models.UptimeIncidentUpdate{
			ID:         hex.EncodeToString(updateID),
			IncidentID: incident.ID,
			Status:     models.IncidentStatusInvestigating,
			Message:    strings.TrimSpace(input.Message),
			CreatedAt:  now,
		}
		_ = s.repo.AddIncidentUpdate(ctx, &initialUpdate)
		incident.Updates = append(incident.Updates, initialUpdate)
	}

	return incident, nil
}

func (s *UptimeService) AddIncidentUpdate(ctx context.Context, incidentID string, status models.IncidentStatus, message string) error {
	incident, err := s.repo.GetIncident(ctx, incidentID)
	if err != nil {
		return err
	}

	now := time.Now().UTC()
	updateID := make([]byte, 16)
	if _, err := rand.Read(updateID); err != nil {
		return fmt.Errorf("generate update id: %w", err)
	}

	update := models.UptimeIncidentUpdate{
		ID:         hex.EncodeToString(updateID),
		IncidentID: incidentID,
		Status:     status,
		Message:    strings.TrimSpace(message),
		CreatedAt:  now,
	}

	if err := s.repo.AddIncidentUpdate(ctx, &update); err != nil {
		return err
	}

	incident.Status = status
	incident.UpdatedAt = now
	if status == models.IncidentStatusResolved {
		incident.ResolvedAt = &now
	}
	return s.repo.UpdateIncident(ctx, incident)
}

func (s *UptimeService) ListOrgIncidents(ctx context.Context, orgID string) ([]models.UptimeIncident, error) {
	return s.repo.ListOrgIncidents(ctx, orgID)
}

type PublicStatusData struct {
	Page            models.UptimeStatusPage `json:"page"`
	OverallStatus   string                  `json:"overallStatus"`
	Monitors        []models.UptimeMonitor  `json:"monitors"`
	ActiveIncidents []models.UptimeIncident `json:"activeIncidents"`
	PastIncidents   []models.UptimeIncident `json:"pastIncidents"`
}

func (s *UptimeService) GetPublicStatusData(ctx context.Context, slug string) (PublicStatusData, error) {
	page, err := s.repo.GetStatusPageBySlug(ctx, slug)
	if err != nil {
		return PublicStatusData{}, fmt.Errorf("status page %q not found: %w", slug, err)
	}

	monitors, err := s.repo.ListOrgMonitors(ctx, page.OrganizationID)
	if err != nil {
		return PublicStatusData{}, fmt.Errorf("load monitors: %w", err)
	}

	incidents, err := s.repo.ListOrgIncidents(ctx, page.OrganizationID)
	if err != nil {
		return PublicStatusData{}, fmt.Errorf("load incidents: %w", err)
	}

	var active []models.UptimeIncident
	var past []models.UptimeIncident
	overall := "operational"

	for _, inc := range incidents {
		if inc.Status == models.IncidentStatusResolved {
			past = append(past, inc)
		} else {
			active = append(active, inc)
			if inc.Severity == models.IncidentSeverityCritical {
				overall = "major_outage"
			} else if overall != "major_outage" {
				overall = "degraded_performance"
			}
		}
	}

	for _, m := range monitors {
		if m.Status == models.MonitorStatusDown && overall == "operational" {
			overall = "partial_outage"
		}
	}

	return PublicStatusData{
		Page:            *page,
		OverallStatus:   overall,
		Monitors:        monitors,
		ActiveIncidents: active,
		PastIncidents:   past,
	}, nil
}

func (s *UptimeService) UpsertStatusPage(ctx context.Context, orgID, slug, title, description, customDomain string) (models.UptimeStatusPage, error) {
	page, err := s.repo.GetStatusPageByOrg(ctx, orgID)
	now := time.Now().UTC()
	if err != nil {
		randomID := make([]byte, 16)
		_, _ = rand.Read(randomID)
		newPage := models.UptimeStatusPage{
			ID:             hex.EncodeToString(randomID),
			OrganizationID: orgID,
			Slug:           strings.ToLower(strings.TrimSpace(slug)),
			Title:          strings.TrimSpace(title),
			Description:    strings.TrimSpace(description),
			CustomDomain:   strings.TrimSpace(customDomain),
			Published:      true,
			CreatedAt:      now,
			UpdatedAt:      now,
		}
		if err := s.repo.UpsertStatusPage(ctx, &newPage); err != nil {
			return models.UptimeStatusPage{}, err
		}
		return newPage, nil
	}

	page.Slug = strings.ToLower(strings.TrimSpace(slug))
	page.Title = strings.TrimSpace(title)
	page.Description = strings.TrimSpace(description)
	page.CustomDomain = strings.TrimSpace(customDomain)
	page.UpdatedAt = now

	if err := s.repo.UpsertStatusPage(ctx, page); err != nil {
		return models.UptimeStatusPage{}, err
	}
	return *page, nil
}

func (s *UptimeService) Subscribe(ctx context.Context, slug, email string) error {
	page, err := s.repo.GetStatusPageBySlug(ctx, slug)
	if err != nil {
		return fmt.Errorf("status page %q not found: %w", slug, err)
	}

	randomID := make([]byte, 16)
	_, _ = rand.Read(randomID)
	now := time.Now().UTC()

	subscriber := models.UptimeSubscriber{
		ID:           hex.EncodeToString(randomID),
		StatusPageID: page.ID,
		Email:        strings.ToLower(strings.TrimSpace(email)),
		Confirmed:    true,
		ConfirmedAt:  &now,
		CreatedAt:    now,
	}

	return s.repo.AddSubscriber(ctx, &subscriber)
}
