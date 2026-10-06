import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '#/components/ui/card';
import type { Tunnel } from '#/interfaces/tunnel';
import { publicTunnelEndpoint } from '../lib/tunnel-command';
import { TunnelDetailItem } from './tunnel-detail-item';

type TunnelDetailCardProps = {
  tunnel: Tunnel;
};

export function TunnelDetailCard({ tunnel }: TunnelDetailCardProps) {
  const publicEndpoint = publicTunnelEndpoint(tunnel);

  return (
    <Card className="border-white/10 bg-white/2.5 py-0 text-white shadow-none">
      <CardHeader className="border-b border-white/10 px-5 py-5 sm:px-6">
        <CardTitle>Connection details</CardTitle>
        <CardDescription className="text-white/45">
          The address exposed by Outpipe and the local target it reaches.
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-5 px-5 py-5 sm:grid-cols-2 sm:px-6">
        <TunnelDetailItem
          label={
            tunnel.protocol === 'tcp' || tunnel.protocol === 'udp'
              ? 'Public endpoint'
              : 'Public address'
          }
          value={
            publicEndpoint ??
            `${tunnel.publicHostname} (assigned on connection)`
          }
          copyValue={
            tunnel.protocol === 'tcp' || tunnel.protocol === 'udp'
              ? tunnel.publicPort
                ? (publicEndpoint ?? undefined)
                : undefined
              : (publicEndpoint ?? undefined)
          }
        />
        <TunnelDetailItem
          label="Local target"
          value={`${tunnel.targetHost}:${tunnel.targetPort}`}
        />
        <TunnelDetailItem
          label="Protocol"
          value={tunnel.protocol.toUpperCase()}
        />
        <TunnelDetailItem
          label="Access policy"
          value={formatPolicy(tunnel.accessPolicy)}
        />
        <TunnelDetailItem
          label="Created"
          value={formatDate(tunnel.createdAt)}
        />
        <TunnelDetailItem
          label="Last heartbeat"
          value={formatDate(tunnel.lastActiveAt)}
        />
      </CardContent>
    </Card>
  );
}

function formatPolicy(policy: string) {
  return policy === '{}' ? 'Default policy' : 'Custom policy';
}

function formatDate(value: string | undefined) {
  if (!value) return 'No heartbeat yet';
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}
