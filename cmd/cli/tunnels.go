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

type TunnelDTO struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Protocol       string `json:"protocol"`
	Subdomain      string `json:"subdomain"`
	Status         string `json:"status"`
	PublicURL      string `json:"publicUrl"`
	PublicHostname string `json:"publicHostname"`
}

func printOutput(jsonOutput bool, value any) {

	if jsonOutput {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(value)
		return
	}

	switch v := value.(type) {
	case []TunnelDTO:
		for _, t := range v {
			fmt.Printf("%s\t%s\t%s\t%s\t%s\n", t.ID, t.Name, t.Protocol, t.Status, t.PublicURL)
		}
	case TunnelDTO:
		fmt.Printf("ID: %s\nName: %s\nProtocol: %s\nStatus: %s\nPublic URL: %s\n", v.ID, v.Name, v.Protocol, v.Status, v.PublicURL)
	default:
		fmt.Println(v)
	}
}

func newTunnelCommand(cfg config.CLIConfig, name, short string, args cobra.PositionalArgs) *cobra.Command {
	command := &cobra.Command{
		Use:   name,
		Short: short,
		Args:  args,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runTunnelAction(cmd, cfg, name, args)
		},
	}

	command.Flags().Bool("json", false, "output in JSON format")
	command.Flags().String("organization", "", "organization ID")
	command.Flags().String("target-host", "127.0.0.1", "target host")
	command.Flags().Int("target-port", 3000, "target port")
	command.Flags().String("hostname", "", "public hostname")
	command.Flags().String("password", cfg.Password, "require this password for HTTP access")

	if name == "create" || name == "list" {
		_ = command.MarkFlagRequired("organization")
	}

	return command
}

func runTunnelAction(cmd *cobra.Command, cfg config.CLIConfig, action string, args []string) error {
	jsonOutput, err := cmd.Flags().GetBool("json")

	if err != nil {
		return err
	}

	organizationID, err := cmd.Flags().GetString("organization")

	if err != nil {
		return err
	}

	apiClient, err := client.New(client.Config{BaseURL: cfg.APIURL, APIKey: cfg.APIKey})

	if err != nil {
		return fmt.Errorf("initialize client: %w", err)
	}

	switch action {
	case "list":
		var tunnels []TunnelDTO

		if err := apiClient.Do(cmd.Context(), http.MethodGet, "/api/v1/organizations/"+organizationID+"/tunnels", nil, &tunnels); err != nil {
			return fmt.Errorf("list tunnels: %w", err)
		}

		printOutput(jsonOutput, tunnels)
	case "inspect":
		var tunnel TunnelDTO

		if err := apiClient.Do(cmd.Context(), http.MethodGet, "/api/v1/tunnels/"+args[0], nil, &tunnel); err != nil {
			return fmt.Errorf("inspect tunnel: %w", err)
		}

		printOutput(jsonOutput, tunnel)
	case "create":
		targetHost, err := cmd.Flags().GetString("target-host")

		if err != nil {
			return err
		}

		targetPort, err := cmd.Flags().GetInt("target-port")

		if err != nil {
			return err
		}

		publicHostname, err := cmd.Flags().GetString("hostname")

		if err != nil {
			return err
		}

		password, err := cmd.Flags().GetString("password")

		if err != nil {
			return err
		}

		protocolName := "http"

		if len(args) > 1 {
			protocolName = args[1]
		}

		var tunnel TunnelDTO
		payload := map[string]any{"name": args[0], "protocol": protocolName, "targetHost": targetHost, "targetPort": targetPort, "publicHostname": publicHostname, "password": password}

		if err := apiClient.Do(cmd.Context(), http.MethodPost, "/api/v1/organizations/"+organizationID+"/tunnels", payload, &tunnel); err != nil {
			return fmt.Errorf("create tunnel: %w", err)
		}

		printOutput(jsonOutput, tunnel)
	case "start", "stop":
		status := "active"

		if action == "stop" {
			status = "disconnected"
		}

		if err := apiClient.Do(cmd.Context(), http.MethodPatch, "/api/v1/tunnels/"+args[0]+"/status", map[string]string{"status": status}, nil); err != nil {
			return fmt.Errorf("%s tunnel: %w", action, err)
		}

		verb := "started"

		if action == "stop" {
			verb = "stopped"
		}

		fmt.Printf("tunnel %s\n", verb)
	case "revoke":
		if err := apiClient.Do(cmd.Context(), http.MethodDelete, "/api/v1/tunnels/"+args[0], nil, nil); err != nil {
			return fmt.Errorf("revoke tunnel: %w", err)
		}

		fmt.Println("tunnel revoked")
	default:
		return fmt.Errorf("unknown action %q", action)
	}

	return nil
}
