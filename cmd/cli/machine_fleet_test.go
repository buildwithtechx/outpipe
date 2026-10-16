package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRunMachineFleetCommands(t *testing.T) {
	var seen []string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		if r.Header.Get("Authorization") != "Bearer key-1" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/organizations/org-1/machine-fleets":
			_, _ = w.Write([]byte(`{"id":"fleet-1","name":"edge"}`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/organizations/org-1/machine-fleets":
			_, _ = w.Write([]byte(`[{"id":"fleet-1","name":"edge"}]`))
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/organizations/org-1/machine-fleets/fleet-1":
			_, _ = w.Write([]byte(`{"fleet":{"id":"fleet-1","name":"edge"},"tokens":[],"desiredTunnels":[],"liveTunnels":[],"drift":[{"tunnelId":"t-1","name":"node-a","want":"active","got":"disconnected"}]}`))
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/organizations/org-1/machine-fleets/fleet-1/reconcile":
			_, _ = w.Write([]byte(`{"reactivated":1,"recreated":0,"skipped":0}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()
	cliTestEnv(t, server.URL)
	t.Setenv("OUTPIPE_API_KEY", "key-1")

	for _, args := range [][]string{
		{"machine-fleet", "create", "--organization", "org-1", "edge"},
		{"machine-fleet", "list", "--organization", "org-1"},
		{"machine-fleet", "status", "--organization", "org-1", "fleet-1"},
		{"machine-fleet", "reconcile", "--organization", "org-1", "fleet-1"},
	} {
		if err := run(args); err != nil {
			t.Fatalf("%v: %v", args, err)
		}
	}
	want := []string{
		"POST /api/v1/organizations/org-1/machine-fleets",
		"GET /api/v1/organizations/org-1/machine-fleets",
		"GET /api/v1/organizations/org-1/machine-fleets/fleet-1",
		"POST /api/v1/organizations/org-1/machine-fleets/fleet-1/reconcile",
	}
	if len(seen) != len(want) {
		t.Fatalf("requests = %#v", seen)
	}
	for i := range want {
		if seen[i] != want[i] {
			t.Errorf("request[%d] = %q, want %q", i, seen[i], want[i])
		}
	}
}

func TestRunMachineFleetStatusJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"fleet":{"id":"fleet-1","name":"edge"},"tokens":[],"desiredTunnels":[],"liveTunnels":[],"drift":[]}`))
	}))
	defer server.Close()
	cliTestEnv(t, server.URL)
	t.Setenv("OUTPIPE_API_KEY", "key-1")
	output := captureStdout(t, func() {
		if err := run([]string{"machine-fleet", "status", "--organization", "org-1", "fleet-1", "--json"}); err != nil {
			t.Fatal(err)
		}
	})
	if !strings.Contains(output, `"name": "edge"`) {
		t.Errorf("json output missing fleet: %s", output)
	}
}
