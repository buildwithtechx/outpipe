package relay

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/contrib/websocket"
	"outpipe.dev/outpipe/pkg/protocol"
)

func (h *Handler) setOrganizationLimit(organizationID string, limit int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if limit > 0 {
		h.orgLimits[organizationID] = limit
	}
}

func (h *Handler) allowConnection(tunnelID string) bool {
	organizationID, ok := h.router.OrganizationID(tunnelID)
	if !ok {
		return false
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	limit := h.orgLimits[organizationID]
	return limit <= 0 || h.orgConnections[organizationID] < limit
}

func (h *Handler) updateOrganizationConnections(organizationID string, delta int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.orgConnections[organizationID] += delta
	if h.orgConnections[organizationID] <= 0 {
		delete(h.orgConnections, organizationID)
	}
}

func splitOrigins(value string) []string {
	var origins []string
	for _, item := range strings.Split(value, ",") {
		if item = strings.TrimSpace(item); item != "" {
			origins = append(origins, item)
		}
	}
	return origins
}

func (h *Handler) originAllowed(origin string) bool {
	if origin == "" || len(h.allowedOrigins) == 0 {
		return true
	}
	for _, allowed := range h.allowedOrigins {
		if origin == allowed {
			return true
		}
	}
	return false
}

func (h *Handler) acquireConnection() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.connections >= h.maxSessions {
		return false
	}
	h.connections++
	return true
}

func (h *Handler) releaseConnection() {
	h.mu.Lock()
	h.connections--
	h.mu.Unlock()
}

func (h *Handler) sendHeartbeats(ctx context.Context, connection *websocket.Conn, organizationID string) {
	ticker := time.NewTicker(h.heartbeat)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			if err := h.writeJSON(connection, protocol.Envelope{Version: protocol.Version, Type: protocol.MessageTypeHeartbeat, Payload: []byte(fmt.Sprintf(`{"timestamp":%d}`, now.Unix()))}); err != nil {
				_ = connection.Close()
				return
			}
			for _, session := range h.sessions.Snapshot() {
				if session.OrganizationID == organizationID {
					h.sessions.Touch(session.TunnelID)
				}
			}
		}
	}
}
