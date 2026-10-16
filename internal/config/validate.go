package config

import (
	"fmt"
	"net/mail"
	"net/url"
	"strings"
)

func validateAPI(cfg APIConfig) error {
	if err := validateAPIApp(cfg.App); err != nil {
		return err
	}
	if err := validateDatabase(cfg.Database); err != nil {
		return err
	}
	if strings.EqualFold(cfg.App.Environment, "production") && isDevelopmentDatabase(cfg.Database.URL) {
		return fmt.Errorf("production database url must not use development credentials or localhost")
	}
	if err := validateService(cfg.Service, cfg.App); err != nil {
		return err
	}
	return validateAuth(cfg.Auth, cfg.App)
}

func validateRelay(cfg RelayConfig) error {
	if err := validateApp(cfg.App); err != nil {
		return err
	}
	if cfg.Tunnel.MaxConnections < 1 {
		return fmt.Errorf("tunnel max connections must be positive")
	}
	if cfg.Tunnel.MaxTunnels < 1 || cfg.Tunnel.MaxFrameBytes < 1 {
		return fmt.Errorf("tunnel limits must be positive")
	}
	if cfg.Tunnel.Heartbeat <= 0 || cfg.Tunnel.ReadTimeout <= cfg.Tunnel.Heartbeat {
		return fmt.Errorf("tunnel heartbeat and read timeout are invalid")
	}
	if err := validateService(cfg.Service, cfg.App); err != nil {
		return err
	}
	return nil
}

func validateApp(cfg AppConfig) error {
	if cfg.Port == "" {
		return fmt.Errorf("port is required")
	}
	if cfg.PublicAPIURL == "" {
		return fmt.Errorf("public api url is required")
	}
	if cfg.DashboardURL == "" {
		return fmt.Errorf("dashboard url is required")
	}
	if cfg.RequireTLS && (cfg.TLSCertFile == "" || cfg.TLSKeyFile == "") {
		return fmt.Errorf("tls certificate and key files are required when tls is enabled")
	}
	return nil
}

func validateAPIApp(cfg AppConfig) error {
	if err := validateApp(cfg); err != nil {
		return err
	}
	if cfg.ACMEEnabled {
		address, err := mail.ParseAddress(cfg.ACMEEmail)
		if err != nil {
			return fmt.Errorf("ACME requires a valid contact email: %w", err)
		}
		if address.Address != cfg.ACMEEmail {
			return fmt.Errorf("ACME contact email must be a bare email address")
		}
	}
	if !strings.EqualFold(cfg.Environment, "production") {
		return nil
	}
	for name, rawURL := range map[string]string{"public api url": cfg.PublicAPIURL, "dashboard url": cfg.DashboardURL, "cors origin": cfg.CORSOrigin} {
		parsed, err := url.Parse(rawURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("%s must use an https URL in production", name)
		}
	}
	if cfg.AllowedOrigins != "" {
		for _, origin := range strings.Split(cfg.AllowedOrigins, ",") {
			origin = strings.TrimSpace(origin)
			if origin == "" {
				continue
			}
			if strings.Contains(strings.ToLower(origin), "localhost") || strings.Contains(strings.ToLower(origin), "127.0.0.1") {
				return fmt.Errorf("localhost origins are not allowed in production")
			}
			parsed, err := url.Parse(origin)
			if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
				return fmt.Errorf("allowed origin %q must use an https URL in production", origin)
			}
		}
	}
	return nil
}

func validateService(cfg ServiceConfig, app AppConfig) error {
	if !strings.EqualFold(app.Environment, "production") {
		return nil
	}
	if len(cfg.InternalAPISecret) < 32 || isDevelopmentSecret(cfg.InternalAPISecret) {
		return fmt.Errorf("internal api secret must be at least 32 characters in production")
	}
	return nil
}

func isDevelopmentSecret(value string) bool {
	value = strings.ToLower(strings.TrimSpace(value))
	return value == "development-secret" || value == "change-me" || strings.Contains(value, "changeme") || strings.Contains(value, "generate_a_secure_random_string_here")
}

func isDevelopmentDatabase(raw string) bool {
	value := strings.ToLower(strings.TrimSpace(raw))
	return strings.Contains(value, "localhost") || strings.Contains(value, "127.0.0.1") || strings.Contains(value, "outpipe:outpipe")
}

func validateAuth(cfg AuthConfig, app AppConfig) error {
	if !strings.EqualFold(app.Environment, "production") {
		return nil
	}
	if !cfg.CookieSecure {
		return fmt.Errorf("auth cookies must be secure in production")
	}
	if length := len([]byte(cfg.EncryptionKey)); length != 16 && length != 24 && length != 32 {
		return fmt.Errorf("auth encryption key must be 16, 24, or 32 bytes in production")
	}
	if (cfg.GoogleClientID == "") != (cfg.GoogleClientSecret == "") {
		return fmt.Errorf("google oauth client id and secret must be configured together")
	}
	if (cfg.GitHubClientID == "") != (cfg.GitHubClientSecret == "") {
		return fmt.Errorf("github oauth client id and secret must be configured together")
	}
	return nil
}

func validateDatabase(cfg DatabaseConfig) error {
	if cfg.URL == "" {
		return fmt.Errorf("database url is required")
	}
	if cfg.MaxConns < 1 {
		return fmt.Errorf("database max connections must be positive")
	}
	return nil
}
