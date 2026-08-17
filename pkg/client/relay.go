package client

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"outpipe.dev/outpipe/pkg/protocol"
)

type RelayConfig struct {
	URL         string
	Token       string
	HTTPHeaders http.Header
}

type RelayConnection struct {
	conn       *websocket.Conn
	TunnelID   string
	PublicURL  string
	PublicPort int
	Protocol   string
	writeMu    sync.Mutex
	tcpMu      sync.Mutex
	tcpConns   map[string]net.Conn
	udpMu      sync.Mutex
	udpConn    *net.UDPConn
	udpQueue   []string
}

func OpenRelay(ctx context.Context, cfg RelayConfig, open protocol.OpenTunnel) (*RelayConnection, error) {
	if strings.TrimSpace(cfg.URL) == "" || strings.TrimSpace(cfg.Token) == "" {
		return nil, fmt.Errorf("relay url and token are required")
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
	if err := negotiate(ctx, connection, cfg.Token); err != nil {
		_ = connection.Close()
		return nil, err
	}
	open.Token = cfg.Token
	message, err := protocol.EncodePayload(protocol.MessageTypeOpenTunnel, "", open)
	if err != nil {
		_ = connection.Close()
		return nil, err
	}
	if err := connection.WriteMessage(websocket.TextMessage, message); err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("send tunnel open: %w", err)
	}
	_, response, err := connection.ReadMessage()
	if err != nil {
		_ = connection.Close()
		return nil, fmt.Errorf("read tunnel acknowledgement: %w", err)
	}
	envelope, err := protocol.Decode(response)
	if err != nil {
		_ = connection.Close()
		return nil, err
	}
	if envelope.Type == protocol.MessageTypeError {
		var failure protocol.ErrorMessage
		_ = protocol.DecodePayload(envelope, &failure)
		_ = connection.Close()
		return nil, fmt.Errorf("relay rejected tunnel: %s", failure.Message)
	}
	if envelope.Type != protocol.MessageTypeOpenTunnelAck {
		_ = connection.Close()
		return nil, fmt.Errorf("unexpected relay response %q", envelope.Type)
	}
	var ack protocol.OpenTunnelAck
	if err := protocol.DecodePayload(envelope, &ack); err != nil {
		_ = connection.Close()
		return nil, err
	}
	return &RelayConnection{conn: connection, TunnelID: ack.TunnelID, PublicURL: ack.PublicURL, PublicPort: ack.PublicPort, Protocol: open.Protocol, tcpConns: make(map[string]net.Conn)}, nil
}

func negotiate(ctx context.Context, connection *websocket.Conn, token string) error {
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

func (c *RelayConnection) SendHeartbeat() error {
	if c == nil || c.conn == nil {
		return fmt.Errorf("relay connection is required")
	}
	message, err := protocol.EncodePayload(protocol.MessageTypeHeartbeat, "", protocol.Heartbeat{Timestamp: time.Now().Unix()})
	if err != nil {
		return err
	}
	return c.write(message)
}

func (c *RelayConnection) ServeLocal(ctx context.Context, targetURL string) error {
	for {
		_, data, err := c.conn.ReadMessage()
		if err != nil {
			return fmt.Errorf("read relay message: %w", err)
		}
		message, err := protocol.Decode(data)
		if err != nil {
			return err
		}
		switch message.Type {
		case protocol.MessageTypeHTTPRequest:
			var request protocol.HTTPRequest
			if err := protocol.DecodePayload(message, &request); err != nil {
				return err
			}
			response := c.forwardHTTP(ctx, targetURL, request)
			payload, err := protocol.EncodePayload(protocol.MessageTypeHTTPResponse, message.RequestID, response)
			if err != nil {
				return err
			}
			if err := c.write(payload); err != nil {
				return err
			}
		case protocol.MessageTypeTCPData:
			if err := c.handleTCPData(targetURL, message); err != nil {
				return err
			}
		case protocol.MessageTypeTCPClose:
			var closeMessage protocol.TCPClose
			if err := protocol.DecodePayload(message, &closeMessage); err != nil {
				return err
			}
			c.closeTCP(closeMessage.ConnectionID)
		case protocol.MessageTypeUDPData:
			if err := c.handleUDPData(targetURL, message); err != nil {
				return err
			}
		}
	}
}

func (c *RelayConnection) write(data []byte) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

func (c *RelayConnection) Close() error {
	if c == nil || c.conn == nil {
		return nil
	}
	c.tcpMu.Lock()
	for connectionID, connection := range c.tcpConns {
		_ = connection.Close()
		delete(c.tcpConns, connectionID)
	}
	c.tcpMu.Unlock()
	c.udpMu.Lock()
	if c.udpConn != nil {
		_ = c.udpConn.Close()
		c.udpConn = nil
	}
	c.udpQueue = nil
	c.udpMu.Unlock()
	return c.conn.Close()
}
