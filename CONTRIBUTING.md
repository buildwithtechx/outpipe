# Contributing to Outpipe Core Engine

Contributions are welcome for the standalone tunnel server, relay agent, developer CLI, cron worker, health checker, protocol schemas, and deployment configurations.

Frontend applications live in `buildwithtechx/outpipe-frontend`.
Multi-language SDKs live in `buildwithtechx/outpipe-sdks`.

## Before contributing

- Read the project documentation and security policy (`SECURITY.md`).
- Search existing issues and pull requests.
- Open an issue for significant architectural or protocol changes before implementation.
- Never include credentials, private tunnel URLs, customer data, or production configuration.

## Development setup

Requirements:

- Go 1.26.8 or newer.
- Docker & Docker Compose (for Postgres and Timescale integration testing).

Compile and test all core binaries:

```sh
# Compile all 5 core binaries
make build
# Or directly via Go:
go build -o bin/ ./cmd/server ./cmd/tunnel ./cmd/cron ./cmd/check ./cmd/cli

# Run formatting, vetting, and tests
gofmt -w .
go vet ./...
go test -race ./...
```

## Pull requests

- Use a focused branch and describe the user-visible or operational impact.
- Include tests for protocol, authentication, routing, limits, or persistence changes.
- Update relevant documentation when behavior or architecture changes.
- Run `gofmt -w .` and `go vet ./...` before requesting review.
- Keep files under 350 lines per `AGENTS.md`.

## Commit guidance

Use concise imperative commit messages that describe one change, such as:

```text
Add tunnel session expiry handling
```
