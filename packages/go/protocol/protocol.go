package protocol

import (
	"encoding/json"
	"fmt"
)

const Version = 1
const MaxFrameSize = 32 * 1024 * 1024

type MessageType string

const (
	MessageTypeAuth                MessageType = "auth"
	MessageTypeAuthResponse        MessageType = "auth_response"
	MessageTypeVersionNegotiate    MessageType = "version_negotiate"
	MessageTypeVersionNegotiateAck MessageType = "version_negotiate_ack"
	MessageTypeFlowControl         MessageType = "flow_control"
	MessageTypeOpenTunnel          MessageType = "open_tunnel"
	MessageTypeOpenTunnelAck       MessageType = "open_tunnel_ack"
	MessageTypeCloseTunnel         MessageType = "close_tunnel"
	MessageTypeData                MessageType = "data"
	MessageTypeHeartbeat           MessageType = "heartbeat"
	MessageTypeError               MessageType = "error"
	MessageTypeHTTPRequest         MessageType = "http_request"
	MessageTypeHTTPResponse        MessageType = "http_response"
	MessageTypeTCPData             MessageType = "tcp_data"
	MessageTypeTCPClose            MessageType = "tcp_close"
	MessageTypeUDPData             MessageType = "udp_data"
	MessageTypeUDPResponse         MessageType = "udp_response"
)

var messageTypes = map[MessageType]struct{}{
	MessageTypeAuth: {}, MessageTypeAuthResponse: {},
	MessageTypeVersionNegotiate: {}, MessageTypeVersionNegotiateAck: {},
	MessageTypeFlowControl: {}, MessageTypeOpenTunnel: {},
	MessageTypeOpenTunnelAck: {}, MessageTypeCloseTunnel: {},
	MessageTypeData: {}, MessageTypeHeartbeat: {}, MessageTypeError: {},
	MessageTypeHTTPRequest: {}, MessageTypeHTTPResponse: {},
	MessageTypeTCPData: {}, MessageTypeTCPClose: {},
	MessageTypeUDPData: {}, MessageTypeUDPResponse: {},
}

type Envelope struct {
	Version   int             `json:"version"`
	Type      MessageType     `json:"type"`
	RequestID string          `json:"request_id,omitempty"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

type AuthRequest struct {
	Token                 string   `json:"token"`
	AgentID               string   `json:"agent_id,omitempty"`
	RequestedCapabilities []string `json:"requested_capabilities,omitempty"`
}
type AuthResponse struct {
	Authenticated       bool     `json:"authenticated"`
	AgentID             string   `json:"agent_id,omitempty"`
	OrganizationID      string   `json:"organization_id,omitempty"`
	GrantedCapabilities []string `json:"granted_capabilities,omitempty"`
	Error               string   `json:"error,omitempty"`
}
type VersionRequest struct {
	MinVersion    int    `json:"min_version"`
	MaxVersion    int    `json:"max_version"`
	ClientName    string `json:"client_name,omitempty"`
	ClientVersion string `json:"client_version,omitempty"`
}
type VersionAck struct {
	NegotiatedVersion int   `json:"negotiated_version"`
	SupportedVersions []int `json:"supported_versions"`
}
type OpenRequest struct {
	Token        string `json:"token"`
	TunnelID     string `json:"tunnel_id,omitempty"`
	LocalPort    int    `json:"local_port"`
	Protocol     string `json:"protocol"`
	Subdomain    string `json:"subdomain,omitempty"`
	CustomDomain string `json:"custom_domain,omitempty"`
	Password     string `json:"password,omitempty"`
}
type OpenAck struct {
	TunnelID   string `json:"tunnel_id"`
	PublicURL  string `json:"public_url"`
	PublicPort int    `json:"public_port,omitempty"`
}
type CloseRequest struct {
	TunnelID string `json:"tunnel_id"`
	Reason   string `json:"reason,omitempty"`
}
type HeartbeatRequest struct {
	Timestamp int64 `json:"timestamp"`
}
type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type FlowControl struct {
	StreamID   string `json:"stream_id"`
	Action     string `json:"action"`
	WindowSize int64  `json:"window_size,omitempty"`
}
type Data struct {
	TunnelID string `json:"tunnel_id"`
	StreamID string `json:"stream_id"`
	Data     string `json:"data"`
}
type HTTPRequest struct {
	Method  string              `json:"method"`
	Path    string              `json:"path"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body,omitempty"`
}
type HTTPResponse struct {
	StatusCode int                 `json:"status_code"`
	Headers    map[string][]string `json:"headers"`
	Body       string              `json:"body,omitempty"`
	Error      string              `json:"error,omitempty"`
}
type TCPData struct {
	TunnelID     string `json:"tunnel_id,omitempty"`
	ConnectionID string `json:"connection_id"`
	Data         string `json:"data"`
}
type TCPClose struct {
	TunnelID     string `json:"tunnel_id,omitempty"`
	ConnectionID string `json:"connection_id"`
	Reason       string `json:"reason,omitempty"`
}
type UDPData struct {
	TunnelID      string `json:"tunnel_id,omitempty"`
	PacketID      string `json:"packet_id"`
	SourceAddress string `json:"source_address"`
	SourcePort    int    `json:"source_port"`
	Data          string `json:"data"`
}
type UDPResponse struct {
	TunnelID      string `json:"tunnel_id,omitempty"`
	PacketID      string `json:"packet_id"`
	TargetAddress string `json:"target_address"`
	TargetPort    int    `json:"target_port"`
	Data          string `json:"data"`
}

func Encode(message Envelope) ([]byte, error) {
	if message.Version == 0 {
		message.Version = Version
	}
	if message.Type == "" {
		return nil, fmt.Errorf("protocol message type is required")
	}
	encoded, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("encode protocol message: %w", err)
	}
	if len(encoded) > MaxFrameSize {
		return nil, fmt.Errorf("protocol frame exceeds %d bytes", MaxFrameSize)
	}
	return encoded, nil
}

func Decode(data []byte) (Envelope, error) {
	if len(data) > MaxFrameSize {
		return Envelope{}, fmt.Errorf("protocol frame exceeds %d bytes", MaxFrameSize)
	}
	var message Envelope
	if err := json.Unmarshal(data, &message); err != nil {
		return Envelope{}, fmt.Errorf("decode protocol message: %w", err)
	}
	if message.Version != Version || message.Type == "" {
		return Envelope{}, fmt.Errorf("invalid protocol envelope")
	}
	if _, ok := messageTypes[message.Type]; !ok {
		return Envelope{}, fmt.Errorf("invalid protocol message type %q", message.Type)
	}
	return message, nil
}

func EncodePayload(messageType MessageType, requestID string, payload any) ([]byte, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("encode protocol payload: %w", err)
	}
	return Encode(Envelope{Type: messageType, RequestID: requestID, Payload: body})
}

func DecodePayload(message Envelope, target any) error {
	if err := json.Unmarshal(message.Payload, target); err != nil {
		return fmt.Errorf("decode protocol payload: %w", err)
	}
	return nil
}
