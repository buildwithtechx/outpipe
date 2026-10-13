package dns

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type cloudflareFake struct {
	mu      sync.Mutex
	zones   map[string]string
	records map[string]map[string]string
	next    int
}

func newCloudflareFake() *cloudflareFake {
	return &cloudflareFake{zones: map[string]string{"example.com": "zone-1"}, records: map[string]map[string]string{}}
}

func (f *cloudflareFake) handler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		write := func(result any) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "errors": []any{}, "result": result})
		}
		path := request.URL.Path
		if request.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("cloudflare token missing")
		}
		switch {
		case request.Method == http.MethodGet && strings.HasSuffix(path, "/zones"):
			name := request.URL.Query().Get("name")
			result := []map[string]string{}
			if id, ok := f.zones[name]; ok {
				result = append(result, map[string]string{"id": id, "name": name})
			}
			write(result)
		case request.Method == http.MethodGet && strings.Contains(path, "dns_records"):
			name := request.URL.Query().Get("name")
			result := []map[string]string{}
			for id, content := range f.records[name] {
				result = append(result, map[string]string{"id": id, "content": content})
			}
			write(result)
		case request.Method == http.MethodPost && strings.Contains(path, "dns_records"):
			var input struct {
				Type    string `json:"type"`
				Name    string `json:"name"`
				Content string `json:"content"`
			}
			if err := json.NewDecoder(request.Body).Decode(&input); err != nil {
				t.Error(err)
			}
			if input.Type != "TXT" {
				t.Errorf("challenge record must be TXT, got %s", input.Type)
			}
			f.next++
			id := strings.Repeat("0", 31) + string(rune('1'+f.next))
			if f.records[input.Name] == nil {
				f.records[input.Name] = map[string]string{}
			}
			f.records[input.Name][id] = input.Content
			write(map[string]string{"id": id})
		case request.Method == http.MethodDelete && strings.Contains(path, "dns_records"):
			parts := strings.Split(path, "/")
			id := parts[len(parts)-1]
			for name, records := range f.records {
				delete(records, id)
				if len(records) == 0 {
					delete(f.records, name)
				}
			}
			write(map[string]string{"id": id})
		default:
			t.Errorf("unexpected cloudflare call %s %s", request.Method, path)
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func TestCloudflarePresentAndCleanup(t *testing.T) {
	fake := newCloudflareFake()
	server := httptest.NewServer(fake.handler(t))
	defer server.Close()
	provider, err := NewCloudflareProvider("test-token", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	provider.baseURL = server.URL
	ctx := context.Background()
	if err := provider.Present(ctx, "_acme-challenge.api.example.com", "digest-a"); err != nil {
		t.Fatal(err)
	}
	if err := provider.Present(ctx, "_acme-challenge.api.example.com", "digest-a"); err != nil {
		t.Fatal(err)
	}
	fake.mu.Lock()
	count := len(fake.records["_acme-challenge.api.example.com"])
	fake.mu.Unlock()
	if count != 1 {
		t.Fatalf("duplicate challenge record was not deduplicated: %d", count)
	}
	if err := provider.Present(ctx, "_acme-challenge.api.example.com", "digest-b"); err != nil {
		t.Fatal(err)
	}
	if err := provider.Cleanup(ctx, "_acme-challenge.api.example.com"); err != nil {
		t.Fatal(err)
	}
	fake.mu.Lock()
	remaining := len(fake.records)
	fake.mu.Unlock()
	if remaining != 0 {
		t.Fatalf("cleanup left %d records", remaining)
	}
	if provider.Name() != "cloudflare" {
		t.Fatal("provider name must be cloudflare")
	}
	if err := provider.Present(ctx, "api.other.example", "x"); err == nil {
		t.Fatal("record outside the provider zone was accepted")
	}
}
