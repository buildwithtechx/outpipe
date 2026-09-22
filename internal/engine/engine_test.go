package engine

import (
	"context"
	"net"
	"testing"
	"time"

	"outpipe.dev/outpipe/pkg/protocol"
)

func TestPortAllocatorHonorsRequestedPorts(t *testing.T) {
	allocator := NewPortAllocator(20000, 20001)
	first, ok := allocator.Allocate(20001)

	if !ok || first != 20001 {
		t.Fatalf("expected requested port 20001, got %d", first)
	}

	if _, ok := allocator.Allocate(20001); ok {
		t.Fatal("expected allocated port to be unavailable")
	}

	allocator.Release(20001)

	if allocator.InUse(20001) {
		t.Fatal("expected released port to be available")
	}
}

func TestRequestRouterForwardsAndResolvesHTTP(t *testing.T) {
	sessions := NewSessionRegistry()
	router, err := NewRequestRouter(sessions, time.Second)

	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	session := Session{
		ID:       "session-1",
		TunnelID: "tunnel-1",
		Send: func(_ context.Context, message protocol.Envelope) error {

			if message.Type != protocol.MessageTypeHTTPRequest {
				t.Fatalf("unexpected message type: %s", message.Type)
			}

			response, encodeErr := protocol.EncodePayload(protocol.MessageTypeHTTPResponse, message.RequestID, protocol.HTTPResponse{StatusCode: 204})

			if encodeErr != nil {
				t.Fatalf("encode response: %v", encodeErr)
			}

			encoded, decodeErr := protocol.Decode(response)

			if decodeErr != nil {
				t.Fatalf("decode response: %v", decodeErr)
			}

			go router.Handle(encoded)
			return nil
		},
	}
	if err := sessions.Reserve(session, false); err != nil {
		t.Fatalf("reserve session: %v", err)
	}

	response, err := router.ForwardHTTP(context.Background(), "tunnel-1", protocol.HTTPRequest{Method: "GET", Path: "/"})

	if err != nil {
		t.Fatalf("forward request: %v", err)
	}

	if response.StatusCode != 204 {
		t.Fatalf("expected 204 response, got %d", response.StatusCode)
	}
}

func TestBandwidthLimiterRejectsOverflow(t *testing.T) {
	limiter := NewBandwidthLimiter()

	if err := limiter.Consume("org-1", 10, 7); err != nil {
		t.Fatalf("consume bandwidth: %v", err)
	}

	if err := limiter.Consume("org-1", 10, 4); err == nil {
		t.Fatal("expected bandwidth overflow")
	}
}

func TestPipeConnectionsMovesData(t *testing.T) {
	left, leftPeer := net.Pipe()
	right, rightPeer := net.Pipe()
	defer leftPeer.Close()
	defer rightPeer.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result := make(chan error, 1)
	go func() { result <- PipeConnections(ctx, left, right, 1024) }()
	go func() { _, _ = leftPeer.Write([]byte("hello")) }()
	buffer := make([]byte, 5)

	if _, err := rightPeer.Read(buffer); err != nil {
		t.Fatalf("read piped data: %v", err)
	}

	if string(buffer) != "hello" {
		t.Fatalf("unexpected piped data: %s", buffer)
	}

	cancel()
	leftPeer.Close()
	rightPeer.Close()
	<-result
}
