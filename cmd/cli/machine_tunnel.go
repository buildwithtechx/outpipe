package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/pkg/client"
)

func newMachineTunnelCommand(cfg config.CLIConfig) *cobra.Command {
	port := 3000
	protocolName := "http"
	var originTLS client.OriginTLSConfig
	command := &cobra.Command{Use: "machine-tunnel NAME", Short: "create and connect a tunnel using a scoped machine credential", Args: cobra.ExactArgs(1), RunE: func(cmd *cobra.Command, args []string) error {
		originClient, err := prepareOriginClient(protocolName, originTLS)
		if err != nil {
			return err
		}
		defer originClient.CloseIdleConnections()
		token := cliEnvValue("OUTPIPE_MACHINE_TOKEN")
		if token == "" {
			token = cliEnvValue("OUTPIPE_TOKEN")
		}
		if token == "" {
			return fmt.Errorf("set OUTPIPE_TOKEN to a machine token with tunnels:write scope")
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
		tunnelID := response.Tunnel.ID
		relayToken := response.RelayToken
		expiresAt := response.ExpiresAt
		refreshDelay := 2 * time.Second
		cfg.APIKey = ""
		for {
			if refreshAt := machineRefreshAt(expiresAt); time.Until(refreshAt) > 0 {
				ctx, cancel := context.WithDeadline(cmd.Context(), refreshAt)
				err := openTunnel(ctx, cfg, port, protocolName, "", "", relayToken, tunnelID, originClient)
				cancel()
				if err != nil {
					return err
				}
				if cmd.Context().Err() != nil || time.Now().Before(refreshAt) {
					return nil
				}
			}
			token, next, fatal, err := refreshMachineRelayToken(cmd.Context(), api, tunnelID)
			if err != nil {
				if fatal {
					return err
				}
				fmt.Fprintf(os.Stderr, "outpipe: relay credential refresh failed: %v; retrying in %s\n", err, refreshDelay)
				select {
				case <-cmd.Context().Done():
					return nil
				case <-time.After(refreshDelay):
				}
				if refreshDelay < 30*time.Second {
					refreshDelay *= 2
				}
				continue
			}
			relayToken = token
			expiresAt = next
			refreshDelay = 2 * time.Second
		}
	}}
	command.Flags().IntVar(&port, "port", port, "local port")
	command.Flags().StringVar(&protocolName, "protocol", protocolName, "tunnel protocol (http, https, tcp, udp)")
	command.Flags().StringVar(&originTLS.CAFile, "origin-ca", "", "PEM CA certificate file for the local HTTPS service")
	command.Flags().StringVar(&originTLS.ServerName, "origin-server-name", "", "TLS certificate hostname for the local HTTPS service")
	return command
}

func machineRefreshAt(expiresAt time.Time) time.Time {
	lead := time.Until(expiresAt) / 10
	if lead < time.Minute {
		lead = time.Minute
	}
	if lead > 2*time.Minute {
		lead = 2 * time.Minute
	}
	return expiresAt.Add(-lead)
}

func refreshMachineRelayToken(ctx context.Context, api *client.Client, tunnelID string) (string, time.Time, bool, error) {
	var response struct {
		RelayToken string    `json:"relayToken"`
		ExpiresAt  time.Time `json:"expiresAt"`
	}
	if err := api.Do(ctx, http.MethodPost, "/api/v1/machines/tunnels/"+tunnelID+"/relay-token", nil, &response); err != nil {
		fatal := strings.Contains(err.Error(), "status 401") || strings.Contains(err.Error(), "status 403") || strings.Contains(err.Error(), "status 404")
		if fatal {
			return "", time.Time{}, true, fmt.Errorf("machine credential is expired, revoked, or unauthorized; rotate OUTPIPE_MACHINE_TOKEN: %w", err)
		}
		return "", time.Time{}, false, fmt.Errorf("refresh machine relay credential: %w", err)
	}
	if response.RelayToken == "" || !response.ExpiresAt.After(time.Now()) {
		return "", time.Time{}, true, fmt.Errorf("machine relay credential response is incomplete")
	}
	return response.RelayToken, response.ExpiresAt, false, nil
}
