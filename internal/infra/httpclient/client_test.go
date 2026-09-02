package httpclient

import (
	"testing"
	"time"
)

func TestNewUsesDefaultTimeout(t *testing.T) {
	client := New(0)

	if client.Timeout != DefaultTimeout {
		t.Fatalf("timeout = %s, want %s", client.Timeout, DefaultTimeout)
	}
}

func TestNewUsesExplicitTimeout(t *testing.T) {
	want := 3 * time.Second
	client := New(want)

	if client.Timeout != want {
		t.Fatalf("timeout = %s, want %s", client.Timeout, want)
	}
}
