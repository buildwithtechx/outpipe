package engine

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"outpipe.dev/outpipe/pkg/protocol"
)

type memoryCaptureRecorder struct {
	enabled  bool
	captures []RequestCapture
}

func (r *memoryCaptureRecorder) Enabled(context.Context, string, string) (bool, error) {
	return r.enabled, nil
}
func (r *memoryCaptureRecorder) RecordCapture(_ context.Context, capture RequestCapture) error {
	r.captures = append(r.captures, capture)
	return nil
}

func TestCaptureToggleAppliesToConnectedTunnel(t *testing.T) {
	sessions := NewSessionRegistry()
	router, err := NewRequestRouter(sessions, time.Second)
	if err != nil {
		t.Fatal(err)
	}
	if err := sessions.Reserve(Session{ID: "session", OrganizationID: "org", TunnelID: "tunnel", Send: func(_ context.Context, request protocol.Envelope) error {
		payload, err := json.Marshal(protocol.HTTPResponse{StatusCode: 200, Headers: map[string][]string{"Set-Cookie": {"secret"}}, Body: ""})
		if err != nil {
			return err
		}
		response := protocol.Envelope{Type: protocol.MessageTypeHTTPResponse, RequestID: request.RequestID, Payload: payload}
		if router.HandleOwned(response, "other-org", map[string]string{"tunnel": "session"}) {
			t.Error("foreign tenant supplied response")
		}
		if !router.HandleOwned(response, "org", map[string]string{"tunnel": "session"}) {
			t.Error("owner response rejected")
		}
		return nil
	}}, false); err != nil {
		t.Fatal(err)
	}
	proxy, err := NewHTTPProxy("outpipe.localhost", router, 4096)
	if err != nil {
		t.Fatal(err)
	}
	recorder := &memoryCaptureRecorder{}
	proxy.SetCaptureRecorder(recorder)
	for _, enabled := range []bool{false, true, false} {
		recorder.enabled = enabled
		request := httptest.NewRequest("POST", "http://tunnel.outpipe.localhost/example?token=private", strings.NewReader(`{"token":"private","value":"safe"}`))
		request.Header.Set("Authorization", "Bearer private")
		response := httptest.NewRecorder()
		proxy.ServeHTTP(response, request)
		if response.Code != 200 {
			t.Fatalf("HTTP %d", response.Code)
		}
	}
	if len(recorder.captures) != 1 {
		t.Fatalf("expected 1 capture, got %d", len(recorder.captures))
	}
	capture := recorder.captures[0]
	if capture.OrganizationID != "org" || capture.TunnelID != "tunnel" || capture.Path != "/example" {
		t.Fatalf("unexpected capture metadata: %+v", capture)
	}
	if strings.Contains(capture.RequestHeaders+capture.ResponseHeaders+capture.RequestBody, "private") || strings.Contains(capture.ResponseHeaders, "secret") {
		t.Fatal("credentials persisted in capture")
	}
}
