package main

import (
	"context"
	"testing"
	"time"

	"outpipe.dev/outpipe/internal/engine"
	"outpipe.dev/outpipe/internal/validation"
)

func TestRecordCaptureHonorsPolicyRules(t *testing.T) {
	rules, err := validation.ParseCaptureRules(`[{"id":"deny-admin","methods":["GET"],"pathPrefix":"/admin","capture":false}]`)
	if err != nil {
		t.Fatal(err)
	}
	recorder := &captureRecorder{
		policies: map[string]capturePolicy{
			"tunnel-1": {enabled: true, expires: time.Now().Add(time.Minute), orgID: "org", rules: rules},
		},
		queue: make(chan captureUpload, 128),
	}
	denied := engine.RequestCapture{ID: "cap-denied", OrganizationID: "org", TunnelID: "tunnel-1", Method: "GET", Path: "/admin/users"}
	if err := recorder.RecordCapture(context.Background(), denied); err != nil {
		t.Fatalf("denied capture must be skipped silently: %v", err)
	}
	if len(recorder.queue) != 0 {
		t.Fatal("denied capture was queued")
	}
	if recorder.dropped.Load() != 0 {
		t.Fatal("denied capture counted as a drop")
	}
	allowed := engine.RequestCapture{ID: "cap-allowed", OrganizationID: "org", TunnelID: "tunnel-1", Method: "GET", Path: "/api/users"}
	if err := recorder.RecordCapture(context.Background(), allowed); err != nil {
		t.Fatal(err)
	}
	if len(recorder.queue) != 0 {
		t.Fatal("unmatched path with rules present must be denied (fail-closed)")
	}
	// Add an allow rule for /api and verify it passes
	rules2, _ := validation.ParseCaptureRules(`[{"id":"deny-admin","methods":["GET"],"pathPrefix":"/admin","capture":false},{"id":"allow-api","methods":["GET"],"pathPrefix":"/api","capture":true}]`)
	recorder2 := &captureRecorder{
		policies: map[string]capturePolicy{
			"tunnel-2": {enabled: true, expires: time.Now().Add(time.Minute), orgID: "org", rules: rules2},
		},
		queue: make(chan captureUpload, 128),
	}
	allowed2 := engine.RequestCapture{ID: "cap-allowed2", OrganizationID: "org", TunnelID: "tunnel-2", Method: "GET", Path: "/api/users"}
	if err := recorder2.RecordCapture(context.Background(), allowed2); err != nil {
		t.Fatal(err)
	}
	if len(recorder2.queue) != 1 {
		t.Fatal("explicit allow rule must permit capture")
	}
	stale := &captureRecorder{
		policies: map[string]capturePolicy{
			"tunnel-1": {enabled: true, expires: time.Now().Add(-time.Minute), orgID: "org", rules: rules},
		},
		queue: make(chan captureUpload, 128),
	}
	if err := stale.RecordCapture(context.Background(), denied); err != nil {
		t.Fatal(err)
	}
	if len(stale.queue) != 1 {
		t.Fatal("stale policy must fail open to the server check")
	}
}
