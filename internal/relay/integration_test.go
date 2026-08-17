package relay

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/engine"
	"outpipe.dev/outpipe/pkg/protocol"
)

func TestLocalHTTPTargetIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok-http-target"))
	}))
	defer server.Close()

	sessions := engine.NewSessionRegistry()
	router, err := engine.NewRequestRouter(sessions, 5*time.Second)
	if err != nil {
		t.Fatalf("create router: %v", err)
	}

	dummySend := func(ctx context.Context, envelope protocol.Envelope) error {
		if envelope.Type == protocol.MessageTypeHTTPRequest {
			resp := protocol.HTTPResponse{StatusCode: 200, Body: "b2staHR0cC10YXJnZXQ="}
			ack, _ := protocol.EncodePayload(protocol.MessageTypeHTTPResponse, envelope.RequestID, resp)
			env, _ := protocol.Decode(ack)
			router.Handle(env)
		}
		return nil
	}

	session := engine.Session{ID: "sess-http", OrganizationID: "org-1", TunnelID: "tunnel-http", Send: dummySend}
	if err := sessions.Reserve(session, false); err != nil {
		t.Fatalf("reserve session: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	resp, err := router.ForwardHTTP(ctx, "tunnel-http", protocol.HTTPRequest{Method: "GET", Path: "/"})
	if err != nil {
		t.Fatalf("forward http failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}
}

func TestLocalTCPTargetIntegration(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen tcp: %v", err)
	}
	defer listener.Close()

	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		buf := make([]byte, 1024)
		n, _ := conn.Read(buf)
		_, _ = conn.Write(buf[:n])
	}()

	tcpMgr := NewTCPManager()
	tunnelID := "tunnel-tcp"
	dummySend := func(ctx context.Context, envelope protocol.Envelope) error { return nil }

	port, err := tcpMgr.Open(tunnelID, dummySend)
	if err != nil {
		t.Fatalf("open tcp tunnel: %v", err)
	}
	if port <= 0 {
		t.Fatalf("expected positive public port, got %d", port)
	}

	tcpMgr.CloseTunnel(tunnelID)
}

func TestLocalUDPTargetIntegration(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("resolve udp addr: %v", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		t.Fatalf("listen udp: %v", err)
	}
	defer conn.Close()

	udpMgr := NewUDPManager()
	dummySend := func(ctx context.Context, envelope protocol.Envelope) error { return nil }

	port, err := udpMgr.Open("tunnel-udp", dummySend)
	if err != nil {
		t.Fatalf("open udp tunnel: %v", err)
	}
	if port <= 0 {
		t.Fatalf("expected positive public port, got %d", port)
	}

	udpMgr.CloseTunnel("tunnel-udp")
}

func TestGracefulShutdownAndDrain(t *testing.T) {
	auth := &mockAuthenticator{tokens: map[string]AgentIdentity{
		"token-1": {AgentID: "agent-1", OrganizationID: "org-1"},
	}}
	sessions := engine.NewSessionRegistry()
	router, _ := engine.NewRequestRouter(sessions, 5*time.Second)
	tcp := NewTCPManager()
	udp := NewUDPManager()

	handler, err := NewHandler(auth, sessions, router, tcp, udp, 10)
	if err != nil {
		t.Fatalf("create handler: %v", err)
	}

	session := engine.Session{ID: "sess-shutdown", OrganizationID: "org-1", TunnelID: "tunnel-shutdown", Send: dummySend}
	if err := sessions.Reserve(session, false); err != nil {
		t.Fatalf("reserve session: %v", err)
	}

	handler.CloseAll()

	if _, ok := sessions.Get("tunnel-shutdown"); ok {
		t.Fatal("expected session to be cleared after CloseAll")
	}
}
