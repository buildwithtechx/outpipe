package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

func (s *UptimeService) GetStatusPageByOrg(ctx context.Context, orgID string) (*models.UptimeStatusPage, error) {
	return s.repo.GetStatusPageByOrg(ctx, orgID)
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
	if !page.Published {
		return PublicStatusData{}, repositories.ErrNotFound
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

func (s *UptimeService) UpsertStatusPage(ctx context.Context, orgID, slug, title, description, customDomain string, published bool) (models.UptimeStatusPage, error) {
	page, err := s.repo.GetStatusPageByOrg(ctx, orgID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && !errors.Is(err, repositories.ErrNotFound) {
		return models.UptimeStatusPage{}, fmt.Errorf("load status page: %w", err)
	}
	now := time.Now().UTC()
	if err != nil {
		randomID := make([]byte, 16)
		if _, err := rand.Read(randomID); err != nil {
			return models.UptimeStatusPage{}, fmt.Errorf("generate status page id: %w", err)
		}
		newPage := models.UptimeStatusPage{
			ID:                   hex.EncodeToString(randomID),
			OrganizationID:       orgID,
			Slug:                 strings.ToLower(strings.TrimSpace(slug)),
			Title:                strings.TrimSpace(title),
			Description:          strings.TrimSpace(description),
			CustomDomain:         strings.TrimSpace(customDomain),
			Published:            published,
			SubscriptionsEnabled: true,
			CreatedAt:            now,
			UpdatedAt:            now,
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
	page.Published = published
	page.UpdatedAt = now

	if err := s.repo.UpsertStatusPage(ctx, page); err != nil {
		return models.UptimeStatusPage{}, err
	}
	return *page, nil
}
