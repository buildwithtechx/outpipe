package http

import (
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"outpipe.dev/outpipe/internal/config"
	"outpipe.dev/outpipe/internal/handlers"
	"outpipe.dev/outpipe/internal/infra/billing"
	"outpipe.dev/outpipe/internal/infra/certificates"
)

type Server struct {
	app             *fiber.App
	internalApp     *fiber.App
	requireTLS      bool
	certFile        string
	keyFile         string
	internalAddress string
}

func NewServer(cfg config.APIConfig, deps Dependencies) (*Server, error) {
	deps.PublicAPIURL = cfg.App.PublicAPIURL
	deps.DashboardURL = cfg.App.DashboardURL
	cookie := handlers.SessionCookieConfig{Name: cfg.Auth.CookieName, Secure: cfg.Auth.CookieSecure, Domain: cfg.Auth.CookieDomain}
	handlers, err := buildHandlers(deps, cookie)

	if err != nil {
		return nil, err
	}

	app := fiber.New(fiber.Config{AppName: cfg.App.Name, DisableStartupMessage: true, ErrorHandler: errorHandler})
	app.Use(recover.New())
	app.Use(helmet.New())
	app.Use(cors.New(cors.Config{AllowOrigins: cfg.App.AllowedOrigins, AllowHeaders: "Origin, Content-Type, Accept, Authorization, X-Internal-Secret", AllowCredentials: true}))

	if err := RegisterRoutes(app, handlers, RouterOptions{CookieName: cfg.Auth.CookieName, CookieSecure: cfg.Auth.CookieSecure, BillingWebhookSecret: cfg.Billing.WebhookSecret, RateLimiter: deps.RateLimiter}); err != nil {
		return nil, err
	}

	var internalApp *fiber.App

	if cfg.Service.InternalAPISecret != "" {
		internalApp = fiber.New(fiber.Config{AppName: cfg.App.Name + "-internal", DisableStartupMessage: true, ErrorHandler: errorHandler})
		internalApp.Use(recover.New())

		if err := RegisterInternalRoutes(internalApp, handlers, RouterOptions{InternalAPISecret: cfg.Service.InternalAPISecret}); err != nil {
			return nil, err
		}
	}

	if handlers.Billing != nil {
		var paystackClient *billing.PaystackClient

		if cfg.Billing.PaystackSecret != "" {
			paystackClient, err = billing.NewPaystack(billing.PaystackConfig{BaseURL: cfg.Billing.PaystackBaseURL, SecretKey: cfg.Billing.PaystackSecret})

			if err != nil {
				return nil, err
			}
		}

		handlers.Billing.SetProviderSecrets(cfg.Billing.PolarWebhookSecret, paystackClient)
	}

	return &Server{app: app, internalApp: internalApp, requireTLS: cfg.App.RequireTLS, certFile: cfg.App.TLSCertFile, keyFile: cfg.App.TLSKeyFile, internalAddress: cfg.App.InternalListenAddress}, nil
}

func (s *Server) App() *fiber.App {
	return s.app
}

func (s *Server) InternalApp() *fiber.App {
	return s.internalApp
}

func (s *Server) Listen(address string) error {

	if s == nil || s.app == nil {
		return fmt.Errorf("http server is not initialized")
	}

	if s.requireTLS {
		listener, err := certificates.NewTLSListener(address, s.certFile, s.keyFile)

		if err != nil {
			return err
		}

		return s.app.Listener(listener)
	}

	return s.app.Listen(address)
}

func (s *Server) ListenInternal(address string) error {

	if s == nil || s.internalApp == nil {
		return fmt.Errorf("internal http server is not initialized")
	}

	return s.internalApp.Listen(address)
}

func (s *Server) Shutdown() error {

	if s == nil || s.app == nil {
		return nil
	}

	if err := s.app.Shutdown(); err != nil {
		return err
	}

	if s.internalApp != nil {
		return s.internalApp.Shutdown()
	}

	return nil
}

func errorHandler(c *fiber.Ctx, err error) error {
	status := fiber.StatusInternalServerError

	if fiberErr, ok := err.(*fiber.Error); ok {
		status = fiberErr.Code
	}

	message := strings.TrimSpace(err.Error())

	if status >= fiber.StatusInternalServerError {
		message = "internal server error"
	}

	return c.Status(status).JSON(fiber.Map{"error": message})
}
