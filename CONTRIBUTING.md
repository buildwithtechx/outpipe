# Contributing

Contributions are welcome for the standalone tunnel server, agent, CLI, dashboard, desktop app, SDKs, protocol, documentation, and deployment tooling.

## Before contributing

- Read the project documentation and security policy.
- Search existing issues and pull requests.
- Open an issue for significant design changes before implementation.
- Never include credentials, private tunnel URLs, customer data, or production configuration.

## Development setup

Requirements:

- Go 1.25 or newer.
- Node.js 22 or newer.
- npm 10 or newer.
- Rust and Tauri prerequisites for desktop work.

Install dependencies and run the available checks:

```sh
npm install
npm run fmt
npm run typecheck
npm run test
```

## Project boundaries

- Keep the tunnel core independent from Outpipe.
- Put Go server code in the appropriate `internal/` layer.
- Keep reusable public Go APIs under `pkg/`.
- Keep reusable TypeScript packages under `packages/`.
- Treat Outpipe support as an optional integration.
- Do not edit generated route or protocol files by hand.

## Pull requests

- Use a focused branch and describe the user-visible or operational impact.
- Include tests for protocol, authentication, routing, limits, or persistence changes.
- Update documentation and `TODO.md` when behavior or architecture changes.
- Run formatting and relevant checks before requesting review.
- Keep files within the project size limits.
- Explain migrations, configuration changes, and rollout requirements.

## Commit guidance

Use concise imperative commit messages that describe one change, such as:

```text
Add tunnel session expiry handling
```
