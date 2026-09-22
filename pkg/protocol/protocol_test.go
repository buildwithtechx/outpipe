package protocol

import "testing"

func TestEncodeDecode(t *testing.T) {
	data, err := EncodePayload(MessageTypeHeartbeat, "request-1", Heartbeat{Timestamp: 123})

	if err != nil {
		t.Fatalf("encode message: %v", err)
	}

	message, err := Decode(data)

	if err != nil {
		t.Fatalf("decode message: %v", err)
	}

	var heartbeat Heartbeat

	if err := DecodePayload(message, &heartbeat); err != nil {
		t.Fatalf("decode payload: %v", err)
	}

	if heartbeat.Timestamp != 123 {
		t.Fatalf("expected timestamp 123, got %d", heartbeat.Timestamp)
	}
}
