package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/cobra"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/pkg/client"
)

func newLoginCommand(cfg config.CLIConfig) *cobra.Command {
	command := &cobra.Command{
		Use:   "login",
		Short: "save an agent token issued by the dashboard",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := cmd.Flags().GetString("agent-token")

			if err != nil {
				return err
			}

			code, err := cmd.Flags().GetString("code")

			if err != nil {
				return err
			}

			apiKey, err := cmd.Flags().GetString("api-key")

			if err != nil {
				return err
			}

			return runLogin(cmd.Context(), cfg, token, code, apiKey)
		},
	}

	command.Flags().String("agent-token", "", "agent token issued by the dashboard")
	command.Flags().String("code", "", "existing browser device code")
	command.Flags().String("api-key", cfg.APIKey, "API key for management commands")
	return command
}

func runLogin(ctx context.Context, cfg config.CLIConfig, token, code, apiKey string) error {

	if token != "" {
		return saveLogin(cfg, token, apiKey)
	}

	apiClient, err := client.New(client.Config{BaseURL: cfg.APIURL})

	if err != nil {
		return fmt.Errorf("initialize client: %w", err)
	}

	deviceCode := code

	if deviceCode == "" {
		var started struct {
			Code string `json:"code"`
		}
		if err := apiClient.Do(ctx, http.MethodPost, "/api/v1/auth/device/start", nil, &started); err != nil {
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
		if err := apiClient.Do(ctx, http.MethodGet, "/api/v1/auth/device/poll?code="+deviceCode, nil, &result); err != nil {
			return fmt.Errorf("poll device login: %w", err)
		}

		if result.Token != "" {
			return saveLogin(cfg, result.Token, apiKey)
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
