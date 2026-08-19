# Outpipe TODO

Remaining work: production hardening and tests, dashboard functionality, desktop integration, optional integrations, and release operations.

## Standalone product principles

- [ ] Provide Outpipe integration as an optional adapter, plugin, or external API client.
- [ ] Document the hosted standalone product and optional Outpipe integration.

## Product surfaces

- [ ] Build a standalone tunnel dashboard with its own authentication and organization model.
- [ ] Keep the Tauri desktop application usable with standalone tunnel servers.

## Route structure

Dashboard routes should remain browser-only and organization-scoped:

```text
/
├── pricing
├── docs/*
├── login
├── signup
├── cli/login
├── admin/
│   ├── index
│   ├── users
│   ├── users/$userId
│   ├── organizations
│   ├── organizations/$organizationID
│   ├── tunnels
│   ├── subscriptions
│   ├── usage
│   ├── charts
│   ├── audit-logs
│   └── actions
└── $orgSlug/
    ├── index
    ├── tunnels/
    │   ├── index
    │   └── $tunnelId
    ├── agents
    ├── domains
    ├── requests
    ├── usage
    ├── billing
    ├── members
    ├── api-keys
    └── settings/
        ├── index
        ├── profile
        └── organization
```

The dashboard application should keep page routing separate from reusable
application code:

```text
apps/web/src/
├── routes/
│   ├── __root.tsx
│   ├── index.tsx
│   ├── login.tsx
│   ├── signup.tsx
│   └── $orgSlug/
├── components/
│   ├── ui/
│   ├── layout/
│   ├── navigation/
│   ├── feedback/
│   └── data-display/
├── features/
│   ├── auth/
│   │   ├── components/
│   │   │   ├── auth-page-shell.tsx
│   │   │   ├── oauth-provider-button.tsx
│   │   │   └── auth-notice.tsx
│   │   ├── hooks/
│   │   │   ├── use-auth-session.ts
│   │   │   └── use-oauth-sign-in.ts
│   │   ├── services/
│   │   │   └── auth-service.ts
│   │   ├── login-page.tsx
│   │   ├── signup-page.tsx
│   │   └── index.ts
│   ├── admin/
│   │   ├── users/
│   │   ├── organizations/
│   │   ├── tunnels/
│   │   ├── subscriptions/
│   │   ├── usage/
│   │   ├── charts/
│   │   └── actions/
│   ├── organizations/
│   ├── tunnels/
│   ├── agents/
│   ├── domains/
│   ├── usage/
│   ├── billing/
│   ├── audit-logs/
│   └── api-keys/
├── interfaces/
│   ├── api.ts
│   ├── auth.ts
│   ├── organization.ts
│   ├── tunnel.ts
│   └── billing.ts
├── hooks/
├── stores/
│   ├── auth-store.ts
│   ├── organization-store.ts
│   ├── tunnel-store.ts
│   └── ui-store.ts
├── lib/
│   ├── api-client.ts
│   ├── route-guards.ts
│   └── utils.ts
├── integrations/
│   └── telemetry/
└── env.ts
```

- [ ] Keep route files focused on loaders, route metadata, and page composition.
- [ ] Protect `/admin/*` with a separate platform-admin authorization guard.
- [ ] Keep admin features and components isolated from organization-member features.
- [ ] Keep domain behavior inside `features/`, not inside route files.
- [ ] Keep shared visual primitives inside `components/ui/`.
- [ ] Keep API contracts and frontend DTOs inside `interfaces/`.
- [ ] Keep cross-feature clients, guards, query setup, and utilities inside `lib/`.
- [ ] Keep global client state in focused Zustand stores under `stores/`.
- [ ] Keep feature-specific hooks beside their feature unless shared by multiple domains.

### Dashboard authentication

The login and signup routes should remain thin compositions that import their
feature page from `features/auth/`. OAuth navigation, session loading, and
query invalidation belong in the feature service and hooks; the shared
Axios instance stays in `lib/api-client.ts`, contracts stay in
`interfaces/auth.ts`, and the authenticated browser state stays in
`stores/auth-store.ts`.

```text
apps/web/src/features/auth/
├── components/
│   ├── auth-page-shell.tsx        # Shared login/signup visual frame
│   ├── oauth-provider-button.tsx  # Google or GitHub action
│   └── auth-notice.tsx            # Provider and callback errors
├── hooks/
│   ├── use-auth-session.ts        # Session query and Zustand synchronization
│   └── use-oauth-sign-in.ts       # Browser redirect to the selected provider
├── services/
│   └── auth-service.ts            # OAuth URL and session requests
├── login-page.tsx
├── signup-page.tsx
└── index.ts
```

- [ ] Add logout and session-expiry UI states when the authenticated dashboard shell is implemented.
- [ ] Add route guards only after session loading distinguishes unauthenticated from pending state.
- [ ] Add focused tests for provider selection, redirect construction, session synchronization, and callback errors.

The browser flow requires these API changes before the dashboard UI is wired:

- [ ] Set API cookie `Secure`, `HttpOnly`, `SameSite=Lax`, and an intentional shared-domain cookie policy for hosted dashboard/API subdomains.

- [ ] Keep `integrations/outpipe/` as an optional external adapter.

## Future SDKs

- [ ] Create an official Laravel/PHP SDK over the public HTTP and relay APIs.
- [ ] Create an official Rust SDK for native services and desktop tooling.
- [ ] Create an official Go SDK for Go services and tunnel-aware workers.
- [ ] Create an official Angular adapter over the shared client contract.
- [ ] Keep every SDK aligned with the versioned protocol and authentication model.
- [ ] Publish language-specific SDKs only after compatibility, security, and conformance tests pass.

```text
packages/
├── php/
│   ├── src/
│   │   ├── Client/
│   │   ├── Contracts/
│   │   ├── Exceptions/
│   │   ├── Resources/
│   │   └── Laravel/
│   │       ├── Console/
│   │       ├── Facades/
│   │       ├── Http/
│   │       ├── Services/
│   │       └── OutpipeServiceProvider.php
│   ├── config/
│   ├── tests/
│   ├── composer.json
│   └── README.md
├── rust/
│   ├── src/
│   │   ├── client.rs
│   │   ├── error.rs
│   │   ├── models.rs
│   │   └── protocol.rs
│   ├── tests/
│   ├── Cargo.toml
│   └── README.md
├── go/
│   ├── client/
│   ├── protocol/
│   ├── internal/
│   ├── tests/
│   ├── go.mod
│   └── README.md
└── angular/
    ├── src/
    │   ├── guards/
    │   ├── interceptors/
    │   ├── models/
    │   ├── services/
    │   └── tokens/
    ├── tests/
    ├── package.json
    └── README.md
```

## Dashboard

- [ ] Create the tunnel dashboard workspace.
- [ ] Implement tunnel creation and target configuration.
- [ ] Display connection status, public URL, client version, and last heartbeat.
- [ ] Add start, stop, rotate credentials, and revoke actions.
- [ ] Display active connections and bandwidth usage.
- [ ] Add tunnel expiration and access policy controls.
- [ ] Add audit history for tunnel and credential actions.
- [ ] Add accessible loading, empty, error, and disconnected states.

## Tauri desktop app

- [ ] Define how the desktop app starts and supervises the Go CLI tunnel client.
- [ ] Keep tunnel data-plane logic in Go rather than duplicating it in Rust or TypeScript.
- [ ] Store credentials using the native operating-system secret store.
- [ ] Add tray controls for tunnel status and quick start/stop.
- [ ] Add native notifications for disconnects, expiry, and authentication failures.
- [ ] Package the CLI and desktop application for macOS, Windows, and Linux.

## Codedock integration

- [ ] Define an optional Codedock adapter that consumes the standalone tunnel API.
- [ ] Allow Codedock to create scoped tunnel credentials through the public integration API.
- [ ] Add optional tunnel metadata links to Codedock projects and services.
- [ ] Add tunnel status synchronization through polling or signed webhooks.
- [ ] Add optional tunnel creation after local-directory deployment.
- [ ] Add optional MCP tools for listing, creating, inspecting, and revoking tunnels.
- [ ] Keep Codedock usable when the tunnel service is unavailable.
- [ ] Ensure uninstalling or disabling Codedock integration never deletes standalone tunnel data.
- [ ] Add integration tests proving a standalone tunnel server works without Codedock.
