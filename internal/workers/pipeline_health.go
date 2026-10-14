package workers

import (
	"context"
	"fmt"

	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

type PipelineHealthChecker interface {
	Health(context.Context, string) (services.PipelineHealth, error)
}

type PipelineOrganizationLister interface {
	ListOrganizations(context.Context) ([]string, error)
}

type PipelineAlertSender interface {
	AlertPipelineUnhealthy(ctx context.Context, organizationID, queue string, depth, deadLettered int64) error
}

type PipelineHealthJob struct {
	health  PipelineHealthChecker
	orgs    PipelineOrganizationLister
	alerts  PipelineAlertSender
	checked map[string]bool
}

func NewPipelineHealthJob(health PipelineHealthChecker, orgs PipelineOrganizationLister, alerts PipelineAlertSender) (*PipelineHealthJob, error) {
	if health == nil || orgs == nil {
		return nil, fmt.Errorf("pipeline health checker and organization lister are required")
	}
	return &PipelineHealthJob{health: health, orgs: orgs, alerts: alerts, checked: make(map[string]bool)}, nil
}

func (j *PipelineHealthJob) Name() string {
	return "pipeline-health"
}

func (j *PipelineHealthJob) Run(ctx context.Context) error {
	orgIDs, err := j.orgs.ListOrganizations(ctx)
	if err != nil {
		return fmt.Errorf("list organizations for pipeline health: %w", err)
	}
	for _, orgID := range orgIDs {
		health, err := j.health.Health(ctx, orgID)
		if err != nil {
			continue
		}
		if health.Healthy {
			delete(j.checked, orgID)
			continue
		}
		if j.checked[orgID] {
			continue
		}
		j.checked[orgID] = true
		if j.alerts == nil {
			continue
		}
		for _, queue := range health.Queues {
			if queue.Healthy {
				continue
			}
			_ = j.alerts.AlertPipelineUnhealthy(ctx, orgID, queue.Queue, queue.Depth, queue.DeadLettered)
		}
	}
	return nil
}

type OrganizationListAdapter struct {
	organizations repositories.OrganizationRepository
}

func NewOrganizationListAdapter(organizations repositories.OrganizationRepository) (*OrganizationListAdapter, error) {
	if organizations == nil {
		return nil, fmt.Errorf("organization repository is required")
	}
	return &OrganizationListAdapter{organizations: organizations}, nil
}

func (a *OrganizationListAdapter) ListOrganizations(ctx context.Context) ([]string, error) {
	orgs, err := a.organizations.List(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(orgs))
	for _, org := range orgs {
		ids = append(ids, org.ID)
	}
	return ids, nil
}
