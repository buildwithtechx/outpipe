import type { RelayConnectionOptions } from '@outpipe/sdk';

export type NextTunnelOptions = RelayConnectionOptions & {
  localPort: number;
  subdomain?: string;
  password?: string;
  enabled?: boolean;
};

export type NextTunnel = {
  start: () => Promise<NextTunnelState>;
  stop: (reason?: string) => Promise<void>;
  state: () => NextTunnelState;
};

export type NextTunnelState = {
  status: 'idle' | 'connecting' | 'active' | 'closed' | 'error';
  tunnelId?: string;
  publicUrl?: string;
  publicPort?: number;
  error?: Error;
};
