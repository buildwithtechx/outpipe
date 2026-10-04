import type { RelayConnectionOptions } from '@outpipe/sdk';

export type TanStackStartTunnelOptions = RelayConnectionOptions & {
  localPort: number;
  subdomain?: string;
  password?: string;
  autoStart?: boolean;
};

export type TanStackStartTunnel = {
  start: () => Promise<TanStackStartTunnelState>;
  stop: (reason?: string) => Promise<void>;
  state: () => TanStackStartTunnelState;
};

export type TanStackStartTunnelState = {
  status: 'idle' | 'connecting' | 'active' | 'closed' | 'error';
  tunnelId?: string;
  publicUrl?: string;
  publicPort?: number;
  error?: Error;
};
