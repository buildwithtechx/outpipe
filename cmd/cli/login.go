package main

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/pkg/client"
)

func runLogin(cfg config.CLIConfig, args []string) error {
	flags := flag.NewFlagSet("login", flag.ContinueOnError)
	flags.SetOutput(os.Stderr)
	token := flags.String("agent-token", "", "agent token issued by the dashboard")
	code := flags.String("code", "", "existing browser device code")
	apiKey := flags.String("api-key", cfg.APIKey, "API key for management commands")

	if err := flags.Parse(args); err != nil {
		return fmt.Errorf("parse login flags: %w", err)
	}

	if *token != "" {
		return saveLogin(cfg, *token, *apiKey)
	}

	apiClient, err := client.New(client.Config{BaseURL: cfg.APIURL})

	if err != nil {
		return fmt.Errorf("initialize client: %w", err)
	}

	deviceCode := *code

	if deviceCode == "" {
		var started struct {
			Code string `json:"code"`
		}
		if err := apiClient.Do(context.Background(), http.MethodPost, "/api/v1/auth/device/start", nil, &started); err != nil {
			return fmt.Errorf("start device login: %w", err)
		}

		deviceCode = started.Code
		fmt.Printf("Authorize this device in the dashboard using code %s\n", deviceCode)
	}

	for {
		var result struct {
			Status string `json:"status"`
			Token  string `json:"token"`
		}
		if err := apiClient.Do(context.Background(), http.MethodGet, "/api/v1/auth/device/poll?code="+deviceCode, nil, &result); err != nil {
			return fmt.Errorf("poll device login: %w", err)
		}

		if result.Token != "" {
			return saveLogin(cfg, result.Token, *apiKey)
		}

		time.Sleep(2 * time.Second)
	}
}

func saveLogin(cfg config.CLIConfig, token, apiKey string) error {
	cfg.AgentToken = token
	cfg.APIKey = apiKey

	if err := config.SaveCLI(cfg); err != nil {
		return fmt.Errorf("save credentials: %w", err)
	}

	fmt.Println("credentials saved")
	return nil
}
