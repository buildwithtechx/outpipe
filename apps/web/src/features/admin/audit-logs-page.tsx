import { useAdminAuditLogs } from './hooks/use-admin-audit-logs';

export function AdminAuditLogsPage() {
  const query = useAdminAuditLogs();

  if (query.isLoading) {
    return (
      <p className="p-8 text-sm text-white/55">Loading platform audit logs…</p>
    );
  }

  if (query.isError) {
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load platform audit logs.
      </p>
    );
  }

  const events = query.data?.items ?? [];

  return (
    <main className="mx-auto w-full max-w-6xl px-6 py-12 text-white sm:px-8 lg:py-16">
      <header className="border-b border-white/10 pb-8">
        <p className="mb-3 text-sm text-indigo-200">Platform administration</p>
        <h1 className="text-3xl font-semibold tracking-tight">Audit logs</h1>
        <p className="mt-3 text-sm text-white/55">
          Review sensitive actions across the Outpipe platform.
        </p>
      </header>
      <section className="mt-8 overflow-hidden rounded-2xl border border-white/10 bg-white/[0.025]">
        {events.length ? (
          events.map((event) => (
            <div
              key={event.id}
              className="grid gap-2 border-b border-white/5 px-5 py-4 sm:grid-cols-[1fr_200px] last:border-0"
            >
              <div>
                <p className="text-sm text-white/80">{event.action}</p>
                <p className="mt-1 text-xs text-white/40">
                  {event.resourceType} · {event.userId ?? 'system'}
                </p>
              </div>
              <time className="text-xs text-white/40 sm:text-right">
                {formatDate(event.occurredAt)}
              </time>
            </div>
          ))
        ) : (
          <p className="px-5 py-12 text-center text-sm text-white/50">
            No platform activity recorded.
          </p>
        )}
      </section>
    </main>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, {
    dateStyle: 'medium',
    timeStyle: 'short',
  }).format(new Date(value));
}
