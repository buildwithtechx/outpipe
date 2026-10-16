package services

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"outpipe.dev/outpipe/internal/models"
)

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
		tokens, _ := s.secrets.ListMachineTokens(ctx, orgID)
		activeTokens := make(map[string]bool, len(tokens))
		for _, tok := range tokens {
			if tok.RevokedAt == nil && (tok.ExpiresAt == nil || tok.ExpiresAt.After(time.Now())) && strings.Contains(tok.Scopes, "tunnels:write") {
				activeTokens[tok.ID] = true
			}
		}
		for _, att := range attachments {
			if activeTokens[att.MachineTokenID] {
				tokenID = att.MachineTokenID
				break
			}
		}
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
	var errs []error
	for _, item := range desired {
		if seen[item.FleetID] {
			continue
		}
		seen[item.FleetID] = true
		fleet, err := s.fleets.FindFleetByID(ctx, item.FleetID)
		if err != nil {
			errs = append(errs, fmt.Errorf("find fleet %s: %w", item.FleetID, err))
			continue
		}
		result, err := s.Reconcile(ctx, fleet.OrganizationID, fleet.ID)
		if err != nil {
			errs = append(errs, fmt.Errorf("reconcile fleet %s: %w", fleet.ID, err))
			continue
		}
		if result.Reactivated > 0 || result.Recreated > 0 || len(result.Errors) > 0 {
			results[fleet.ID] = result
		}
	}
	if len(errs) > 0 {
		return results, errors.Join(errs...)
	}
	return results, nil
}
