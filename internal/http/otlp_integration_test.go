package http

import (
	"bytes"
	"compress/gzip"
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/handlers"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func TestOTLPHTTPEncodingTenantAndLimits(t *testing.T) {
	stack := newVerificationStack(t)
	if err := stack.db.AutoMigrate(&models.TelemetryMetric{}); err != nil {
		t.Fatal(err)
	}
	svc := services.NewObservabilityService(repositories.NewGormObservabilityRepository(stack.db))
	handler := handlers.NewObservabilityHandler(svc)
	app := fiber.New(fiber.Config{BodyLimit: services.MaxOTLPBodyBytes + 1024})
	app.Post("/metrics", ingestionRequired(stack.keys, stack.auth, stack.organizations), handler.IngestOTLPMetrics)
	expires := time.Now().Add(time.Hour)
	raw, _, err := stack.keys.CreateForOrganization(context.Background(), stack.userID, stack.organizationID, "OTLP", []string{"telemetry:write"}, &expires, "")
	if err != nil {
		t.Fatal(err)
	}
	body := []byte(`{"resourceMetrics":[{"resource":{"attributes":[{"key":"organization_id","value":{"stringValue":"foreign-org"}}]},"scopeMetrics":[{"metrics":[{"name":"requests","sum":{"dataPoints":[{"asInt":"9007199254740993","timeUnixNano":"1000000000"}]}}]}]}]}`)
	var compressed bytes.Buffer
	writer := gzip.NewWriter(&compressed)
	if _, err := writer.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("POST", "/metrics", &compressed)
	request.Header.Set("Authorization", "Bearer "+raw)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "GZip")
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		t.Fatalf("gzip JSON ingestion returned %d", response.StatusCode)
	}
	rows, err := svc.ListMetrics(context.Background(), stack.organizationID, "", 10)
	if err != nil || len(rows) != 1 || !bytes.Contains([]byte(rows[0].Data), []byte("9007199254740993")) {
		t.Fatalf("metric tenant or integer precision lost: %v", err)
	}
	rows, err = svc.ListMetrics(context.Background(), "foreign-org", "", 10)
	if err != nil || len(rows) != 0 {
		t.Fatalf("resource attribute overrode authenticated tenant: %v", err)
	}
	compressed.Reset()
	writer = gzip.NewWriter(&compressed)
	if _, err := writer.Write(bytes.Repeat([]byte("x"), services.MaxOTLPBodyBytes+1)); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request = httptest.NewRequest("POST", "/metrics", &compressed)
	request.Header.Set("Authorization", "Bearer "+raw)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Content-Encoding", "gzip")
	response, err = app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	if err := response.Body.Close(); err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != 413 {
		t.Fatalf("decompression limit returned %d", response.StatusCode)
	}
	for _, tc := range []struct {
		media  string
		body   []byte
		status int
	}{{"application/x-protobuf", nil, 200}, {"text/plain", []byte("x"), 415}, {"application/x-protobuf", []byte{255}, 400}} {
		request := httptest.NewRequest("POST", "/metrics", bytes.NewReader(tc.body))
		request.Header.Set("Authorization", "Bearer "+raw)
		request.Header.Set("Content-Type", tc.media)
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		if err := response.Body.Close(); err != nil {
			t.Fatal(err)
		}
		if response.StatusCode != tc.status {
			t.Fatalf("%s: expected %d, got %d", tc.media, tc.status, response.StatusCode)
		}
	}
}
