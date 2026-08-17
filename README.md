# Codedock Tunnel

Codedock Tunnel is an independent tunneling platform for exposing local and private services through secure public endpoints.

It is designed to work without Codedock. Codedock is an optional integration that can create and manage tunnels through the public Codedock Tunnel API.

## Hosted domains

- `codedock-tunnel.dev` serves the Tunnel web dashboard.
- `api.codedock-tunnel.dev` serves the Tunnel control-plane API.
- `tunnel.codedock-tunnel.dev` serves the public tunnel relay and generated tunnel endpoints.
- `cli.codedock-tunnel.dev` serves the CLI installer and release assets.
- `desktop.codedock-tunnel.dev` serves Tunnel Desktop installers.

## Product surfaces

- Go tunnel server and public relay
- Standalone `codedock-tunnel` CLI
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
- `integrations/codedock` contains the optional Codedock adapter.

## Development

Requirements:

- Go 1.25 or newer
- Node.js 22 or newer
- npm 10 or newer

```sh
npm install
npm run dev
```

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
docker build -f docker/Dockerfile.api -t codedock-api .
docker build -f docker/Dockerfile.tunnel -t codedock-tunnel-server .
docker build -f docker/Dockerfile.cron -t codedock-tunnel-cron .
docker build -f docker/Dockerfile.check -t codedock-tunnel-check .
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

## Provisioning platform administrators

Users authenticate with Google or GitHub before they can be granted platform-admin access. The server does not promote users automatically during signup. After the user has signed in once, provision an administrator explicitly:

```sh
./bin/codedock-api bootstrap-admin --email owner@example.com
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
curl -fsSL https://cli.codedock-tunnel.dev | bash
```

`cli.codedock-tunnel.dev` should serve this installer, while `cli.codedock-tunnel.dev/releases/cli` should serve versioned release assets. The installer falls back to GitHub Releases if the downloads path is unavailable. It installs the `codedock-tunnel` CLI to `$HOME/.local/bin` by default. Windows users download the Tunnel CLI release asset directly.

The CLI and desktop app are separate products. The CLI is a terminal binary for local tunnels, automation, and CI. The desktop app is a Tauri GUI distributed through platform installers; installing one does not install the other.

```sh
codedock-tunnel login --server https://api.codedock-tunnel.dev
codedock-tunnel http 3000
```

## Design boundary

The tunnel core owns tunnel identity, credentials, sessions, routing, quotas, analytics, and audit history. It must not import Codedock models, routes, authentication, or database packages.

Codedock integrations communicate with the public tunnel API and may be disabled without affecting standalone tunnel operation.

Edge routing, wildcard DNS, certificate mounting, and optional DNS automation are documented in [docs/edge-routing.md](docs/edge-routing.md).
