package protocol

import "testing"

func TestRoundTrip(t *testing.T) {
	encoded, err := EncodePayload(MessageTypeHeartbeat, "request-1", HeartbeatRequest{Timestamp: 123})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	message, err := Decode(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	var heartbeat HeartbeatRequest
	if err := DecodePayload(message, &heartbeat); err != nil {
		t.Fatalf("payload: %v", err)
	}
	if heartbeat.Timestamp != 123 {
		t.Fatalf("timestamp = %d", heartbeat.Timestamp)
	}
}
