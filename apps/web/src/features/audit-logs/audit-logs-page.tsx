import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { useAuditLogs } from './hooks/use-audit-logs';

export function AuditLogsPage({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const organizationId = organizationQuery.organization?.id;

  const query = useAuditLogs(organizationId);

  if (organizationQuery.isLoading || query.isLoading) {
    return <p className="p-8 text-sm text-white/55">Loading audit history…</p>;
  }

  if (
    organizationQuery.isError ||
    query.isError ||
    !organizationQuery.organization
  ) {
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load audit history.
      </p>
    );
  }

  const events = query.data?.events ?? [];

  return (
    <div className="w-full max-w-6xl space-y-6 pb-12 text-white">
      <header className="border-b border-white/10 pb-6">
        <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-indigo-400">
          {organizationQuery.organization.name}
        </p>
        <h1 className="text-3xl font-semibold tracking-[-0.04em]">
          Audit history
        </h1>
        <p className="mt-1 text-sm text-white/55">
          A record of sensitive workspace and tunnel activity.
        </p>
      </header>
      <section className="mt-6 overflow-hidden rounded-2xl border border-white/10 bg-white/2.5">
        {events.length ? (
          events.map((event) => (
            <div
              key={event.id}
              className="grid gap-2 border-b border-white/5 px-5 py-4 sm:grid-cols-[1fr_180px] last:border-0"
            >
              <div>
                <p className="text-sm text-white/80">{event.action}</p>
                <p className="mt-1 text-xs text-white/40">
                  <span className="font-mono text-indigo-300/80">
                    {event.userId || 'system'}
                  </span>
                  {' · '}
                  {event.resourceType}
                  {event.resourceId ? ` · ${event.resourceId}` : ''}
                </p>
              </div>
              <time className="text-xs text-white/40 sm:text-right">
                {formatDate(event.occurredAt)}
              </time>
            </div>
          ))
        ) : (
          <p className="px-5 py-12 text-center text-sm text-white/50">
            No audit activity recorded yet.
          </p>
        )}
      </section>
    </div>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}
