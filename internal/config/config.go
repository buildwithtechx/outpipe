package config

import "time"

type APIConfig struct {
	Analytics AnalyticsConfig `envPrefix:"OUTPIPE_" json:"analytics" yaml:"analytics"`
	App       AppConfig       `envPrefix:"OUTPIPE_" json:"app" yaml:"app"`
	Auth      AuthConfig      `envPrefix:"OUTPIPE_" json:"auth" yaml:"auth"`
	Database  DatabaseConfig  `envPrefix:"OUTPIPE_" json:"database" yaml:"database"`
	Redis     RedisConfig     `envPrefix:"OUTPIPE_" json:"redis" yaml:"redis"`
	Mail      MailConfig      `envPrefix:"OUTPIPE_" json:"mail" yaml:"mail"`
	Service   ServiceConfig   `envPrefix:"OUTPIPE_" json:"service" yaml:"service"`
	Billing   BillingConfig   `envPrefix:"OUTPIPE_" json:"billing" yaml:"billing"`
	Tunnel    TunnelConfig    `envPrefix:"OUTPIPE_" json:"tunnel" yaml:"tunnel"`
	Agent     AgentConfig     `json:"agent" yaml:"agent"`
}

type RelayConfig struct {
	App     AppConfig     `envPrefix:"OUTPIPE_" json:"app" yaml:"app"`
	Redis   RedisConfig   `envPrefix:"OUTPIPE_" json:"redis" yaml:"redis"`
	Tunnel  TunnelConfig  `envPrefix:"OUTPIPE_" json:"tunnel" yaml:"tunnel"`
	Service ServiceConfig `envPrefix:"OUTPIPE_" json:"service" yaml:"service"`
	RelayID string        `env:"OUTPIPE_RELAY_ID" json:"relayId" yaml:"relay_id"`
}

type CronConfig struct {
	App       AppConfig       `envPrefix:"OUTPIPE_" json:"app" yaml:"app"`
	Auth      AuthConfig      `envPrefix:"OUTPIPE_" json:"auth" yaml:"auth"`
	Analytics AnalyticsConfig `envPrefix:"OUTPIPE_" json:"analytics" yaml:"analytics"`
	Database  DatabaseConfig  `envPrefix:"OUTPIPE_" json:"database" yaml:"database"`
	Redis     RedisConfig     `envPrefix:"OUTPIPE_" json:"redis" yaml:"redis"`
	Service   ServiceConfig   `envPrefix:"OUTPIPE_" json:"service" yaml:"service"`
	Backup    BackupConfig    `envPrefix:"OUTPIPE_" json:"backup" yaml:"backup"`
	Mail      MailConfig      `envPrefix:"OUTPIPE_" json:"mail" yaml:"mail"`
	Metrics   bool            `json:"metrics" yaml:"metrics" env:"OUTPIPE_CRON_METRICS" envDefault:"false"`
}

type CheckConfig struct {
	App      AppConfig      `envPrefix:"OUTPIPE_" json:"app" yaml:"app"`
	Service  ServiceConfig  `envPrefix:"OUTPIPE_" json:"service" yaml:"service"`
	Database DatabaseConfig `envPrefix:"OUTPIPE_" json:"database" yaml:"database"`
	Redis    RedisConfig    `envPrefix:"OUTPIPE_" json:"redis" yaml:"redis"`
	RelayURL string         `env:"OUTPIPE_CHECK_RELAY_URL" json:"relayUrl" yaml:"relay_url"`
}

type CLIConfig struct {
	Version      int    `json:"version" yaml:"version" env:"-"`
	APIURL       string `json:"apiUrl" yaml:"api_url" env:"OUTPIPE_API_URL" envDefault:"https://api.outpipe.dev"`
	RelayURL     string `json:"relayUrl" yaml:"relay_url" env:"OUTPIPE_RELAY_URL" envDefault:"wss://relay.outpipe.app/v1/connect"`
	PublicDomain string `json:"publicDomain" yaml:"public_domain" env:"OUTPIPE_DOMAIN" envDefault:"outpipe.app"`
	APIKey       string `json:"apiKey" yaml:"api_key" env:"OUTPIPE_API_KEY"`
	AgentToken   string `json:"agentToken" yaml:"agent_token" env:"OUTPIPE_AGENT_TOKEN"`
	Password     string `json:"password" yaml:"password" env:"OUTPIPE_PASSWORD"`
	ConfigPath   string `json:"configPath" yaml:"config_path" env:"OUTPIPE_CONFIG_PATH"`
}

const CurrentCLIConfigVersion = 1

type AgentConfig struct {
	DefaultProvider string `json:"defaultProvider" yaml:"default_provider" env:"DEFAULT_AGENT_PROVIDER" envDefault:"mock"`
	Model           string `json:"model" yaml:"model" env:"AGENT_MODEL"`
	OpenAIKey       string `json:"-" yaml:"openai_api_key" env:"OPENAI_API_KEY"`
	XAIKey          string `json:"-" yaml:"xai_api_key" env:"XAI_API_KEY"`
	AnthropicKey    string `json:"-" yaml:"anthropic_api_key" env:"ANTHROPIC_API_KEY"`
}

type ServiceConfig struct {
	InternalAPIURL    string `json:"internalApiUrl" yaml:"internal_api_url" env:"INTERNAL_API_URL" envDefault:"http://127.0.0.1:9090"`
	InternalAPISecret string `json:"-" yaml:"internal_api_secret" env:"INTERNAL_API_SECRET"`
}

type BackupConfig struct {
	Directory     string `json:"directory" yaml:"directory" env:"BACKUP_DIRECTORY"`
	DatabaseURL   string `json:"-" yaml:"database_url" env:"DATABASE_URL"`
	Keep          int    `json:"keep" yaml:"keep" env:"BACKUP_KEEP" envDefault:"7"`
	PgDumpPath    string `json:"pgDumpPath" yaml:"pg_dump_path" env:"BACKUP_PG_DUMP_PATH" envDefault:"pg_dump"`
	PgRestorePath string `json:"pgRestorePath" yaml:"pg_restore_path" env:"BACKUP_PG_RESTORE_PATH" envDefault:"pg_restore"`
}

type AppConfig struct {
	Port                   string        `json:"port" yaml:"port" env:"PORT" envDefault:"8080"`
	InternalListenAddress  string        `json:"internalListenAddress" yaml:"internal_listen_address" env:"INTERNAL_LISTEN_ADDRESS" envDefault:"127.0.0.1:9090"`
	Name                   string        `json:"name" yaml:"name" env:"APP_NAME" envDefault:"outpipe"`
	Environment            string        `json:"environment" yaml:"environment" env:"ENV" envDefault:"development"`
	LogLevel               string        `json:"logLevel" yaml:"log_level" env:"LOG_LEVEL" envDefault:"info"`
	ShutdownTimeout        time.Duration `json:"shutdownTimeout" yaml:"shutdown_timeout" env:"SHUTDOWN_TIMEOUT" envDefault:"10s"`
	AllowedOrigins         string        `json:"allowedOrigins" yaml:"allowed_origins" env:"ALLOWED_ORIGINS" envDefault:"http://localhost:3000,http://localhost:3001"`
	CORSOrigin             string        `json:"corsOrigin" yaml:"cors_origin" env:"CORS_ORIGIN" envDefault:"http://localhost:3000"`
	PublicAPIURL           string        `json:"publicApiUrl" yaml:"public_api_url" env:"PUBLIC_API_URL" envDefault:"http://localhost:8080"`
	DashboardURL           string        `json:"dashboardUrl" yaml:"dashboard_url" env:"DASHBOARD_URL" envDefault:"http://localhost:3000"`
	StatusBaseURL          string        `json:"statusBaseUrl" yaml:"status_base_url" env:"STATUS_BASE_URL" envDefault:"https://status.outpipe.dev"`
	ACMEEnabled            bool          `json:"acmeEnabled" yaml:"acme_enabled" env:"ACME_ENABLED" envDefault:"false"`
	ACMEEmail              string        `json:"acmeEmail" yaml:"acme_email" env:"ACME_EMAIL"`
	ACMEDirectory          string        `json:"acmeDirectory" yaml:"acme_directory" env:"ACME_DIRECTORY"`
	CertificateCache       string        `json:"certificateCacheDir" yaml:"certificate_cache_dir" env:"CERTIFICATE_CACHE_DIR" envDefault:".data/acme"`
	ChallengeZone          string        `json:"challengeZone" yaml:"challenge_zone" env:"CHALLENGE_ZONE"`
	ChallengeDNSProvider   string        `json:"challengeDnsProvider" yaml:"challenge_dns_provider" env:"CHALLENGE_DNS_PROVIDER"`
	ChallengeDNSCredential string        `json:"-" yaml:"challenge_dns_credential" env:"CHALLENGE_DNS_CREDENTIAL"`
	RequireTLS             bool          `json:"requireTls" yaml:"require_tls" env:"REQUIRE_TLS" envDefault:"false"`
	TLSCertFile            string        `json:"tlsCertFile" yaml:"tls_cert_file" env:"TLS_CERT_FILE"`
	TLSKeyFile             string        `json:"tlsKeyFile" yaml:"tls_key_file" env:"TLS_KEY_FILE"`
}

type AuthConfig struct {
	SessionTTL         time.Duration `json:"sessionTtl" yaml:"session_ttl" env:"SESSION_TTL" envDefault:"720h"`
	DeviceLoginTTL     time.Duration `json:"deviceLoginTtl" yaml:"device_login_ttl" env:"DEVICE_LOGIN_TTL" envDefault:"10m"`
	InvitationTTL      time.Duration `json:"invitationTtl" yaml:"invitation_ttl" env:"INVITATION_TTL" envDefault:"168h"`
	OAuthStateTTL      time.Duration `json:"oauthStateTtl" yaml:"oauth_state_ttl" env:"OAUTH_STATE_TTL" envDefault:"10m"`
	CookieName         string        `json:"cookieName" yaml:"cookie_name" env:"AUTH_COOKIE_NAME" envDefault:"outpipe_session"`
	CookieSecure       bool          `json:"cookieSecure" yaml:"cookie_secure" env:"AUTH_COOKIE_SECURE" envDefault:"false"`
	CookieDomain       string        `json:"cookieDomain" yaml:"cookie_domain" env:"AUTH_COOKIE_DOMAIN"`
	GoogleClientID     string        `json:"googleClientId" yaml:"google_client_id" env:"GOOGLE_CLIENT_ID"`
	GoogleClientSecret string        `json:"-" yaml:"google_client_secret" env:"GOOGLE_CLIENT_SECRET"`
	GitHubClientID     string        `json:"githubClientId" yaml:"github_client_id" env:"GITHUB_CLIENT_ID"`
	GitHubClientSecret string        `json:"-" yaml:"github_client_secret" env:"GITHUB_CLIENT_SECRET"`
	EncryptionKey      string        `json:"-" yaml:"encryption_key" env:"AUTH_ENCRYPTION_KEY"`
}

type DatabaseConfig struct {
	URL         string        `json:"-" yaml:"url" env:"DATABASE_URL" envDefault:"postgres://outpipe:outpipe@localhost:5432/outpipe?sslmode=disable"`
	MaxConns    int           `json:"maxConns" yaml:"max_conns" env:"DATABASE_MAX_CONNS" envDefault:"50"`
	MaxLifetime time.Duration `json:"maxLifetime" yaml:"max_lifetime" env:"DB_CONN_MAX_LIFETIME" envDefault:"30m"`
	MaxIdleTime time.Duration `json:"maxIdleTime" yaml:"max_idle_time" env:"DB_CONN_MAX_IDLE_TIME" envDefault:"5m"`
}

type RedisConfig struct {
	Host     string `json:"host" yaml:"host" env:"REDIS_HOST" envDefault:"localhost"`
	Port     string `json:"port" yaml:"port" env:"REDIS_PORT" envDefault:"6379"`
	Password string `json:"-" yaml:"password" env:"REDIS_PASSWORD"`
	DB       int    `json:"db" yaml:"db" env:"REDIS_DB" envDefault:"0"`
}

type MailConfig struct {
	FromAddress string `json:"fromAddress" yaml:"from_address" env:"MAIL_FROM" envDefault:"noreply@localhost"`
	Support     string `json:"support" yaml:"support" env:"SUPPORT_EMAIL" envDefault:"support@outpipe.dev"`
	ZeptoAPIKey string `json:"-" yaml:"zepto_api_key" env:"ZEPTO_API_KEY"`
	ZeptoURL    string `json:"zeptoUrl" yaml:"zepto_url" env:"ZEPTO_URL" envDefault:"https://api.zeptomail.com/v1.1/email"`
}

type TunnelConfig struct {
	Domain          string        `json:"domain" yaml:"domain" env:"TUNNEL_DOMAIN" envDefault:"outpipe.app"`
	TokenTTL        time.Duration `json:"tokenTtl" yaml:"token_ttl" env:"TUNNEL_TOKEN_TTL" envDefault:"24h"`
	MaxConnections  int           `json:"maxConnections" yaml:"max_connections" env:"TUNNEL_MAX_CONNECTIONS" envDefault:"1000"`
	MaxTunnels      int           `json:"maxTunnels" yaml:"max_tunnels" env:"TUNNEL_MAX_TUNNELS" envDefault:"1000"`
	MaxBytes        int64         `json:"maxBytes" yaml:"max_bytes" env:"TUNNEL_MAX_BYTES" envDefault:"0"`
	MaxBandwidth    int64         `json:"maxBandwidth" yaml:"max_bandwidth" env:"TUNNEL_MAX_BANDWIDTH_BYTES" envDefault:"0"`
	RequireTLS      bool          `json:"requireTls" yaml:"require_tls" env:"TUNNEL_REQUIRE_TLS" envDefault:"false"`
	AgentInactivity time.Duration `json:"agentInactivityTimeout" yaml:"agent_inactivity_timeout" env:"AGENT_INACTIVITY_TIMEOUT" envDefault:"90s"`
	Heartbeat       time.Duration `json:"heartbeatInterval" yaml:"heartbeat_interval" env:"TUNNEL_HEARTBEAT_INTERVAL" envDefault:"20s"`
	ReadTimeout     time.Duration `json:"readTimeout" yaml:"read_timeout" env:"TUNNEL_READ_TIMEOUT" envDefault:"90s"`
	DrainTimeout    time.Duration `json:"drainTimeout" yaml:"drain_timeout" env:"TUNNEL_DRAIN_TIMEOUT" envDefault:"10s"`
	MaxFrameBytes   int64         `json:"maxFrameBytes" yaml:"max_frame_bytes" env:"TUNNEL_MAX_FRAME_BYTES" envDefault:"16777216"`
	CaptureSpoolDir string        `json:"captureSpoolDir" yaml:"capture_spool_dir" env:"TUNNEL_CAPTURE_SPOOL_DIR" envDefault:"data/capture-spool"`
}

type BillingConfig struct {
	GracePeriod             time.Duration `json:"gracePeriod" yaml:"grace_period" env:"BILLING_GRACE_PERIOD" envDefault:"72h"`
	PolarServer             string        `json:"polarServer" yaml:"polar_server" env:"POLAR_SERVER" envDefault:"sandbox"`
	PolarBaseURL            string        `json:"polarBaseUrl" yaml:"polar_base_url" env:"POLAR_BASE_URL" envDefault:"https://sandbox-api.polar.sh"`
	PolarAccessToken        string        `json:"-" yaml:"polar_access_token" env:"POLAR_ACCESS_TOKEN"`
	PolarWebhookSecret      string        `json:"-" yaml:"polar_webhook_secret" env:"POLAR_WEBHOOK_SECRET"`
	PolarProductLink        string        `json:"polarProductLink" yaml:"polar_product_link" env:"POLAR_PRODUCT_LINK"`
	PolarProductRoute       string        `json:"polarProductRoute" yaml:"polar_product_route" env:"POLAR_PRODUCT_ROUTE"`
	PolarProductEdge        string        `json:"polarProductEdge" yaml:"polar_product_edge" env:"POLAR_PRODUCT_EDGE"`
	PolarProductLinkYearly  string        `json:"polarProductLinkYearly" yaml:"polar_product_link_yearly" env:"POLAR_PRODUCT_LINK_YEARLY"`
	PolarProductRouteYearly string        `json:"polarProductRouteYearly" yaml:"polar_product_route_yearly" env:"POLAR_PRODUCT_ROUTE_YEARLY"`
	PolarProductEdgeYearly  string        `json:"polarProductEdgeYearly" yaml:"polar_product_edge_yearly" env:"POLAR_PRODUCT_EDGE_YEARLY"`
	PaystackBaseURL         string        `json:"paystackBaseUrl" yaml:"paystack_base_url" env:"PAYSTACK_BASE_URL" envDefault:"https://api.paystack.co"`
	PaystackSecret          string        `json:"-" yaml:"paystack_secret_key" env:"PAYSTACK_SECRET_KEY"`
	WebhookSecret           string        `json:"-" yaml:"billing_webhook_secret" env:"BILLING_WEBHOOK_SECRET"`
}
