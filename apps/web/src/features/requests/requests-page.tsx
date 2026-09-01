import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { useRequestEvents } from './hooks/use-request-events';

export function RequestsPage({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const query = useRequestEvents(organizationQuery.organization?.id);
  if (organizationQuery.isLoading || query.isLoading)
    return <p className="p-8 text-sm text-white/55">Loading requests…</p>;
  if (
    organizationQuery.isError ||
    query.isError ||
    !organizationQuery.organization
  )
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load request activity.
      </p>
    );
  const organization = organizationQuery.organization;
  const events = query.data?.events ?? [];
  return (
    <main className="mx-auto w-full max-w-6xl px-6 py-12 text-white sm:px-8 lg:py-16">
      <header className="border-b border-white/10 pb-8">
        <p className="mb-3 text-sm text-indigo-200">{organization.name}</p>
        <h1 className="text-3xl font-semibold tracking-tight">Requests</h1>
        <p className="mt-3 text-sm text-white/55">
          Recent traffic observed across your public endpoints.
        </p>
      </header>
      <section className="mt-8 overflow-hidden rounded-2xl border border-white/10 bg-white/[0.025]">
        <div className="grid grid-cols-[80px_minmax(0,1fr)_80px_80px] gap-4 border-b border-white/10 px-5 py-3 text-xs uppercase tracking-wider text-white/35">
          <span>Method</span>
          <span>Path</span>
          <span>Status</span>
          <span>Time</span>
        </div>
        {events.length ? (
          events.map((event) => (
            <div
              key={event.id}
              className="grid grid-cols-[80px_minmax(0,1fr)_80px_80px] gap-4 border-b border-white/5 px-5 py-4 font-mono text-sm last:border-0"
            >
              <span className="text-white/55">{event.method ?? '—'}</span>
              <span className="truncate text-white/80">
                {event.path ?? event.eventType}
              </span>
              <span className={statusColor(event.statusCode)}>
                {event.statusCode ?? '—'}
              </span>
              <span className="text-right text-white/45">
                {event.durationMillis ? `${event.durationMillis}ms` : '—'}
              </span>
            </div>
          ))
        ) : (
          <p className="px-5 py-12 text-center text-sm text-white/50">
            No requests recorded in the last 24 hours.
          </p>
        )}
      </section>
    </main>
  );
}

function statusColor(status?: number) {
  return status && status >= 500
    ? 'text-rose-200'
    : status && status >= 400
      ? 'text-amber-200'
      : 'text-emerald-200';
}
