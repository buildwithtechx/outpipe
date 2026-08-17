# Agent Instructions

## Project boundary

Codedock Tunnel is an independent tunneling product. Its core must work without Codedock and must not import Codedock models, routes, authentication, or database packages.

Codedock integration belongs under `integrations/codedock/` and communicates through the public tunnel API.

## Code style

- Keep files under 350 lines.
- Keep one component per TypeScript file.
- Use kebab-case for TypeScript files and snake_case for Go files.
- Use named exports in TypeScript.
- Avoid comments, GoDoc, and JSDoc unless required by a tool or language directive.
- Always check and wrap Go errors with useful context.
- Add JSON tags to exported Go struct fields.
- Use `modernc.org/sqlite` for SQLite and the official Docker client where needed.
- Avoid global mutable state and `init()` functions.

## Go architecture

- `internal/models/` contains domain models and DTOs.
- `internal/repositories/` contains persistence implementations.
- `internal/services/` contains business logic and integrations.
- `internal/handlers/` contains HTTP controllers.
- `internal/http/` contains server, routes, and middleware wiring.
- `internal/engine/` contains relay, proxy, routing, and runtime engines.
- `pkg/` contains intentionally reusable public packages.
- `cmd/` contains binary entrypoints.

## Dashboard and desktop

- Dashboard routes live in `apps/web/src/routes/`.
- Dashboard components live in `apps/web/src/components/` and domain feature folders.
- Dashboard hooks live in `apps/web/src/hooks/`.
- Dashboard utilities live in `apps/web/src/lib/`.
- Do not edit generated route files by hand.
- Keep Tauri native code in `apps/desktop/src-tauri/`.
- Keep tunnel networking and data-plane logic in Go rather than duplicating it in Tauri or TypeScript.
- Use Biome for TypeScript formatting. Never use Prettier.

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
- Do not run builds or tests after every small change.
- Run `npm run fmt` before finishing a change session.
- Run relevant Go, TypeScript, protocol, security, and integration checks for the affected area.
- Do not commit or push unless explicitly requested.
