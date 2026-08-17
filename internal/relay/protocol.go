package relay

import (
	"fmt"

	"outpipe.dev/outpipe/pkg/protocol"
)

func isDataMessage(messageType protocol.MessageType) bool {
	return messageType == protocol.MessageTypeHTTPRequest || messageType == protocol.MessageTypeHTTPResponse || messageType == protocol.MessageTypeTCPData || messageType == protocol.MessageTypeUDPData || messageType == protocol.MessageTypeUDPResponse
}

func decodeOpenTunnel(message protocol.Envelope) (protocol.OpenTunnel, error) {
	var open protocol.OpenTunnel
	if err := protocol.DecodePayload(message, &open); err != nil {
		return open, err
	}
	if open.Protocol == "" || open.LocalPort < 1 || open.LocalPort > 65535 {
		return open, fmt.Errorf("invalid tunnel open request")
	}
	return open, nil
}

func decodeTCPData(message protocol.Envelope) (protocol.TCPData, error) {
	var data protocol.TCPData
	if err := protocol.DecodePayload(message, &data); err != nil {
		return data, err
	}
	if data.TunnelID == "" {
		return data, fmt.Errorf("tcp tunnel id is required")
	}
	return data, nil
}

func decodeTCPClose(message protocol.Envelope) (protocol.TCPClose, error) {
	var closeMessage protocol.TCPClose
	if err := protocol.DecodePayload(message, &closeMessage); err != nil {
		return closeMessage, err
	}
	if closeMessage.TunnelID == "" {
		return closeMessage, fmt.Errorf("tcp tunnel id is required")
	}
	return closeMessage, nil
}

func decodeUDPResponse(message protocol.Envelope) (protocol.UDPResponse, error) {
	var response protocol.UDPResponse
	if err := protocol.DecodePayload(message, &response); err != nil {
		return response, err
	}
	return response, nil
}
