package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type QueueHealth struct {
	Queue         string `json:"queue"`
	Depth         int64  `json:"depth"`
	OldestAgeSecs *int64 `json:"oldestAgeSecs,omitempty"`
	DeadLettered  int64  `json:"deadLettered"`
	Healthy       bool   `json:"healthy"`
}

type PipelineHealth struct {
	OrganizationID string        `json:"organizationId"`
	Queues         []QueueHealth `json:"queues"`
	Healthy        bool          `json:"healthy"`
	CheckedAt      time.Time     `json:"checkedAt"`
}

type DeadLetter struct {
	Queue     string    `json:"queue"`
	ID        string    `json:"id"`
	EventType string    `json:"eventType"`
	Attempts  int       `json:"attempts"`
	LastError string    `json:"lastError"`
	UpdatedAt time.Time `json:"updatedAt"`
}

type PipelineHealthService struct {
	webhooks     repositories.WebhookRepository
	exports      repositories.TelemetryExportRepository
	uptime       repositories.UptimeRepository
	now          func() time.Time
	staleAfter   time.Duration
	deadLettered bool
}

func NewPipelineHealthService(webhooks repositories.WebhookRepository, exports repositories.TelemetryExportRepository, uptime repositories.UptimeRepository) (*PipelineHealthService, error) {
	if webhooks == nil || exports == nil || uptime == nil {
		return nil, fmt.Errorf("webhook, export, and uptime repositories are required")
	}
	return &PipelineHealthService{webhooks: webhooks, exports: exports, uptime: uptime, now: time.Now, staleAfter: 15 * time.Minute}, nil
}

func (s *PipelineHealthService) Health(ctx context.Context, orgID string) (PipelineHealth, error) {
	if strings.TrimSpace(orgID) == "" {
		return PipelineHealth{}, fmt.Errorf("organization is required")
	}
	health := PipelineHealth{OrganizationID: orgID, CheckedAt: s.now().UTC(), Healthy: true}
	webhook, err := s.webhookHealth(ctx, orgID)
	if err != nil {
		return PipelineHealth{}, err
	}
	health.Queues = append(health.Queues, webhook)
	export, err := s.exportHealth(ctx, orgID)
	if err != nil {
		return PipelineHealth{}, err
	}
	health.Queues = append(health.Queues, export)
	notify, err := s.notificationHealth(ctx, orgID)
	if err != nil {
		return PipelineHealth{}, err
	}
	health.Queues = append(health.Queues, notify)
	for _, queue := range health.Queues {
		if !queue.Healthy {
			health.Healthy = false
			break
		}
	}
	return health, nil
}

func (s *PipelineHealthService) webhookHealth(ctx context.Context, orgID string) (QueueHealth, error) {
	subscriptions, err := s.webhooks.ListByOrganization(ctx, orgID)
	if err != nil {
		return QueueHealth{}, fmt.Errorf("list webhook subscriptions: %w", err)
	}
	health := QueueHealth{Queue: "webhook-deliveries", Healthy: true}
	var oldest *time.Time
	for _, subscription := range subscriptions {
		deliveries, err := s.webhooks.ListDeliveries(ctx, subscription.ID)
		if err != nil {
			return QueueHealth{}, fmt.Errorf("list webhook deliveries: %w", err)
		}
		for i := range deliveries {
			delivery := deliveries[i]
			switch delivery.Status {
			case models.WebhookDeliveryPending, models.WebhookDeliverySending:
				health.Depth++
				if oldest == nil || delivery.CreatedAt.Before(*oldest) {
					at := delivery.CreatedAt
					oldest = &at
				}
			case models.WebhookDeliveryFailed:
				if delivery.Attempts >= webhookMaxDeliveryAttempts {
					health.DeadLettered++
				} else {
					health.Depth++
				}
				if oldest == nil || delivery.CreatedAt.Before(*oldest) {
					at := delivery.CreatedAt
					oldest = &at
				}
			}
		}
	}
	health.Healthy = health.DeadLettered == 0 && !s.stale(oldest)
	if oldest != nil {
		age := int64(s.now().UTC().Sub(*oldest).Seconds())
		health.OldestAgeSecs = &age
	}
	return health, nil
}

func (s *PipelineHealthService) exportHealth(ctx context.Context, orgID string) (QueueHealth, error) {
	health := QueueHealth{Queue: "telemetry-exports", Healthy: true}
	depth, oldest, dead, err := s.exports.OrgStats(ctx, orgID)
	if err != nil {
		return QueueHealth{}, err
	}
	health.Depth = depth
	health.DeadLettered = dead
	health.Healthy = dead == 0 && !s.stale(oldest)
	if oldest != nil {
		age := int64(s.now().UTC().Sub(*oldest).Seconds())
		health.OldestAgeSecs = &age
	}
	return health, nil
}

func (s *PipelineHealthService) notificationHealth(ctx context.Context, orgID string) (QueueHealth, error) {
	health := QueueHealth{Queue: "subscriber-notifications", Healthy: true}
	page, err := s.uptime.GetStatusPageByOrg(ctx, orgID)
	if err != nil {
		health.Healthy = true
		return health, nil
	}
	counts, err := s.uptime.CountSubscribersByStatus(ctx, page.ID)
	if err != nil {
		return QueueHealth{}, fmt.Errorf("count subscribers: %w", err)
	}
	health.Depth = counts[models.SubscriberStatusPending]
	if !page.SubscriptionsEnabled && health.Depth > 0 {
		health.Healthy = false
	}
	return health, nil
}

func (s *PipelineHealthService) stale(oldest *time.Time) bool {
	if oldest == nil {
		return false
	}
	return s.now().UTC().Sub(*oldest) > s.staleAfter
}

func (s *PipelineHealthService) DeadLetters(ctx context.Context, orgID string, limit int) ([]DeadLetter, error) {
	if strings.TrimSpace(orgID) == "" {
		return nil, fmt.Errorf("organization is required")
	}
	if limit <= 0 || limit > 100 {
		limit = 25
	}
	letters := make([]DeadLetter, 0, limit)
	exports, err := s.exports.ListDeadLettered(ctx, orgID, limit)
	if err != nil {
		return nil, err
	}
	for _, job := range exports {
		letters = append(letters, DeadLetter{Queue: "telemetry-exports", ID: job.ID, EventType: job.Signal, Attempts: job.Attempts, LastError: job.LastError, UpdatedAt: job.UpdatedAt})
		if len(letters) >= limit {
			return letters, nil
		}
	}
	subscriptions, err := s.webhooks.ListByOrganization(ctx, orgID)
	if err != nil {
		return nil, fmt.Errorf("list webhook subscriptions: %w", err)
	}
	for _, subscription := range subscriptions {
		deliveries, err := s.webhooks.ListDeliveries(ctx, subscription.ID)
		if err != nil {
			return nil, fmt.Errorf("list webhook deliveries: %w", err)
		}
		for i := range deliveries {
			delivery := deliveries[i]
			if delivery.Status != models.WebhookDeliveryFailed || delivery.Attempts < webhookMaxDeliveryAttempts {
				continue
			}
			letters = append(letters, DeadLetter{Queue: "webhook-deliveries", ID: delivery.ID, EventType: delivery.EventType, Attempts: delivery.Attempts, LastError: delivery.Error, UpdatedAt: delivery.UpdatedAt})
			if len(letters) >= limit {
				return letters, nil
			}
		}
	}
	return letters, nil
}
