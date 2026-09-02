package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

func LoadAPI() (APIConfig, error) {
	var cfg APIConfig

	if err := parse(&cfg); err != nil {
		return APIConfig{}, err
	}

	if err := validateAPI(cfg); err != nil {
		return APIConfig{}, err
	}

	return cfg, nil
}

func LoadRelay() (RelayConfig, error) {
	var cfg RelayConfig

	if err := parse(&cfg); err != nil {
		return RelayConfig{}, err
	}

	if err := validateRelay(cfg); err != nil {
		return RelayConfig{}, err
	}

	return cfg, nil
}

func LoadCron() (CronConfig, error) {
	var cfg CronConfig

	if err := parse(&cfg); err != nil {
		return CronConfig{}, err
	}

	if err := validateDatabase(cfg.Database); err != nil {
		return CronConfig{}, err
	}

	return cfg, nil
}

func LoadCheck() (CheckConfig, error) {
	var cfg CheckConfig

	if err := parse(&cfg); err != nil {
		return CheckConfig{}, err
	}

	if err := validateApp(cfg.App); err != nil {
		return CheckConfig{}, err
	}

	return cfg, nil
}

func LoadCLI() (CLIConfig, error) {
	var cfg CLIConfig

	if err := parse(&cfg); err != nil {
		return CLIConfig{}, err
	}

	if cfg.APIURL == "" || cfg.RelayURL == "" {
		return CLIConfig{}, fmt.Errorf("tunnel api and relay urls are required")
	}

	return cfg, nil
}

func SaveCLI(cfg CLIConfig) error {

	if cfg.ConfigPath == "" {
		return fmt.Errorf("config path is required")
	}

	dir := filepath.Dir(cfg.ConfigPath)

	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}

	cfg.Version = CurrentCLIConfigVersion
	data, err := json.MarshalIndent(cfg, "", "  ")

	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	return os.WriteFile(cfg.ConfigPath, data, 0600)
}

func LoadCLIFile(path string) (CLIConfig, error) {

	if path == "" {
		return CLIConfig{}, fmt.Errorf("config path is required")
	}

	data, err := os.ReadFile(path)

	if err != nil {
		return CLIConfig{}, fmt.Errorf("read cli config: %w", err)
	}

	return decodeCLIFile(data)
}

func decodeCLIFile(data []byte) (CLIConfig, error) {
	var cfg CLIConfig

	if err := json.Unmarshal(data, &cfg); err != nil {
		return CLIConfig{}, fmt.Errorf("decode cli config: %w", err)
	}

	if cfg.Version > CurrentCLIConfigVersion {
		return CLIConfig{}, fmt.Errorf("cli config version %d is newer than supported version %d", cfg.Version, CurrentCLIConfigVersion)
	}

	if cfg.Version == 0 {
		return migrateCLIFile(&cfg), nil
	}

	return cfg, nil
}

func migrateCLIFile(cfg *CLIConfig) CLIConfig {
	cfg.Version = CurrentCLIConfigVersion
	return *cfg
}

func parse[T any](cfg *T) error {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}

	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("parse environment: %w", err)
	}

	return nil
}

func validateAPI(cfg APIConfig) error {

	if err := validateAPIApp(cfg.App); err != nil {
		return err
	}

	if err := validateDatabase(cfg.Database); err != nil {
		return err
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
	if !strings.EqualFold(cfg.Environment, "production") {
		return nil
	}
	for name, rawURL := range map[string]string{"public api url": cfg.PublicAPIURL, "dashboard url": cfg.DashboardURL, "cors origin": cfg.CORSOrigin} {
		parsed, err := url.Parse(rawURL)
		if err != nil || parsed.Scheme != "https" || parsed.Host == "" {
			return fmt.Errorf("%s must use an https URL in production", name)
		}
	}
	if strings.Contains(strings.ToLower(cfg.AllowedOrigins), "localhost") || strings.Contains(strings.ToLower(cfg.AllowedOrigins), "127.0.0.1") {
		return fmt.Errorf("localhost origins are not allowed in production")
	}
	return nil
}

func validateService(cfg ServiceConfig, app AppConfig) error {
	if !strings.EqualFold(app.Environment, "production") {
		return nil
	}
	if len(cfg.InternalAPISecret) < 32 {
		return fmt.Errorf("internal api secret must be at least 32 characters in production")
	}
	return nil
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
