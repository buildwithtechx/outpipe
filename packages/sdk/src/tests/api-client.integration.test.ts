import {
  createServer,
  type IncomingMessage,
  type ServerResponse,
} from 'node:http';
import { describe, expect, it } from 'vitest';
import { TunnelAPIClient } from '../services/api-client';

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
        for await (const _chunk of request) {
          // Consume request bodies so the test exercises normal HTTP request handling.
        }

        if (method === 'DELETE') {
          response.writeHead(204);
          response.end();
          return;
        }

        const body =
          method === 'POST'
            ? JSON.stringify({
                id: 'tunnel-1',
                publicUrl: 'https://preview.outpipe.app',
                status: 'active',
              })
            : method === 'GET' && path.includes('/tunnels/tunnel-1')
              ? JSON.stringify({
                  id: 'tunnel-1',
                  publicUrl: 'https://preview.outpipe.app',
                  status: 'active',
                })
              : JSON.stringify([{ id: 'tunnel-1', status: 'active' }]);
        response.writeHead(200, {
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
      });
      const tunnels = await client.listTunnels('org/one');
      const inspected = await client.inspectTunnel('tunnel-1');
      await client.closeTunnel('tunnel-1');

      expect(created.publicUrl).toBe('https://preview.outpipe.app');
      expect(tunnels).toHaveLength(1);
      expect(inspected.status).toBe('active');
      expect(requests.map(({ method, path }) => `${method} ${path}`)).toEqual([
        'POST /api/v1/organizations/org%2Fone/tunnels',
        'GET /api/v1/organizations/org%2Fone/tunnels',
        'GET /api/v1/tunnels/tunnel-1',
        'DELETE /api/v1/tunnels/tunnel-1',
      ]);
      expect(
        requests.every(
          (request) => request.authorization === 'Bearer integration-key',
        ),
      ).toBe(true);
    } finally {
      await new Promise<void>((resolve, reject) =>
        server.close((error) => (error ? reject(error) : resolve())),
      );
    }
  });
});
