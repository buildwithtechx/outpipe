package relay

import (
	"context"
	"fmt"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/security"
	"outpipe.dev/outpipe/pkg/protocol"
	"strings"
)

func (h *Handler) resolveManagedPolicy(ctx context.Context, identity AgentIdentity, open *protocol.OpenTunnel) (string, error) {
	if identity.MachineTokenID != "" && (open.TunnelID == "" || h.managedTunnels == nil) {
		return "", fmt.Errorf("machine tunnels require managed authorization")
	}

	if open.TunnelID != "" && h.managedTunnels != nil {
		policy, err := h.managedTunnels.Resolve(ctx, open.TunnelID)

		if err != nil {
			return "", fmt.Errorf("resolve managed tunnel: %w", err)
		}

		if policy.OrganizationID != identity.OrganizationID || policy.Status == string(models.TunnelStatusRevoked) {
			return "", fmt.Errorf("managed tunnel is not available")
		}
		if identity.MachineTokenID != "" && (policy.MachineTokenID == nil || *policy.MachineTokenID != identity.MachineTokenID) {
			return "", fmt.Errorf("machine does not own tunnel")
		}
		if policy.MachineTokenID != nil && identity.MachineTokenID != *policy.MachineTokenID {
			return "", fmt.Errorf("machine credential is required")
		}
		if identity.MachineTokenID != "" && (open.LocalPort != policy.TargetPort || open.Protocol != policy.Protocol) {
			return "", fmt.Errorf("machine tunnel target does not match policy")
		}
		if identity.MachineTokenID != "" {
			hostname := open.CustomDomain
			if open.Subdomain != "" {
				hostname = open.Subdomain + "." + h.publicDomain
			}
			if hostname != "" && hostname != policy.PublicHostname {
				return "", fmt.Errorf("machine tunnel hostname does not match policy")
			}
		}

		if open.Subdomain == "" && open.CustomDomain == "" {

			if suffix := "." + h.publicDomain; strings.HasSuffix(policy.PublicHostname, suffix) {
				open.Subdomain = strings.TrimSuffix(policy.PublicHostname, suffix)
			} else {
				open.CustomDomain = policy.PublicHostname
			}
		}

		if policy.PasswordProtected {

			if verifier, ok := h.managedTunnels.(ManagedTunnelPasswordVerifier); ok {
				valid, err := verifier.VerifyPassword(ctx, open.TunnelID, open.Password)

				if err != nil {
					return "", fmt.Errorf("verify managed tunnel password: %w", err)
				}

				if !valid {
					return "", fmt.Errorf("invalid managed tunnel password")
				}

				return hashRelayPassword(open.Password)
			}

			if policy.PasswordHash == "" || !security.VerifyPassword(open.Password, policy.PasswordHash) {
				return "", fmt.Errorf("invalid managed tunnel password")
			}
		}

		return policy.PasswordHash, nil
	}

	return hashRelayPassword(open.Password)
}
