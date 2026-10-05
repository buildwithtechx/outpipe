# Outpipe

Outpipe is a hosted SaaS for exposing local and private services through secure public endpoints, with open-source clients and server code.

Users connect to the hosted service through the dashboard, CLI, desktop app or SDKs. Database, Redis, internal authentication, TLS, backups and provider credentials belong to the service operator.

## Hosted domains

- `outpipe.dev` serves the Outpipe web dashboard.
- `api.outpipe.dev` serves the Outpipe control-plane API.
- `relay.outpipe.app` serves the public tunnel relay, and `*.outpipe.app` serves generated tunnel endpoints (`myapp.outpipe.app`, for example).
- `cli.outpipe.dev` serves the CLI installer and release assets.
- `desktop.outpipe.dev` serves Outpipe Desktop installers.

## Product surfaces

- Go tunnel server and public relay
- Standalone `outpipe` CLI
- React dashboard for accounts, organizations, tunnels, access policies, and analytics
- Tauri desktop application
- Go and TypeScript client SDKs
- Hosted deployment with open-source clients and integrations

## Repository layout

- `cmd/server` runs the control-plane API.
- `cmd/tunnel` runs the public tunnel relay and data plane.
- `cmd/cli` provides the standalone CLI.
- `internal` contains private server, relay, routing, storage, and authentication code.
- `protocol` contains language-neutral protocol schemas and generated bindings.
- `apps/web` contains the standalone web interface.
- `apps/desktop` contains the Tauri desktop shell.
- `integrations/outpipe` contains the optional Outpipe adapter.

## Hosted usage

The CLI defaults to the hosted API and relay; users do not need an environment
file or server configuration. Run `outpipe login`, then
`outpipe open --port 3000`. Credentials are saved in your OS user configuration
directory and reused across projects. SDK clients need only a scoped credential.
Machine automation uses `OUTPIPE_TOKEN` for secrets and tunnel creation, with
the appropriate scopes. Endpoint overrides are available for local development.

## Development

The root and server environment examples below configure the service operator's
development environment. They are not requirements for SaaS customers.

Requirements:

- Go 1.26.8 or newer
- Node.js 22 or newer
- npm 10 or newer

SDK clients default to `https://api.outpipe.dev`; only credentials are needed
for the hosted service. Override `apiUrl` (TypeScript/Angular), `BaseURL`
(Go), `baseUrl` (PHP), or `.base_url(...)` (Rust) for a self-hosted API.

After updating Go dependencies, run `go mod vendor` if you have a local
`vendor/` directory. Vendored dependencies are not tracked in this repository.

Generate the local self-signed TLS certificate and key with
`go run ./scripts/generate_local_tls.go`. The generated files live in
`data/tls/`, are ignored by Git, and are preserved when the command is rerun.
The environment examples point to these files; TLS remains disabled until
explicitly enabled. ACME also requires `OUTPIPE_ACME_ENABLED=true` and an
actual contact email. Examples use Let's Encrypt staging; production issuance
requires the production ACME directory.

The examples configure relay readiness at `http://127.0.0.1:8081` and backups
in `data/backups/`. Backups require installed `pg_dump` and `pg_restore`
executables. API, agent and machine-token example values are placeholders;
issue real credentials for your organization before using them. Keep generated
secrets in your ignored `.env`.

```sh
npm install
npm run dev
```

This starts the API on port 8080, tunnel relay on 8081, dashboard on 3000,
secrets-share on 4321, and status app on 4322. Use `npm run dev:web`,
`npm run dev:status`, or `npm run dev:secrets-share` to run an individual app.

## Common commands

```sh
npm run build
npm run fmt
npm run typecheck
npm run test
make docker-build
```

## Deployment roles

The Go commands are independently deployable and do not all need to run on the same machine.

| Command      | Deployment                                                      | Lifecycle                                |
| ------------ | --------------------------------------------------------------- | ---------------------------------------- |
| `cmd/server` | Public VPS or container                                         | Long-running control-plane API           |
| `cmd/tunnel` | Public VPS or container                                         | Long-running tunnel relay and data plane |
| `cmd/cron`   | Long-running worker under a process manager or worker container | Scheduled maintenance jobs               |
| `cmd/check`  | Internal service container or process-manager service           | HTTP domain and edge verification        |
| `cmd/cli`    | User workstation binary or package manager                      | Opens tunnels and calls the API          |

Each server-side command has its own Dockerfile:

```sh
docker build -f docker/Dockerfile.api -t outpipe-api .
docker build -f docker/Dockerfile.tunnel -t outpipe-server .
docker build -f docker/Dockerfile.cron -t outpipe-cron .
docker build -f docker/Dockerfile.check -t outpipe-check .
```

The CLI is distributed as a platform binary. It runs on the user’s workstation or CI runner and connects to the API and relay; it is not deployed as a server process.

Each independently deployed command has a focused environment example:

| Command       | Environment example       |
| ------------- | ------------------------- |
| API server    | `cmd/server/.env.example` |
| Tunnel relay  | `cmd/tunnel/.env.example` |
| Cron worker   | `cmd/cron/.env.example`   |
| Check service | `cmd/check/.env.example`  |
| CLI           | `cmd/cli/.env.example`    |

See [docs/configuration.md](docs/configuration.md) for descriptions of every variable.

For local development, use `docker compose -f docker/docker-compose.yml up`. That
file intentionally uses development credentials, localhost URLs, and a source
mount. Production deployments should build and run each service from its own
Dockerfile; provide `OUTPIPE_*` values through the deployment environment or a
secret manager, never inside the image.

## Provisioning platform administrators

Users authenticate with Google or GitHub before they can be granted platform-admin access. The server does not promote users automatically during signup. After the user has signed in once, provision an administrator explicitly:

```sh
./bin/outpipe-api bootstrap-admin --email owner@example.com
```

Use `--name` to set a separate admin display name. The command assigns the `owner` role in the `platform_admins` table.

## Runtime communication

The API owns accounts, organizations, OAuth sessions, billing, and tunnel metadata. The tunnel relay owns public ingress and data-plane forwarding. The CLI calls the API for authentication and tunnel management, then opens the tunnel WebSocket to the relay. Cron shares backend storage and providers, while check exposes a private HTTP endpoint that an edge proxy can call before accepting a custom domain.

```text
CLI ── HTTPS ─────▶ API ── control ──▶ TUNNEL RELAY ◀── TLS/WebSocket ── CLI ──▶ local application

CRON ── PostgreSQL / Redis ──▶ backend state
CHECK ◀── private HTTP request ── edge proxy
```

## Standalone usage

The CLI must support a configurable tunnel server URL so users can connect to the hosted service, a private installation, or a local development server.

After a release, Unix users can install the CLI with the branded domain:

```sh
curl -fsSL https://cli.outpipe.dev | bash
```

`cli.outpipe.dev` should serve this installer, while `cli.outpipe.dev/releases/cli` should serve versioned release assets. The installer falls back to GitHub Releases if the downloads path is unavailable. It installs the `outpipe` CLI to `$HOME/.local/bin` by default. Windows users download the Outpipe CLI release asset directly.

The CLI and desktop app are separate products. The CLI is a terminal binary for local tunnels, automation, and CI. The desktop app is a Tauri GUI distributed through platform installers; installing one does not install the other.

```sh
outpipe login --server https://api.outpipe.dev
outpipe http 3000
```

## Design boundary

The tunnel core owns tunnel identity, credentials, sessions, routing, quotas, analytics, and audit history. It must not import Outpipe models, routes, authentication, or database packages.

Outpipe integrations communicate with the public tunnel API and may be disabled without affecting standalone tunnel operation.

Edge routing, wildcard DNS, certificate mounting, and optional DNS automation are documented in [docs/edge-routing.md](docs/edge-routing.md).
