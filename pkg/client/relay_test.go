package client

import (
	"fmt"
	"testing"
)

func TestRelayOwnerFromError(t *testing.T) {

	if owner := RelayOwnerFromError(nil); owner != "" {
		t.Fatalf("expected no owner for a nil error, got %q", owner)
	}

	if owner := RelayOwnerFromError(fmt.Errorf("connect relay: connection refused")); owner != "" {
		t.Fatalf("expected no owner for unrelated errors, got %q", owner)
	}

	rejected := &RelayRejectedError{Message: "tunnel is connected through relay relay-abc"}

	if owner := RelayOwnerFromError(rejected); owner != "relay-abc" {
		t.Fatalf("expected owning relay relay-abc, got %q", owner)
	}

	if owner := RelayOwnerFromError(&RelayRejectedError{Message: "tunnel capacity reached"}); owner != "" {
		t.Fatalf("expected no owner for non-handoff rejections, got %q", owner)
	}
}
