package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/engine"
)

func TestCaptureSpoolSurvivesRestart(t *testing.T) {
	var mu sync.Mutex
	received := map[string]bool{}
	failing := atomic.Bool{}
	failing.Store(true)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		var capture engine.RequestCapture
		if err := json.NewDecoder(request.Body).Decode(&capture); err != nil {
			t.Error(err)
		}
		if failing.Load() {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		mu.Lock()
		received[capture.ID] = true
		mu.Unlock()
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()
	dir := t.TempDir()
	first := &captureRecorder{baseURL: server.URL, secret: "secret", client: server.Client()}
	if err := first.setSpoolDir(dir); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		if err := first.RecordCapture(context.Background(), engine.RequestCapture{OrganizationID: "org", TunnelID: "tunnel", Method: "GET", Path: "/", StatusCode: 200}); err != nil {
			t.Fatal(err)
		}
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 3 {
		t.Fatalf("spool holds %d files", len(entries))
	}
	second := &captureRecorder{baseURL: server.URL, secret: "secret", client: server.Client(), sweepInterval: 20 * time.Millisecond}
	if err := second.setSpoolDir(dir); err != nil {
		t.Fatal(err)
	}
	second.spool.delay = func(int) time.Duration { return 20 * time.Millisecond }
	ctx, cancel := context.WithCancel(context.Background())
	second.start(ctx)
	failing.Store(false)
	deadline := time.Now().Add(10 * time.Second)
	for {
		mu.Lock()
		count := len(received)
		mu.Unlock()
		if count == 3 {
			break
		}
		if time.Now().After(deadline) {
			cancel()
			second.wait.Wait()
			t.Fatalf("spool replay delivered %d of 3", count)
		}
		time.Sleep(20 * time.Millisecond)
	}
	cancel()
	second.wait.Wait()
	if left, err := filepath.Glob(filepath.Join(dir, "*.json")); err != nil || len(left) != 0 {
		t.Fatalf("spool left %d files: %v", len(left), err)
	}
	if second.dropped.Load() != 0 {
		t.Fatal("durable replay recorded drops")
	}
}

func TestCaptureSpoolBoundsDrop(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusCreated) }))
	defer server.Close()
	recorder := &captureRecorder{baseURL: server.URL, secret: "secret", client: server.Client()}
	if err := recorder.setSpoolDir(t.TempDir()); err != nil {
		t.Fatal(err)
	}
	recorder.spool.maxFiles = 1
	if err := recorder.RecordCapture(context.Background(), engine.RequestCapture{OrganizationID: "org"}); err != nil {
		t.Fatal(err)
	}
	if err := recorder.RecordCapture(context.Background(), engine.RequestCapture{OrganizationID: "org"}); err == nil {
		t.Fatal("overflowing spool accepted")
	}
	if recorder.dropped.Load() != 1 {
		t.Fatalf("spool overflow drops = %d", recorder.dropped.Load())
	}
}
