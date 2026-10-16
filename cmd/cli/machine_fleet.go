package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/spf13/cobra"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/pkg/client"
)

type fleetStatusDTO struct {
	Fleet struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"fleet"`
	Tokens []struct {
		ID        string  `json:"id"`
		Name      string  `json:"name"`
		RevokedAt *string `json:"revokedAt"`
	} `json:"tokens"`
	DesiredTunnels []struct {
		TunnelID string `json:"tunnelId"`
		Name     string `json:"name"`
	} `json:"desiredTunnels"`
	LiveTunnels []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Status string `json:"status"`
	} `json:"liveTunnels"`
	Drift []struct {
		TunnelID string `json:"tunnelId"`
		Name     string `json:"name"`
		Want     string `json:"want"`
		Got      string `json:"got"`
	} `json:"drift"`
}

func newMachineFleetCommand(cfg config.CLIConfig) *cobra.Command {
	command := &cobra.Command{
		Use:   "machine-fleet",
		Short: "manage machine tunnel fleets",
	}
	command.AddCommand(
		newMachineFleetCreateCommand(cfg),
		newMachineFleetListCommand(cfg),
		newMachineFleetStatusCommand(cfg),
		newMachineFleetReconcileCommand(cfg),
	)
	return command
}

func fleetAPIClient(cfg config.CLIConfig) (*client.Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("set OUTPIPE_API_KEY to a credential with organization scope")
	}
	api, err := client.New(client.Config{BaseURL: cfg.APIURL, APIKey: cfg.APIKey})
	if err != nil {
		return nil, fmt.Errorf("initialize fleet API: %w", err)
	}
	return api, nil
}

func newMachineFleetCreateCommand(cfg config.CLIConfig) *cobra.Command {
	var organization, description string
	command := &cobra.Command{
		Use:   "create NAME",
		Short: "create a machine fleet",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			api, err := fleetAPIClient(cfg)
			if err != nil {
				return err
			}
			var fleet struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			}
			if err := api.Do(cmd.Context(), http.MethodPost, "/api/v1/organizations/"+organization+"/machine-fleets", map[string]string{"name": args[0], "description": description}, &fleet); err != nil {
				return fmt.Errorf("create fleet: %w", err)
			}
			fmt.Printf("fleet %s (%s)\n", fleet.Name, fleet.ID)
			return nil
		},
	}
	command.Flags().StringVar(&organization, "organization", "", "organization ID")
	command.Flags().StringVar(&description, "description", "", "fleet description")
	_ = command.MarkFlagRequired("organization")
	return command
}

func newMachineFleetListCommand(cfg config.CLIConfig) *cobra.Command {
	var organization string
	var jsonOutput bool
	command := &cobra.Command{
		Use:   "list",
		Short: "list machine fleets",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			api, err := fleetAPIClient(cfg)
			if err != nil {
				return err
			}
			var fleets []map[string]any
			if err := api.Do(cmd.Context(), http.MethodGet, "/api/v1/organizations/"+organization+"/machine-fleets", nil, &fleets); err != nil {
				return fmt.Errorf("list fleets: %w", err)
			}
			if jsonOutput {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(fleets)
			}
			for _, fleet := range fleets {
				fmt.Printf("%s\t%s\n", fleet["id"], fleet["name"])
			}
			return nil
		},
	}
	command.Flags().StringVar(&organization, "organization", "", "organization ID")
	command.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	_ = command.MarkFlagRequired("organization")
	return command
}

func newMachineFleetStatusCommand(cfg config.CLIConfig) *cobra.Command {
	var organization string
	var jsonOutput bool
	command := &cobra.Command{
		Use:   "status FLEET_ID",
		Short: "show fleet desired-state, live tunnels, and drift",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			api, err := fleetAPIClient(cfg)
			if err != nil {
				return err
			}
			var status fleetStatusDTO
			if err := api.Do(cmd.Context(), http.MethodGet, "/api/v1/organizations/"+organization+"/machine-fleets/"+args[0], nil, &status); err != nil {
				return fmt.Errorf("fleet status: %w", err)
			}
			if jsonOutput {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(status)
			}
			fmt.Printf("fleet %s (%s)\n", status.Fleet.Name, status.Fleet.ID)
			fmt.Printf("tokens: %d desired: %d live: %d drift: %d\n", len(status.Tokens), len(status.DesiredTunnels), len(status.LiveTunnels), len(status.Drift))
			for _, drift := range status.Drift {
				fmt.Printf("drift %s (%s): want %s got %s\n", drift.Name, drift.TunnelID, drift.Want, drift.Got)
			}
			return nil
		},
	}
	command.Flags().StringVar(&organization, "organization", "", "organization ID")
	command.Flags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	_ = command.MarkFlagRequired("organization")
	return command
}

func newMachineFleetReconcileCommand(cfg config.CLIConfig) *cobra.Command {
	var organization string
	command := &cobra.Command{
		Use:   "reconcile FLEET_ID",
		Short: "reconcile fleet tunnels toward desired state",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			api, err := fleetAPIClient(cfg)
			if err != nil {
				return err
			}
			var result map[string]any
			if err := api.Do(cmd.Context(), http.MethodPost, "/api/v1/organizations/"+organization+"/machine-fleets/"+args[0]+"/reconcile", nil, &result); err != nil {
				return fmt.Errorf("reconcile fleet: %w", err)
			}
			fmt.Printf("reactivated=%v recreated=%v skipped=%v\n", result["reactivated"], result["recreated"], result["skipped"])
			return nil
		},
	}
	command.Flags().StringVar(&organization, "organization", "", "organization ID")
	_ = command.MarkFlagRequired("organization")
	return command
}
