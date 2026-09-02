# Configuration reference

All binaries read configuration from environment variables. Every variable is
prefixed with `OUTPIPE_`; the Go `env` package applies a `/`-free key naming
convention (lowercase) internally. Variables that are not present use the
defaults below, so an empty file or empty environment is always valid in
development.

Copy the relevant `.env.example` file next to your binary or systemd unit as a
starting point. The examples contain only variable names and defaults; the
explanations live in this document.

## Shared services

| Variable                        | Default                                                             | Applies to                  | Purpose                                                                    |
| ------------------------------- | ------------------------------------------------------------------- | --------------------------- | -------------------------------------------------------------------------- |
| `OUTPIPE_DATABASE_URL`          | `postgres://outpipe:outpipe@localhost:5432/outpipe?sslmode=disable` | server, cron                | PostgreSQL DSN.                                                            |
| `OUTPIPE_DATABASE_MAX_CONNS`    | `25`                                                                | server, cron                | Maximum database connections in the pool.                                  |
| `OUTPIPE_DB_CONN_MAX_LIFETIME`  | `30m`                                                               | server, cron                | Maximum connection lifetime before recycling.                              |
| `OUTPIPE_DB_CONN_MAX_IDLE_TIME` | `5m`                                                                | server, cron                | Maximum idle time before a connection is closed.                           |
| `OUTPIPE_REDIS_HOST`            | `localhost`                                                         | server, tunnel, cron, check | Redis host.                                                                |
| `OUTPIPE_REDIS_PORT`            | `6379`                                                              | server, tunnel, cron, check | Redis port.                                                                |
| `OUTPIPE_REDIS_PASSWORD`        | empty                                                               | server, tunnel, cron, check | Redis password.                                                            |
| `OUTPIPE_REDIS_DB`              | `0`                                                                 | server, tunnel, cron, check | Redis logical database index.                                              |
| `OUTPIPE_ENV`                   | `development`                                                       | all binaries                | Runtime environment label (`development`, `production`, ...).              |
| `OUTPIPE_APP_NAME`              | `outpipe`                                                           | server, tunnel, cron, check | Process name used in logs and startup banners.                             |
| `OUTPIPE_LOG_LEVEL`             | `info`                                                              | server                      | Logging verbosity.                                                         |
| `OUTPIPE_SHUTDOWN_TIMEOUT`      | `10s`                                                               | server                      | Grace period for in-flight work during shutdown.                           |
| `OUTPIPE_PUBLIC_API_URL`        | `http://localhost:8080`                                             | server                      | Externally reachable control-plane URL, used to build OAuth links.         |
| `OUTPIPE_DASHBOARD_URL`         | `http://localhost:3000`                                             | server                      | Dashboard origin used for post-OAuth redirects and email links.            |
| `OUTPIPE_ALLOWED_ORIGINS`       | `http://localhost:3000,http://localhost:3001`                       | server                      | CORS allow list. In production the hosted dashboard domain must be listed. |
| `OUTPIPE_CORS_ORIGIN`           | `http://localhost:3000`                                             | server                      | Legacy CORS origin used by the web build.                                  |

### Private internal API

The control plane exposes a second, loopback-only HTTP listener for
server-to-server traffic (agent authentication, managed tunnel policy, usage
ingestion, readiness). Only the API traffic itself and the loopback binding
protect it; a shared secret authenticates every caller.

| Variable                          | Default                 | Applies to                  | Purpose                                                                                 |
| --------------------------------- | ----------------------- | --------------------------- | --------------------------------------------------------------------------------------- |
| `OUTPIPE_INTERNAL_LISTEN_ADDRESS` | `127.0.0.1:9090`        | server                      | Bind address of the private internal listener. Keep it on loopback in production.       |
| `OUTPIPE_INTERNAL_API_URL`        | `http://127.0.0.1:9090` | tunnel, cron, check, cli    | Base URL used by relay, workers, and checks to reach the internal listener.             |
| `OUTPIPE_INTERNAL_API_SECRET`     | empty                   | server, tunnel, cron, check | Shared secret sent as `X-Internal-Secret` on internal requests. Required on both sides. |

## Server (cmd/server)

### Listen and TLS

| Variable                | Default | Purpose                                             |
| ----------------------- | ------- | --------------------------------------------------- |
| `OUTPIPE_PORT`          | `8080`  | Public control-plane listen port.                   |
| `OUTPIPE_REQUIRE_TLS`   | `false` | Serve HTTPS directly when true.                     |
| `OUTPIPE_TLS_CERT_FILE` | empty   | PEM certificate path (rotating; reloads on change). |
| `OUTPIPE_TLS_KEY_FILE`  | empty   | PEM key path.                                       |

### Automated certificates (ACME)

When `OUTPIPE_ACME_EMAIL` is set, the domain service issues Let's Encrypt
certificates for verified custom domains with `autocert`.

| Variable                        | Default      | Purpose                                                |
| ------------------------------- | ------------ | ------------------------------------------------------ |
| `OUTPIPE_ACME_EMAIL`            | empty        | Account email; empty disables ACME issuance.           |
| `OUTPIPE_ACME_DIRECTORY`        | empty        | Custom ACME directory URL (test/LetsEncrypt staging).  |
| `OUTPIPE_CERTIFICATE_CACHE_DIR` | `.data/acme` | Where issued certificates and account keys are cached. |

### Authentication and cookies

OAuth is the only authentication mechanism. Sessions use an HTTP-only cookie;
the encryption key protects OAuth state values.

| Variable                                                    | Default           | Purpose                                                                                                                                                                 |
| ----------------------------------------------------------- | ----------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `OUTPIPE_SESSION_TTL`                                       | `720h`            | Session validity after login.                                                                                                                                           |
| `OUTPIPE_DEVICE_LOGIN_TTL`                                  | `10m`             | Device-login code validity.                                                                                                                                             |
| `OUTPIPE_INVITATION_TTL`                                    | `168h`            | Organization invitation validity.                                                                                                                                       |
| `OUTPIPE_OAUTH_STATE_TTL`                                   | `10m`             | OAuth state validity.                                                                                                                                                   |
| `OUTPIPE_AUTH_COOKIE_NAME`                                  | `outpipe_session` | Cookie name.                                                                                                                                                            |
| `OUTPIPE_AUTH_COOKIE_SECURE`                                | `false`           | `Secure` flag; must be true once HTTPS terminates the cookie path.                                                                                                      |
| `OUTPIPE_AUTH_COOKIE_DOMAIN`                                | empty             | Shared cookie domain (for example `.outpipe.dev`) so the browser sends the session cookie to both the dashboard and the API subdomains. Empty keeps a host-only cookie. |
| `OUTPIPE_AUTH_ENCRYPTION_KEY`                               | empty             | 16/24/32-byte AES key for OAuth state and device flows. Required in production.                                                                                         |
| `OUTPIPE_GOOGLE_CLIENT_ID` / `OUTPIPE_GOOGLE_CLIENT_SECRET` | empty             | Google OAuth application.                                                                                                                                               |
| `OUTPIPE_GITHUB_CLIENT_ID` / `OUTPIPE_GITHUB_CLIENT_SECRET` | empty             | GitHub OAuth application.                                                                                                                                               |

### Mail

Zepto Mail is the only transactional email provider.

| Variable                | Default                                | Purpose                                        |
| ----------------------- | -------------------------------------- | ---------------------------------------------- |
| `OUTPIPE_MAIL_FROM`     | `noreply@localhost`                    | Sender address.                                |
| `OUTPIPE_SUPPORT_EMAIL` | `support@outpipe.dev`                  | Recipient for contact messages and bug reports. |
| `OUTPIPE_ZEPTO_URL`     | `https://api.zeptomail.com/v1.1/email` | Zepto API base URL.                            |
| `OUTPIPE_ZEPTO_API_KEY` | empty                                  | Zepto API token. Empty disables email sending. |

### Billing

The API owns plans and billing. Polar is the primary provider; Paystack is a
fallback provider. Product IDs are server-side only.

| Variable                                                               | Default                        | Purpose                                                      |
| ---------------------------------------------------------------------- | ------------------------------ | ------------------------------------------------------------ |
| `OUTPIPE_BILLING_GRACE_PERIOD`                                         | `72h`                          | Time after payment failure before the subscription is reset. |
| `OUTPIPE_POLAR_SERVER`                                                 | `sandbox`                      | `sandbox` or `production`.                                   |
| `OUTPIPE_POLAR_BASE_URL`                                               | `https://sandbox-api.polar.sh` | Polar API base URL.                                          |
| `OUTPIPE_POLAR_ACCESS_TOKEN`                                           | empty                          | Polar authentication token.                                  |
| `OUTPIPE_POLAR_WEBHOOK_SECRET`                                         | empty                          | Polar webhook verification secret.                           |
| `OUTPIPE_POLAR_PRODUCT_LINK` / `_ROUTE` / `_EDGE`                      | empty                          | Monthly product IDs.                                         |
| `OUTPIPE_POLAR_PRODUCT_LINK_YEARLY` / `_ROUTE_YEARLY` / `_EDGE_YEARLY` | empty                          | Annual product IDs for the yearly pricing toggle.            |
| `OUTPIPE_PAYSTACK_BASE_URL`                                            | `https://api.paystack.co`      | Paystack API base URL.                                       |
| `OUTPIPE_PAYSTACK_SECRET_KEY`                                          | empty                          | Paystack secret key.                                         |
| `OUTPIPE_BILLING_WEBHOOK_SECRET`                                       | empty                          | Shared webhook secret for non-provider-signed webhooks.      |

## Relay (cmd/tunnel)

The relay terminates `wss://relay.<domain>/v1/connect` agent connections and
serves public `*.outpipe.app` HTTP traffic through the engine proxy. The
public data plane is entirely separate from the control-plane handlers.

| Variable                                         | Default       | Purpose                                                                                                                                                                                           |
| ------------------------------------------------ | ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `OUTPIPE_PORT`                                   | `8080`        | Public relay listen port (use `8081` when colocated with the API).                                                                                                                                |
| `OUTPIPE_RELAY_ID`                               | generated     | Stable identity used for Redis affinity claims. Set it explicitly when scaling out.                                                                                                               |
| `OUTPIPE_TUNNEL_DOMAIN`                          | `outpipe.app` | Tunnel base domain. One wildcard DNS record (`*.outpipe.app`) must point at the relay. Use `outpipe.localhost` in local development. (The CLI calls its own copy of this value `OUTPIPE_DOMAIN`.) |
| `OUTPIPE_REQUIRE_TLS`                            | `false`       | Serve HTTPS directly when true. When a reverse proxy or CDN terminates HTTPS, leave false.                                                                                                        |
| `OUTPIPE_TLS_CERT_FILE` / `OUTPIPE_TLS_KEY_FILE` | empty         | PEM certificate and key for direct TLS.                                                                                                                                                           |
| `OUTPIPE_TUNNEL_TOKEN_TTL`                       | `24h`         | Agent token validity when issued by the API.                                                                                                                                                      |
| `OUTPIPE_TUNNEL_MAX_CONNECTIONS`                 | `1000`        | Concurrent agent connections per relay.                                                                                                                                                           |
| `OUTPIPE_TUNNEL_MAX_TUNNELS`                     | `1000`        | Concurrent open tunnels per relay.                                                                                                                                                                |
| `OUTPIPE_TUNNEL_MAX_BYTES`                       | `0`           | Maximum HTTP request body proxied per request (`0` = unlimited).                                                                                                                                  |
| `OUTPIPE_TUNNEL_MAX_BANDWIDTH_BYTES`             | `0`           | Per-organization bandwidth budget per relay (`0` = unlimited).                                                                                                                                    |
| `OUTPIPE_TUNNEL_HEARTBEAT_INTERVAL`              | `20s`         | Agent heartbeat interval.                                                                                                                                                                         |
| `OUTPIPE_TUNNEL_READ_TIMEOUT`                    | `90s`         | Stale connection deadline. Must exceed the heartbeat interval.                                                                                                                                    |
| `OUTPIPE_TUNNEL_MAX_FRAME_BYTES`                 | `16777216`    | Maximum single WebSocket frame.                                                                                                                                                                   |
| `OUTPIPE_TUNNEL_DRAIN_TIMEOUT`                   | `10s`         | Grace period when an agent reconnects into an existing tunnel.                                                                                                                                    |
| `OUTPIPE_AGENT_INACTIVITY_TIMEOUT`               | `90s`         | Sessions without heartbeats are reclaimed after this.                                                                                                                                             |
| `OUTPIPE_ALLOWED_ORIGINS`                        | empty         | Optional WebSocket `Origin` allow list; empty allows all.                                                                                                                                         |

## CLI (cmd/cli)

| Variable              | Default                       | Purpose                                                           |
| --------------------- | ----------------------------- | ----------------------------------------------------------------- |
| `OUTPIPE_API_URL`     | `http://localhost:8080`       | Control-plane API base URL.                                       |
| `OUTPIPE_RELAY_URL`   | `ws://localhost:8081`         | Relay WebSocket URL. May be `wss://` in production.               |
| `OUTPIPE_DOMAIN`      | `outpipe.app`                 | Public tunnel domain, used to normalize managed-tunnel hostnames. |
| `OUTPIPE_API_KEY`     | empty                         | API key for managed-tunnel and management commands.               |
| `OUTPIPE_AGENT_TOKEN` | empty                         | Agent token for ephemeral CI/CD tunnel opening.                   |
| `OUTPIPE_PASSWORD`    | empty                         | Default tunnel access password.                                   |
| `OUTPIPE_CONFIG_PATH` | `.config/outpipe/config.json` | Where `outpipe login` stores credentials (0600).                  |

## Cron (cmd/cron)

The cron binary runs maintenance jobs (cleanup, usage aggregation, retention,
billing reconciliation, Redis reconciliation, backups) under a Redis lease.
Schedule it on an interval (for example hourly); `RunOnce` executes every due
job once per invocation.

| Variable                         | Default      | Purpose                                                                                            |
| -------------------------------- | ------------ | -------------------------------------------------------------------------------------------------- |
| `OUTPIPE_BACKUP_DIRECTORY`       | empty        | Enables the backup job: `pg_dump` custom-format archives are written here. Empty disables backups. |
| `OUTPIPE_BACKUP_KEEP`            | `7`          | Number of archives to retain; older ones are pruned.                                               |
| `OUTPIPE_BACKUP_PG_DUMP_PATH`    | `pg_dump`    | Path to the `pg_dump` binary.                                                                      |
| `OUTPIPE_BACKUP_PG_RESTORE_PATH` | `pg_restore` | Path to the `pg_restore` binary, used to verify each archive (`--list`).                           |
| `OUTPIPE_CRON_METRICS`           | unset        | Set to `1` to print Prometheus `worker_health` gauges after the run.                               |

## Check (cmd/check)

The check binary probes readiness end-to-end and exits non-zero when any
configured check fails, for orchestration use. Checks run concurrently with a
shared timeout. `-json` switches to JSON output for monitoring systems.

| Variable                  | Default          | Purpose                                                                       |
| ------------------------- | ---------------- | ----------------------------------------------------------------------------- |
| `OUTPIPE_CHECK_RELAY_URL` | empty            | Relay URL to probe (`http(s)://` or `ws(s)://`); empty skips the relay check. |
| `OUTPIPE_DATABASE_URL`    | `postgres://...` | Database to probe directly; empty skips the database check.                   |
| `OUTPIPE_REDIS_HOST`      | `localhost`      | Redis to probe; empty skips the Redis check.                                  |

Flags override the environment: `-api-url`, `-relay-url`, `-database-url`,
`-redis-host`, `-timeout`, `-json`.
