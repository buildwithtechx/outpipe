import type { RelayConnectionOptions } from '@outpipe/sdk';

export type OutpipePluginOptions = RelayConnectionOptions & {
  enabled?: boolean;
  localPort?: number;
  subdomain?: string;
  password?: string;
};
