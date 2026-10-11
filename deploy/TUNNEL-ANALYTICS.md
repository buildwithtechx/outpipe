# Tunnel Analytics in Tinybird & ClickHouse

The Outpipe analytics runtime uses Tinybird / ClickHouse for HTTP request history,
captured requests, TCP/UDP events, tunnel metrics, and active-tunnel snapshots.
Timescale is no longer a runtime dependency. SQLite / PostgreSQL remains the
source of truth for organizations, permissions, tunnel configuration, Uptime,
and Secrets; Redis remains the live tunnel state and telemetry delivery queue.

## Rollout order

1. Validate and deploy the committed Tinybird project to an isolated development
   workspace/branch first. Then test HTTP, TCP, UDP, request capture, filters,
   dashboard metrics, and recovery after a simulated Tinybird outage.
2. Deploy the same definitions to the intended production workspace before
   deploying any new tunnel, cron, or dashboard runtime. This is an additive
   change; it does not migrate relational schemas.
3. Configure `TINYBIRD_API_HOST` and scoped credentials in each runtime:
   - Tunnel: `TINYBIRD_TUNNEL_INGEST_TOKEN` or fallback `TINYBIRD_INGEST_TOKEN`.
   - Cron: the same append-token choice and `TINYBIRD_QUERY_TOKEN` (alerts).
   - Dashboard: `TINYBIRD_QUERY_TOKEN`.
   - Observability ingest: its existing `TINYBIRD_INGEST_TOKEN`.
4. Append scopes must include `tunnel_events`, `tunnel_request_captures`,
   `tunnel_protocol_events`, and `tunnel_active_snapshots`. Query scopes must
   include the 13 `tunnel_*` endpoints. Never give runtime services a
   workspace-admin or deployment token.
5. Promote the new runtimes together after verification.

## Existing history

This change switches readers and writers; it does **not** transfer existing
Timescale history. Old history will not appear in the new dashboard unless it
is backfilled. Keep existing backups until you have accepted fresh analytics history.

When exporting history:

- Export only unexpired rows, mapping old `request_captures` to `tunnel_request_captures`
  and `protocol_events` to `tunnel_protocol_events`.
- Serialize UTC timestamps with millisecond precision and headers as JSON strings.
- Assign stable event identities before importing so retries are deduplicated.

## Safety and retention

All organization endpoints require an authenticated organization context; tunnel
detail and capture lookups additionally verify tunnel ownership. Tinybird read
tokens stay server-side. HTTP/protocol events retain stable identities for
retry deduplication; query pipes apply per-row retention even before physical
TTL cleanup. Storage keeps the existing maximum of 90 days.

Captured payloads remain sensitive data: protect the Tinybird workspace and tokens
appropriately. Redis also temporarily holds sensitive capture payloads while
delivering or retrying them, so restrict network access and protect credentials.

Unsent capture storage is intentionally bounded: the queue and dead-letter
stream each hold at most 64 messages of at most 4 MiB each, with a three-day cutoff.
A capture larger than 3 MiB is rejected individually rather than poisoning neighboring
captures; ordinary request metadata continues to be recorded.

## Endpoints Summary

### Organization Endpoints

1. `tunnel_overview_stats.pipe`: Organization-wide HTTP & protocol totals compared with preceding window.
2. `tunnel_overview_chart.pipe`: Time-series aggregate of HTTP requests and protocol traffic.
3. `tunnel_http_stats.pipe`: HTTP request status counts, error rates, and latency p50/p90/p99.
4. `tunnel_http_chart.pipe`: Time-bucketed HTTP status codes and response volumes.
5. `tunnel_protocol_stats.pipe`: Protocol-level metrics (TCP, UDP) for raw streams.
6. `tunnel_protocol_chart.pipe`: Time-bucketed protocol events and bandwidth.
7. `tunnel_protocol_recent.pipe`: Recent protocol connection activity log.
8. `tunnel_requests.pipe`: Paginated and filterable HTTP requests log with substring search.
9. `tunnel_capture.pipe`: Payloads and header inspection for individual captured requests.

### Administrator Endpoints

10. `tunnel_admin_usage.pipe`: Platform-wide tenant bandwidth and request metrics.
11. `tunnel_admin_http_chart.pipe`: Global HTTP traffic volume across all active tunnels.
12. `tunnel_admin_active_series.pipe`: Concurrent tunnel connection snapshots.
13. `tunnel_admin_active_chart.pipe`: Global active tunnel count time series.
