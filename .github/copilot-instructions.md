# Outpipe

Outpipe is an independent tunneling platform. It exposes local and private services through secure public endpoints and works with hosted or self-hosted tunnel servers.

## Product boundary

The tunnel core must not depend on Outpipe. Outpipe support is an optional adapter under `integrations/outpipe/` that uses the public tunnel API.

## Repository

- Go binaries live in `cmd/`.
- Private Go code lives in `internal/`.
- Reusable Go packages live in `pkg/`.
- Protocol schemas and generated bindings live in `protocol/`.
- The standalone dashboard lives in `apps/web/`.
- The Tauri desktop shell lives in `apps/desktop/`.
- TypeScript SDKs live in `packages/`.
- Migrations, tests, deployment files, and documentation live in their top-level directories.

## Go conventions

- Use layered packages for API, auth, relay, routing, sessions, storage, telemetry, and workers.
- Use `modernc.org/sqlite` and checked error handling.
- Wrap errors with operation context.
- Use dependency injection instead of global mutable state.
- Avoid `init()`.
- Keep Go files below 350 lines and use snake_case filenames.

## Dashboard and desktop conventions

- Use TanStack Router file conventions under `apps/web/src/routes/`.
- Do not edit generated route trees manually.
- Use Zustand, TanStack Query, Zod, Radix UI, and Tailwind utilities consistently.
- Use kebab-case TypeScript filenames and named exports.
- Keep Tauri as a desktop shell; tunnel networking belongs in Go.
- Format TypeScript with Biome and never use Prettier.

## Security

- Use short-lived, scoped credentials for users and agents.
- Validate tunnel ownership for every operation and stream.
- Prevent cross-tunnel access and unintended private-network access.
- Add tests for replay, SSRF, authorization bypass, resource exhaustion, and credential leakage.
- Never commit secrets, private tunnel URLs, customer data, or production configuration.
