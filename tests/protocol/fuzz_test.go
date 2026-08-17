package protocol_test

import (
	"testing"

	"outpipe.dev/outpipe/pkg/protocol"
)

func FuzzProtocolDecode(f *testing.F) {
	seeds := [][]byte{
		[]byte(`{"version":1,"type":"heartbeat","payload":"{}"}`),
		[]byte(`{"version":1,"type":"open_tunnel","payload":"{\"local_port\":3000,\"protocol\":\"http\"}"}`),
		[]byte(`invalid json`),
		[]byte(``),
		[]byte(`{"version":999,"type":"unknown"}`),
	}
	for _, seed := range seeds {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = protocol.Decode(data)
	})
}
