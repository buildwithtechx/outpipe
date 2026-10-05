# @outpipe/sdk

Framework-neutral TypeScript and Node.js client for the standalone Outpipe API.

The SDK owns client authentication, tunnel lifecycle operations, connection state, reconnect behavior, and protocol communication. Framework packages should remain thin adapters over this package.

The SDK bundles the shared protocol bindings and supports API-key
authentication for server-side and CI usage. Browser applications should use
short-lived tokens issued by the API.

```ts
import { TunnelAPIClient } from '@outpipe/sdk';

const client = new TunnelAPIClient({ apiKey: process.env.OUTPIPE_API_KEY });
const tunnels = await client.listTunnels(organizationId);
```

Clients default to `https://api.outpipe.dev`. Set `apiUrl` only when connecting
to a self-hosted API.
