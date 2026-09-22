import type { RelayConnectionOptions } from '@outpipe/sdk';

export const OUTPIPE_OPTIONS = Symbol('OUTPIPE_OPTIONS');

export type NestTunnelOptions = RelayConnectionOptions & {
  localPort: number;
  subdomain?: string;
  password?: string;
  autoStart?: boolean;
};
