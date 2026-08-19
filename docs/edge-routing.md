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
OUTPIPE_DOMAIN=outpipe.app
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

Provider-specific DNS automation is intentionally outside the core service. A future user-authorized integration may create records through a customer-owned DNS token, but the hosted service does not store or use a global DNS provider credential.

## Local development

Use the local example values:

```text
OUTPIPE_REQUIRE_TLS=false
OUTPIPE_REQUIRE_TLS=false
OUTPIPE_DOMAIN=outpipe.localhost
```
