import { AdminShell } from './admin-overview-page';
import { useAdminSubscriptions } from './hooks/use-admin-resources';

export function AdminSubscriptionsPage() {
  const query = useAdminSubscriptions();

  if (query.isLoading) {
    return <p className="p-8 text-sm text-white/55">Loading subscriptions…</p>;
  }

  if (query.isError) {
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load subscriptions.
      </p>
    );
  }

  return (
    <AdminShell
      title="Subscriptions"
      subtitle={`${query.data?.total ?? 0} subscription records to monitor.`}
    >
      <div className="overflow-hidden rounded-2xl border border-white/10 bg-white/[0.025]">
        {query.data?.items.map((subscription) => (
          <div
            key={subscription.id}
            className="grid gap-2 border-b border-white/5 px-5 py-4 last:border-0 sm:grid-cols-[1fr_auto_auto] sm:items-center"
          >
            <span className="font-mono text-xs text-white/45">
              {subscription.organizationId}
            </span>
            <span className="text-sm text-white/70">
              {subscription.provider}
            </span>
            <span className="text-xs text-emerald-200">
              {subscription.status}
            </span>
          </div>
        ))}
      </div>
    </AdminShell>
  );
}
