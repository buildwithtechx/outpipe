package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/pkg/client"
	"outpipe.dev/outpipe/pkg/protocol"
)

var version = "dev"

func main() {
	cfg, err := config.LoadCLI()
	if err != nil {
		log.Fatal(err)
	}
	if stored, loadErr := config.LoadCLIFile(cfg.ConfigPath); loadErr == nil {
		if _, ok := os.LookupEnv("OUTPIPE_API_KEY"); !ok {
			cfg.APIKey = stored.APIKey
		}
		if _, ok := os.LookupEnv("OUTPIPE_AGENT_TOKEN"); !ok {
			cfg.AgentToken = stored.AgentToken
		}
	}
	if len(os.Args) < 2 || os.Args[1] == "help" {
		printUsage()
		return
	}

	command := os.Args[1]
	switch command {
	case "version":
		fmt.Println(version)
	case "login":
		runLogin(cfg)
	case "open", "http", "tcp":
		openTunnel(cfg, command, os.Args[2:])
	case "create", "list", "inspect", "start", "stop", "revoke":
		runTunnelsCommand(cfg, os.Args[1:])
	case "completion":
		runCompletion(os.Args[2:])
	case "health":
		runHealth(cfg, os.Args[2:])
	default:
		log.Fatalf("unknown command %q", command)
	}
}

func runHealth(cfg config.CLIConfig, args []string) {
	flags := flag.NewFlagSet("health", flag.ExitOnError)
	apiURL := flags.String("api-url", cfg.APIURL, "tunnel API URL")
	_ = flags.Parse(args)
	apiClient, err := client.New(client.Config{BaseURL: *apiURL, APIKey: cfg.APIKey})
	if err != nil {
		log.Fatal(err)
	}
	if err := apiClient.Do(context.Background(), http.MethodGet, "/readyz", nil, nil); err != nil {
		log.Fatal(err)
	}
	fmt.Println("ready")
}

func openTunnel(cfg config.CLIConfig, cmdName string, args []string) {
	flags := flag.NewFlagSet(cmdName, flag.ExitOnError)
	port := flags.Int("port", 3000, "local port")
	defaultProtocol := "http"
	if cmdName == "tcp" {
		defaultProtocol = "tcp"
	}
	protocolName := flags.String("protocol", defaultProtocol, "tunnel protocol (http, tcp, udp)")
	subdomain := flags.String("subdomain", "", "requested subdomain")
	password := flags.String("password", cfg.Password, "require this password for HTTP access")
	agentToken := flags.String("agent-token", cfg.AgentToken, "agent token for CI/CD usage")
	tunnelIDFlag := flags.String("tunnel-id", "", "resume a managed tunnel")
	_ = flags.Parse(args)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	target := "http://127.0.0.1:" + fmt.Sprint(*port)
	if *protocolName == "tcp" || *protocolName == "udp" {
		target = "127.0.0.1:" + fmt.Sprint(*port)
	}
	delay := 2 * time.Second
	tunnelID := *tunnelIDFlag
	resolvedHostname := ""
	if tunnelID != "" {
		resolvedSubdomain, err := resolveManagedTunnel(ctx, cfg, tunnelID, *subdomain)
		if err != nil {
			log.Fatal(err)
		}
		resolvedHostname = resolvedSubdomain
		if strings.TrimSpace(*subdomain) == "" {
			publicDomain := strings.TrimSuffix(cfg.PublicDomain, ".")
			if strings.HasSuffix(resolvedSubdomain, "."+publicDomain) || resolvedSubdomain == publicDomain {
				*subdomain = strings.TrimSuffix(strings.TrimSuffix(resolvedSubdomain, "."+publicDomain), ".")
			} else {
				*subdomain = ""
			}
		}
	}
	for ctx.Err() == nil {
		open := protocol.OpenTunnel{TunnelID: tunnelID, LocalPort: *port, Protocol: *protocolName, Subdomain: *subdomain, Password: *password}
		if *subdomain == "" && resolvedHostname != "" {
			open.CustomDomain = resolvedHostname
		}
		connection, err := client.OpenRelay(ctx, client.RelayConfig{URL: cfg.RelayURL, Token: *agentToken}, open)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			log.Printf("relay connection failed: %v; retrying in %s", err, delay)
			select {
			case <-ctx.Done():
				return
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
			log.Printf("relay connection closed: %v; reconnecting", serveErr)
		}
	}
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

func printUsage() {
	fmt.Println("outpipe <command>")
	fmt.Println("commands: login, open, create, list, inspect, start, stop, revoke, http, tcp, health, completion, version")
}
