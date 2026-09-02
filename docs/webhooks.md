# Webhooks

Webhook delivery is queued durably by Outpipe and retried with bounded
exponential backoff. A receiver should acknowledge a valid event quickly and
process it asynchronously. Delivery can be repeated after a timeout, a
non-2xx response, or worker restart.

Use `X-Outpipe-Event-ID` as the idempotency key. Store processed event IDs for
at least the maximum retry/replay window and safely ignore an event ID that has
already been applied. `X-Outpipe-Event-Type` identifies the event, while
`X-Outpipe-Signature` is an HMAC-SHA256 signature over the exact raw request
body. Verify the signature before parsing JSON.

Outpipe currently does not support customer-private webhook destinations.
Webhook URLs must be public HTTP(S) endpoints; loopback, private, link-local,
multicast, metadata-service, and unspecified addresses are rejected, including
when DNS resolves to more than one address.

The delivery response body is discarded after a small bounded read. Do not put
credentials, cookies, or sensitive data in URLs or response bodies.
