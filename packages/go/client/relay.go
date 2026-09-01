package client

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"outpipe.dev/outpipe-go/protocol"
)

type RelayConfig struct {
	URL         string
	Token       string
	AgentID     string
	HTTPHeaders http.Header
}
type RelayConnection struct {
	conn   *websocket.Conn
	Tunnel protocol.OpenAck
}

func OpenRelay(ctx context.Context, cfg RelayConfig, open protocol.OpenRequest) (*RelayConnection, error) {
	if strings.TrimSpace(cfg.URL) == "" || strings.TrimSpace(cfg.Token) == "" {
		return nil, fmt.Errorf("relay URL and token are required")
	}
	if open.LocalPort < 1 || open.LocalPort > 65535 || open.Protocol == "" {
		return nil, fmt.Errorf("valid local port and protocol are required")
	}
	headers := cfg.HTTPHeaders
	if headers == nil {
		headers = make(http.Header)
	}
	headers.Set("Authorization", "Bearer "+cfg.Token)
	connection, _, err := websocket.DefaultDialer.DialContext(ctx, cfg.URL, headers)
	if err != nil {
		return nil, fmt.Errorf("connect relay: %w", err)
	}
	if err := exchange(connection, protocol.MessageTypeVersionNegotiate, protocol.VersionRequest{MinVersion: protocol.Version, MaxVersion: protocol.Version, ClientName: "outpipe-go", ClientVersion: "0.1.0"}, protocol.MessageTypeVersionNegotiateAck, nil); err != nil {
		_ = connection.Close()
		return nil, err
	}
	open.Token = cfg.Token
	var auth protocol.AuthResponse
	if err := exchange(connection, protocol.MessageTypeAuth, protocol.AuthRequest{Token: cfg.Token, AgentID: cfg.AgentID}, protocol.MessageTypeAuthResponse, &auth); err != nil {
		_ = connection.Close()
		return nil, err
	}
	if !auth.Authenticated {
		_ = connection.Close()
		return nil, fmt.Errorf("relay authentication rejected: %s", auth.Error)
	}
	if err := writePayload(connection, protocol.MessageTypeOpenTunnel, "", open); err != nil {
		_ = connection.Close()
		return nil, err
	}
	_, data, err := connection.ReadMessage()
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("read tunnel acknowledgement: %w", err)
	}
	message, err := protocol.Decode(data)
	if err != nil {
		_ = connection.Close()
		return nil, err
	}
	if message.Type != protocol.MessageTypeOpenTunnelAck {
		_ = connection.Close()
		return nil, fmt.Errorf("expected %q, received %q", protocol.MessageTypeOpenTunnelAck, message.Type)
	}
	var ack protocol.OpenAck
	if err := protocol.DecodePayload(message, &ack); err != nil {
		_ = connection.Close()
		return nil, err
	}
	return &RelayConnection{conn: connection, Tunnel: ack}, nil
}

func (c *RelayConnection) Heartbeat() error {
	return writePayload(c.conn, protocol.MessageTypeHeartbeat, "", protocol.HeartbeatRequest{Timestamp: time.Now().Unix()})
}

func (c *RelayConnection) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func exchange(connection *websocket.Conn, messageType protocol.MessageType, payload any, expected protocol.MessageType, target any) error {
	if err := writePayload(connection, messageType, "request", payload); err != nil {
		return err
	}
	_, data, err := connection.ReadMessage()
	if err != nil {
		return fmt.Errorf("read %s response: %w", expected, err)
	}
	message, err := protocol.Decode(data)
	if err != nil {
		return err
	}
	if message.Type != expected {
		return fmt.Errorf("expected %q, received %q", expected, message.Type)
	}
	if target != nil {
		return protocol.DecodePayload(message, target)
	}
	return nil
}

func writePayload(connection *websocket.Conn, messageType protocol.MessageType, requestID string, payload any) error {
	data, err := protocol.EncodePayload(messageType, requestID, payload)
	if err != nil {
		return err
	}
	if err := connection.WriteMessage(websocket.TextMessage, data); err != nil {
		return fmt.Errorf("send %s: %w", messageType, err)
	}
	return nil
}
