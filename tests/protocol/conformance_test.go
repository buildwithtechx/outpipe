package protocol_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"outpipe.dev/outpipe/pkg/protocol"
)

type fixtureItem struct {
	Name  string `json:"name"`
	Valid bool   `json:"valid"`
	Raw   string `json:"raw"`
}

type fixturesFile struct {
	Fixtures []fixtureItem `json:"fixtures"`
}

func TestProtocolConformance(t *testing.T) {
	fixturePath := filepath.Join("..", "..", "protocol", "fixtures", "conformance_fixtures.json")
	data, err := os.ReadFile(fixturePath)

	if err != nil {
		t.Fatalf("failed to read conformance fixtures: %v", err)
	}

	var file fixturesFile

	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("failed to parse conformance fixtures JSON: %v", err)
	}

	for _, fixture := range file.Fixtures {
		t.Run(fixture.Name, func(t *testing.T) {
			envelope, err := protocol.Decode([]byte(fixture.Raw))

			if fixture.Valid {

				if err != nil {
					t.Fatalf("expected valid envelope for %s, got error: %v", fixture.Name, err)
				}

				if envelope.Version < protocol.MinSupportedVersion || envelope.Version > protocol.MaxSupportedVersion {
					t.Fatalf("expected valid version for %s, got %d", fixture.Name, envelope.Version)
				}

			} else if err == nil {
				t.Fatalf("expected decode error for invalid fixture %s, got nil", fixture.Name)
			}

		})
	}
}

func TestVersionNegotiation(t *testing.T) {
	ack, err := protocol.NegotiateVersion(protocol.VersionNegotiate{MinVersion: 1, MaxVersion: 1, ClientName: "test", ClientVersion: "1.0"})

	if err != nil {
		t.Fatal(err)
	}

	if ack.NegotiatedVersion != 1 {
		t.Fatalf("expected negotiated version 1, got %d", ack.NegotiatedVersion)
	}

	if _, err := protocol.NegotiateVersion(protocol.VersionNegotiate{MinVersion: 2, MaxVersion: 3}); err == nil {
		t.Fatal("expected error negotiating incompatible versions")
	}
}

func TestMaxFrameSizeEnforcement(t *testing.T) {
	hugeData := make([]byte, protocol.AbsoluteMaxFrameSize+100)

	if _, err := protocol.Encode(protocol.Envelope{Type: protocol.MessageTypeData, Version: protocol.Version, Payload: hugeData}); err == nil {
		t.Fatal("expected error encoding frame exceeding max size")
	}
}
