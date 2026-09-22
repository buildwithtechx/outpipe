import { RelayConnection } from '@outpipe/sdk';
import type {
  ExpressTunnel,
  ExpressTunnelOptions,
  ExpressTunnelState,
} from '../interfaces/options';

export function createExpressTunnel(
  options: ExpressTunnelOptions,
): ExpressTunnel {
  const connection = new RelayConnection(options);
  let current: ExpressTunnelState = { status: 'idle' };
  let startPromise: Promise<ExpressTunnelState> | undefined;
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
    if (!options.reconnect && current.status !== 'closed') {
      current = { status: 'closed' };
    }
  });
  connection.on('reconnect_exhausted', () => {
    if (current.status !== 'closed') current = { status: 'closed' };
  });
  connection.on('error', (error) => {
    if (current.status !== 'closed') {
      current = { status: 'error', error };
    }
  });

  return {
    start: async () => {
      if (options.autoStart === false) {
        return current;
      }
      if (current.status === 'active') {
        return current;
      }
      if (startPromise) {
        return startPromise;
      }
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
          if (startGeneration !== generation) {
            return current;
          }
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
        if (tunnelId) {
          await connection.closeTunnel(tunnelId, reason);
        }
      } catch (error) {
        closeError = error;
      } finally {
        connection.close();
        current = { status: 'closed' };
      }
      if (closeError) {
        throw closeError;
      }
    },
    state: () => current,
  };
}

function normalizeError(value: unknown): Error {
  return value instanceof Error ? value : new Error(String(value));
}
