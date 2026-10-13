package http

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"outpipe.dev/outpipe/internal/handlers"
	"outpipe.dev/outpipe/internal/infra/storage"
	"outpipe.dev/outpipe/internal/models"
	"outpipe.dev/outpipe/internal/repositories"
	"outpipe.dev/outpipe/internal/services"
)

func TestDNSProviderCredentialEndpointsHideSecrets(t *testing.T) {
	stack := newVerificationStack(t)
	if err := stack.db.AutoMigrate(&models.Domain{}, &models.DNSProviderCredential{}); err != nil {
		t.Fatal(err)
	}
	domainRepo, err := repositories.NewDomainRepository(stack.db)
	if err != nil {
		t.Fatal(err)
	}
	dnsRepo, err := repositories.NewDNSProviderRepository(stack.db)
	if err != nil {
		t.Fatal(err)
	}
	domains, err := services.NewDomainService(domainRepo)
	if err != nil {
		t.Fatal(err)
	}
	credentials, err := services.NewDNSCredentialService(dnsRepo)
	if err != nil {
		t.Fatal(err)
	}
	secretBox, err := storage.NewSecretBox([]byte("0123456789abcdef0123456789abcdef"))
	if err != nil {
		t.Fatal(err)
	}
	credentials.SetProtector(secretBox)
	domains.SetDNSCredentials(credentials)
	handler, err := handlers.NewDomainHandler(domains)
	if err != nil {
		t.Fatal(err)
	}
	app := fiber.New()
	app.Post("/organizations/:organizationID/dns-providers", handler.StoreDNSProvider)
	app.Get("/organizations/:organizationID/dns-providers", handler.ListDNSProviders)
	app.Delete("/organizations/:organizationID/dns-providers/:credentialID", handler.DeleteDNSProvider)
	post := func(body string) (int, string) {
		request := httptest.NewRequest("POST", "/organizations/"+stack.organizationID+"/dns-providers", strings.NewReader(body))
		request.Header.Set("Content-Type", "application/json")
		response, err := app.Test(request)
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		raw, err := io.ReadAll(response.Body)
		if err != nil {
			t.Fatal(err)
		}
		return response.StatusCode, string(raw)
	}
	if status, _ := post(`{"provider":"route53","token":"x"}`); status != 400 {
		t.Fatalf("unsupported provider returned %d", status)
	}
	status, body := post(`{"provider":"cloudflare","token":"cf-test-token-1234","label":"primary"}`)
	if status != 201 {
		t.Fatalf("credential creation returned %d: %s", status, body)
	}
	if strings.Contains(body, "cf-test-token") || strings.Contains(body, "ciphertext") || strings.Contains(body, `"token"`) {
		t.Fatalf("credential secret leaked in response: %s", body)
	}
	if !strings.Contains(body, "1234") {
		t.Fatalf("credential suffix hint missing: %s", body)
	}
	var created struct {
		ID string `json:"id"`
	}
	if err := json.Unmarshal([]byte(body), &created); err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest("GET", "/organizations/"+stack.organizationID+"/dns-providers", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	var listed []map[string]any
	if err := json.NewDecoder(response.Body).Decode(&listed); err != nil {
		t.Fatal(err)
	}
	if len(listed) != 1 || listed[0]["provider"] != "cloudflare" {
		t.Fatalf("credential list returned %v", listed)
	}
	for key := range listed[0] {
		if key == "ciphertext" || key == "token" {
			t.Fatalf("credential secret leaked in list: %v", listed[0])
		}
	}
	request = httptest.NewRequest("DELETE", "/organizations/"+stack.organizationID+"/dns-providers/"+created.ID, nil)
	response, err = app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 204 {
		t.Fatalf("credential deletion returned %d", response.StatusCode)
	}
	request = httptest.NewRequest("DELETE", "/organizations/"+stack.organizationID+"/dns-providers/"+created.ID, nil)
	response, err = app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()
	if response.StatusCode != 404 {
		t.Fatalf("repeat deletion returned %d", response.StatusCode)
	}
}
