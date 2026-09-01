import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { useDomains } from './hooks/use-domains';

export function DomainsPage({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const query = useDomains(organizationQuery.organization?.id);
  if (organizationQuery.isLoading || query.isLoading)
    return <p className="p-8 text-sm text-white/55">Loading domains…</p>;
  if (
    organizationQuery.isError ||
    query.isError ||
    !organizationQuery.organization
  )
    return (
      <p className="p-8 text-sm text-rose-200">We could not load domains.</p>
    );
  const domains = query.data ?? [];
  return (
    <main className="mx-auto w-full max-w-6xl px-6 py-12 text-white sm:px-8 lg:py-16">
      <header className="border-b border-white/10 pb-8">
        <p className="mb-3 text-sm text-indigo-200">
          {organizationQuery.organization.name}
        </p>
        <h1 className="text-3xl font-semibold tracking-tight">Domains</h1>
        <p className="mt-3 text-sm text-white/55">
          Stable hostnames for previews and production-like callbacks.
        </p>
      </header>
      <section className="mt-8 overflow-hidden rounded-2xl border border-white/10 bg-white/[0.025]">
        {domains.length ? (
          domains.map((domain) => (
            <div
              key={domain.id}
              className="flex flex-wrap items-center justify-between gap-4 border-b border-white/5 px-5 py-4 last:border-0"
            >
              <div>
                <p className="font-mono text-sm text-white/85">
                  {domain.hostname}
                </p>
                <p className="mt-1 text-xs text-white/40">
                  {domain.verificationMethod} verification
                </p>
              </div>
              <span className="rounded-full border border-white/10 px-3 py-1 text-xs text-white/60">
                {domain.status}
              </span>
            </div>
          ))
        ) : (
          <p className="px-5 py-12 text-center text-sm text-white/50">
            No custom domains connected yet.
          </p>
        )}
      </section>
    </main>
  );
}
