import { AdminShell } from './admin-overview-page';
import { useAdminUsage } from './hooks/use-admin-resources';

export function AdminUsagePage() {
  const query = useAdminUsage();
  if (query.isLoading)
    return <p className="p-8 text-sm text-white/55">Loading platform usage…</p>;
  if (query.isError || !query.data)
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load platform usage.
      </p>
    );
  const metrics = [
    ['Requests', query.data.requestCount.toLocaleString()],
    ['Errors', query.data.errorCount.toLocaleString()],
    ['Bandwidth', formatBytes(query.data.bandwidthBytes)],
  ];
  return (
    <AdminShell
      title="Platform usage"
      subtitle="Traffic totals aggregated from recorded usage events."
    >
      <div className="grid gap-4 sm:grid-cols-3">
        {metrics.map(([label, value]) => (
          <div
            key={label}
            className="rounded-2xl border border-white/10 bg-white/[0.025] p-5"
          >
            <p className="text-sm text-white/50">{label}</p>
            <p className="mt-3 text-3xl font-semibold">{value}</p>
          </div>
        ))}
      </div>
    </AdminShell>
  );
}
function formatBytes(value: number) {
  if (value < 1024 ** 2) return `${Math.round(value / 1024)} KB`;
  if (value < 1024 ** 3) return `${(value / 1024 ** 2).toFixed(1)} MB`;
  return `${(value / 1024 ** 3).toFixed(1)} GB`;
}
