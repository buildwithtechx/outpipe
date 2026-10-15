package dns

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type route53Fake struct {
	mu      sync.Mutex
	zones   map[string]string
	records map[string]map[string]bool
}

func newRoute53Fake() *route53Fake {
	return &route53Fake{zones: map[string]string{"example.com.": "Z123"}, records: map[string]map[string]bool{}}
}

func (f *route53Fake) handler(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if !strings.HasPrefix(request.Header.Get("Authorization"), "AWS4-HMAC-SHA256 Credential=AKID/") {
			t.Error("route53 request is not SigV4 signed")
		}
		if request.Header.Get("X-Amz-Date") == "" {
			t.Error("route53 request is missing Amz-Date")
		}
		write := func(body string) {
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(body))
		}
		path := request.URL.Path
		switch {
		case request.Method == http.MethodGet && path == "/2013-04-01/hostedzone":
			var zones strings.Builder
			for name, id := range f.zones {
				zones.WriteString(`<HostedZone><Id>/hostedzone/` + id + `</Id><Name>` + name + `</Name></HostedZone>`)
			}
			write(`<ListHostedZonesResponse xmlns="https://route53.amazonaws.com/doc/2013-04-01/"><HostedZones>` + zones.String() + `</HostedZones><IsTruncated>false</IsTruncated></ListHostedZonesResponse>`)
		case request.Method == http.MethodGet && strings.HasSuffix(path, "/rrset"):
			name := request.URL.Query().Get("name")
			var sets strings.Builder
			for value := range f.records[name] {
				sets.WriteString(`<ResourceRecord><Value>"` + value + `"</Value></ResourceRecord>`)
			}
			records := ""
			if len(f.records[name]) > 0 {
				records = `<ResourceRecordSet><Name>` + name + `</Name><Type>TXT</Type><TTL>120</TTL><ResourceRecords>` + sets.String() + `</ResourceRecords></ResourceRecordSet>`
			}
			write(`<ListResourceRecordSetsResponse xmlns="https://route53.amazonaws.com/doc/2013-04-01/"><ResourceRecordSets>` + records + `</ResourceRecordSets></ListResourceRecordSetsResponse>`)
		case request.Method == http.MethodPost && strings.HasSuffix(path, "/rrset"):
			raw, err := io.ReadAll(request.Body)
			if err != nil {
				t.Error(err)
			}
			var batch struct {
				XMLName xml.Name `xml:"ChangeResourceRecordSetsRequest"`
				Changes []struct {
					Action string `xml:"Action"`
					Set    struct {
						Name   string `xml:"Name"`
						Type   string `xml:"Type"`
						Values []struct {
							Value string `xml:"Value"`
						} `xml:"ResourceRecords>ResourceRecord"`
					} `xml:"ResourceRecordSet"`
				} `xml:"ChangeBatch>Changes>Change"`
			}
			if err := xml.Unmarshal(raw, &batch); err != nil {
				t.Error(err)
			}
			for _, change := range batch.Changes {
				name := strings.TrimSuffix(change.Set.Name, ".")
				switch change.Action {
				case "UPSERT":
					replaced := map[string]bool{}
					for _, record := range change.Set.Values {
						replaced[unquoteTXT(record.Value)] = true
					}
					f.records[name] = replaced
				case "DELETE":
					delete(f.records, name)
				default:
					t.Errorf("unexpected route53 action %s", change.Action)
				}
			}
			write(`<ChangeResourceRecordSetsResponse xmlns="https://route53.amazonaws.com/doc/2013-04-01/"><ChangeInfo><Id>/change/C1</Id><Status>PENDING</Status></ChangeResourceRecordSetsResponse>`)
		default:
			t.Errorf("unexpected route53 call %s %s", request.Method, path)
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

func TestRoute53PresentAndCleanupValue(t *testing.T) {
	fake := newRoute53Fake()
	server := httptest.NewServer(fake.handler(t))
	defer server.Close()
	provider, err := NewRoute53Provider("AKID", "SECRET", "", server.Client())
	if err != nil {
		t.Fatal(err)
	}
	provider.baseURL = server.URL
	ctx := context.Background()
	if err := provider.Present(ctx, "_acme-challenge.api.example.com", "digest-a"); err != nil {
		t.Fatal(err)
	}
	if err := provider.Present(ctx, "_acme-challenge.api.example.com", "digest-b"); err != nil {
		t.Fatal(err)
	}
	fake.mu.Lock()
	count := len(fake.records["_acme-challenge.api.example.com"])
	fake.mu.Unlock()
	if count != 2 {
		t.Fatalf("route53 RRset holds %d values, want both digests", count)
	}
	if err := provider.CleanupValue(ctx, "_acme-challenge.api.example.com", "digest-a"); err != nil {
		t.Fatal(err)
	}
	fake.mu.Lock()
	remaining := fake.records["_acme-challenge.api.example.com"]
	fake.mu.Unlock()
	if len(remaining) != 1 || !remaining["digest-b"] {
		t.Fatalf("value-targeted cleanup left %v", remaining)
	}
	if err := provider.Cleanup(ctx, "_acme-challenge.api.example.com"); err != nil {
		t.Fatal(err)
	}
	fake.mu.Lock()
	left := len(fake.records)
	fake.mu.Unlock()
	if left != 0 {
		t.Fatal("cleanup left records behind")
	}
	if provider.Name() != "route53" {
		t.Fatal("provider name must be route53")
	}
	if err := provider.Present(ctx, "host.other.example", "x"); err == nil {
		t.Fatal("record outside the hosted zone was accepted")
	}
}
