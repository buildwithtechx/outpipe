package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestServeLocalContextCancellationClosesIdleRelay(t *testing.T) {
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
	go func() { done <- client.ServeLocal(ctx, "http://localhost:3000") }()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("idle relay remained blocked after cancellation")
	}
}
