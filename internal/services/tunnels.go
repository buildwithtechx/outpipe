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

type TunnelService struct {
	tunnels         repositories.TunnelRepository
	capturePolicies repositories.CapturePolicyRepository
	billing         *BillingService
	allocator       *HostnameAllocator
	webhooks        *WebhookService
	now             func() time.Time
}

func NewTunnelService(tunnels repositories.TunnelRepository) (*TunnelService, error) {
	if tunnels == nil {
		return nil, fmt.Errorf("tunnel repository is required")
	}
	return &TunnelService{tunnels: tunnels, now: time.Now}, nil
}

func (s *TunnelService) SetBilling(billing *BillingService) {
	s.billing = billing
}

func (s *TunnelService) SetHostnameAllocator(allocator *HostnameAllocator) {
	s.allocator = allocator
}

func (s *TunnelService) SetWebhooks(webhooks *WebhookService) {
	s.webhooks = webhooks
}

func (s *TunnelService) SetCapturePolicies(policies repositories.CapturePolicyRepository) {
	s.capturePolicies = policies
}

func (s *TunnelService) Find(ctx context.Context, id string) (models.Tunnel, error) {
	if id == "" {
		return models.Tunnel{}, fmt.Errorf("tunnel id is required")
	}
	tunnel, err := s.tunnels.FindByID(ctx, id)
	if err != nil {
		return models.Tunnel{}, fmt.Errorf("find tunnel: %w", err)
	}
	return tunnel, nil
}

func (s *TunnelService) Policy(ctx context.Context, id string) (models.Tunnel, error) {
	return s.Find(ctx, id)
}

func (s *TunnelService) UpdateConfiguration(ctx context.Context, id, accessPolicy string, expiresAt *time.Time) (models.Tunnel, error) {
	if id == "" {
		return models.Tunnel{}, fmt.Errorf("tunnel id is required")
	}
	trimmedPolicy := strings.TrimSpace(accessPolicy)
	if trimmedPolicy == "" || !json.Valid([]byte(trimmedPolicy)) || trimmedPolicy[0] != '{' {
		return models.Tunnel{}, fmt.Errorf("access policy must be a JSON object")
	}
	if len(trimmedPolicy) > 16*1024 {
		return models.Tunnel{}, fmt.Errorf("access policy is too large")
	}
	if expiresAt != nil && !expiresAt.After(s.now()) {
		return models.Tunnel{}, fmt.Errorf("expiration must be in the future")
	}
	tunnel, err := s.Find(ctx, id)
	if err != nil {
		return models.Tunnel{}, err
	}
	tunnel.AccessPolicy = trimmedPolicy
	tunnel.ExpiresAt = expiresAt
	if err := s.tunnels.Update(ctx, &tunnel); err != nil {
		return models.Tunnel{}, fmt.Errorf("update tunnel configuration: %w", err)
	}
	return tunnel, nil
}

func (s *TunnelService) List(ctx context.Context, organizationID string) ([]models.Tunnel, error) {
	if organizationID == "" {
		return nil, fmt.Errorf("organization id is required")
	}
	tunnels, err := s.tunnels.FindByOrganization(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("list tunnels: %w", err)
	}
	return tunnels, nil
}

func (s *TunnelService) Create(ctx context.Context, organizationID, name string, protocol models.TunnelProtocol, targetHost string, targetPort int, publicHostname, password, metadata string) (models.Tunnel, error) {
	return s.createOwned(ctx, organizationID, name, protocol, targetHost, targetPort, publicHostname, password, metadata, nil)
}

func (s *TunnelService) createOwned(ctx context.Context, organizationID, name string, protocol models.TunnelProtocol, targetHost string, targetPort int, publicHostname, password, metadata string, machineTokenID *string) (models.Tunnel, error) {
	if organizationID == "" || strings.TrimSpace(name) == "" || strings.TrimSpace(targetHost) == "" || !validTunnelProtocol(protocol) || targetPort < 1 || targetPort > 65535 {
		return models.Tunnel{}, fmt.Errorf("invalid tunnel configuration")
	}

	passwordHash, err := hashTunnelPassword(password)
	if err != nil {
		return models.Tunnel{}, err
	}

	metadata = strings.TrimSpace(metadata)
	if metadata == "" || !json.Valid([]byte(metadata)) {
		metadata = `{}`
	}

	if s.allocator != nil {
		allocated, err := s.allocator.Allocate(ctx, strings.TrimSpace(publicHostname))
		if err != nil {
			return models.Tunnel{}, fmt.Errorf("allocate public hostname: %w", err)
		}
		publicHostname = allocated
	}

	if strings.TrimSpace(publicHostname) == "" {
		return models.Tunnel{}, fmt.Errorf("public hostname is required")
	}

	tunnel := models.Tunnel{OrganizationID: organizationID, Name: strings.TrimSpace(name), Protocol: protocol, Status: models.TunnelStatusCreated, TargetHost: strings.TrimSpace(targetHost), TargetPort: targetPort, PublicHostname: strings.ToLower(strings.TrimSpace(publicHostname)), AccessPolicy: `{}`, Metadata: metadata, PasswordHash: passwordHash}
	if machineTokenID != nil && *machineTokenID != "" {
		tunnel.MachineOwned = true
		tunnel.MachineTokenID = machineTokenID
	}

	if s.billing != nil {
		plan, _, err := s.billing.Entitlements(ctx, organizationID)
		if err != nil {
			return models.Tunnel{}, fmt.Errorf("check tunnel entitlement: %w", err)
		}
		count, err := s.tunnels.CountByOrganization(ctx, organizationID)
		if err != nil {
			return models.Tunnel{}, fmt.Errorf("count organization tunnels: %w", err)
		}
		if plan.MaxTunnels > 0 && count >= int64(plan.MaxTunnels) {
			return models.Tunnel{}, fmt.Errorf("tunnel plan limit reached")
		}
	}

	if err := s.tunnels.Create(ctx, &tunnel); err != nil {
		return models.Tunnel{}, fmt.Errorf("create tunnel: %w", err)
	}
	return tunnel, nil
}

func (s *TunnelService) CreateWithMachineOwner(ctx context.Context, organizationID, name string, protocol models.TunnelProtocol, targetHost string, targetPort int, publicHostname, password, metadata string, machineTokenID *string) (models.Tunnel, error) {
	return s.createOwned(ctx, organizationID, name, protocol, targetHost, targetPort, publicHostname, password, metadata, machineTokenID)
}

func (s *TunnelService) SetStatus(ctx context.Context, id string, status models.TunnelStatus) error {
	if !validTunnelStatus(status) {
		return fmt.Errorf("invalid tunnel status")
	}
	deliveries, err := s.statusDeliveries(ctx, id, status)
	if err != nil {
		return err
	}
	if err := s.tunnels.TransitionWithDeliveries(ctx, id, status, nil, deliveries); err != nil {
		return fmt.Errorf("set tunnel status: %w", err)
	}
	if s.webhooks != nil {
		s.webhooks.afterEnqueue(ctx, len(deliveries))
	}
	return nil
}

func (s *TunnelService) Touch(ctx context.Context, id string) error {
	if err := s.tunnels.Touch(ctx, id, s.now()); err != nil {
		return fmt.Errorf("touch tunnel: %w", err)
	}
	return nil
}

func (s *TunnelService) RepointMachineToken(ctx context.Context, orgID, oldTokenID, newTokenID string) (int64, error) {
	if orgID == "" || oldTokenID == "" || newTokenID == "" {
		return 0, fmt.Errorf("organization and token ids are required")
	}
	count, err := s.tunnels.RepointMachineToken(ctx, orgID, oldTokenID, newTokenID)
	if err != nil {
		return 0, fmt.Errorf("repoint machine tunnels: %w", err)
	}
	return count, nil
}

func (s *TunnelService) Revoke(ctx context.Context, id string) error {
	tunnel, err := s.tunnels.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("find tunnel for revocation: %w", err)
	}

	var deliveries []models.WebhookDelivery
	if s.webhooks != nil {
		tunnel.Status = models.TunnelStatusRevoked
		deliveries, err = s.webhooks.prepareDeliveries(ctx, tunnel.OrganizationID, models.WebhookEventTunnelRevoked, map[string]any{"tunnel": tunnel})
		if err != nil {
			return fmt.Errorf("prepare revocation webhook deliveries: %w", err)
		}
	}

	now := s.now()
	if err := s.tunnels.TransitionWithDeliveries(ctx, id, models.TunnelStatusRevoked, &now, deliveries); err != nil {
		return fmt.Errorf("revoke tunnel: %w", err)
	}

	if s.webhooks != nil {
		s.webhooks.afterEnqueue(ctx, len(deliveries))
	}
	return nil
}
