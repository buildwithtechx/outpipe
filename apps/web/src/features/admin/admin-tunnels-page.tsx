import { AdminShell } from './admin-overview-page';
import { useAdminTunnels } from './hooks/use-admin-resources';
export function AdminTunnelsPage() {
  const query = useAdminTunnels();
  if (query.isLoading)
    return (
      <p className="p-8 text-sm text-white/55">Loading platform tunnels…</p>
    );
  if (query.isError)
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load platform tunnels.
      </p>
    );
  return (
    <AdminShell
      title="All tunnels"
      subtitle={`${query.data?.total ?? 0} tunnels across every workspace.`}
    >
      <div className="overflow-hidden rounded-2xl border border-white/10 bg-white/[0.025]">
        {query.data?.items.map((tunnel) => (
          <div
            key={tunnel.id}
            className="grid gap-2 border-b border-white/5 px-5 py-4 last:border-0 sm:grid-cols-[1fr_auto] sm:items-center"
          >
            <div>
              <p className="text-sm text-white/85">{tunnel.name}</p>
              <p className="mt-1 font-mono text-xs text-white/45">
                {tunnel.publicHostname}
              </p>
            </div>
            <span className="text-xs text-indigo-200">{tunnel.status}</span>
          </div>
        ))}
      </div>
    </AdminShell>
  );
}
