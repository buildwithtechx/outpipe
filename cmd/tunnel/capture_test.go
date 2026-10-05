package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/engine"
	"outpipe.dev/outpipe/internal/relay"
)

func TestCaptureRejectsRedirectWithoutForwardingPrivateData(t *testing.T) {
	var calls atomic.Int32
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { calls.Add(1); w.WriteHeader(201) }))
	defer target.Close()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) { http.Redirect(w, request, target.URL, 307) }))
	defer server.Close()
	recorder := &captureRecorder{baseURL: server.URL, secret: "private", client: server.Client()}
	if err := recorder.sendCapture(context.Background(), engine.RequestCapture{RequestBody: "private"}); err == nil {
		t.Fatal("redirect accepted")
	}
	if calls.Load() != 0 {
		t.Fatal("capture credentials followed redirect")
	}
}

func TestCaptureQueueDoesNotBlockRequestsOnSlowControlPlane(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		select {
		case <-release:
		case <-request.Context().Done():
			return
		}
		if _, err := w.Write([]byte(`{"organizationId":"org","status":"active","captureEnabled":true}`)); err != nil {
			t.Error(err)
		}
	}))
	defer server.Close()
	defer close(release)
	resolver, err := relay.NewInternalTunnelResolver(server.URL, "secret", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	recorder := &captureRecorder{resolver: resolver, baseURL: server.URL, secret: "secret", client: server.Client()}
	ctx, cancel := context.WithCancel(context.Background())
	recorder.start(ctx)
	defer func() { cancel(); recorder.wait.Wait() }()
	done := make(chan struct{})
	go func() {
		if _, err := recorder.Enabled(ctx, "tunnel", "org"); err != nil {
			t.Error(err)
		}
		for range 200 {
			if err := recorder.RecordCapture(ctx, engine.RequestCapture{}); err != nil {
				t.Error(err)
			}
		}
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("capture blocked tunnel traffic")
	}
}
