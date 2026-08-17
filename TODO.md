# Outpipe TODO

The Go backend, CLI, protocol package, and current TypeScript framework adapters are implemented. Remaining work is dashboard functionality, desktop integration, optional integrations, and release operations.

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
- [x] Add `outpiped bootstrap-admin --email ...` to provision the first platform administrator explicitly.
- [x] Add platform-admin roles without promoting users automatically during signup.
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

- [x] Implement the shared OAuth-only authentication feature for Google and GitHub.
- [x] Keep `/login` and `/signup` as distinct intent pages with shared OAuth controls.
- [x] Navigate the browser to the API OAuth start endpoint; do not exchange OAuth tokens in browser code.
- [x] Synchronize a successful authenticated session into `auth-store` through a feature-scoped TanStack Query hook.
- [ ] Add logout and session-expiry UI states when the authenticated dashboard shell is implemented.
- [x] Add provider-unavailable and callback-failure states.
- [x] Redirect an authenticated visitor away from `/login` and `/signup` to organization selection, their last organization, or their only organization.
- [ ] Add route guards only after session loading distinguishes unauthenticated from pending state.
- [ ] Add focused tests for provider selection, redirect construction, session synchronization, and callback errors.

The browser flow requires these API changes before the dashboard UI is wired:

- [x] Let OAuth start accept a validated dashboard-relative return path and store it in server-side OAuth state.
- [x] After OAuth callback creates the HTTP-only API session cookie, redirect to the configured dashboard origin and validated return path instead of returning JSON from the API domain.
- [x] Reject absolute or cross-origin return paths to prevent open redirects.
- [x] Provide the current authenticated user alongside session data, or add a protected current-account endpoint so `AuthUser` can populate `auth-store`.
- [x] Provide a protected organization list endpoint for post-login organization selection.
- [ ] Set API cookie `Secure`, `HttpOnly`, `SameSite=Lax`, and an intentional shared-domain cookie policy for hosted dashboard/API subdomains.

The control-plane API should remain versioned under `/api/v1`:

```text
/api/v1/
├── auth/
│   ├── oauth/:provider
│   ├── oauth/:provider/callback
│   ├── device/start
│   ├── device/poll
│   ├── device/complete
│   ├── session
│   └── logout
├── account
├── organizations
│   ├── :organizationID
│   ├── :organizationID/members
│   ├── :organizationID/members/:memberID
│   └── :organizationID/transfer
├── organizations/:organizationID/
│   ├── tunnels
│   ├── agents
│   ├── domains
│   ├── usage/events
│   ├── usage/snapshot
│   ├── usage/requests
│   ├── billing/{status,checkout,portal,cancel,resume,invoices}
│   ├── api-keys
│   ├── audit-logs
│   └── settings
├── tunnels/:tunnelID/
│   ├── status
│   ├── revoke
│   ├── connections
│   └── requests
├── domains/:domainID/verify
├── agents/:agentID/
│   ├── heartbeat
│   └── revoke
├── admin/
│   ├── users
│   ├── users/:userID
│   ├── organizations
│   ├── organizations/:organizationID
│   ├── tunnels
│   ├── subscriptions
│   ├── usage
│   ├── charts
│   └── actions
└── webhooks/billing/:provider
```

- [x] Keep liveness, readiness, and metrics outside the versioned API at `/healthz`, `/readyz`, and `/metrics`.
- [x] Keep relay WebSocket transport separate at `wss://tunnel.outpipe.dev/v1/connect`.
- [ ] Add organization detail, member listing/removal, profile, API-key, audit-log, request-log, and invoice routes.
- [x] Add platform-admin authorization and admin users, organizations, tunnels, subscriptions, usage, audit, and action routes.
- [ ] Move agent heartbeats away from browser-session authentication to agent or relay authentication.
- [ ] Keep internal health, usage ingestion, relay handoff, and agent authentication routes private.
- [ ] Route wildcard public tunnel traffic through `*.tunnel.outpipe.dev`, not through control-plane handlers.

- [ ] Keep `integrations/outpipe/` as an optional external adapter.

## Transactional email

- [x] Use Zepto Mail as the only transactional email provider.
- [x] Keep HTML email templates in the root `templates/` directory.
- [x] Add welcome, account-update, billing-update, organization-invite, payment-failed, and subscription-reset templates.
- [x] Escape user-controlled template values through Go's HTML template renderer.
- [x] Send the welcome email after the first successful OAuth account creation.
- [x] Send account-update email after account deletion.
- [x] Send billing-update email after subscription state changes.
- [x] Add organization invitation persistence, expiration, acceptance, and delivery workflow.
- [x] Connect payment-failed events to retry-count and billing-page email data.
- [x] Connect subscription downgrade/reset events to the subscription-reset email.

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
