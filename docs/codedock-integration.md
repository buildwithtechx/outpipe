# Codedock integration (platform contract)

This document is the contract the Outpipe platform offers to Codedock
(`~/Dev/TechX/codedock`, a Go daemon + TypeScript dashboard). All integration
code lives in the Codedock repository as an optional adapter; Outpipe exposes
the API surface below and knows nothing about Codedock.

Everything here is optional: Codedock must remain fully functional with the
integration disabled and when the tunnel service is unreachable.

## Authentication

The adapter authenticates with an organization-scoped API key created through
the dashboard or the API:

```text
POST /api/v1/organizations/:organizationID/api-keys
Authorization: Bearer <session-token>
{
  "name": "codedock",
  "scopes": ["organization:admin", "tunnels:read", "tunnels:write"],
  "source": "codedock"
}
```

- `source` is an optional marker (max 40 chars) recorded for audit clarity;
  use `"codedock"`.
- Recommended scopes: `organization:admin` (covers webhook management and the
  role required to create tunnels), plus `tunnels:read`/`tunnels:write`.
  Create one key per Codedock project so tunnels can be revoked independently.
- The raw token (`cdk_....`) is returned once, on creation.
- Authenticate with `Authorization: Bearer <raw token>` on every request.

## Tunnels

Create:

```text
POST /api/v1/organizations/:organizationID/tunnels
{
  "name": "my-service",
  "protocol": "https",
  "targetHost": "127.0.0.1",
  "targetPort": 8080,
  "publicHostname": "my-service.example.com",
  "metadata": {"codedock": {"project": "p1", "service": "s1"}}
}
```

- `metadata` (optional JSON, max 16 KB) is stored verbatim and echoed in every
  tunnel response. Codedock records `codedock.project` and `codedock.service`.
- `publicHostname` is optional; when omitted the platform allocates one.
- The tunnel is created in status `created`. A tunnel is live when its status
  is `active`.

Inspect / list:

```text
GET /api/v1/tunnels/:tunnelID          # full tunnel object incl. metadata
GET /api/v1/organizations/:organizationID/tunnels
```

Status transitions (agent connects/disconnects through the CLI or API):

```text
PATCH /api/v1/tunnels/:tunnelID/status   {"status": "active"}        # connected
PATCH /api/v1/tunnels/:tunnelID/status   {"status": "disconnected"}
```

Revoke (terminal, soft: the row stays readable with status `revoked`):

```text
DELETE /api/v1/tunnels/:tunnelID
```

## Webhooks

Signed webhook subscriptions per organization. Events are delivered
synchronously with a 5-second timeout; every delivery is recorded and can be
inspected.

Create (returns `secret` exactly once — store it):

```text
POST /api/v1/organizations/:organizationID/webhooks
{
  "name": "codedock",
  "url": "https://codedock.example.com/hooks/outpipe",
  "events": ["tunnel.connected", "tunnel.disconnected", "tunnel.revoked"]
}
```

- `events` defaults to all three when omitted.
- Delete: `DELETE /api/v1/organizations/:organizationID/webhooks/:webhookID`
- List: `GET /api/v1/organizations/:organizationID/webhooks`
- Delivery history (last 100):
  `GET /api/v1/organizations/:organizationID/webhooks/:webhookID/deliveries`

### Events

| Event                 | Meaning                                       |
| --------------------- | --------------------------------------------- |
| `tunnel.connected`    | tunnel became active                          |
| `tunnel.disconnected` | tunnel left the active state                  |
| `tunnel.revoked`      | tunnel was revoked (including platform admin) |

### Delivery

```text
POST <subscription url>
Content-Type: application/json
X-Outpipe-Event-ID:      <uuid, equals envelope id>
X-Outpipe-Event-Type:    <event name>
X-Outpipe-Signature:     hex(hmac_sha256(raw_body, subscription_secret))
```

Envelope:

```json
{
  "id": "8f1e...",
  "type": "tunnel.connected",
  "occurredAt": "2026-08-19T12:00:00Z",
  "data": {
    "tunnel": { "id": "...", "status": "active", "metadata": {"codedock": {...}}, "...": "..." }
  }
}
```

Verification (Go, Codedock daemon):

```go
import "crypto/hmac"
import "crypto/sha256"
import "encoding/hex"

func valid(body []byte, signature, secret string) bool {
 digest := hmac.New(sha256.New, []byte(secret))
 _, _ = digest.Write(body)
 expected, err := hex.DecodeString(signature)
 return err == nil && hmac.Equal(digest.Sum(nil), expected)
}
```

- **Idempotency:** deliver at least once per event; dedupe on `id` (event id
  is unique per emission, shared across subscriptions) plus `data.tunnel.id`.
- Reply 2xx to accept; anything else is recorded as a failed delivery.

## Deploy flow (proposal)

After a successful local-directory deploy, when the project has
`tunnel.enabled: true`:

1. Ensure a scoped key exists for the project (create if missing).
2. `POST /api/v1/organizations/:organizationID/tunnels` with
   `metadata.codedock.project` set; store the returned `id` and
   `publicHostname` on the project.
3. Start the outpipe CLI (or agent) pointing at the service port; the
   platform emits `tunnel.connected` on success.
4. If the tunnel service is unreachable: skip the tunnel step, keep the
   deploy successful, log a warning, and retry on the next deploy.

## MCP tools (proposal)

Extend the Codedock daemon MCP command surface (`cmd/codedockd`):

- `tunnel_list` → `GET /api/v1/organizations/:organizationID/tunnels`
- `tunnel_inspect <id>` → `GET /api/v1/tunnels/:tunnelID`
- `tunnel_create <name> <port> [hostname] [protocol]` → create + metadata link
- `tunnel_revoke <id>` → `DELETE /api/v1/tunnels/:tunnelID`

All tools use the project's scoped key and never require platform-wide scopes.

## Resilience rules (hard requirements)

- All adapter HTTP calls are timeout-bounded (2 s) and degrade to warnings.
- No Codedock model or migration may depend on Outpipe data; the link is
  informational (`tunnel_id` column, no FK).
- Disabling or uninstalling the integration must never call outpipe's
  `DELETE` on tunnels: tunnels keep running until explicitly revoked by a
  user action (project delete with confirmed cleanup is allowed).
- The dashboard renders "tunnel service unreachable" instead of failing when
  Outpipe is down.

## Testing (Codedock `tests/`)

Prove standalone behavior with the real outpipe binaries (`cmd/server` +
`cmd/tunnel`) or a fake API server:

- deploy without the integration enabled works;
- deploy with outpipe down degrades gracefully;
- webhook delivery updates project tunnel status;
- scoped keys can create/revoke only their own tunnels;
- uninstalling the adapter leaves all outpipe tunnels running.
