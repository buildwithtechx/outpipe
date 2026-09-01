export type TunnelProtocol = 'http' | 'https' | 'tcp' | 'udp';

export type TunnelStatus =
  | 'created'
  | 'connecting'
  | 'active'
  | 'disconnected'
  | 'expired'
  | 'revoked';

import type { Entity } from '#/interfaces/api';

export type Tunnel = Entity & {
  organizationId: string;
  agentId?: string;
  name: string;
  protocol: TunnelProtocol;
  status: TunnelStatus;
  targetHost: string;
  targetPort: number;
  publicHostname: string;
  publicPort?: number;
  accessPolicy: string;
  expiresAt?: string;
  lastActiveAt?: string;
  revokedAt?: string;
};

export type CreateTunnelRequest = {
  name: string;
  protocol: TunnelProtocol;
  targetHost: string;
  targetPort: number;
  publicHostname?: string;
  password?: string;
};

export type UpdateTunnelConfigurationRequest = {
  accessPolicy: string;
  expiresAt?: string;
};
