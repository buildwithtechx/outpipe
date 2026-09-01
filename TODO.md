# Outpipe TODO

Remaining work: optional SDKs, desktop integration, optional integrations, and release operations.

## Integration delivery workflow

- [x] Run the API, relay, and web application together while building dashboard features.
- [x] Connect each dashboard feature to its real API endpoint before marking it complete.
- [x] Add loading, empty, authorization, and failure states from observed API responses.
- [x] Fix Go API contracts, validation, authorization, persistence, and relay behavior when end-to-end testing exposes a defect.
- [x] Add or update focused Go and web tests for every defect fixed during integration.
- [x] Mark a feature complete only after its browser flow works against local PostgreSQL, Redis, API, and relay services.

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

- [x] Keep route files focused on loaders, route metadata, and page composition.
- [x] Protect `/admin/*` with a separate platform-admin authorization guard.
- [x] Keep admin features and components isolated from organization-member features.
- [x] Keep domain behavior inside `features/`, not inside route files.
- [x] Keep shared visual primitives inside `components/ui/`.
- [x] Keep API contracts and frontend DTOs inside `interfaces/`.
- [x] Keep cross-feature clients, guards, query setup, and utilities inside `lib/`.
- [x] Keep global client state in focused Zustand stores under `stores/`.
- [x] Keep feature-specific hooks beside their feature unless shared by multiple domains.

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

- [x] Add logout and session-expiry UI states when the authenticated dashboard shell is implemented.
- [x] Add route guards only after session loading distinguishes unauthenticated from pending state.
- [x] Add focused tests for provider selection, redirect construction, session synchronization, and callback errors.

The browser flow requires these API changes before the dashboard UI is wired:

- [x] Set API cookie `Secure`, `HttpOnly`, `SameSite=Lax`, and an intentional shared-domain cookie policy for hosted dashboard/API subdomains.

## Future SDKs

- [x] Create an official Laravel/PHP SDK over the public HTTP and relay APIs (`packages/php`).
- [x] Create an official Rust SDK for native services and desktop tooling (`packages/rust`).
- [x] Create an official Go SDK for Go services and tunnel-aware workers (`packages/go`).
- [x] Create an official Angular adapter over the shared client contract (`packages/angular`).
- [x] Keep every SDK aligned with the versioned protocol and authentication model.
- [x] Publish language-specific SDKs only after compatibility, security, and conformance tests pass.

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
    │   ├── interfaces.ts
    │   ├── providers.ts
    │   ├── services/
    │   ├── tests/
    │   └── tokens.ts
    ├── package.json
    ├── tsconfig.json
    ├── tsup.config.ts
    └── README.md
```

## Dashboard

- [x] Create the tunnel dashboard workspace.
- [x] Implement tunnel creation and target configuration.
- [x] Display connection status, public URL, client version, and last heartbeat.
- [x] Add start, stop, rotate credentials, and revoke actions.
- [x] Display active connections and bandwidth usage.
- [x] Add tunnel expiration and access policy controls.
- [x] Add audit history for tunnel and credential actions.
- [x] Add accessible loading, empty, error, and disconnected states.
