package workers

import (
	"context"
	"testing"

	"outpipe.dev/outpipe/internal/services"
)

type stubPipelineHealth struct {
	health map[string]services.PipelineHealth
	err    error
}

func (s *stubPipelineHealth) Health(_ context.Context, orgID string) (services.PipelineHealth, error) {
	if s.err != nil {
		return services.PipelineHealth{}, s.err
	}
	if health, ok := s.health[orgID]; ok {
		return health, nil
	}
	return services.PipelineHealth{OrganizationID: orgID, Healthy: true, Queues: []services.QueueHealth{{Queue: "webhook-deliveries", Healthy: true}}}, nil
}

type stubOrgLister struct {
	orgs []string
	err  error
}

func (s *stubOrgLister) ListOrganizations(context.Context) ([]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.orgs, nil
}

type stubPipelineAlerts struct {
	calls []string
}

func (s *stubPipelineAlerts) AlertPipelineUnhealthy(_ context.Context, orgID, queue string, _, _ int64) error {
	s.calls = append(s.calls, orgID+"/"+queue)
	return nil
}

func TestPipelineHealthJobAlertsOncePerOutage(t *testing.T) {
	health := &stubPipelineHealth{health: map[string]services.PipelineHealth{
		"org-1": {OrganizationID: "org-1", Healthy: false, Queues: []services.QueueHealth{{Queue: "telemetry-exports", Depth: 10, DeadLettered: 2, Healthy: false}}},
	}}
	orgs := &stubOrgLister{orgs: []string{"org-1", "org-2"}}
	alerts := &stubPipelineAlerts{}
	job, err := NewPipelineHealthJob(health, orgs, alerts)
	if err != nil {
		t.Fatal(err)
	}
	if err := job.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(alerts.calls) != 1 || alerts.calls[0] != "org-1/telemetry-exports" {
		t.Fatalf("alert calls = %v", alerts.calls)
	}
	if err := job.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(alerts.calls) != 1 {
		t.Fatalf("repeat run re-alerted: %v", alerts.calls)
	}
	health.health["org-1"] = services.PipelineHealth{OrganizationID: "org-1", Healthy: true}
	if err := job.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	health.health["org-1"] = services.PipelineHealth{OrganizationID: "org-1", Healthy: false, Queues: []services.QueueHealth{{Queue: "telemetry-exports", Depth: 11, DeadLettered: 2, Healthy: false}}}
	if err := job.Run(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(alerts.calls) != 2 {
		t.Fatalf("recovery did not reset alert latch: %v", alerts.calls)
	}
	if _, err := NewPipelineHealthJob(nil, orgs, alerts); err == nil {
		t.Fatal("job accepted nil checker")
	}
}
