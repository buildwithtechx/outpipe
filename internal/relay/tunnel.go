package relay

import (
	"context"
	"fmt"
	"strings"

	"github.com/gofiber/contrib/websocket"
	"github.com/google/uuid"
	"outpipe.dev/outpipe/internal/engine"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/security"
	"outpipe.dev/outpipe/pkg/protocol"
)

func (h *Handler) openTunnel(ctx context.Context, connection *websocket.Conn, identity AgentIdentity, message protocol.Envelope, owned map[string]string, state *connectionState) error {

	if !state.negotiated || !state.authenticated {
		return fmt.Errorf("protocol authentication and negotiation are required")
	}

	open, err := decodeOpenTunnel(message)

	if err != nil {
		return err
	}

	tunnelID := open.TunnelID

	if tunnelID == "" {
		tunnelID = uuid.NewString()
	}

	previous, exists := h.sessions.Get(tunnelID)

	if exists && previous.OrganizationID != identity.OrganizationID {
		return fmt.Errorf("tunnel belongs to another organization")
	}

	if err := h.checkTunnelCapacity(identity, exists); err != nil {
		return err
	}

	if err := h.claimTunnel(ctx, tunnelID); err != nil {
		return err
	}

	passwordHash, err := h.resolveManagedPolicy(ctx, identity, &open)

	if err != nil {
		return err
	}

	if passwordHash == "" && exists {
		passwordHash = previous.PasswordHash
	}

	session := h.newSession(connection, identity, tunnelID, passwordHash)

	if err := h.sessions.ReserveWithDrain(session, exists, h.drainTimeout); err != nil {
		return err
	}

	owned[tunnelID] = session.ID
	h.recordUsage(ctx, identity.OrganizationID, tunnelID, "tunnel_open", 0, 1)

	if !exists {
		h.metrics.AddTunnel(1)
	}

	if err := h.bindTunnelAlias(open, tunnelID, session.ID, owned, exists); err != nil {
		return err
	}

	publicPort, err := h.openDataListener(open, tunnelID, session.Send, exists)

	if err != nil {
		h.rollbackTunnel(tunnelID, session.ID, owned, exists)
		return err
	}

	payload, err := protocol.EncodePayload(protocol.MessageTypeOpenTunnelAck, message.RequestID, protocol.OpenTunnelAck{TunnelID: tunnelID, PublicURL: publicURL(open, tunnelID, h.publicDomain), PublicPort: publicPort})

	if err != nil {
		h.rollbackTunnel(tunnelID, session.ID, owned, exists)
		return err
	}

	return h.writeMessage(connection, websocket.TextMessage, payload)
}

func (h *Handler) resolveManagedPolicy(ctx context.Context, identity AgentIdentity, open *protocol.OpenTunnel) (string, error) {

	if open.TunnelID != "" && h.managedTunnels != nil {
		policy, err := h.managedTunnels.Resolve(ctx, open.TunnelID)

		if err != nil {
			return "", fmt.Errorf("resolve managed tunnel: %w", err)
		}

		if policy.OrganizationID != identity.OrganizationID || policy.Status == string(models.TunnelStatusRevoked) {
			return "", fmt.Errorf("managed tunnel is not available")
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

func (h *Handler) checkTunnelCapacity(identity AgentIdentity, exists bool) error {

	if exists {
		return nil
	}

	maxTunnels := identity.MaxTunnels

	if maxTunnels == 0 {
		maxTunnels = h.maxTunnels
	}

	organizationTunnels := 0

	for _, active := range h.sessions.Snapshot() {

		if active.OrganizationID == identity.OrganizationID {
			organizationTunnels++
		}
	}

	if len(h.sessions.Snapshot()) >= h.maxTunnels || (maxTunnels > 0 && organizationTunnels >= maxTunnels) {
		return fmt.Errorf("tunnel capacity reached")
	}

	return nil
}

func (h *Handler) claimTunnel(ctx context.Context, tunnelID string) error {

	if h.affinity == nil {
		return nil
	}

	claimed, err := h.affinity.Claim(ctx, tunnelID, h.relayID, h.affinityTTL)

	if err != nil {
		return fmt.Errorf("claim relay affinity: %w", err)
	}

	if !claimed {
		return fmt.Errorf("tunnel is connected through another relay")
	}

	return nil
}

func (h *Handler) newSession(connection *websocket.Conn, identity AgentIdentity, tunnelID, passwordHash string) engine.Session {
	return engine.Session{ID: uuid.NewString(), OrganizationID: identity.OrganizationID, TunnelID: tunnelID, PasswordHash: passwordHash, Send: func(sendCtx context.Context, outgoing protocol.Envelope) error {

		if err := sendCtx.Err(); err != nil {
			return err
		}

		limit := identity.BandwidthBytes

		if limit == 0 {
			limit = h.maxBandwidth
		}

		if limit > 0 && isDataMessage(outgoing.Type) {

			if err := h.bandwidth.Consume(identity.OrganizationID, limit, int64(len(outgoing.Payload))); err != nil {
				return err
			}
		}

		h.recordMessageUsage(sendCtx, identity.OrganizationID, outgoing)
		return h.writeJSON(connection, outgoing)
	}, Close: func() { _ = connection.Close() }}
}

func hashRelayPassword(password string) (string, error) {

	if strings.TrimSpace(password) == "" {

		if password != "" {
			return "", fmt.Errorf("tunnel password cannot contain only whitespace")
		}

		return "", nil
	}

	if len(password) < 8 || len(password) > 256 {
		return "", fmt.Errorf("tunnel password must be between 8 and 256 characters")
	}

	hash, err := security.HashPassword(password)

	if err != nil {
		return "", fmt.Errorf("hash tunnel password: %w", err)
	}

	return hash, nil
}

func (h *Handler) bindTunnelAlias(open protocol.OpenTunnel, tunnelID, sessionID string, owned map[string]string, exists bool) error {
	alias := open.CustomDomain

	if alias == "" {
		alias = open.Subdomain
	}

	if alias == "" {
		return nil
	}

	if err := h.sessions.BindAlias(alias, tunnelID); err != nil {
		h.sessions.Remove(tunnelID, sessionID)
		delete(owned, tunnelID)

		if !exists {
			h.metrics.AddTunnel(-1)
		}

		return err
	}

	return nil
}

func (h *Handler) openDataListener(open protocol.OpenTunnel, tunnelID string, send func(context.Context, protocol.Envelope) error, exists bool) (int, error) {

	if open.Protocol == "tcp" {

		if exists {
			h.tcp.SetSender(tunnelID, send)
			return h.tcp.Port(tunnelID), nil
		}

		return h.tcp.Open(tunnelID, send)
	}

	if open.Protocol == "udp" {

		if exists {
			h.udp.SetSender(tunnelID, send)
			return h.udp.Port(tunnelID), nil
		}

		return h.udp.Open(tunnelID, send)
	}

	return 0, nil
}

func (h *Handler) rollbackTunnel(tunnelID, sessionID string, owned map[string]string, exists bool) {
	h.sessions.Remove(tunnelID, sessionID)
	delete(owned, tunnelID)
	h.tcp.CloseTunnel(tunnelID)
	h.udp.CloseTunnel(tunnelID)

	if !exists {
		h.metrics.AddTunnel(-1)
	}
}

func (h *Handler) touchOrganizationSessions(organizationID string) error {

	for _, session := range h.sessions.Snapshot() {

		if session.OrganizationID == organizationID {
			h.sessions.Touch(session.TunnelID)
		}
	}

	return nil
}

func (h *Handler) closeTunnel(organizationID string, message protocol.Envelope, owned map[string]string) error {
	var closeMessage protocol.CloseTunnel

	if err := protocol.DecodePayload(message, &closeMessage); err != nil {
		return err
	}

	session, ok := h.sessions.Get(closeMessage.TunnelID)

	if !ok || session.OrganizationID != organizationID || !h.sessions.Remove(closeMessage.TunnelID, session.ID) {
		return fmt.Errorf("tunnel not found")
	}

	delete(owned, closeMessage.TunnelID)
	h.metrics.AddTunnel(-1)
	h.router.RemoveTunnel(closeMessage.TunnelID)
	h.tcp.CloseTunnel(closeMessage.TunnelID)
	h.udp.CloseTunnel(closeMessage.TunnelID)
	return nil
}
