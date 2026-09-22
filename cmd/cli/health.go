package main

import (
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/pkg/client"
)

func newHealthCommand(cfg config.CLIConfig) *cobra.Command {
	command := &cobra.Command{
		Use:   "health",
		Short: "check API readiness",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			apiURL, err := cmd.Flags().GetString("api-url")

			if err != nil {
				return err
			}

			apiClient, err := client.New(client.Config{BaseURL: apiURL, APIKey: cfg.APIKey})

			if err != nil {
				return fmt.Errorf("initialize client: %w", err)
			}

			if err := apiClient.Do(cmd.Context(), http.MethodGet, "/readyz", nil, nil); err != nil {
				return fmt.Errorf("health check: %w", err)
			}

			cmd.Println("ready")
			return nil
		},
	}

	command.Flags().String("api-url", cfg.APIURL, "tunnel API URL")
	return command
}
