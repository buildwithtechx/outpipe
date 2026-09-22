package relay

import (
	"context"

	"outpipe.dev/outpipe/internal/engine"
	"outpipe.dev/outpipe/pkg/protocol"
	"outpipe.dev/outpipe/pkg/utils"
)

func (h *Handler) recordMessageUsage(ctx context.Context, organizationID string, message protocol.Envelope) {

	if h.usage == nil {
		return
	}

	var eventType string
	var encoded string

	switch message.Type {
	case protocol.MessageTypeTCPData:
		var data protocol.TCPData

		if protocol.DecodePayload(message, &data) == nil {
			eventType, encoded = "tcp", data.Data
		}
	case protocol.MessageTypeUDPData:
		var data protocol.UDPData

		if protocol.DecodePayload(message, &data) == nil {
			eventType, encoded = "udp", data.Data
		}
	case protocol.MessageTypeUDPResponse:
		var data protocol.UDPResponse

		if protocol.DecodePayload(message, &data) == nil {
			eventType, encoded = "udp", data.Data
		}
	default:
		return
	}

	data, err := utils.DecodeBase64(encoded)

	if err != nil {
		return
	}

	h.recordUsage(ctx, organizationID, tunnelIDFromMessage(message), eventType, int64(len(data)), 0)
}

func (h *Handler) recordUsage(ctx context.Context, organizationID, tunnelID, eventType string, bytes int64, connections int) {

	if h.usage == nil || organizationID == "" {
		return
	}

	_ = h.usage.Record(ctx, engine.UsageMeasurement{OrganizationID: organizationID, TunnelID: tunnelID, EventType: eventType, Bytes: bytes, Connections: connections})
}

func tunnelIDFromMessage(message protocol.Envelope) string {

	switch message.Type {
	case protocol.MessageTypeTCPData:
		var data protocol.TCPData

		if protocol.DecodePayload(message, &data) == nil {
			return data.TunnelID
		}
	case protocol.MessageTypeUDPData:
		var data protocol.UDPData

		if protocol.DecodePayload(message, &data) == nil {
			return data.TunnelID
		}
	case protocol.MessageTypeUDPResponse:
		var data protocol.UDPResponse

		if protocol.DecodePayload(message, &data) == nil {
			return data.TunnelID
		}
	}

	return ""
}
