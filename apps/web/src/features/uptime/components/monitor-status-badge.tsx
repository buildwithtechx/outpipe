import { Badge } from '#/components/ui/badge';
import type { UptimeMonitor } from '#/interfaces';

export function MonitorStatusBadge({
  status,
}: {
  status: UptimeMonitor['status'];
}) {
  const colors = {
    up: ['bg-emerald-500/15 text-emerald-400', 'bg-emerald-400'],
    down: ['bg-rose-500/15 text-rose-400', 'bg-rose-400'],
    degraded: ['bg-amber-500/15 text-amber-400', 'bg-amber-400'],
    paused: ['bg-zinc-500/15 text-zinc-400', 'bg-zinc-400'],
  }[status];
  return (
    <Badge
      variant={status === 'down' ? 'destructive' : 'default'}
      className={`text-[11px] font-medium gap-1.5 ${colors[0]}`}
    >
      <span className={`size-1.5 rounded-full ${colors[1]}`} />
      {status}
    </Badge>
  );
}
