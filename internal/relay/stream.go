package relay

import (
	"fmt"

	"outpipe.dev/outpipe/pkg/protocol"
	"outpipe.dev/outpipe/pkg/utils"
)

func (h *Handler) handleTCPData(message protocol.Envelope) error {
	data, err := decodeTCPData(message)

	if err != nil {
		return err
	}

	decoded, err := utils.DecodeBase64(data.Data)

	if err != nil {
		return fmt.Errorf("decode tcp data: %w", err)
	}

	return h.tcp.Write(data.TunnelID, data.ConnectionID, decoded)
}

func (h *Handler) handleTCPClose(message protocol.Envelope) error {
	closeMessage, err := decodeTCPClose(message)

	if err != nil {
		return err
	}

	h.tcp.CloseConnection(closeMessage.TunnelID, closeMessage.ConnectionID)
	return nil
}

func (h *Handler) handleUDPResponse(message protocol.Envelope) error {
	response, err := decodeUDPResponse(message)

	if err != nil {
		return err
	}

	return h.udp.Write(response.TunnelID, response)
}
