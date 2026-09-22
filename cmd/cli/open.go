package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/pkg/client"
	"outpipe.dev/outpipe/pkg/protocol"
)

func newOpenCommand(cfg config.CLIConfig) *cobra.Command {
	command := openTunnelCommand(cfg, "open", "open a tunnel from a local port to the public network")
	command.Aliases = []string{"http"}
	return command
}

func newTCPCommand(cfg config.CLIConfig) *cobra.Command {
	command := openTunnelCommand(cfg, "tcp", "open a raw TCP tunnel from a local port")
	command.Flags().Lookup("protocol").DefValue = "tcp"
	command.Flags().Lookup("protocol").Usage = "tunnel protocol (tcp, udp, http)"
	return command
}

func openTunnelCommand(cfg config.CLIConfig, name, short string) *cobra.Command {
	command := &cobra.Command{
		Use:   name,
		Short: short,
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			port, err := cmd.Flags().GetInt("port")

			if err != nil {
				return err
			}

			protocolName, err := cmd.Flags().GetString("protocol")

			if err != nil {
				return err
			}

			subdomain, err := cmd.Flags().GetString("subdomain")

			if err != nil {
				return err
			}

			password, err := cmd.Flags().GetString("password")

			if err != nil {
				return err
			}

			agentToken, err := cmd.Flags().GetString("agent-token")

			if err != nil {
				return err
			}

			tunnelID, err := cmd.Flags().GetString("tunnel-id")

			if err != nil {
				return err
			}

			return openTunnel(cmd.Context(), cfg, port, protocolName, subdomain, password, agentToken, tunnelID)
		},
	}

	command.Flags().Int("port", 3000, "local port")
	command.Flags().String("protocol", "http", "tunnel protocol (http, tcp, udp)")
	command.Flags().String("subdomain", "", "requested subdomain")
	command.Flags().String("password", cfg.Password, "require this password for HTTP access")
	command.Flags().String("agent-token", cfg.AgentToken, "agent token for CI/CD usage")
	command.Flags().String("tunnel-id", "", "resume a managed tunnel")
	return command
}

func openTunnel(ctx context.Context, cfg config.CLIConfig, port int, protocolName, subdomain, password, agentToken, tunnelID string) error {
	ctx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
	defer stop()
	target := "http://127.0.0.1:" + fmt.Sprint(port)

	if protocolName == "tcp" || protocolName == "udp" {
		target = "127.0.0.1:" + fmt.Sprint(port)
	}

	delay := 2 * time.Second
	resolvedHostname := ""

	if tunnelID != "" {
		resolvedSubdomain, err := resolveManagedTunnel(ctx, cfg, tunnelID, subdomain)

		if err != nil {
			return err
		}

		resolvedHostname = resolvedSubdomain

		if strings.TrimSpace(subdomain) == "" {
			publicDomain := strings.TrimSuffix(cfg.PublicDomain, ".")

			if strings.HasSuffix(resolvedSubdomain, "."+publicDomain) || resolvedSubdomain == publicDomain {
				subdomain = strings.TrimSuffix(strings.TrimSuffix(resolvedSubdomain, "."+publicDomain), ".")
			} else {
				subdomain = ""
			}
		}
	}

	for ctx.Err() == nil {
		open := protocol.OpenTunnel{TunnelID: tunnelID, LocalPort: port, Protocol: protocolName, Subdomain: subdomain, Password: password}

		if subdomain == "" && resolvedHostname != "" {
			open.CustomDomain = resolvedHostname
		}

		connection, err := client.OpenRelay(ctx, client.RelayConfig{URL: cfg.RelayURL, Token: agentToken}, open)

		if err != nil {

			if ctx.Err() != nil {
				return nil
			}

			if ownerRelay := client.RelayOwnerFromError(err); ownerRelay != "" {
				fmt.Fprintf(os.Stderr, "outpipe: tunnel %q is connected through relay %s; retrying on the owning relay in %s\n", tunnelID, ownerRelay, 2*time.Second)
				delay = 2 * time.Second
			} else {
				fmt.Fprintf(os.Stderr, "outpipe: relay connection failed: %v; retrying in %s\n", err, delay)
			}

			select {
			case <-ctx.Done():
				return nil
			case <-time.After(delay):
			}

			if delay < 30*time.Second {
				delay *= 2
			}

			continue
		}

		tunnelID = connection.TunnelID
		delay = 2 * time.Second

		if connection.PublicPort > 0 {
			fmt.Printf("tunnel %s %s:%d\n", connection.TunnelID, connection.PublicURL, connection.PublicPort)
		} else {
			fmt.Printf("tunnel %s %s\n", connection.TunnelID, connection.PublicURL)
		}

		ticker := time.NewTicker(20 * time.Second)
		done := make(chan struct{})
		go func() {
			defer close(done)

			for {

				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if err := connection.SendHeartbeat(); err != nil {
						return
					}
				}
			}

		}()
		serveErr := connection.ServeLocal(ctx, target)
		ticker.Stop()
		connection.Close()
		<-done

		if serveErr != nil && ctx.Err() == nil {
			fmt.Fprintf(os.Stderr, "outpipe: relay connection closed: %v; reconnecting\n", serveErr)
		}
	}

	return nil
}

func resolveManagedTunnel(ctx context.Context, cfg config.CLIConfig, tunnelID, requestedSubdomain string) (string, error) {

	if strings.TrimSpace(cfg.APIKey) == "" {
		return "", fmt.Errorf("--tunnel-id requires OUTPIPE_API_KEY to validate the managed tunnel")
	}

	apiClient, err := client.New(client.Config{BaseURL: cfg.APIURL, APIKey: cfg.APIKey})

	if err != nil {
		return "", fmt.Errorf("initialize tunnel API client: %w", err)
	}

	var tunnel TunnelDTO

	if err := apiClient.Do(ctx, http.MethodGet, "/api/v1/tunnels/"+tunnelID, nil, &tunnel); err != nil {
		return "", fmt.Errorf("validate managed tunnel: %w", err)
	}

	if tunnel.ID != tunnelID || tunnel.Status == "revoked" {
		return "", fmt.Errorf("managed tunnel %q is not available", tunnelID)
	}

	if strings.TrimSpace(requestedSubdomain) != "" {
		return requestedSubdomain, nil
	}

	hostname := strings.TrimSpace(tunnel.PublicHostname)

	if hostname == "" {
		return "", fmt.Errorf("managed tunnel %q has no public hostname", tunnelID)
	}

	return hostname, nil
}
