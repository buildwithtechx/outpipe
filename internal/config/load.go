package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

func LoadAPI() (APIConfig, error) {
	var cfg APIConfig
	if err := parse(&cfg, "outpipe"); err != nil {
		return APIConfig{}, err
	}
	if err := validateAPI(cfg); err != nil {
		return APIConfig{}, err
	}
	return cfg, nil
}

func LoadRelay() (RelayConfig, error) {
	var cfg RelayConfig
	if err := parse(&cfg, "relay"); err != nil {
		return RelayConfig{}, err
	}
	cfg.App.RequireTLS = cfg.App.RequireTLS || cfg.Tunnel.RequireTLS
	if err := validateRelay(cfg); err != nil {
		return RelayConfig{}, err
	}
	return cfg, nil
}

func LoadCron() (CronConfig, error) {
	var cfg CronConfig
	if err := parse(&cfg, "cron", "outpipe"); err != nil {
		return CronConfig{}, err
	}
	if err := validateDatabase(cfg.Database); err != nil {
		return CronConfig{}, err
	}
	return cfg, nil
}

func LoadCheck() (CheckConfig, error) {
	var cfg CheckConfig
	if err := parse(&cfg, "check", "outpipe"); err != nil {
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
	if _, err := CanonicalAPIOrigin(cfg.APIURL); err != nil {
		return CLIConfig{}, err
	}
	if cfg.ConfigPath == "" {
		path, err := defaultCLIConfigPath(cfg.APIURL)
		if err != nil {
			return CLIConfig{}, err
		}
		cfg.ConfigPath = path
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
		return CLIConfig{}, &CLIConfigVersionError{Version: cfg.Version}
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

func parse[T any](cfg *T, defaultPaths ...string) error {
	if err := env.Parse(cfg); err != nil {
		return fmt.Errorf("parse defaults: %w", err)
	}
	if err := loadYAMLConfig(cfg, defaultPaths...); err != nil {
		return err
	}
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("load .env: %w", err)
	}
	if err := env.ParseWithOptions(cfg, env.Options{
		DefaultValueTagName: "-",
	}); err != nil {
		return fmt.Errorf("parse environment overrides: %w", err)
	}
	return nil
}
