import { createServer } from 'node:http';
import { afterEach, describe, expect, it } from 'vitest';
import { WebSocket, WebSocketServer } from 'ws';
import type { WebSocketLike } from '../interfaces/relay';
import { decodeMessage, encodeMessage } from '../protocol';
import { RelayConnection } from '../services/relay-connection';

describe('RelayConnection integration', () => {
  let httpServer: ReturnType<typeof createServer> | undefined;
  let webSocketServer: WebSocketServer | undefined;

  afterEach(async () => {
    await new Promise<void>((resolve) =>
      webSocketServer?.close(() => resolve()),
    );
    await new Promise<void>((resolve) => httpServer?.close(() => resolve()));
    webSocketServer = undefined;
    httpServer = undefined;
  });

  it('negotiates, authenticates, opens, and closes through a real WebSocket server', async () => {
    const received: Array<{ type: string; payload: Record<string, unknown> }> =
      [];
    httpServer = createServer();
    webSocketServer = new WebSocketServer({ server: httpServer });
    webSocketServer.on('connection', (socket) => {
      socket.on('message', (raw) => {
        const message = decodeMessage(raw.toString());
        received.push({
          type: message.type,
          payload: message.payload as Record<string, unknown>,
        });
        const response =
          message.type === 'version_negotiate'
            ? {
                type: 'version_negotiate_ack' as const,
                payload: { negotiated_version: 1, supported_versions: [1] },
              }
            : message.type === 'auth'
              ? {
                  type: 'auth_response' as const,
                  payload: {
                    authenticated: true,
                    agent_id: 'agent-1',
                    organization_id: 'org-1',
                  },
                }
              : message.type === 'open_tunnel'
                ? {
                    type: 'open_tunnel_ack' as const,
                    payload: {
                      tunnel_id: 'tunnel-1',
                      public_url: 'https://preview.outpipe.app',
                    },
                  }
                : undefined;
        if (response) {
          socket.send(
            encodeMessage({
              version: 1,
              type: response.type,
              request_id: message.request_id,
              payload: response.payload,
            }),
          );
        }
      });
    });

    await new Promise<void>((resolve) =>
      httpServer?.listen(0, '127.0.0.1', resolve),
    );
    const address = httpServer?.address();
    if (!address || typeof address === 'string') {
      throw new Error('local WebSocket server did not expose an address');
    }

    const connection = new RelayConnection({
      relayUrl: `ws://127.0.0.1:${address.port}`,
      agentToken: 'integration-token',
      webSocket: (url) => new WebSocket(url) as unknown as WebSocketLike,
      heartbeatIntervalMs: 60_000,
    });

    await connection.connect();
    const tunnel = await connection.openTunnel({
      local_port: 3000,
      protocol: 'http',
    });
    await connection.closeTunnel('tunnel-1');
    await new Promise((resolve) => setTimeout(resolve, 25));
    connection.close();

    expect(tunnel).toMatchObject({
      tunnel_id: 'tunnel-1',
      public_url: 'https://preview.outpipe.app',
    });
    expect(received.map(({ type }) => type)).toEqual([
      'version_negotiate',
      'auth',
      'open_tunnel',
      'close_tunnel',
    ]);
    expect(received[1]?.payload.token).toBe('integration-token');
    expect(received[2]?.payload).toMatchObject({
      token: 'integration-token',
      local_port: 3000,
      protocol: 'http',
    });
  }, 10_000);
});
