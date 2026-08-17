package client

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Credentials struct {
	APIURL     string `json:"api_url"`
	RelayURL   string `json:"relay_url"`
	AgentToken string `json:"agent_token"`
}

func SaveCredentials(path string, credentials Credentials) error {

	if path == "" || credentials.AgentToken == "" {
		return fmt.Errorf("credential path and agent token are required")
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return fmt.Errorf("create credential directory: %w", err)
	}

	data, err := json.Marshal(credentials)

	if err != nil {
		return fmt.Errorf("encode credentials: %w", err)
	}

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write credentials: %w", err)
	}

	return nil
}

func LoadCredentials(path string) (Credentials, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return Credentials{}, fmt.Errorf("read credentials: %w", err)
	}

	var credentials Credentials

	if err := json.Unmarshal(data, &credentials); err != nil {
		return Credentials{}, fmt.Errorf("decode credentials: %w", err)
	}

	if credentials.AgentToken == "" {
		return Credentials{}, fmt.Errorf("stored agent token is empty")
	}

	return credentials, nil
}
