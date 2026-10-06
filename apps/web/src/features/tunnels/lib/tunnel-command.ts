import type { Tunnel } from '#/interfaces/tunnel';

export function tunnelConnectCommand(
  tunnel: Pick<Tunnel, 'id' | 'protocol' | 'targetPort'>,
): string {
  if (
    !['http', 'https', 'tcp', 'udp'].includes(tunnel.protocol) ||
    !/^[a-zA-Z0-9_-]+$/.test(tunnel.id) ||
    !Number.isInteger(tunnel.targetPort) ||
    tunnel.targetPort < 1 ||
    tunnel.targetPort > 65535
  ) {
    throw new Error('Invalid tunnel connection details');
  }
  const protocol = tunnel.protocol === 'https' ? 'http' : tunnel.protocol;
  return `outpipe open --port ${tunnel.targetPort} --protocol ${protocol} --tunnel-id ${tunnel.id}`;
}

export function publicTunnelEndpoint(
  tunnel: Pick<Tunnel, 'protocol' | 'publicHostname' | 'publicPort'>,
): string | null {
  if (tunnel.protocol === 'tcp' || tunnel.protocol === 'udp') {
    return tunnel.publicPort
      ? `${tunnel.publicHostname}:${tunnel.publicPort}`
      : null;
  }
  const port =
    tunnel.publicPort && tunnel.publicPort !== 443 && tunnel.publicPort !== 80
      ? `:${tunnel.publicPort}`
      : '';
  return `https://${tunnel.publicHostname}${port}`;
}
