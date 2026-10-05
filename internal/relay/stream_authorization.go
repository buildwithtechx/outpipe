package relay

import (
	"context"
	"fmt"
	"time"
)

func (h *Handler) authorizeStream(ctx context.Context, identity AgentIdentity, tunnelID string, owned map[string]string) error {
	session, ok := h.sessions.Get(tunnelID)
	if !ok || session.OrganizationID != identity.OrganizationID || owned[tunnelID] != session.ID {
		return fmt.Errorf("stream is not owned by this connection")
	}
	return h.authorizeMachine(ctx, identity, tunnelID)
}

func (h *Handler) authorizeMachine(ctx context.Context, identity AgentIdentity, tunnelID string) error {
	if identity.ExpiresAt != 0 && time.Now().Unix() >= identity.ExpiresAt {
		return fmt.Errorf("relay credential expired")
	}
	if identity.MachineTokenID == "" {
		return nil
	}
	if identity.TunnelID != tunnelID || h.managedTunnels == nil {
		return fmt.Errorf("machine credential is not authorized for this tunnel")
	}
	policy, err := h.managedTunnels.Resolve(ctx, tunnelID)
	if err != nil {
		return fmt.Errorf("authorize machine tunnel: %w", err)
	}
	if policy.Status == "revoked" || policy.OrganizationID != identity.OrganizationID || policy.MachineTokenID == nil || *policy.MachineTokenID != identity.MachineTokenID {
		return fmt.Errorf("machine tunnel is unavailable")
	}
	return nil
}
