import { AdminShell } from './admin-overview-page';
import { useAdminOverview, useAdminUsage } from './hooks/use-admin-resources';

export function AdminChartsPage() {
  const overview = useAdminOverview();
  const usage = useAdminUsage();
  if (overview.isLoading || usage.isLoading)
    return (
      <p className="p-8 text-sm text-white/55">Loading platform trends…</p>
    );
  if (overview.isError || usage.isError || !overview.data || !usage.data)
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load platform trends.
      </p>
    );
  const values: Array<[string, number]> = [
    ['Accounts', overview.data.users],
    ['Workspaces', overview.data.organizations],
    ['Tunnels', overview.data.tunnels],
    ['Requests', usage.data.requestCount],
    ['Errors', usage.data.errorCount],
  ];
  const max = Math.max(...values.map(([, value]) => value), 1);
  return (
    <AdminShell
      title="Platform trends"
      subtitle="A compact operational snapshot for the control plane."
    >
      <div className="rounded-2xl border border-white/10 bg-white/[0.025] p-6">
        <div className="grid gap-5">
          {values.map(([label, value]) => (
            <div key={label as string}>
              <div className="flex justify-between text-sm">
                <span className="text-white/60">{label}</span>
                <span className="text-white/85">{value.toLocaleString()}</span>
              </div>
              <div className="mt-2 h-2 rounded-full bg-white/10">
                <div
                  className="h-2 rounded-full bg-indigo-300"
                  style={{ width: `${Math.max(4, (value / max) * 100)}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      </div>
    </AdminShell>
  );
}
