package config

import "testing"

func productionApp() AppConfig {
	return AppConfig{
		Port:           "8080",
		Environment:    "production",
		PublicAPIURL:   "https://api.outpipe.dev",
		DashboardURL:   "https://outpipe.dev",
		AllowedOrigins: "https://outpipe.dev",
	}
}

func productionAuth() AuthConfig {
	return AuthConfig{CookieSecure: true, EncryptionKey: "0123456789abcdef0123456789abcdef"}
}

func TestValidateProductionAppRejectsInsecureURLs(t *testing.T) {
	app := productionApp()
	app.PublicAPIURL = "http://api.outpipe.dev"

	if err := validateAPIApp(app); err == nil {
		t.Fatal("validateApp succeeded with an insecure production URL")
	}
}

func TestValidateProductionAuthRequiresSecureSecrets(t *testing.T) {
	app := productionApp()
	auth := productionAuth()
	auth.CookieSecure = false
	if err := validateAuth(auth, app); err == nil {
		t.Fatal("validateAuth succeeded with insecure cookies")
	}

	auth = productionAuth()
	auth.EncryptionKey = "short"
	if err := validateAuth(auth, app); err == nil {
		t.Fatal("validateAuth succeeded with a short encryption key")
	}
}

func TestValidateProductionServiceRequiresSecret(t *testing.T) {
	app := productionApp()
	if err := validateService(ServiceConfig{}, app); err == nil {
		t.Fatal("validateService succeeded without an internal secret")
	}
	if err := validateService(ServiceConfig{InternalAPISecret: "01234567890123456789012345678901"}, app); err != nil {
		t.Fatalf("validateService rejected a valid secret: %v", err)
	}
}

func TestValidateDevelopmentDefaultsRemainAllowed(t *testing.T) {
	app := AppConfig{Port: "8080", Environment: "development", PublicAPIURL: "http://localhost:8080", DashboardURL: "http://localhost:3000"}
	if err := validateApp(app); err != nil {
		t.Fatalf("validateApp rejected development defaults: %v", err)
	}
}
