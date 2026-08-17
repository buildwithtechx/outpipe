import { RelayConnection } from '@outpipe/sdk';
import type {
  NextTunnel,
  NextTunnelOptions,
  NextTunnelState,
} from '../interfaces/options';

export function createNextTunnel(options: NextTunnelOptions): NextTunnel {
  const connection = new RelayConnection(options);
  let current: NextTunnelState = { status: 'idle' };
  let startPromise: Promise<NextTunnelState> | undefined;
  let generation = 0;

  connection.on('tunnel_opened', (tunnel) => {
    current = {
      status: 'active',
      tunnelId: tunnel.tunnel_id,
      publicUrl: tunnel.public_url,
      publicPort: tunnel.public_port,
    };
  });
  connection.on('disconnected', () => {
    return;
  });
  connection.on('error', (error) => {
    if (current.status !== 'closed') current = { status: 'error', error };
  });

  return {
    start: async () => {
      if (options.enabled === false) {
        return current;
      }
      if (current.status === 'active') return current;
      if (startPromise) return startPromise;
      const startGeneration = ++generation;
      current = { status: 'connecting' };
      startPromise = connection
        .openTunnel({
          local_port: options.localPort,
          protocol: 'http',
          subdomain: options.subdomain,
          password: options.password,
        })
        .then((tunnel) => {
          if (startGeneration !== generation) return current;
          current = {
            status: 'active',
            tunnelId: tunnel.tunnel_id,
            publicUrl: tunnel.public_url,
            publicPort: tunnel.public_port,
          };
          return current;
        })
        .catch((value) => {
          if (startGeneration === generation) {
            current = { status: 'error', error: normalizeError(value) };
          }
          throw value;
        })
        .finally(() => {
          startPromise = undefined;
        });
      return startPromise;
    },
    stop: async (reason) => {
      generation += 1;
      const tunnelId = current.tunnelId;
      let closeError: unknown;
      try {
        if (tunnelId) await connection.closeTunnel(tunnelId, reason);
      } catch (error) {
        closeError = error;
      } finally {
        connection.close();
        current = { status: 'closed' };
      }
      if (closeError) throw closeError;
    },
    state: () => current,
  };
}

function normalizeError(value: unknown): Error {
  return value instanceof Error ? value : new Error(String(value));
}
