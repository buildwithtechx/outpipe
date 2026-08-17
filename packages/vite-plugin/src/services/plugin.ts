import { RelayConnection } from '@outpipe/sdk';
import type { Plugin, ViteDevServer } from 'vite';
import type { OutpipePluginOptions } from '../interfaces/options';

export function outpipeTunnel(options: OutpipePluginOptions): Plugin {
  let connection: RelayConnection | undefined;

  return {
    name: 'outpipe',
    apply: 'serve',
    configureServer(server: ViteDevServer) {
      if (options.enabled === false) {
        return;
      }
      const start = () =>
        startTunnel(server, options, (nextConnection) => {
          connection = nextConnection;
        });
      if (server.httpServer?.listening) {
        void start().catch((error: unknown) => {
          server.config.logger.error(`Outpipe failed: ${String(error)}`);
        });
      } else if (server.httpServer) {
        server.httpServer.once('listening', () => {
          void start().catch((error: unknown) => {
            server.config.logger.error(`Outpipe failed: ${String(error)}`);
          });
        });
      } else {
        if (options.localPort) {
          void start().catch((error: unknown) => {
            server.config.logger.error(`Outpipe failed: ${String(error)}`);
          });
        } else {
          server.config.logger.warn(
            'Outpipe requires localPort when Vite has no HTTP server',
          );
        }
      }
    },
    closeBundle() {
      connection?.close();
      connection = undefined;
    },
  };
}

async function startTunnel(
  server: ViteDevServer,
  options: OutpipePluginOptions,
  assign: (connection: RelayConnection) => void,
): Promise<void> {
  const localPort = resolveLocalPort(server, options);
  const connection = new RelayConnection({ ...options, localPort });
  assign(connection);
  try {
    const tunnel = await connection.openTunnel({
      local_port: localPort,
      protocol: 'http',
      subdomain: options.subdomain,
      password: options.password,
    });
    server.config.logger.info(`Outpipe: ${tunnel.public_url}`);
    server.httpServer?.once('close', () => connection.close());
  } catch (error) {
    connection.close();
    throw error;
  }
}

function resolveLocalPort(
  server: ViteDevServer,
  options: OutpipePluginOptions,
): number {
  if (options.localPort) {
    return options.localPort;
  }
  const address = server.httpServer?.address();
  if (address && typeof address === 'object') {
    return address.port;
  }
  return server.config.server.port ?? 5173;
}
