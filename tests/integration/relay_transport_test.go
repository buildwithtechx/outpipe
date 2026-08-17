package integration_test

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/engine"
	"outpipe.dev/outpipe/internal/relay"
	"outpipe.dev/outpipe/pkg/protocol"
)

func TestHTTPProxyWithRealLocalTarget(t *testing.T) {
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { _, _ = w.Write([]byte("real-http")) }))
	defer target.Close()
	sessions := engine.NewSessionRegistry()
	router, err := engine.NewRequestRouter(sessions, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	session := engine.Session{ID: "http-session", OrganizationID: "org", TunnelID: "http-tunnel", Send: func(ctx context.Context, message protocol.Envelope) error {
		var request protocol.HTTPRequest
		if err := protocol.DecodePayload(message, &request); err != nil {
			return err
		}
		response, err := http.Get(target.URL + "/")
		if err != nil {
			return err
		}
		defer response.Body.Close()
		body, _ := io.ReadAll(response.Body)
		payload, _ := protocol.EncodePayload(protocol.MessageTypeHTTPResponse, message.RequestID, protocol.HTTPResponse{StatusCode: response.StatusCode, Body: base64.StdEncoding.EncodeToString(body)})
		decoded, _ := protocol.Decode(payload)
		router.Handle(decoded)
		return nil
	}}
	if err := sessions.Reserve(session, false); err != nil {
		t.Fatal(err)
	}
	proxy, err := engine.NewHTTPProxy("tunnel.localhost", router, 1<<20)
	if err != nil {
		t.Fatal(err)
	}
	record := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "http://http-tunnel.tunnel.localhost/", nil)
	proxy.ServeHTTP(record, request)
	if record.Code != http.StatusOK || record.Body.String() != "real-http" {
		t.Fatalf("unexpected response: %d %q", record.Code, record.Body.String())
	}
}

func TestTCPManagerRoundTripWithRealLocalTarget(t *testing.T) {
	target, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	go func() {
		connection, acceptErr := target.Accept()
		if acceptErr == nil {
			defer connection.Close()
			data := make([]byte, 3)
			_, _ = io.ReadFull(connection, data)
			_, _ = connection.Write(data)
		}
	}()
	manager := relay.NewTCPManager()
	manager.SetMaxConnections(4)
	ready := make(chan protocol.TCPData, 1)
	port, err := manager.Open("tcp-integration", func(ctx context.Context, envelope protocol.Envelope) error {
		var data protocol.TCPData
		if err := protocol.DecodePayload(envelope, &data); err != nil {
			return err
		}
		ready <- data
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.CloseTunnel("tcp-integration")
	client, err := net.Dial("tcp", net.JoinHostPort("127.0.0.1", fmt.Sprint(port)))
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()
	_, _ = client.Write([]byte("tcp"))
	data := <-ready
	targetConn, err := net.Dial("tcp", target.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	_, _ = targetConn.Write([]byte("tcp"))
	response := make([]byte, 3)
	_, _ = io.ReadFull(targetConn, response)
	_ = targetConn.Close()
	if err := manager.Write("tcp-integration", data.ConnectionID, response); err != nil {
		t.Fatal(err)
	}
	actual := make([]byte, 3)
	if _, err := io.ReadFull(client, actual); err != nil {
		t.Fatal(err)
	}
	if string(actual) != "tcp" {
		t.Fatalf("unexpected tcp response %q", actual)
	}
}

func TestUDPManagerRoundTripWithRealLocalTarget(t *testing.T) {
	target, err := net.ListenUDP("udp", &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 0})
	if err != nil {
		t.Fatal(err)
	}
	defer target.Close()
	manager := relay.NewUDPManager()
	port, err := manager.Open("udp-integration", func(ctx context.Context, envelope protocol.Envelope) error {
		var data protocol.UDPData
		if err := protocol.DecodePayload(envelope, &data); err != nil {
			return err
		}
		payload, err := base64.StdEncoding.DecodeString(data.Data)
		if err != nil {
			return err
		}
		if _, err := target.WriteToUDP(payload, target.LocalAddr().(*net.UDPAddr)); err != nil {
			return err
		}
		buffer := make([]byte, 32)
		count, address, err := target.ReadFromUDP(buffer)
		if err != nil {
			return err
		}
		return manager.Write("udp-integration", protocol.UDPResponse{TunnelID: "udp-integration", PacketID: data.PacketID, Data: base64.StdEncoding.EncodeToString(buffer[:count]), TargetAddress: address.IP.String(), TargetPort: address.Port})
	})
	if err != nil {
		t.Fatal(err)
	}
	defer manager.CloseTunnel("udp-integration")
	connection, err := net.DialUDP("udp", nil, &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: port})
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	_ = connection.SetReadDeadline(time.Now().Add(time.Second))
	if _, err := connection.Write([]byte("udp")); err != nil {
		t.Fatal(err)
	}
	buffer := make([]byte, 16)
	if _, _, err := connection.ReadFromUDP(buffer); err != nil {
		t.Fatal(err)
	}
	if string(buffer[:3]) != "udp" {
		t.Fatalf("unexpected udp response %q", buffer[:3])
	}
}
