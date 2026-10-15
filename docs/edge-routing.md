# Tunnel edge routing

The tunnel relay owns public routing. DNS automation is optional and is not part of the tunnel protocol.

## Hosted relay

Create these records once for the hosted installation:

```text
relay.outpipe.app        A/AAAA <relay-address>
*.outpipe.app            A/AAAA <relay-address>
```

The relay then assigns URLs such as:

```text
https://abc123.outpipe.app
```

No DNS record is created for an individual tunnel.

When the relay is directly exposed on the public network, provision a certificate containing `*.outpipe.app` and `relay.outpipe.app`. Mount the certificate and key into the relay and set:

```text
OUTPIPE_REQUIRE_TLS=true
OUTPIPE_TLS_CERT_FILE=/run/secrets/tunnel/fullchain.pem
OUTPIPE_TLS_KEY_FILE=/run/secrets/tunnel/privkey.pem
OUTPIPE_TUNNEL_DOMAIN=outpipe.app
```

The relay reloads the certificate on the next TLS handshake after either mounted file changes. A deployment can renew the certificate atomically by writing new files beside the existing files and replacing them with a rename.

Wildcard certificates require DNS-01 validation. The current `autocert` helper is intended for individual host certificates and should not be used as the wildcard issuance mechanism.

## Outpipe PaaS or another reverse proxy

When Outpipe PaaS or another reverse proxy owns the public domain, TLS belongs to that proxy. Configure the relay as an internal HTTP service and leave these values disabled:

```text
OUTPIPE_REQUIRE_TLS=false
OUTPIPE_TLS_CERT_FILE=
OUTPIPE_TLS_KEY_FILE=
```

Configure the proxy with the relay domain and forward WebSocket upgrades. The external client still uses `wss://`; TLS is terminated at the proxy and the internal hop uses HTTP/WebSocket. The relay container must not publish its internal port directly to the internet.

HTTP tunnels work through the normal HTTPS reverse-proxy route. Raw TCP and UDP tunnels require explicit public TCP/UDP port mappings or a proxy that supports TCP/UDP entrypoints; an ordinary HTTPS domain route cannot carry raw TCP or UDP traffic.

## Custom domains

Custom domains are verified independently from the relay wildcard. Users manage their DNS records themselves and provide either the DNS TXT challenge or HTTP challenge required by the dashboard.

For automated verification and wildcard certificates, connect a customer-owned DNS provider credential with `POST /api/v1/organizations/{organizationID}/dns-providers` (Cloudflare API token, or Route53 JSON with `access_key_id`/`secret_access_key`/`session_token`), then register the domain with verification method `dns-automation`. The platform sets and cleans up `_acme-challenge` TXT records through the customer credential, completes DNS-01, and issues the certificate; wildcard hostnames (`*.example.com`) require automation. Register apex plus wildcard together with `POST .../domains/bundle` (`{"hostname":"example.com","wildcard":true}`) for a single certificate covering both names. Tokens are encrypted at rest, never returned by the API, and only need DNS-edit scope on the relevant zone (Cloudflare Zone-DNS-edit; Route53 `ChangeResourceRecordSets` on the zone ARN only). Certificates are renewed automatically by the cron `domain-certificates` worker starting 30 days before expiry. The hosted service never uses a global DNS provider credential for customer zones.

Customers who cannot share API tokens use CNAME delegation instead (`dns-delegation`): create `_acme-challenge.<host> CNAME <target>` where `<target>` is the delegation target returned at registration, then verify. The platform publishes challenge TXT records only inside its own operator-owned challenge zone (`OUTPIPE_CHALLENGE_ZONE`, written with a server-held credential scoped to that zone alone) — the sole exception to the no-global-credential rule. Bundles share one delegation target because both authorizations resolve under `_acme-challenge.<apex>`.

### Relay SNI certificate serving

Issued certificates are served to relays over the internal API so custom-domain TLS works end to end on a stock relay deploy:

- `GET /internal/domains/certificate?hostname=<host>` (or `?tunnelId=<id>`) returns the PEM chain, key, and expiry for the calling relay. Access is scope-checked per organization: the domain's organization must match the tunnel's organization (or the resolved domain's when queried by hostname), and only `active`/`ready`, unexpired, non-revoked certificates are served. Exact hostnames win; otherwise a `*.parent` wildcard certificate covers subdomains.
- The relay selects certificates by SNI on every TLS handshake (`SNIDispatcher`): per-host certificates from the internal API with a 1-minute cache, falling back to the mounted `TLSCertFile`/`TLSKeyFile` when no custom certificate exists or the API is unreachable. In-memory entries expire and refresh; revocation stops serving within the cache window.
- Revoke with `POST /api/v1/domains/:domainID/revoke` (or delete the domain): the domain is marked `revoked`, certificate files are removed from the cache dir, and relays stop serving it once cached entries expire. Deleting a domain also removes its `<host>.crt`/`<host>.key` files.

To enable it, run the relay with TLS and the internal API configured — no extra flags are needed:

```text
OUTPIPE_REQUIRE_TLS=true
OUTPIPE_TLS_CERT_FILE=/run/secrets/tunnel/fullchain.pem
OUTPIPE_TLS_KEY_FILE=/run/secrets/tunnel/privkey.pem
OUTPIPE_INTERNAL_API_URL=http://127.0.0.1:9090
OUTPIPE_INTERNAL_API_SECRET=<shared secret>
OUTPIPE_TUNNEL_DOMAIN=outpipe.app
```

The API side needs `OUTPIPE_CERTIFICATE_CACHE_DIR` pointing at the same certificate cache the issuance and renewal workers write to. When the internal API is unreachable, the relay keeps serving the mounted fallback certificate.

## Local development

Use the local example values:

```text
OUTPIPE_REQUIRE_TLS=false
OUTPIPE_TUNNEL_DOMAIN=outpipe.localhost
```
