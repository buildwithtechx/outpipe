import { Copy } from 'lucide-react';
import { Button } from '#/components/ui/button';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '#/components/ui/card';
import type { Tunnel } from '#/interfaces/tunnel';

type TunnelDetailCardProps = {
  tunnel: Tunnel;
};

export function TunnelDetailCard({ tunnel }: TunnelDetailCardProps) {
  const publicURL = buildPublicURL(tunnel);

  return (
    <Card className="border-white/10 bg-white/[0.025] py-0 text-white shadow-none">
      <CardHeader className="border-b border-white/10 px-5 py-5 sm:px-6">
        <CardTitle>Connection details</CardTitle>
        <CardDescription className="text-white/45">
          The address exposed by Outpipe and the local target it reaches.
        </CardDescription>
      </CardHeader>
      <CardContent className="grid gap-5 px-5 py-5 sm:grid-cols-2 sm:px-6">
        <DetailItem
          label="Public address"
          value={publicURL}
          copyValue={publicURL}
        />
        <DetailItem
          label="Local target"
          value={`${tunnel.targetHost}:${tunnel.targetPort}`}
        />
        <DetailItem label="Protocol" value={tunnel.protocol.toUpperCase()} />
        <DetailItem
          label="Access policy"
          value={formatPolicy(tunnel.accessPolicy)}
        />
        <DetailItem label="Created" value={formatDate(tunnel.createdAt)} />
        <DetailItem
          label="Last heartbeat"
          value={formatDate(tunnel.lastActiveAt)}
        />
      </CardContent>
    </Card>
  );
}

function DetailItem({
  label,
  value,
  copyValue,
}: {
  label: string;
  value: string;
  copyValue?: string;
}) {
  return (
    <div className="min-w-0">
      <p className="text-xs font-medium uppercase tracking-[0.14em] text-white/35">
        {label}
      </p>
      <div className="mt-2 flex min-w-0 items-center gap-2">
        <p className="truncate font-mono text-sm text-white/85">{value}</p>
        {copyValue && (
          <Button
            type="button"
            variant="ghost"
            size="icon-xs"
            aria-label={`Copy ${label.toLowerCase()}`}
            onClick={() => void navigator.clipboard.writeText(copyValue)}
            className="shrink-0 text-white/45 hover:bg-white/10 hover:text-white"
          >
            <Copy />
          </Button>
        )}
      </div>
    </div>
  );
}

function buildPublicURL(tunnel: Tunnel) {
  const scheme = tunnel.protocol === 'https' ? 'https' : 'http';
  const port = tunnel.publicPort ? `:${tunnel.publicPort}` : '';
  return `${scheme}://${tunnel.publicHostname}${port}`;
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
