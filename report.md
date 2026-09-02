# Outpipe Codebase Audit Report

**Audit date:** 2026-09-02  
**Scope:** Repository-wide architecture, security, API behavior, persistence, SDKs, web application, CI/CD, containers, and documentation  
**Status:** Research complete; remediation not started

## Executive summary

The repository currently passes its available Go, web, SDK, formatting, and typechecking checks. The strongest foundations are the versioned protocol schema, OAuth state and PKCE flow, non-root container runtimes, relay tests, and SDK conformance coverage.

The most important remaining risks are production-hardening issues rather than compilation failures:

1. Webhook destinations can be used as an SSRF path.
2. Webhook delivery is synchronous and has no durable retry worker.
3. Usage-event reads are unbounded.
4. Production configuration can start with development security defaults.
5. Rate limiting is process-local and covers too few expensive endpoints.
6. Billing webhook idempotency has a concurrency gap.
7. API, relay, SDK, documentation, CI, and deployment contracts still need tighter alignment.

No files were changed during the audit other than this report.

## Verification performed

- [x] `go vet ./...`
- [x] `go test ./...`
- [x] Web Biome formatting check
- [x] Web typecheck and generated route checks
- [x] SDK typechecks
- [x] `git diff --check`
- [x] Tracked source-size scan
- [x] CI, release, Docker, Compose, protocol, auth, billing, webhook, usage, and documentation review
- [x] Primary-source security and platform guidance review
- [ ] npm dependency audit — registry DNS was unavailable
- [ ] Composer dependency audit — Packagist/network was unavailable
- [ ] Rust dependency audit — `cargo-audit` is not installed
- [ ] Go vulnerability scan — `govulncheck` is not installed locally

The failed dependency checks are unresolved verification gaps, not evidence that dependencies are safe.

## Severity guide

- **P0:** Exploitable or availability-critical issue requiring immediate attention.
- **P1:** Production security or correctness issue that should be fixed before launch.
- **P2:** Important reliability, maintainability, or release-process improvement.
- **P3:** Cleanup, consistency, or developer-experience improvement.

## P0/P1 remediation checklist

### 1. Secure webhook egress — P0/P1

**Finding:** Webhook creation accepts any HTTP/HTTPS URL and delivery uses the default Go HTTP client. The client follows redirects, and the initial URL is not checked against private, loopback, link-local, metadata, or other unsafe address ranges.

Relevant code:

- [x] Review URL acceptance in [`internal/services/webhooks.go`](internal/services/webhooks.go).
- [x] Reuse or extract one safe outbound-target validator for webhook delivery and domain verification.
- [x] Block loopback, private, link-local, multicast, carrier-grade NAT, metadata-service, and unspecified addresses.
- [x] Resolve hostnames and validate every resolved address.
- [x] Prevent DNS time-of-check/time-of-use bypasses by validating the address used for the actual connection.
- [x] Disable redirects by default, or validate every redirect independently.
- [ ] Limit response-body reads and enforce request and total delivery timeouts.
- [ ] Decide explicitly whether customer-private endpoints are supported; if yes, implement a separate controlled egress policy.
- [x] Add regression tests for IPv4 and IPv6 loopback, RFC1918 ranges, link-local/metadata addresses, malformed URLs, unsafe userinfo, and redirect behavior.
- [ ] Add integration tests for DNS failures, mixed DNS answers, and oversized responses.

Reference: [OWASP SSRF Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html)

### 2. Durable webhook delivery — P1

**Finding:** Tunnel lifecycle dispatch performs network delivery inline, records only one attempt, and does not persist `LastDeliveredAt` after a successful request.

- [x] Introduce a durable webhook outbox record for every emitted event.
- [x] Make tunnel lifecycle requests enqueue work instead of waiting for customer endpoints.
- [x] Add a worker that claims pending deliveries safely across replicas.
- [x] Add exponential backoff and a maximum attempt count.
- [x] Add a permanently-failed state after the maximum attempts.
- [x] Persist delivery status, attempts, error, and timestamps.
- [x] Persist `LastDeliveredAt` through the repository.
- [ ] Add idempotency keys and document receiver behavior.
- [ ] Add metrics and structured event IDs without logging secrets or payloads.
- [ ] Add tests for timeout, retry, duplicate worker claims, restart recovery, and dead-letter behavior.

Reference: [OWASP API4:2023 Unrestricted Resource Consumption](https://owasp.org/API-Security/editions/2023/en/0xa4-unrestricted-resource-consumption/)

### 3. Bound usage and list endpoints — P1

**Finding:** The usage-events endpoint accepts a time window but no page size, and the repository loads all matching events into memory.

- [x] Add a bounded `limit` to usage-event queries.
- [x] Add cursor pagination with stable `(occurred_at, id)` ordering.
- [x] Return a continuation cursor when another page exists.
- [ ] Review every list endpoint for an explicit maximum page size.
- [x] Add an index matching organization and time-based query patterns.
- [x] Add tests for stable cursor behavior and bounded pages.
- [ ] Add tests for maximum handler limits, empty pages, and large datasets.
- [ ] Add response-size and query-latency monitoring.

Relevant code: [`internal/handlers/usage.go`](internal/handlers/usage.go), [`internal/repositories/usage.go`](internal/repositories/usage.go)

### 4. Enforce production security configuration — P1

**Finding:** Defaults permit development mode, HTTP URLs, non-secure cookies, optional auth encryption, and development-style credentials. Validation does not currently reject unsafe production combinations.

- [x] Add explicit production configuration validation.
- [x] Require HTTPS public URLs and CORS origin in production.
- [x] Require secure cookies in production.
- [x] Require `AUTH_ENCRYPTION_KEY` with a valid AES key length where encrypted credentials are used.
- [x] Require a strong internal API secret and paired OAuth credentials where applicable.
- [ ] Reject known development secrets and default database credentials.
- [x] Validate production allowed-origin relationships against localhost usage.
- [x] Fail startup with actionable errors instead of allowing later request-time failures.
- [ ] Document which variables are mandatory for API, relay, worker, and web deployments.
- [x] Add production configuration tests that fail when unsafe values are supplied.

Relevant code: [`internal/config/config.go`](internal/config/config.go), [`internal/config/load.go`](internal/config/load.go)

### 5. Distributed and broader rate limiting — P1

**Finding:** Rate limits are stored in a process-local map and are applied mainly to support and device-login routes. They reset on restart and do not coordinate across replicas.

- [x] Replace process-local buckets with Redis-backed limiting for deployed environments.
- [x] Keep a bounded local fallback only for failure containment, with clear behavior.
- [x] Rate-limit authenticated API traffic and high-cost writes including tunnel creation, organization creation, API-key creation, checkout, and webhook creation.
- [x] Prefer authenticated user identity keys where available.
- [ ] Configure trusted proxy handling before using client IPs.
- [ ] Add endpoint-specific limits based on cost and external side effects.
- [ ] Add tests for concurrent requests, multiple instances, Redis failure, proxy headers, and Redis Lua behavior.

Relevant code: [`internal/http/middleware.go`](internal/http/middleware.go), [`internal/http/router.go`](internal/http/router.go)

### 6. Make billing webhook handling race-safe — P1

**Finding:** Billing processing performs a pre-check followed by an insert. Concurrent duplicate deliveries can race on the unique constraint and return an error instead of being acknowledged idempotently.

- [ ] Change processing to insert-first or use an atomic upsert.
- [ ] Treat a duplicate provider event ID as an idempotent success.
- [ ] Preserve provider retry behavior for genuine processing failures.
- [ ] Record processing failure state and error details safely.
- [ ] Add a concurrent duplicate-delivery test.
- [ ] Add tests for out-of-order events, repeated successful events, and failed transitions.

Relevant code: [`internal/services/billing.go`](internal/services/billing.go), [`internal/repositories/billing.go`](internal/repositories/billing.go)

## P2 architecture and reliability checklist

### 7. Normalize API, relay, and SDK contracts — P2

- [ ] Define the control-plane DTO naming convention, including `publicHostname`.
- [ ] Keep the relay wire field `public_url` only where protocol compatibility requires it.
- [ ] Select one canonical SDK-facing field name.
- [ ] Keep adapter translation explicit and tested.
- [ ] Generate or validate examples from shared contract fixtures.
- [ ] Add compatibility tests for every supported SDK and protocol version.
- [ ] Document versioning and breaking-change policy.

Relevant files include [`protocol/schema/messages.json`](protocol/schema/messages.json), [`packages/sdk/src/interfaces/api.ts`](packages/sdk/src/interfaces/api.ts), and the adapter packages.

### 8. Improve migrations and schema evolution — P2

The project has migration version tracking, but schema changes still rely heavily on GORM `AutoMigrate`.

- [ ] Decide whether GORM AutoMigrate is acceptable for production schema evolution.
- [ ] If not, adopt an explicit migration tool with reviewable forward migrations.
- [ ] Add migration locking and startup coordination for multiple replicas.
- [ ] Add staging migration rehearsal and backup expectations.
- [ ] Document rollback and data-migration procedures.
- [ ] Add CI validation for a clean database and an upgraded database.

Relevant code: [`internal/infra/postgres/migration.go`](internal/infra/postgres/migration.go)

### 9. Add meaningful observability — P2

- [ ] Replace the static metrics response with real Prometheus metrics.
- [ ] Measure HTTP status, latency, and request volume.
- [ ] Measure active relay connections and tunnel lifecycle transitions.
- [ ] Measure webhook queue depth, attempts, failures, retries, and age.
- [ ] Measure billing event processing and duplicate counts.
- [ ] Ensure logs never contain tokens, cookies, credentials, webhook secrets, or request bodies.
- [ ] Decide whether metrics are public, internal-only, or protected by network policy.

Relevant code: [`internal/http/router.go`](internal/http/router.go)

### 10. Separate development and production containers — P2

- [ ] Keep the current Compose setup explicitly development-only.
- [ ] Create production-specific Compose/deployment configuration.
- [ ] Remove source bind mounts and dev-server commands from production.
- [ ] Use secrets files or a deployment secret manager.
- [ ] Pin production image digests instead of mutable tags.
- [ ] Add container image scanning and startup smoke tests.

Relevant file: [`docker-compose.yml`](docker-compose.yml)

Reference: [Docker Compose production guidance](https://docs.docker.com/compose/how-tos/production/)

### 11. Strengthen CI and release supply-chain controls — P2

- [ ] Pin GitHub Actions to full commit SHAs.
- [ ] Add Rust dependency auditing.
- [ ] Add Composer/PHP dependency auditing.
- [ ] Add dependency review on pull requests.
- [ ] Generate SBOMs for release artifacts.
- [ ] Verify npm provenance and trusted-publishing configuration.
- [ ] Verify package contents before publishing each SDK.
- [ ] Ensure every published SDK has behavior/conformance tests rather than only compilation checks.
- [ ] Remove or replace `--passWithNoTests` where a package is expected to contain tests.
- [ ] Add release smoke tests for npm, Go, Rust, and PHP artifacts.

Relevant workflows:

- [`.github/workflows/ci.yml`](.github/workflows/ci.yml)
- [`.github/workflows/dependency-audit.yml`](.github/workflows/dependency-audit.yml)
- [`.github/workflows/publish-packages.yml`](.github/workflows/publish-packages.yml)
- [`.github/workflows/publish-go.yml`](.github/workflows/publish-go.yml)
- [`.github/workflows/publish-rust.yml`](.github/workflows/publish-rust.yml)

References: [GitHub secure use reference](https://docs.github.com/en/actions/reference/security/secure-use?learn=getting_started&learnProduct=actions), [GitHub dependency review](https://docs.github.com/en/code-security/concepts/supply-chain-security/dependency-review)

## P3 consistency and maintainability checklist

### 12. Clean stale documentation and branding — P3

- [ ] Rename or clearly label [`docs/codedock-integration.md`](docs/codedock-integration.md) as an optional legacy integration.
- [ ] Remove obsolete Codedock commands, URLs, package names, and metadata from active documentation.
- [ ] Review README language that still calls Outpipe “Tunnel” where product terminology should now be “Outpipe.”
- [ ] Verify all public tunnel examples use `outpipe.app`.
- [ ] Verify control-plane examples use `outpipe.dev`.
- [ ] Check SDK examples, MDX content, fixtures, and integration tests for inconsistent field names.

### 13. Split oversized source files — P3

The project rule is to keep source files under 350 lines. The following tracked files are at or above 300 lines and should be reviewed:

- [ ] Split [`internal/services/billing.go`](internal/services/billing.go), currently above the 350-line limit.
- [ ] Review [`internal/http/middleware.go`](internal/http/middleware.go).
- [ ] Review [`internal/relay/tunnel.go`](internal/relay/tunnel.go).
- [ ] Review [`internal/http/database.go`](internal/http/database.go).
- [ ] Review [`pkg/client/relay.go`](pkg/client/relay.go).
- [ ] Review [`apps/web/src/components/layout/marketing-header.tsx`](apps/web/src/components/layout/marketing-header.tsx).
- [ ] Review [`internal/relay/handler.go`](internal/relay/handler.go).
- [ ] Review [`packages/sdk/src/services/relay-connection.ts`](packages/sdk/src/services/relay-connection.ts).
- [ ] Leave generated route files generated; do not hand-split [`apps/web/src/routeTree.gen.ts`](apps/web/src/routeTree.gen.ts).

## Suggested implementation sequence

- [ ] Phase 1: Safe outbound networking and webhook SSRF tests.
- [ ] Phase 2: Webhook outbox, worker, retries, idempotency, and persistence.
- [ ] Phase 3: Pagination, query limits, indexes, and broader rate limiting.
- [ ] Phase 4: Production configuration guardrails and deployment separation.
- [ ] Phase 5: Billing concurrency fixes and webhook replay tests.
- [ ] Phase 6: API/relay/SDK contract normalization and documentation cleanup.
- [ ] Phase 7: CI supply-chain controls, dependency audits, SBOMs, and release smoke tests.
- [ ] Phase 8: Observability, migration hardening, and source-file decomposition.

## Audit conclusion

The project is not failing because of basic code quality. Its immediate gap is that several production boundaries are still too trusting: outbound network destinations, unbounded reads, environment defaults, distributed enforcement, and release verification.

The first implementation pass should focus on secure webhook egress and durable webhook delivery. Those changes reduce the most serious security and reliability risks while creating infrastructure that can later support retries, metrics, replay handling, and operational visibility.
