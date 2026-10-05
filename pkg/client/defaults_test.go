package client

import "testing"

func TestClientDefaultsToHostedAPI(t *testing.T) {
	client, err := New(Config{APIKey: "scoped-key"})
	if err != nil {
		t.Fatalf("create hosted client: %v", err)
	}
	if client.baseURL != "https://api.outpipe.dev" {
		t.Fatal("hosted API default was not applied")
	}
}
