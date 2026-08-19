package config

import "time"

type APIConfig struct {
	App      AppConfig      `envPrefix:"OUTPIPE_"`
	Auth     AuthConfig     `envPrefix:"OUTPIPE_"`
	Database DatabaseConfig `envPrefix:"OUTPIPE_"`
	Redis    RedisConfig    `envPrefix:"OUTPIPE_"`
	Mail     MailConfig     `envPrefix:"OUTPIPE_"`
	Service  ServiceConfig  `envPrefix:"OUTPIPE_"`
	Billing  BillingConfig  `envPrefix:"OUTPIPE_"`
	Tunnel   TunnelConfig   `envPrefix:"OUTPIPE_"`
}

type RelayConfig struct {
	App     AppConfig     `envPrefix:"OUTPIPE_"`
	Redis   RedisConfig   `envPrefix:"OUTPIPE_"`
	Tunnel  TunnelConfig  `envPrefix:"OUTPIPE_"`
	Service ServiceConfig `envPrefix:"OUTPIPE_"`
	RelayID string        `env:"RELAY_ID"`
}

type CronConfig struct {
	App      AppConfig      `envPrefix:"OUTPIPE_"`
	Database DatabaseConfig `envPrefix:"OUTPIPE_"`
	Redis    RedisConfig    `envPrefix:"OUTPIPE_"`
	Service  ServiceConfig  `envPrefix:"OUTPIPE_"`
}

type CheckConfig struct {
	App     AppConfig     `envPrefix:"OUTPIPE_"`
	Service ServiceConfig `envPrefix:"OUTPIPE_"`
}

type CLIConfig struct {
	APIURL       string `env:"OUTPIPE_API_URL" envDefault:"http://localhost:8080"`
	RelayURL     string `env:"OUTPIPE_RELAY_URL" envDefault:"ws://localhost:8081"`
	PublicDomain string `env:"OUTPIPE_DOMAIN" envDefault:"outpipe.app"`
	APIKey       string `env:"OUTPIPE_API_KEY"`
	AgentToken   string `env:"OUTPIPE_AGENT_TOKEN"`
	Password     string `env:"OUTPIPE_PASSWORD"`
	ConfigPath   string `env:"OUTPIPE_CONFIG_PATH" envDefault:".config/outpipe/config.json"`
}

type ServiceConfig struct {
	InternalAPIURL    string `env:"INTERNAL_API_URL" envDefault:"http://127.0.0.1:9090"`
	InternalAPISecret string `env:"INTERNAL_API_SECRET"`
}

type AppConfig struct {
	Port                  string        `env:"PORT" envDefault:"8080"`
	InternalListenAddress string        `env:"INTERNAL_LISTEN_ADDRESS" envDefault:"127.0.0.1:9090"`
	Name                  string        `env:"APP_NAME" envDefault:"outpipe"`
	Environment           string        `env:"ENV" envDefault:"development"`
	LogLevel              string        `env:"LOG_LEVEL" envDefault:"info"`
	ShutdownTimeout       time.Duration `env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
	AllowedOrigins        string        `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:3000,http://localhost:3001"`
	CORSOrigin            string        `env:"CORS_ORIGIN" envDefault:"http://localhost:3000"`
	PublicAPIURL          string        `env:"PUBLIC_API_URL" envDefault:"http://localhost:8080"`
	DashboardURL          string        `env:"DASHBOARD_URL" envDefault:"http://localhost:3000"`
	ACMEEmail             string        `env:"ACME_EMAIL"`
	ACMEDirectory         string        `env:"ACME_DIRECTORY"`
	CertificateCache      string        `env:"CERTIFICATE_CACHE_DIR" envDefault:".data/acme"`
	RequireTLS            bool          `env:"REQUIRE_TLS" envDefault:"false"`
	TLSCertFile           string        `env:"TLS_CERT_FILE"`
	TLSKeyFile            string        `env:"TLS_KEY_FILE"`
}

type AuthConfig struct {
	SessionTTL         time.Duration `env:"SESSION_TTL" envDefault:"720h"`
	DeviceLoginTTL     time.Duration `env:"DEVICE_LOGIN_TTL" envDefault:"10m"`
	InvitationTTL      time.Duration `env:"INVITATION_TTL" envDefault:"168h"`
	OAuthStateTTL      time.Duration `env:"OAUTH_STATE_TTL" envDefault:"10m"`
	CookieName         string        `env:"AUTH_COOKIE_NAME" envDefault:"outpipe_session"`
	CookieSecure       bool          `env:"AUTH_COOKIE_SECURE" envDefault:"false"`
	CookieDomain       string        `env:"AUTH_COOKIE_DOMAIN"`
	GoogleClientID     string        `env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string        `env:"GOOGLE_CLIENT_SECRET"`
	GitHubClientID     string        `env:"GITHUB_CLIENT_ID"`
	GitHubClientSecret string        `env:"GITHUB_CLIENT_SECRET"`
	EncryptionKey      string        `env:"AUTH_ENCRYPTION_KEY"`
}

type DatabaseConfig struct {
	URL         string        `env:"DATABASE_URL" envDefault:"postgres://outpipe:outpipe@localhost:5432/outpipe?sslmode=disable"`
	MaxConns    int           `env:"DATABASE_MAX_CONNS" envDefault:"25"`
	MaxLifetime time.Duration `env:"DB_CONN_MAX_LIFETIME" envDefault:"30m"`
	MaxIdleTime time.Duration `env:"DB_CONN_MAX_IDLE_TIME" envDefault:"5m"`
}

type RedisConfig struct {
	Host     string `env:"REDIS_HOST" envDefault:"localhost"`
	Port     string `env:"REDIS_PORT" envDefault:"6379"`
	Password string `env:"REDIS_PASSWORD"`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
}

type MailConfig struct {
	FromAddress string `env:"MAIL_FROM" envDefault:"noreply@localhost"`
	ZeptoAPIKey string `env:"ZEPTO_API_KEY"`
	ZeptoURL    string `env:"ZEPTO_URL" envDefault:"https://api.zeptomail.com/v1.1/email"`
}

type TunnelConfig struct {
	Domain          string        `env:"TUNNEL_DOMAIN" envDefault:"outpipe.app"`
	TokenTTL        time.Duration `env:"TUNNEL_TOKEN_TTL" envDefault:"24h"`
	MaxConnections  int           `env:"TUNNEL_MAX_CONNECTIONS" envDefault:"1000"`
	MaxTunnels      int           `env:"TUNNEL_MAX_TUNNELS" envDefault:"1000"`
	MaxBytes        int64         `env:"TUNNEL_MAX_BYTES" envDefault:"0"`
	MaxBandwidth    int64         `env:"TUNNEL_MAX_BANDWIDTH_BYTES" envDefault:"0"`
	RequireTLS      bool          `env:"TUNNEL_REQUIRE_TLS" envDefault:"false"`
	AgentInactivity time.Duration `env:"AGENT_INACTIVITY_TIMEOUT" envDefault:"90s"`
	Heartbeat       time.Duration `env:"TUNNEL_HEARTBEAT_INTERVAL" envDefault:"20s"`
	ReadTimeout     time.Duration `env:"TUNNEL_READ_TIMEOUT" envDefault:"90s"`
	DrainTimeout    time.Duration `env:"TUNNEL_DRAIN_TIMEOUT" envDefault:"10s"`
	MaxFrameBytes   int64         `env:"TUNNEL_MAX_FRAME_BYTES" envDefault:"16777216"`
}

type BillingConfig struct {
	GracePeriod             time.Duration `env:"BILLING_GRACE_PERIOD" envDefault:"72h"`
	PolarServer             string        `env:"POLAR_SERVER" envDefault:"sandbox"`
	PolarBaseURL            string        `env:"POLAR_BASE_URL" envDefault:"https://sandbox-api.polar.sh"`
	PolarAccessToken        string        `env:"POLAR_ACCESS_TOKEN"`
	PolarWebhookSecret      string        `env:"POLAR_WEBHOOK_SECRET"`
	PolarProductLink        string        `env:"POLAR_PRODUCT_LINK"`
	PolarProductRoute       string        `env:"POLAR_PRODUCT_ROUTE"`
	PolarProductEdge        string        `env:"POLAR_PRODUCT_EDGE"`
	PolarProductLinkYearly  string        `env:"POLAR_PRODUCT_LINK_YEARLY"`
	PolarProductRouteYearly string        `env:"POLAR_PRODUCT_ROUTE_YEARLY"`
	PolarProductEdgeYearly  string        `env:"POLAR_PRODUCT_EDGE_YEARLY"`
	PaystackBaseURL         string        `env:"PAYSTACK_BASE_URL" envDefault:"https://api.paystack.co"`
	PaystackSecret          string        `env:"PAYSTACK_SECRET_KEY"`
	WebhookSecret           string        `env:"BILLING_WEBHOOK_SECRET"`
}
