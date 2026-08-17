package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/google/uuid"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/internal/engine"
	"outpipe.dev/outpipe/internal/infra/certificates"
	"outpipe.dev/outpipe/internal/infra/redis"
	"outpipe.dev/outpipe/internal/relay"
)

var version = "dev"

func main() {
	cfg, err := config.LoadRelay()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.RelayID == "" {
		cfg.RelayID = "relay-" + uuid.NewString()
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	redisClient, err := redis.Open(ctx, redis.Config{Host: cfg.Redis.Host, Port: cfg.Redis.Port, Password: cfg.Redis.Password, DB: cfg.Redis.DB})
	if err != nil {
		log.Fatal(err)
	}
	defer redisClient.Close()
	affinity, err := redis.NewRelayAffinity(redisClient)
	if err != nil {
		log.Fatal(err)
	}
	sessions := engine.NewSessionRegistry()
	requestRouter, err := engine.NewRequestRouter(sessions, cfg.Tunnel.AgentInactivity)
	if err != nil {
		log.Fatal(err)
	}
	tcpManager := relay.NewTCPManager()
	udpManager := relay.NewUDPManager()
	metrics := relay.NewMetrics()
	httpProxy, err := engine.NewHTTPProxy(cfg.Tunnel.Domain, requestRouter, cfg.Tunnel.MaxBytes)
	if err != nil {
		log.Fatal(err)
	}
	usage := newUsageRecorder(cfg.Service.InternalAPIURL, cfg.Service.InternalAPISecret)
	httpProxy.SetUsageRecorder(usage)
	authenticator, err := relay.NewInternalAgentAuthenticator(cfg.Service.InternalAPIURL, cfg.Service.InternalAPISecret, nil)
	if err != nil {
		log.Fatal(err)
	}
	managedTunnels, err := relay.NewInternalTunnelResolver(cfg.Service.InternalAPIURL, cfg.Service.InternalAPISecret, nil)
	if err != nil {
		log.Fatal(err)
	}
	relayHandler, err := relay.NewHandlerWithOptions(authenticator, sessions, requestRouter, tcpManager, udpManager, relay.HandlerOptions{MaxConnections: cfg.Tunnel.MaxConnections, MaxTunnels: cfg.Tunnel.MaxTunnels, MaxBandwidth: cfg.Tunnel.MaxBandwidth, Heartbeat: cfg.Tunnel.Heartbeat, ReadTimeout: cfg.Tunnel.ReadTimeout, DrainTimeout: cfg.Tunnel.DrainTimeout, MaxFrameBytes: cfg.Tunnel.MaxFrameBytes, Logger: slog.Default(), Metrics: metrics, UsageRecorder: usage, Affinity: affinity, ManagedTunnels: managedTunnels, RelayID: cfg.RelayID, AffinityTTL: cfg.Tunnel.AgentInactivity, AllowedOrigins: cfg.App.AllowedOrigins, PublicDomain: cfg.Tunnel.Domain})
	if err != nil {
		log.Fatal(err)
	}
	app := fiber.New(fiber.Config{AppName: cfg.App.Name, DisableStartupMessage: true})
	app.Use(recover.New())
	app.Get("/v1/connect", relayHandler.Upgrade)
	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok", "sessions": len(sessions.Snapshot()), "metrics": metrics.Snapshot()})
	})
	app.Get("/metrics", func(c *fiber.Ctx) error {
		c.Type("text", "plain")
		return c.SendString(metrics.Prometheus())
	})
	app.Get("/readyz", func(c *fiber.Ctx) error {
		if err := redisClient.Raw().Ping(c.UserContext()).Err(); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{"error": "redis unavailable"})
		}
		return c.JSON(fiber.Map{"status": "ready"})
	})
	app.All("/*", adaptor.HTTPHandler(httpProxy))
	go func() {
		var listenErr error
		if cfg.App.RequireTLS {
			listener, err := certificates.NewTLSListener(cfg.App.ListenAddress(), cfg.App.TLSCertFile, cfg.App.TLSKeyFile)
			if err != nil {
				listenErr = err
			} else {
				listenErr = app.Listener(listener)
			}
		} else {
			listenErr = app.Listen(cfg.App.ListenAddress())
		}
		if err := listenErr; err != nil {
			log.Printf("relay stopped: %v", err)
			stop()
		}
	}()
	<-ctx.Done()
	relayHandler.CloseAll()
	if err := app.Shutdown(); err != nil {
		log.Printf("shutdown relay: %v", err)
	}
}
