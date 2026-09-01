import { Link } from '@tanstack/react-router';
import type { Tunnel } from '#/interfaces/tunnel';
import { TunnelStatusBadge } from './tunnel-status-badge';

type TunnelListItemProps = {
  tunnel: Tunnel;
  orgSlug: string;
};

export function TunnelListItem({ tunnel, orgSlug }: TunnelListItemProps) {
  return (
    <Link
      to="/$orgSlug/tunnels/$tunnelId"
      params={{ orgSlug, tunnelId: tunnel.id }}
      className="group grid gap-4 rounded-2xl border border-white/10 bg-white/[0.025] p-5 transition-colors hover:border-indigo-300/40 hover:bg-indigo-300/[0.05] focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-indigo-300 sm:grid-cols-[minmax(0,1fr)_auto] sm:items-center"
    >
      <div className="min-w-0">
        <div className="flex flex-wrap items-center gap-2">
          <h3 className="truncate font-medium text-white">{tunnel.name}</h3>
          <TunnelStatusBadge status={tunnel.status} />
          <span className="rounded-full border border-white/10 px-2 py-0.5 font-mono text-[11px] uppercase tracking-wide text-white/50">
            {tunnel.protocol}
          </span>
        </div>
        <p className="mt-2 truncate font-mono text-sm text-indigo-200">
          {tunnel.publicHostname}
        </p>
      </div>
      <div className="font-mono text-xs text-white/45 sm:text-right">
        {tunnel.targetHost}:{tunnel.targetPort}
      </div>
    </Link>
  );
}
