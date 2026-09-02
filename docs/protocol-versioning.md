# Outpipe protocol and API versioning

Outpipe has two related contracts:

- Control-plane HTTP JSON uses camelCase DTOs. `publicHostname` is the canonical
  field for a tunnel's public host in the API and in SDK models.
- The relay wire protocol retains `public_url` because connected agents and
  older relay implementations use that field. It is translated at the SDK/API
  boundary and must not leak into new control-plane DTOs.

Protocol version 1 is additive by default. New optional fields may be added
without changing the version. A breaking change—such as removing a field,
changing its meaning, or changing authentication semantics—requires a new
protocol version and a compatibility window for the previous version.

Every SDK release must run its behavior tests against the shared protocol
fixtures before publishing. SDK adapters may use idiomatic names such as
`publicUrl`, but their HTTP serialization must explicitly map to
`publicHostname`.

Agent credentials are short-lived and scoped to a tunnel session. API keys
remain management credentials and must not be substituted for relay session
tokens in the wire protocol.
