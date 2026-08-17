import type { RelayConnectionOptions } from '@outpipe/sdk';

export type ExpressTunnelOptions = RelayConnectionOptions & {
  localPort: number;
  subdomain?: string;
  password?: string;
  autoStart?: boolean;
};

export type ExpressTunnel = {
  start: () => Promise<ExpressTunnelState>;
  stop: (reason?: string) => Promise<void>;
  state: () => ExpressTunnelState;
};

export type ExpressTunnelState = {
  status: 'idle' | 'connecting' | 'active' | 'closed' | 'error';
  tunnelId?: string;
  publicUrl?: string;
  publicPort?: number;
  error?: Error;
};
