package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"outpipe.dev/outpipe/internal/config"
)

func cliEnvValue(name string) string {
	value := strings.TrimSpace(os.Getenv(name))
	if strings.HasPrefix(value, "replace_with_") {
		return ""
	}
	return value
}

func loadCLISettings() (config.CLIConfig, error) {
	cfg, err := config.LoadCLI()
	if err != nil {
		return config.CLIConfig{}, err
	}
	if cliEnvValue("OUTPIPE_API_KEY") == "" {
		cfg.APIKey = ""
	}
	if cliEnvValue("OUTPIPE_AGENT_TOKEN") == "" {
		cfg.AgentToken = ""
	}
	if cliEnvValue("OUTPIPE_PASSWORD") == "" {
		cfg.Password = ""
	}
	stored, loadErr := config.LoadCLIFile(cfg.ConfigPath)
	if loadErr == nil {
		if cliEnvValue("OUTPIPE_API_URL") == "" && stored.APIURL != "" {
			cfg.APIURL = stored.APIURL
		}
		if cliEnvValue("OUTPIPE_RELAY_URL") == "" && stored.RelayURL != "" {
			cfg.RelayURL = stored.RelayURL
		}
		if cliEnvValue("OUTPIPE_DOMAIN") == "" && stored.PublicDomain != "" {
			cfg.PublicDomain = stored.PublicDomain
		}
		if cliEnvValue("OUTPIPE_API_KEY") == "" {
			cfg.APIKey = stored.APIKey
		}
		if cliEnvValue("OUTPIPE_AGENT_TOKEN") == "" {
			cfg.AgentToken = stored.AgentToken
		}
	} else if !errors.Is(loadErr, os.ErrNotExist) {
		return config.CLIConfig{}, fmt.Errorf("load saved CLI settings: %w", loadErr)
	}
	return cfg, nil
}
