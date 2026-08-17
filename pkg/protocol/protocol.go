package protocol

import (
	"encoding/json"
	"fmt"

	generated "outpipe.dev/outpipe/protocol/generated/go"
)

const Version = generated.Version
const MinSupportedVersion = generated.MinSupportedVersion
const MaxSupportedVersion = generated.MaxSupportedVersion

const DefaultMaxFrameSize = generated.DefaultMaxFrameSize
const AbsoluteMaxFrameSize = generated.AbsoluteMaxFrameSize
const DefaultIdleTimeoutSeconds = generated.DefaultIdleTimeoutSeconds
const DefaultConnectionTimeoutSeconds = generated.DefaultConnectionTimeoutSeconds

type MessageType = generated.MessageType
type Envelope = generated.Envelope
type AuthRequest = generated.AuthRequest
type AuthResponse = generated.AuthResponse
type VersionNegotiate = generated.VersionNegotiate
type VersionNegotiateAck = generated.VersionNegotiateAck
type FlowControl = generated.FlowControl
type OpenTunnel = generated.OpenTunnel
type OpenTunnelAck = generated.OpenTunnelAck
type CloseTunnel = generated.CloseTunnel
type Data = generated.Data
type Heartbeat = generated.Heartbeat
type ErrorMessage = generated.ErrorMessage
type HTTPRequest = generated.HTTPRequest
type HTTPResponse = generated.HTTPResponse
type TCPData = generated.TCPData
type TCPClose = generated.TCPClose
type UDPData = generated.UDPData
type UDPResponse = generated.UDPResponse

const (
	MessageTypeAuth                = generated.MessageTypeAuth
	MessageTypeAuthResponse        = generated.MessageTypeAuthResponse
	MessageTypeVersionNegotiate    = generated.MessageTypeVersionNegotiate
	MessageTypeVersionNegotiateAck = generated.MessageTypeVersionNegotiateAck
	MessageTypeFlowControl         = generated.MessageTypeFlowControl
	MessageTypeOpenTunnel          = generated.MessageTypeOpenTunnel
	MessageTypeOpenTunnelAck       = generated.MessageTypeOpenTunnelAck
	MessageTypeCloseTunnel         = generated.MessageTypeCloseTunnel
	MessageTypeData                = generated.MessageTypeData
	MessageTypeHeartbeat           = generated.MessageTypeHeartbeat
	MessageTypeError               = generated.MessageTypeError
	MessageTypeHTTPRequest         = generated.MessageTypeHTTPRequest
	MessageTypeHTTPResponse        = generated.MessageTypeHTTPResponse
	MessageTypeTCPData             = generated.MessageTypeTCPData
	MessageTypeTCPClose            = generated.MessageTypeTCPClose
	MessageTypeUDPData             = generated.MessageTypeUDPData
	MessageTypeUDPResponse         = generated.MessageTypeUDPResponse
)

func Encode(message Envelope) ([]byte, error) {
	if message.Version == 0 {
		message.Version = Version
	}
	if message.Type == "" {
		return nil, fmt.Errorf("protocol message type is required")
	}
	data, err := json.Marshal(message)
	if err != nil {
		return nil, fmt.Errorf("encode protocol message: %w", err)
	}
	if int64(len(data)) > AbsoluteMaxFrameSize {
		return nil, fmt.Errorf("frame size %d exceeds max frame size %d", len(data), AbsoluteMaxFrameSize)
	}
	return data, nil
}

func IsValidMessageType(t MessageType) bool {
	switch t {
	case MessageTypeAuth, MessageTypeAuthResponse, MessageTypeVersionNegotiate, MessageTypeVersionNegotiateAck, MessageTypeFlowControl, MessageTypeOpenTunnel, MessageTypeOpenTunnelAck, MessageTypeCloseTunnel, MessageTypeData, MessageTypeHeartbeat, MessageTypeError, MessageTypeHTTPRequest, MessageTypeHTTPResponse, MessageTypeTCPData, MessageTypeTCPClose, MessageTypeUDPData, MessageTypeUDPResponse:
		return true
	default:
		return false
	}
}

func Decode(data []byte) (Envelope, error) {
	if int64(len(data)) > AbsoluteMaxFrameSize {
		return Envelope{}, fmt.Errorf("frame size %d exceeds max frame size %d", len(data), AbsoluteMaxFrameSize)
	}
	var message Envelope
	if err := json.Unmarshal(data, &message); err != nil {
		return Envelope{}, fmt.Errorf("decode protocol message: %w", err)
	}
	if message.Version < MinSupportedVersion || message.Version > MaxSupportedVersion {
		return Envelope{}, fmt.Errorf("unsupported protocol version %d", message.Version)
	}
	if !IsValidMessageType(message.Type) {
		return Envelope{}, fmt.Errorf("invalid protocol message type %q", message.Type)
	}
	return message, nil
}

func NegotiateVersion(req VersionNegotiate) (VersionNegotiateAck, error) {
	if req.MaxVersion < MinSupportedVersion || req.MinVersion > MaxSupportedVersion {
		return VersionNegotiateAck{}, fmt.Errorf("no compatible protocol version: requested range %d-%d, supported range %d-%d", req.MinVersion, req.MaxVersion, MinSupportedVersion, MaxSupportedVersion)
	}
	negotiated := req.MaxVersion
	if negotiated > MaxSupportedVersion {
		negotiated = MaxSupportedVersion
	}
	return VersionNegotiateAck{
		NegotiatedVersion: negotiated,
		SupportedVersions: []int{1},
		ServerName:        "outpipe",
		ServerVersion:     "0.1.0",
	}, nil
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
