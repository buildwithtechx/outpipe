package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckAPIUsesAuthenticatedInternalReadiness(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/internal/health" || r.Header.Get("X-Internal-Secret") != "test-secret" {
			t.Error("incorrect readiness request")
			w.WriteHeader(404)
			return
		}
		w.WriteHeader(200)
	}))
	defer server.Close()
	if result := checkAPI(context.Background(), server.URL, "test-secret"); result.Status != "ok" {
		t.Fatalf("readiness failed: %s", result.Error)
	}
}
