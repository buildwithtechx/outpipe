package client

import (
	"context"
	"fmt"

	"github.com/gorilla/websocket"
	"outpipe.dev/outpipe/pkg/protocol"
)

func negotiate(ctx context.Context, connection *websocket.Conn, token string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if deadline, ok := ctx.Deadline(); ok {
		_ = connection.SetWriteDeadline(deadline)
		_ = connection.SetReadDeadline(deadline)
	}
	versionPayload, err := protocol.EncodePayload(protocol.MessageTypeVersionNegotiate, "version", protocol.VersionNegotiate{MinVersion: protocol.MinSupportedVersion, MaxVersion: protocol.MaxSupportedVersion, ClientName: "outpipe-cli", ClientVersion: "0.1.0"})
	if err != nil {
		return err
	}

	if err := connection.WriteMessage(websocket.TextMessage, versionPayload); err != nil {
		return fmt.Errorf("send protocol negotiation: %w", err)
	}

	if err := readExpected(connection, protocol.MessageTypeVersionNegotiateAck); err != nil {
		return fmt.Errorf("negotiate protocol: %w", err)
	}

	authPayload, err := protocol.EncodePayload(protocol.MessageTypeAuth, "auth", protocol.AuthRequest{Token: token, RequestedCapabilities: []string{"http", "https", "tcp", "udp"}})
	if err != nil {
		return err
	}

	if err := connection.WriteMessage(websocket.TextMessage, authPayload); err != nil {
		return fmt.Errorf("send protocol authentication: %w", err)
	}

	_, data, err := connection.ReadMessage()
	if err != nil {
		return fmt.Errorf("read protocol authentication: %w", err)
	}

	message, err := protocol.Decode(data)
	if err != nil {
		return err
	}

	if message.Type != protocol.MessageTypeAuthResponse {
		return fmt.Errorf("unexpected protocol authentication response %q", message.Type)
	}

	var response protocol.AuthResponse
	if err := protocol.DecodePayload(message, &response); err != nil {
		return err
	}

	if !response.Authenticated {
		return fmt.Errorf("protocol authentication rejected: %s", response.Error)
	}

	return nil
}

func readExpected(connection *websocket.Conn, expected protocol.MessageType) error {
	_, data, err := connection.ReadMessage()
	if err != nil {
		return err
	}

	message, err := protocol.Decode(data)
	if err != nil {
		return err
	}

	if message.Type == protocol.MessageTypeError {
		var failure protocol.ErrorMessage
		_ = protocol.DecodePayload(message, &failure)
		return fmt.Errorf("relay rejected request: %s", failure.Message)
	}

	if message.Type != expected {
		return fmt.Errorf("expected %q, received %q", expected, message.Type)
	}

	return nil
}
