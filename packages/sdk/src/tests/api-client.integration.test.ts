import { readFileSync } from 'node:fs';
import {
  createServer,
  type IncomingMessage,
  type ServerResponse,
} from 'node:http';
import { describe, expect, it } from 'vitest';
import { TunnelAPIClient } from '../services/api-client';
import type { TunnelAPIError } from '../utils/errors';

const contract = JSON.parse(
  readFileSync(
    new URL(
      '../../../../protocol/fixtures/http_tunnel_contract.json',
      import.meta.url,
    ),
    'utf8',
  ),
) as {
  authentication: string;
  tunnel: Record<string, unknown>;
  routes: {
    create: {
      method: string;
      path: string;
      request?: Record<string, unknown>;
      status: number;
    };
    list: { method: string; path: string; status: number };
    inspect: { method: string; path: string; status: number };
    revoke: { method: string; path: string; status: number };
  };
};

describe('TunnelAPIClient integration', () => {
  it('completes a tunnel lifecycle against a real local HTTP server', async () => {
    const requests: Array<{
      method: string;
      path: string;
      authorization: string | undefined;
    }> = [];
    const server = createServer(
      async (request: IncomingMessage, response: ServerResponse) => {
        const method = request.method ?? '';
        const path = request.url ?? '';
        requests.push({
          method,
          path,
          authorization: request.headers.authorization,
        });
        const expectedRoute = Object.values(contract.routes).find(
          (route) => route.method === method && route.path === path,
        );
        if (!expectedRoute && path.includes('/error%2F401')) {
          const body = JSON.stringify({ message: 'authentication required' });
          response.writeHead(401, {
            'Content-Type': 'application/json',
            'Content-Length': Buffer.byteLength(body),
          });
          response.end(body);
          return;
        }
        if (!expectedRoute) {
          response.writeHead(404);
          response.end();
          return;
        }
        const chunks: Buffer[] = [];
        for await (const chunk of request) {
          chunks.push(Buffer.from(chunk));
        }
        if ('request' in expectedRoute && expectedRoute.request) {
          expect(JSON.parse(Buffer.concat(chunks).toString('utf8'))).toEqual(
            expectedRoute.request,
          );
        }

        if (method === 'DELETE') {
          response.writeHead(expectedRoute.status);
          response.end();
          return;
        }

        const body =
          method === 'POST'
            ? JSON.stringify(contract.tunnel)
            : method === 'GET' && path.includes('/tunnels/tunnel-1')
              ? JSON.stringify(contract.tunnel)
              : JSON.stringify([contract.tunnel]);
        response.writeHead(expectedRoute.status, {
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(body),
        });
        response.end(body);
      },
    );

    await new Promise<void>((resolve) =>
      server.listen(0, '127.0.0.1', resolve),
    );
    const address = server.address();
    if (!address || typeof address === 'string') {
      throw new Error('local HTTP server did not expose an address');
    }
    const client = new TunnelAPIClient({
      apiUrl: `http://127.0.0.1:${address.port}`,
      apiKey: 'integration-key',
    });

    try {
      const created = await client.createTunnel('org/one', {
        name: 'preview',
        protocol: 'http',
        targetHost: '127.0.0.1',
        targetPort: 3000,
      });
      const tunnels = await client.listTunnels('org/one');
      const inspected = await client.inspectTunnel('tunnel-1');
      await client.closeTunnel('tunnel-1');

      expect(created.publicHostname).toBe('preview.outpipe.app');
      expect(tunnels).toHaveLength(1);
      expect(inspected.status).toBe('active');
      expect(requests.map(({ method, path }) => `${method} ${path}`)).toEqual([
        `${contract.routes.create.method} ${contract.routes.create.path}`,
        `${contract.routes.list.method} ${contract.routes.list.path}`,
        `${contract.routes.inspect.method} ${contract.routes.inspect.path}`,
        `${contract.routes.revoke.method} ${contract.routes.revoke.path}`,
      ]);
      expect(
        requests.every(
          (request) =>
            request.authorization ===
            `${contract.authentication} integration-key`,
        ),
      ).toBe(true);

      await expect(client.listTunnels('error/401')).rejects.toMatchObject({
        status: 401,
        message: 'authentication required',
      } satisfies Partial<TunnelAPIError>);
    } finally {
      await new Promise<void>((resolve, reject) =>
        server.close((error) => (error ? reject(error) : resolve())),
      );
    }
  });

  it('normalizes legacy tunnel URL fields at the API boundary', async () => {
    const client = new TunnelAPIClient({
      apiUrl: 'https://api.outpipe.dev',
      fetch: async () =>
        new Response(
          JSON.stringify({
            id: 'tunnel-1',
            public_url: 'https://preview.outpipe.app',
            status: 'active',
          }),
          {
            status: 200,
            headers: { 'Content-Type': 'application/json' },
          },
        ),
    });

    const tunnel = await client.inspectTunnel('tunnel-1');

    expect(tunnel.publicHostname).toBe('https://preview.outpipe.app');
  });
});
