package relay

import (
	"context"
	"log/slog"
	"time"

	"github.com/gofiber/contrib/websocket"
	"outpipe.dev/outpipe/pkg/protocol"
)

func (h *Handler) closeOwnedSessions(ctx context.Context, identity AgentIdentity, owned map[string]string) {
	for tunnelID, sessionID := range owned {
		if h.sessions.Remove(tunnelID, sessionID) {
			if h.affinity != nil {
				_ = h.affinity.Release(ctx, tunnelID, h.relayID)
			}

			h.recordUsage(ctx, identity.OrganizationID, tunnelID, "tunnel_close", 0, 0)
			h.metrics.AddTunnel(-1)
			h.router.RemoveTunnel(tunnelID)
			h.tcp.CloseTunnel(tunnelID)
			h.udp.CloseTunnel(tunnelID)
		}
	}

	h.logger.Info("relay connection closed", slog.String("agent_id", identity.AgentID), slog.Int("tunnels", len(owned)))
}

func (h *Handler) writeError(connection *websocket.Conn, code, message string) {
	payload, err := protocol.EncodePayload(protocol.MessageTypeError, "", protocol.ErrorMessage{Code: code, Message: message})
	if err == nil {
		_ = h.writeMessage(connection, websocket.TextMessage, payload)
	}
}

func (h *Handler) writeJSON(connection *websocket.Conn, value any) error {
	h.writeMu.Lock()
	defer h.writeMu.Unlock()
	_ = connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return connection.WriteJSON(value)
}

func (h *Handler) writeMessage(connection *websocket.Conn, messageType int, data []byte) error {
	h.writeMu.Lock()
	defer h.writeMu.Unlock()
	_ = connection.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return connection.WriteMessage(messageType, data)
}

func (h *Handler) CloseAll() {
	for _, session := range h.sessions.Snapshot() {
		h.sessions.Remove(session.TunnelID, session.ID)
		h.router.RemoveTunnel(session.TunnelID)
		h.tcp.CloseTunnel(session.TunnelID)
		h.udp.CloseTunnel(session.TunnelID)
	}
}
