package client

import (
	"context"
	"errors"
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

type RelayRejectedError struct{ Message string }

func (e *RelayRejectedError) Error() string {
	return "relay rejected tunnel: " + e.Message
}

func RelayOwnerFromError(err error) string {
	var rejected *RelayRejectedError
	if err == nil || !errors.As(err, &rejected) {
		return ""
	}
	const marker = "connected through relay "
	_, owner, found := strings.Cut(rejected.Message, marker)
	if found {
		if owner = strings.TrimSpace(owner); owner != "" {
			return owner
		}
	}
	return ""
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

	if deadline, ok := ctx.Deadline(); ok {
		_ = connection.SetWriteDeadline(deadline)
		_ = connection.SetReadDeadline(deadline)
	}
	handshakeDone := make(chan struct{})
	defer close(handshakeDone)
	go func() {
		select {
		case <-ctx.Done():
			_ = connection.Close()
		case <-handshakeDone:
		}
	}()

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
		return nil, &RelayRejectedError{Message: failure.Message}
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

	_ = connection.SetWriteDeadline(time.Time{})
	_ = connection.SetReadDeadline(time.Time{})

	return &RelayConnection{conn: connection, TunnelID: ack.TunnelID, PublicURL: ack.PublicURL, PublicPort: ack.PublicPort, Protocol: open.Protocol, tcpConns: make(map[string]net.Conn)}, nil
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
	return c.ServeLocalWithHTTPClient(ctx, targetURL, &http.Client{Timeout: 90 * time.Second})
}

func (c *RelayConnection) ServeLocalWithHTTPClient(ctx context.Context, targetURL string, originClient *http.Client) error {
	if originClient == nil {
		return fmt.Errorf("origin HTTP client is required")
	}
	stopClose := context.AfterFunc(ctx, func() { _ = c.Close() })
	defer stopClose()

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
			response := c.forwardHTTP(ctx, targetURL, request, originClient)
			payload, err := protocol.EncodePayload(protocol.MessageTypeHTTPResponse, message.RequestID, response)
			if err != nil {
				return err
			}
			if err := c.write(payload); err != nil {
				return err
			}
		case protocol.MessageTypeTCPData:
			if err := c.handleTCPData(ctx, targetURL, message); err != nil {
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
