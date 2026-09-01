import { Badge } from '#/components/ui/badge';
import type { TunnelStatus } from '#/interfaces/tunnel';

const statusClasses: Record<TunnelStatus, string> = {
  active: 'bg-emerald-400/10 text-emerald-200 ring-emerald-300/20',
  connecting: 'bg-amber-400/10 text-amber-100 ring-amber-300/20',
  created: 'bg-indigo-300/10 text-indigo-200 ring-indigo-300/20',
  disconnected: 'bg-white/5 text-white/55 ring-white/10',
  expired: 'bg-amber-400/10 text-amber-100 ring-amber-300/20',
  revoked: 'bg-rose-400/10 text-rose-200 ring-rose-300/20',
};

export function TunnelStatusBadge({ status }: { status: TunnelStatus }) {
  return (
    <Badge
      variant="outline"
      className={`border-0 ring-1 ring-inset ${statusClasses[status]}`}
    >
      {status}
    </Badge>
  );
}
