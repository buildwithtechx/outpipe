package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/pkg/client"
)

func newCapturePolicyCommand(cfg config.CLIConfig) *cobra.Command {
	command := &cobra.Command{Use: "capture-policy", Short: "manage tunnel capture policies", Args: cobra.NoArgs}
	command.AddCommand(
		&cobra.Command{Use: "show TUNNEL_ID", Short: "show a tunnel capture policy", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			return runCapturePolicyAction(cmd, cfg, "show", args)
		}},
		&cobra.Command{Use: "set TUNNEL_ID", Short: "set a tunnel capture policy", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
			return runCapturePolicyAction(cmd, cfg, "set", args)
		}},
	)
	command.PersistentFlags().Bool("json", false, "output in JSON format")
	command.PersistentFlags().String("rules", "", "capture rules JSON (set only)")
	command.PersistentFlags().Bool("enabled", false, "enable capture (set only)")
	command.PersistentFlags().Bool("disabled", false, "disable capture (set only)")
	return command
}

func runCapturePolicyAction(cmd *cobra.Command, cfg config.CLIConfig, action string, args []string) error {
	jsonOutput, err := cmd.Flags().GetBool("json")
	if err != nil {
		return err
	}
	apiClient, err := client.New(client.Config{BaseURL: cfg.APIURL, APIKey: cfg.APIKey})
	if err != nil {
		return fmt.Errorf("initialize client: %w", err)
	}
	printJSON := func(value any) error {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(value)
	}
	switch action {
	case "show":
		var policy struct {
			TunnelID string `json:"tunnelId"`
			Version  int    `json:"version"`
			Rules    string `json:"rules"`
		}
		if err := apiClient.Do(cmd.Context(), http.MethodGet, "/api/v1/tunnels/"+args[0]+"/capture-policy", nil, &policy); err != nil {
			return fmt.Errorf("show capture policy: %w", err)
		}
		if jsonOutput {
			return printJSON(policy)
		}
		fmt.Printf("tunnel %s policy version %d\nrules: %s\n", policy.TunnelID, policy.Version, policy.Rules)
	case "set":
		rules, err := cmd.Flags().GetString("rules")
		if err != nil {
			return err
		}
		if rules == "" {
			return fmt.Errorf("capture rules JSON is required via --rules")
		}
		payload := map[string]any{"policy": map[string]any{"rules": rules}}
		enabled, err := cmd.Flags().GetBool("enabled")
		if err != nil {
			return err
		}
		disabled, err := cmd.Flags().GetBool("disabled")
		if err != nil {
			return err
		}
		if enabled && disabled {
			return fmt.Errorf("only one of --enabled or --disabled may be set")
		}
		if enabled {
			payload["enabled"] = true
		}
		if disabled {
			payload["enabled"] = false
		}
		var policy struct {
			TunnelID string `json:"tunnelId"`
			Version  int    `json:"version"`
			Rules    string `json:"rules"`
		}
		if enabled || disabled {
			var tunnel TunnelDTO
			if err := apiClient.Do(cmd.Context(), http.MethodPatch, "/api/v1/tunnels/"+args[0]+"/capture", payload, &tunnel); err != nil {
				return fmt.Errorf("set capture policy: %w", err)
			}
			if jsonOutput {
				return printJSON(tunnel)
			}
			fmt.Printf("tunnel %s capture enabled\n", tunnel.ID)
			return nil
		}
		if err := apiClient.Do(cmd.Context(), http.MethodPatch, "/api/v1/tunnels/"+args[0]+"/capture", payload, &policy); err != nil {
			return fmt.Errorf("set capture policy: %w", err)
		}
		if jsonOutput {
			return printJSON(policy)
		}
		fmt.Printf("tunnel %s policy version %d\n", policy.TunnelID, policy.Version)
	}
	return nil
}
