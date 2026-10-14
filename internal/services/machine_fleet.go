package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
)

type MachineFleetService struct {
	fleets     repositories.MachineFleetRepository
	secrets    *SecretService
	tunnels    *TunnelService
	now        func() time.Time
	maxTunnels int
}

func NewMachineFleetService(fleets repositories.MachineFleetRepository, secrets *SecretService, tunnels *TunnelService) (*MachineFleetService, error) {
	if fleets == nil {
		return nil, fmt.Errorf("machine fleet repository is required")
	}
	if secrets == nil {
		return nil, fmt.Errorf("secret service is required")
	}
	if tunnels == nil {
		return nil, fmt.Errorf("tunnel service is required")
	}
	return &MachineFleetService{fleets: fleets, secrets: secrets, tunnels: tunnels, now: time.Now, maxTunnels: 100}, nil
}

func (s *MachineFleetService) CreateFleet(ctx context.Context, orgID, name, description string) (models.MachineFleet, error) {
	if strings.TrimSpace(orgID) == "" || strings.TrimSpace(name) == "" {
		return models.MachineFleet{}, fmt.Errorf("organization and fleet name are required")
	}
	if len(name) > 120 {
		return models.MachineFleet{}, fmt.Errorf("fleet name is too long")
	}
	fleet := models.MachineFleet{OrganizationID: orgID, Name: strings.TrimSpace(name), Description: strings.TrimSpace(description)}
	if err := s.fleets.CreateFleet(ctx, &fleet); err != nil {
		return models.MachineFleet{}, fmt.Errorf("create machine fleet: %w", err)
	}
	return fleet, nil
}

func (s *MachineFleetService) ListFleets(ctx context.Context, orgID string) ([]models.MachineFleet, error) {
	return s.fleets.ListFleets(ctx, orgID)
}

func (s *MachineFleetService) DeleteFleet(ctx context.Context, orgID, fleetID string) error {
	if _, err := s.fleets.FindFleet(ctx, orgID, fleetID); err != nil {
		return fmt.Errorf("find machine fleet: %w", err)
	}
	return s.fleets.DeleteFleet(ctx, orgID, fleetID)
}

type FleetStatus struct {
	Fleet          models.MachineFleet         `json:"fleet"`
	Tokens         []models.SecretMachineToken `json:"tokens"`
	DesiredTunnels []models.MachineFleetTunnel `json:"desiredTunnels"`
	LiveTunnels    []models.Tunnel             `json:"liveTunnels"`
	Drift          []FleetDrift                `json:"drift"`
}

type FleetDrift struct {
	TunnelID     string `json:"tunnelId"`
	Name         string `json:"name"`
	Want         string `json:"want"`
	Got          string `json:"got"`
	MachineOwned bool   `json:"machineOwned"`
	CredentialOK bool   `json:"credentialOk"`
}

func (s *MachineFleetService) Status(ctx context.Context, orgID, fleetID string) (FleetStatus, error) {
	var status FleetStatus
	fleet, err := s.fleets.FindFleet(ctx, orgID, fleetID)
	if err != nil {
		return status, fmt.Errorf("find machine fleet: %w", err)
	}
	status.Fleet = fleet
	attachments, err := s.fleets.ListFleetTokens(ctx, orgID, fleetID)
	if err != nil {
		return status, err
	}
	allTokens, err := s.secrets.ListMachineTokens(ctx, orgID)
	if err != nil {
		return status, err
	}
	byID := make(map[string]models.SecretMachineToken, len(allTokens))
	for _, token := range allTokens {
		byID[token.ID] = token
	}
	for _, attachment := range attachments {
		if token, ok := byID[attachment.MachineTokenID]; ok {
			status.Tokens = append(status.Tokens, token)
		}
	}
	desired, err := s.fleets.ListFleetTunnels(ctx, orgID, fleetID)
	if err != nil {
		return status, err
	}
	status.DesiredTunnels = desired
	live, err := s.tunnels.List(ctx, orgID)
	if err != nil {
		return status, err
	}
	byTunnel := make(map[string]models.Tunnel, len(live))
	for _, tunnel := range live {
		byTunnel[tunnel.ID] = tunnel
		if tunnel.MachineOwned {
			status.LiveTunnels = append(status.LiveTunnels, tunnel)
		}
	}
	for _, want := range desired {
		got, ok := byTunnel[want.TunnelID]
		drift := FleetDrift{TunnelID: want.TunnelID, Name: want.Name, Want: want.DesiredStatus}
		if !ok {
			drift.Got = "missing"
			status.Drift = append(status.Drift, drift)
			continue
		}
		drift.Got = string(got.Status)
		drift.MachineOwned = got.MachineOwned
		drift.CredentialOK = s.tunnelCredentialOK(ctx, orgID, &got, byID)
		if drift.Got != drift.Want && !(drift.Want == string(models.TunnelStatusActive) && drift.Got == string(models.TunnelStatusCreated)) {
			status.Drift = append(status.Drift, drift)
			continue
		}
		if !drift.MachineOwned || !drift.CredentialOK {
			status.Drift = append(status.Drift, drift)
		}
	}
	return status, nil
}

func (s *MachineFleetService) tunnelCredentialOK(ctx context.Context, orgID string, tunnel *models.Tunnel, byID map[string]models.SecretMachineToken) bool {
	if tunnel == nil || tunnel.MachineTokenID == nil {
		return false
	}
	token, ok := byID[*tunnel.MachineTokenID]
	if !ok {
		tokens, err := s.secrets.ListMachineTokens(ctx, orgID)
		if err != nil {
			return false
		}
		for _, candidate := range tokens {
			if candidate.ID == *tunnel.MachineTokenID {
				token, ok = candidate, true
				break
			}
		}
		if !ok {
			return false
		}
	}
	return token.RevokedAt == nil && token.ExpiresAt != nil && token.ExpiresAt.After(s.now())
}

type FleetTunnelSpec struct {
	Name           string `json:"name"`
	Protocol       string `json:"protocol"`
	TargetHost     string `json:"targetHost"`
	TargetPort     int    `json:"targetPort"`
	PublicHostname string `json:"publicHostname"`
	DesiredStatus  string `json:"desiredStatus"`
}

func (s *MachineFleetService) BulkCreateTunnels(ctx context.Context, orgID, fleetID string, userID string, specs []FleetTunnelSpec) ([]models.Tunnel, error) {
	if _, err := s.fleets.FindFleet(ctx, orgID, fleetID); err != nil {
		return nil, fmt.Errorf("find machine fleet: %w", err)
	}
	if len(specs) == 0 || len(specs) > s.maxTunnels {
		return nil, fmt.Errorf("spec count must be between 1 and %d", s.maxTunnels)
	}
	attachments, err := s.fleets.ListFleetTokens(ctx, orgID, fleetID)
	if err != nil {
		return nil, err
	}
	if len(attachments) == 0 {
		return nil, fmt.Errorf("fleet has no machine tokens; attach one first")
	}
	tokenID := attachments[0].MachineTokenID
	created := make([]models.Tunnel, 0, len(specs))
	for _, spec := range specs {
		protocol := models.TunnelProtocol(strings.ToLower(strings.TrimSpace(spec.Protocol)))
		if protocol == "" {
			protocol = models.TunnelProtocolHTTP
		}
		hostname := strings.TrimSpace(spec.PublicHostname)
		if hostname == "" {
			hostname = strings.ToLower(strings.TrimSpace(spec.Name)) + "-" + fleetID[:8] + ".fleet"
		}
		tunnel, err := s.tunnels.CreateWithMachineOwner(ctx, orgID, spec.Name, protocol, spec.TargetHost, spec.TargetPort, hostname, "", "{}", &tokenID)
		if err != nil {
			return created, fmt.Errorf("create fleet tunnel %q: %w", spec.Name, err)
		}
		desired := models.MachineFleetTunnel{FleetID: fleetID, TunnelID: tunnel.ID, DesiredStatus: "active", TargetHost: spec.TargetHost, TargetPort: spec.TargetPort, Protocol: string(protocol), Name: spec.Name, PublicHostname: tunnel.PublicHostname}
		if strings.TrimSpace(spec.DesiredStatus) != "" {
			desired.DesiredStatus = strings.TrimSpace(spec.DesiredStatus)
		}
		if err := s.fleets.UpsertFleetTunnel(ctx, &desired); err != nil {
			return created, fmt.Errorf("record fleet tunnel %q: %w", spec.Name, err)
		}
		created = append(created, tunnel)
	}
	return created, nil
}

func (s *MachineFleetService) AttachToken(ctx context.Context, orgID, fleetID, machineTokenID, label string) (models.MachineFleetToken, error) {
	if _, err := s.fleets.FindFleet(ctx, orgID, fleetID); err != nil {
		return models.MachineFleetToken{}, fmt.Errorf("find machine fleet: %w", err)
	}
	tokens, err := s.secrets.ListMachineTokens(ctx, orgID)
	if err != nil {
		return models.MachineFleetToken{}, err
	}
	var found *models.SecretMachineToken
	for i := range tokens {
		if tokens[i].ID == machineTokenID {
			found = &tokens[i]
			break
		}
	}
	if found == nil {
		return models.MachineFleetToken{}, fmt.Errorf("machine token not found in organization")
	}
	if found.RevokedAt != nil {
		return models.MachineFleetToken{}, fmt.Errorf("machine token is revoked")
	}
	attachment := models.MachineFleetToken{OrganizationID: orgID, FleetID: fleetID, Name: found.Name, MachineTokenID: found.ID, Label: strings.TrimSpace(label)}
	if err := s.fleets.AttachToken(ctx, &attachment); err != nil {
		return models.MachineFleetToken{}, fmt.Errorf("attach fleet token: %w", err)
	}
	return attachment, nil
}

func (s *MachineFleetService) RotateFleetTokens(ctx context.Context, orgID, fleetID, userID string) ([]CreatedMachineTokenDTO, error) {
	if _, err := s.fleets.FindFleet(ctx, orgID, fleetID); err != nil {
		return nil, fmt.Errorf("find machine fleet: %w", err)
	}
	attachments, err := s.fleets.ListFleetTokens(ctx, orgID, fleetID)
	if err != nil {
		return nil, err
	}
	if len(attachments) == 0 {
		return nil, fmt.Errorf("fleet has no machine tokens to rotate")
	}
	rotated := make([]CreatedMachineTokenDTO, 0, len(attachments))
	for _, attachment := range attachments {
		created, err := s.secrets.CreateMachineToken(ctx, orgID, "", "", attachment.Name+"-rotated", []string{"tunnels:write"}, userID)
		if err != nil {
			return rotated, fmt.Errorf("create rotated token: %w", err)
		}
		if _, err := s.tunnels.RepointMachineToken(ctx, orgID, attachment.MachineTokenID, created.Token.ID); err != nil {
			return rotated, fmt.Errorf("repoint fleet tunnels: %w", err)
		}
		if err := s.secrets.RevokeMachineToken(ctx, orgID, attachment.MachineTokenID); err != nil {
			return rotated, fmt.Errorf("revoke old token: %w", err)
		}
		if err := s.fleets.DetachToken(ctx, orgID, fleetID, attachment.MachineTokenID); err != nil {
			return rotated, fmt.Errorf("detach old token: %w", err)
		}
		if _, err := s.AttachToken(ctx, orgID, fleetID, created.Token.ID, attachment.Label); err != nil {
			return rotated, fmt.Errorf("attach rotated token: %w", err)
		}
		rotated = append(rotated, created)
	}
	return rotated, nil
}

func (s *MachineFleetService) RevokeFleet(ctx context.Context, orgID, fleetID string) error {
	status, err := s.Status(ctx, orgID, fleetID)
	if err != nil {
		return err
	}
	for _, token := range status.Tokens {
		if err := s.secrets.RevokeMachineToken(ctx, orgID, token.ID); err != nil {
			return fmt.Errorf("revoke fleet token %s: %w", token.ID, err)
		}
	}
	for _, tunnel := range status.LiveTunnels {
		if err := s.tunnels.Revoke(ctx, tunnel.ID); err != nil {
			return fmt.Errorf("revoke fleet tunnel %s: %w", tunnel.ID, err)
		}
	}
	return nil
}

type FleetReconcileResult struct {
	Reactivated int      `json:"reactivated"`
	Recreated   int      `json:"recreated"`
	Skipped     int      `json:"skipped"`
	Errors      []string `json:"errors,omitempty"`
}

func (s *MachineFleetService) Reconcile(ctx context.Context, orgID, fleetID string) (FleetReconcileResult, error) {
	var result FleetReconcileResult
	status, err := s.Status(ctx, orgID, fleetID)
	if err != nil {
		return result, err
	}
	if len(status.Drift) == 0 {
		return result, nil
	}
	live := make(map[string]models.Tunnel, len(status.LiveTunnels))
	for _, tunnel := range status.LiveTunnels {
		live[tunnel.ID] = tunnel
	}
	all, err := s.tunnels.List(ctx, orgID)
	if err != nil {
		return result, err
	}
	byID := make(map[string]models.Tunnel, len(all))
	for _, tunnel := range all {
		byID[tunnel.ID] = tunnel
	}
	attachments, err := s.fleets.ListFleetTokens(ctx, orgID, fleetID)
	if err != nil {
		return result, err
	}
	var tokenID string
	if len(attachments) > 0 {
		tokenID = attachments[0].MachineTokenID
	}
	for _, drift := range status.Drift {
		tunnel, ok := byID[drift.TunnelID]
		if !ok {
			if tokenID == "" {
				result.Skipped++
				continue
			}
			var spec *models.MachineFleetTunnel
			for i := range status.DesiredTunnels {
				if status.DesiredTunnels[i].TunnelID == drift.TunnelID {
					spec = &status.DesiredTunnels[i]
					break
				}
			}
			if spec == nil {
				result.Skipped++
				continue
			}
			hostname := strings.TrimSpace(spec.PublicHostname)
			if hostname == "" {
				hostname = strings.ToLower(strings.TrimSpace(spec.Name)) + "-" + fleetID[:8] + ".fleet"
			}
			recreated, err := s.tunnels.CreateWithMachineOwner(ctx, orgID, spec.Name, models.TunnelProtocol(spec.Protocol), spec.TargetHost, spec.TargetPort, hostname, "", "{}", &tokenID)
			if err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("recreate %s: %v", drift.Name, err))
				continue
			}
			if err := s.fleets.RemoveFleetTunnel(ctx, fleetID, drift.TunnelID); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("remove stale desired %s: %v", drift.Name, err))
				continue
			}
			desired := models.MachineFleetTunnel{FleetID: fleetID, TunnelID: recreated.ID, DesiredStatus: spec.DesiredStatus, TargetHost: spec.TargetHost, TargetPort: spec.TargetPort, Protocol: spec.Protocol, Name: spec.Name, PublicHostname: recreated.PublicHostname}
			if err := s.fleets.UpsertFleetTunnel(ctx, &desired); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("record recreated %s: %v", drift.Name, err))
				continue
			}
			result.Recreated++
			continue
		}
		if tunnel.Status == models.TunnelStatusRevoked || tunnel.RevokedAt != nil {
			result.Skipped++
			continue
		}
		want := models.TunnelStatusActive
		if drift.Want == string(models.TunnelStatusDisconnected) {
			want = models.TunnelStatusDisconnected
		}
		if tunnel.Status == models.TunnelStatusCreated && want == models.TunnelStatusActive {
			result.Skipped++
			continue
		}
		if tunnel.Status != want {
			if err := s.tunnels.SetStatus(ctx, tunnel.ID, want); err != nil {
				result.Errors = append(result.Errors, fmt.Sprintf("set status %s: %v", drift.Name, err))
				continue
			}
			result.Reactivated++
			continue
		}
		result.Skipped++
	}
	return result, nil
}

func (s *MachineFleetService) ReconcileAll(ctx context.Context) (map[string]FleetReconcileResult, error) {
	desired, err := s.fleets.ListAllFleetTunnels(ctx)
	if err != nil {
		return nil, err
	}
	seen := make(map[string]bool, len(desired))
	results := make(map[string]FleetReconcileResult)
	for _, item := range desired {
		if seen[item.FleetID] {
			continue
		}
		seen[item.FleetID] = true
		fleet, err := s.fleets.FindFleetByID(ctx, item.FleetID)
		if err != nil {
			continue
		}
		result, err := s.Reconcile(ctx, fleet.OrganizationID, fleet.ID)
		if err != nil {
			continue
		}
		if result.Reactivated > 0 || result.Recreated > 0 || len(result.Errors) > 0 {
			results[fleet.ID] = result
		}
	}
	return results, nil
}
