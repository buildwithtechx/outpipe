import type { RelayConnectionOptions } from '@outpipe/sdk';

export type HonoTunnelOptions = RelayConnectionOptions & {
  localPort: number;
  subdomain?: string;
  password?: string;
  autoStart?: boolean;
};

export type HonoTunnel = {
  start: () => Promise<HonoTunnelState>;
  stop: (reason?: string) => Promise<void>;
  state: () => HonoTunnelState;
};

export type HonoTunnelState = {
  status: 'idle' | 'connecting' | 'active' | 'closed' | 'error';
  tunnelId?: string;
  publicUrl?: string;
  publicPort?: number;
  error?: Error;
};
