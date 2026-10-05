# Telemetry, capture, machine tunnels, and uptime

## OTLP ingestion

An organization administrator can create a one-hour ingestion credential with `POST /api/v1/organizations/{organizationID}/observability/ingestion-keys`. Configure collectors to send `Authorization: Bearer <raw>` and renew the credential before expiry. Only active organization members with scoped credentials can ingest; the authenticated credential determines the organization. Conflicting organization headers or query parameters are rejected.

OTLP HTTP endpoints are `/api/v1/ingest/otlp/v1/traces`, `/api/v1/ingest/otlp/v1/logs`, and `/api/v1/ingest/otlp/v1/metrics`. Use `application/json` or `application/x-protobuf`; gzip is supported. Batches are limited to 4 MiB after decompression and 10,000 records/data points. All five OTLP metric shapes are retained with their resource and scope. Read metrics through `GET /api/v1/organizations/{organizationID}/observability/metrics`.

## Analytics destinations

Apply `integrations/analytics/clickhouse.sql` or deploy `integrations/analytics/outpipe_telemetry.datasource` before enabling export. Configuration is deployment-owned, never selected by ingestion callers.

| Destination | Environment variables |
| --- | --- |
| Tinybird | `OUTPIPE_TINYBIRD_URL`, `OUTPIPE_TINYBIRD_TOKEN`, `OUTPIPE_TINYBIRD_DATASOURCE` (default `outpipe_telemetry`) |
| ClickHouse | `OUTPIPE_CLICKHOUSE_URL`, `OUTPIPE_CLICKHOUSE_USER`, `OUTPIPE_CLICKHOUSE_PASSWORD`, `OUTPIPE_CLICKHOUSE_DATABASE` (default `default`), `OUTPIPE_CLICKHOUSE_TABLE` (default `outpipe_telemetry`) |

Use HTTPS and credentials restricted to insertion into the configured destination. Events contain `id`, `organization_id`, `signal`, `payload`, and `timestamp`; tenant filtering is required in analytics queries. Tinybird receives NDJSON through its [Events API](https://www.tinybird.co/docs/forward/ingest-data/events-api). ClickHouse receives `JSONEachRow` inserts through its [HTTP interface](https://clickhouse.com/docs/interfaces/http). Both can be enabled together.

Storage or exporter failures return retryable HTTP 503. Local storage occurs before export; retries can duplicate records, and successful delivery to one configured destination is not rolled back if the other fails. This is synchronous delivery without a durable export queue. Test credentials against the deployed destination before production use.

## Request capture and replay

Both analytics schemas expire telemetry after 30 days. They use `ReplacingMergeTree` with stable event IDs and timestamps so capture retries can be deduplicated; use `FINAL` or group by event ID when reading before background merges finish. Tinybird engine and retention settings follow its [datasource reference](https://www.tinybird.co/docs/forward/dev-reference/datafiles/datasource-files). Existing MergeTree tables need a separate replacement table and backfill before switching the configured destination; `CREATE TABLE IF NOT EXISTS` does not change an existing engine.

Capture uploads run through a bounded in-memory queue with four upload workers and three attempts per capture. Retries preserve the capture ID and do not insert duplicate local rows. Export failures return retryable 503. Policy refreshes run in the background and cache their result for one second; expired or failed capture lookups disable capture until refreshed. This is best-effort capture: full queues, exhausted retries, and process shutdown can lose captures. It is not a durable outbox. Machine authorization policies are cached for up to one second, so revocation can take that long to reach active streams; expired JWTs are rejected immediately.

The tunnel details page enables/disables capture through `PATCH /api/v1/tunnels/{tunnelID}/capture` with `{"enabled":true}`. Existing connected HTTP tunnels pick up the setting after the next background policy refresh. The relay authenticates to the internal capture API, which verifies tunnel ownership and the current capture setting again. Capture is disabled by default. Headers and JSON payloads have credential fields redacted and are bounded to 64 KiB each; non-JSON payloads are omitted. Query strings are omitted from captured paths.

Replay permits HTTP(S) public destinations only. The client disables ambient proxies and redirects, validates DNS at connection time, and dials the validated public IP. Loopback, private networks, link-local metadata endpoints, multicast, URL credentials, and non-HTTP schemes are rejected.

## Machine tunnels

Create an organization-wide machine token with the exact `tunnels:write` scope. Newly issued machine tokens expire after 24 hours; old tokens without an expiry must be replaced. Project/environment-bound secret tokens cannot create tunnels.

Set `OUTPIPE_MACHINE_TOKEN`, `OUTPIPE_API_URL`, and `OUTPIPE_RELAY_URL`, then run `outpipe machine-tunnel NAME --port 3000 --protocol http` (also supports TCP/UDP). The command creates a machine-owned tunnel through the public API and connects with a 15-minute relay credential scoped to that tunnel and machine. The command stops when the relay credential expires; automatic credential renewal is not implemented. The target is the local machine. Revocation and ownership are checked again during stream traffic.

## Uptime assertions

The API process runs scheduled checks automatically, polling for due monitors every second with up to 16 concurrent probes per replica. Atomic database claims prevent duplicate scheduled work across replicas; abandoned claims expire after 90 seconds. Paused monitors are skipped. Successful probe persistence schedules the next check after the configured interval. Status-page publication is persisted and enforced for public reads and subscriptions; unavailable pages display an unavailable state.

Anonymous shares cannot attach organization, project, environment, or creator IDs. Organization-bound shares use the authenticated `POST /api/v1/organizations/{organizationID}/secrets/shares` endpoint, which checks membership and derives the creator. Machine secret tokens accept `projectId` and `environmentId` (IDs or slugs); their scopes are validated within the organization. CLI secret retrieval requires the exact `secrets:read` permission.

Configure a nonempty `OUTPIPE_AUTH_ENCRYPTION_KEY` before starting the API. Run `outpipe secrets run` with a short-lived secrets machine token in `OUTPIPE_TOKEN`; saved login credentials and tunnel API keys do not grant secret access. Organization share lists return metadata without ciphertext or verification material. Active secret keys and version numbers are unique, and writes allocate versions transactionally. Migration 21 archives duplicate active entries in trash and normalizes duplicate version numbers while preserving encrypted history. Deleting a project removes its secret data and revokes its scoped credentials.

The root and per-command `.env.example` files document all Go configuration variables, including optional analytics exports and tunnel limits. Copy `apps/status/.env.example` for the public status application's `VITE_API_URL`. Analytics exports remain disabled while their URLs are empty.

The dashboard supports HTTP, HTTPS, TCP, and ICMP monitors. HTTP monitors can assert an exact status code and body regular expression; all protocols can set a maximum latency. API fields are `expectedStatusCode` (0 accepts 200–399), `bodyRegex`, and `maxLatencyMs` (0 disables the assertion), with `intervalSeconds` and `timeoutSeconds`. Body assertions read at most 1 MiB. Monitor targets must resolve to public addresses, and checks pin the validated IP.

ICMP probes require raw-socket privileges on the API process executing the probe (for Linux, grant `CAP_NET_RAW`). A process lacking that capability records a failed check with the socket error. Run it on a host/network that permits outbound ICMP. Monitor reads, probes, deletion, check history, and incident updates enforce organization ownership.

The ingestion encoding follows the [OTLP specification](https://opentelemetry.io/docs/specs/otlp/).
