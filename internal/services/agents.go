package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/auth"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type AgentService struct {
	agents  repositories.AgentRepository
	billing *BillingService
	now     func() time.Time
}

func NewAgentService(agents repositories.AgentRepository) (*AgentService, error) {

	if agents == nil {
		return nil, fmt.Errorf("agent repository is required")
	}

	return &AgentService{agents: agents, now: time.Now}, nil
}

func (s *AgentService) SetBilling(billing *BillingService) {
	s.billing = billing
}

func (s *AgentService) Find(ctx context.Context, id string) (models.Agent, error) {
	agent, err := s.agents.FindByID(ctx, id)

	if err != nil {
		return models.Agent{}, fmt.Errorf("find agent: %w", err)
	}

	return agent, nil
}

func (s *AgentService) Register(ctx context.Context, organizationID, name string) (string, models.Agent, error) {

	if organizationID == "" || strings.TrimSpace(name) == "" {
		return "", models.Agent{}, fmt.Errorf("organization and agent name are required")
	}

	raw, err := auth.NewToken("cda", 32)

	if err != nil {
		return "", models.Agent{}, err
	}

	agent := models.Agent{OrganizationID: organizationID, Name: strings.TrimSpace(name), TokenHash: auth.HashToken(raw), Status: models.AgentStatusPending, Metadata: `{}`}

	if err := s.agents.Create(ctx, &agent); err != nil {
		return "", models.Agent{}, fmt.Errorf("register agent: %w", err)
	}

	return raw, agent, nil
}

func (s *AgentService) Authenticate(ctx context.Context, raw string) (models.Agent, error) {

	if strings.TrimSpace(raw) == "" {
		return models.Agent{}, fmt.Errorf("agent token is required")
	}

	agent, err := s.agents.FindByTokenHash(ctx, auth.HashToken(raw))

	if err != nil {
		return models.Agent{}, fmt.Errorf("find agent: %w", err)
	}

	if agent.RevokedAt != nil || agent.Status == models.AgentStatusRevoked {
		return models.Agent{}, fmt.Errorf("agent is revoked")
	}

	return agent, nil
}

func (s *AgentService) AuthenticateWithPlan(ctx context.Context, raw string) (models.Agent, models.Plan, error) {
	agent, err := s.Authenticate(ctx, raw)

	if err != nil {
		return models.Agent{}, models.Plan{}, err
	}

	if s.billing == nil {
		return models.Agent{}, models.Plan{}, fmt.Errorf("billing service is required")
	}

	plan, _, err := s.billing.Entitlements(ctx, agent.OrganizationID)

	if err != nil {
		return models.Agent{}, models.Plan{}, fmt.Errorf("resolve agent entitlements: %w", err)
	}

	return agent, plan, nil
}

func (s *AgentService) Heartbeat(ctx context.Context, id, version, hostname, platform string) error {

	if err := s.agents.Touch(ctx, id, s.now(), version, hostname, platform); err != nil {
		return fmt.Errorf("heartbeat agent: %w", err)
	}

	return nil
}

func (s *AgentService) Revoke(ctx context.Context, id string) error {

	if err := s.agents.Revoke(ctx, id, s.now()); err != nil {
		return fmt.Errorf("revoke agent: %w", err)
	}

	return nil
}
