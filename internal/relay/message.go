package relay

import (
	"context"
	"fmt"

	"github.com/gofiber/contrib/websocket"
	"outpipe.dev/outpipe/pkg/protocol"
)

func (h *Handler) handleMessage(ctx context.Context, connection *websocket.Conn, identity AgentIdentity, message protocol.Envelope, owned map[string]string, states ...*connectionState) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	state := &connectionState{negotiated: true, authenticated: true}
	if len(states) > 0 && states[0] != nil {
		state = states[0]
	}
	if state.identity.OrganizationID != "" {
		identity = state.identity
	}
	bandwidthLimit := identity.BandwidthBytes
	if bandwidthLimit == 0 {
		bandwidthLimit = h.maxBandwidth
	}
	if bandwidthLimit > 0 && isDataMessage(message.Type) {
		if err := h.bandwidth.Consume(identity.OrganizationID, bandwidthLimit, int64(len(message.Payload))); err != nil {
			return err
		}
	}
	switch message.Type {
	case protocol.MessageTypeVersionNegotiate:
		return h.handleVersionNegotiation(connection, message, state)
	case protocol.MessageTypeAuth:
		return h.handleAuthentication(ctx, connection, message, state)
	case protocol.MessageTypeFlowControl:
		return h.handleFlowControl(ctx, message)
	case protocol.MessageTypeOpenTunnel:
		return h.openTunnel(ctx, connection, identity, message, owned, state)
	case protocol.MessageTypeHeartbeat:
		return h.touchOrganizationSessions(identity.OrganizationID)
	case protocol.MessageTypeCloseTunnel:
		return h.closeTunnel(identity.OrganizationID, message, owned)
	case protocol.MessageTypeHTTPResponse:
		if !h.router.Handle(message) {
			return fmt.Errorf("unmatched http response")
		}
	case protocol.MessageTypeTCPData:
		return h.handleTCPData(message)
	case protocol.MessageTypeTCPClose:
		return h.handleTCPClose(message)
	case protocol.MessageTypeUDPResponse:
		return h.handleUDPResponse(message)
	default:
		return fmt.Errorf("unsupported relay message type %q", message.Type)
	}
	return nil
}

func (h *Handler) handleFlowControl(ctx context.Context, message protocol.Envelope) error {
	var flowControl protocol.FlowControl
	if err := protocol.DecodePayload(message, &flowControl); err != nil {
		return err
	}
	h.logger.DebugContext(ctx, "flow control message received", "stream_id", flowControl.StreamID, "action", flowControl.Action)
	return nil
}

func (h *Handler) handleVersionNegotiation(connection *websocket.Conn, message protocol.Envelope, state *connectionState) error {
	var request protocol.VersionNegotiate
	if err := protocol.DecodePayload(message, &request); err != nil {
		return err
	}
	ack, err := protocol.NegotiateVersion(request)
	if err != nil {
		return err
	}
	state.negotiated = true
	payload, err := protocol.EncodePayload(protocol.MessageTypeVersionNegotiateAck, message.RequestID, ack)
	if err != nil {
		return err
	}
	return h.writeMessage(connection, websocket.TextMessage, payload)
}

func (h *Handler) handleAuthentication(ctx context.Context, connection *websocket.Conn, message protocol.Envelope, state *connectionState) error {
	if !state.negotiated {
		return fmt.Errorf("protocol version must be negotiated before authentication")
	}
	var request protocol.AuthRequest
	if err := protocol.DecodePayload(message, &request); err != nil {
		return err
	}
	identity, err := h.authenticator.Authenticate(ctx, request.Token)
	if err != nil {
		payload, _ := protocol.EncodePayload(protocol.MessageTypeAuthResponse, message.RequestID, protocol.AuthResponse{Authenticated: false, Error: err.Error()})
		_ = h.writeMessage(connection, websocket.TextMessage, payload)
		return err
	}
	state.authenticated = true
	state.identity = identity
	payload, err := protocol.EncodePayload(protocol.MessageTypeAuthResponse, message.RequestID, protocol.AuthResponse{Authenticated: true, AgentID: identity.AgentID, OrganizationID: identity.OrganizationID, GrantedCapabilities: []string{"http", "https", "tcp", "udp"}})
	if err != nil {
		return err
	}
	return h.writeMessage(connection, websocket.TextMessage, payload)
}
