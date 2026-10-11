# Agent Instructions

## Project boundary

Outpipe is an independent tunneling engine. Its core provides the tunnel server (`cmd/server`), tunnel relay client (`cmd/tunnel`), cron worker (`cmd/cron`), health check worker (`cmd/check`), and developer CLI (`cmd/cli`).

Frontend web and desktop applications live in the separate `outpipe-frontend` repository.
Multi-language client SDKs live in the separate `outpipe-sdks` repository.

## Code style

- Keep files under 350 lines.
- Use snake_case for Go files.
- Avoid comments and GoDoc unless required by a tool or language directive.
- Always check and wrap Go errors with useful context.
- Add JSON tags to exported Go struct fields.
- Use `modernc.org/sqlite` for SQLite and the official Docker client where needed.
- Avoid global mutable state and `init()` functions.
- Format all Go files with standard `gofmt`.

## Go architecture

- `internal/models/`: Domain models and DTOs.
- `internal/repositories/`: Persistence implementations (Postgres, Timescale, SQLite, Redis).
- `internal/services/`: Business logic, tunnel lifecycle, and background orchestration.
- `internal/handlers/`: HTTP/WS controllers.
- `internal/http/`: Server wiring, route registrations, and middleware.
- `internal/engine/`: Relay, reverse proxy, routing, and connection multiplexing engines.
- `pkg/`: Intentionally reusable public packages (`client`, `protocol`, `utils`, `version`).
- `cmd/`: Binary entrypoints (`server`, `tunnel`, `cron`, `check`, `cli`).

## Protocol and security

- Protocol schemas are language-neutral and live under `protocol/schema/`.
- Generated protocol bindings must not be edited by hand.
- Authenticate agents and users with short-lived, scoped credentials.
- Validate tunnel ownership and target authorization for every stream.
- Do not log tokens, cookies, credentials, request bodies, or private tunnel data.
- Treat public tunnel input as hostile and test SSRF, replay, cross-tunnel access, and resource exhaustion.

## Workflow

- Read files before editing them.
- Use targeted patches for edits.
- Run `gofmt -w .` before finishing a change session.
- Run `go vet ./...` and relevant `go test` checks for the affected area.
- Do not commit or push unless explicitly requested.
