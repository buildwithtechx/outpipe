package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"outpipe.dev/outpipe/pkg/protocol"
)

func TestServeLocalContextCancellationClosesIdleRelay(t *testing.T) {
	processed := make(chan struct{})
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		connection, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Error(err)
			return
		}
		defer connection.Close()
		if err := connection.SetReadDeadline(time.Now().Add(3 * time.Second)); err != nil {
			t.Error(err)
			return
		}
		payload, err := protocol.EncodePayload(protocol.MessageTypeHTTPRequest, "request", protocol.HTTPRequest{Method: http.MethodGet, Path: "/"})
		if err != nil {
			t.Error(err)
			return
		}
		if err := connection.WriteMessage(websocket.TextMessage, payload); err != nil {
			t.Error(err)
			return
		}
		_, data, err := connection.ReadMessage()
		if err != nil {
			t.Error(err)
			return
		}
		message, err := protocol.Decode(data)
		if err != nil || message.Type != protocol.MessageTypeHTTPResponse {
			t.Errorf("relay frame was not processed: %v", err)
			return
		}
		close(processed)
		_, _, _ = connection.ReadMessage()
	}))
	defer server.Close()
	connection, _, err := websocket.DefaultDialer.Dial("ws"+strings.TrimPrefix(server.URL, "http"), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()
	client := &RelayConnection{conn: connection}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan error, 1)
	go func() { done <- client.ServeLocal(ctx, target.URL) }()
	select {
	case <-processed:
	case <-time.After(2 * time.Second):
		t.Fatal("relay frame was not processed before cancellation")
	}
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("idle relay remained blocked after cancellation")
	}
}
