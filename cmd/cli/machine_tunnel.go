package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/pkg/client"
)

func newMachineTunnelCommand(cfg config.CLIConfig) *cobra.Command {
	port := 3000
	protocolName := "http"
	command := &cobra.Command{Use: "machine-tunnel NAME", Short: "create and connect a tunnel using OUTPIPE_MACHINE_TOKEN", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		token := os.Getenv("OUTPIPE_MACHINE_TOKEN")
		if token == "" {
			return fmt.Errorf("OUTPIPE_MACHINE_TOKEN is required")
		}
		api, err := client.New(client.Config{BaseURL: cfg.APIURL, APIKey: token})
		if err != nil {
			return fmt.Errorf("initialize machine API: %w", err)
		}
		var response struct {
			Tunnel struct {
				ID string `json:"id"`
			} `json:"tunnel"`
			RelayToken string    `json:"relayToken"`
			ExpiresAt  time.Time `json:"expiresAt"`
		}
		input := map[string]any{"name": args[0], "protocol": protocolName, "targetHost": "127.0.0.1", "targetPort": port}
		if err := api.Do(cmd.Context(), http.MethodPost, "/api/v1/machines/tunnels", input, &response); err != nil {
			return fmt.Errorf("create machine tunnel: %w", err)
		}
		if response.Tunnel.ID == "" || response.RelayToken == "" || !response.ExpiresAt.After(time.Now()) {
			return fmt.Errorf("machine tunnel response is incomplete")
		}
		ctx, cancel := context.WithDeadline(cmd.Context(), response.ExpiresAt)
		defer cancel()
		cfg.APIKey = ""
		if err := openTunnel(ctx, cfg, port, protocolName, "", "", response.RelayToken, response.Tunnel.ID); err != nil {
			return fmt.Errorf("connect machine tunnel: %w", err)
		}
		return nil
	}}
	command.Flags().IntVar(&port, "port", port, "local port")
	command.Flags().StringVar(&protocolName, "protocol", protocolName, "tunnel protocol (http, tcp, udp)")
	return command
}
