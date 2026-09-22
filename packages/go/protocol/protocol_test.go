package protocol

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

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

func TestConformanceFixtures(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("..", "..", "..", "protocol", "fixtures", "conformance_fixtures.json"))
	if err != nil {
		t.Fatalf("read fixtures: %v", err)
	}
	var file struct {
		Fixtures []struct {
			Name  string `json:"name"`
			Raw   string `json:"raw"`
			Valid bool   `json:"valid"`
		} `json:"fixtures"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("decode fixtures: %v", err)
	}
	for _, fixture := range file.Fixtures {
		_, decodeErr := Decode([]byte(fixture.Raw))
		if fixture.Valid && decodeErr != nil {
			t.Errorf("%s: expected valid fixture: %v", fixture.Name, decodeErr)
		}
		if !fixture.Valid && decodeErr == nil {
			t.Errorf("%s: expected invalid fixture", fixture.Name)
		}
	}
}
