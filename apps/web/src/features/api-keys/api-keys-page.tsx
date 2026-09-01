import { useOrganization } from '#/features/organizations/hooks/use-organization';
import { useApiKeys } from './hooks/use-api-keys';

export function ApiKeysPage({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const query = useApiKeys(organizationQuery.organization?.id);
  if (organizationQuery.isLoading || query.isLoading)
    return <p className="p-8 text-sm text-white/55">Loading API keys…</p>;
  if (
    organizationQuery.isError ||
    query.isError ||
    !organizationQuery.organization
  )
    return (
      <p className="p-8 text-sm text-rose-200">We could not load API keys.</p>
    );
  const organization = organizationQuery.organization;
  return (
    <main className="mx-auto w-full max-w-6xl px-6 py-12 text-white sm:px-8 lg:py-16">
      <header className="border-b border-white/10 pb-8">
        <p className="mb-3 text-sm text-indigo-200">{organization.name}</p>
        <h1 className="text-3xl font-semibold tracking-tight">API keys</h1>
        <p className="mt-3 text-sm text-white/55">
          Credentials used by tools and automation to access this workspace.
        </p>
      </header>
      <section className="mt-8 overflow-hidden rounded-2xl border border-white/10 bg-white/[0.025]">
        {query.data?.length ? (
          query.data.map((key) => (
            <div
              key={key.id}
              className="flex flex-wrap items-center justify-between gap-4 border-b border-white/5 px-5 py-4 last:border-0"
            >
              <div>
                <p className="font-medium">{key.name}</p>
                <p className="mt-1 font-mono text-sm text-indigo-200">
                  {key.prefix}••••••••
                </p>
              </div>
              <div className="text-right text-xs text-white/45">
                <p>{key.revokedAt ? 'Revoked' : 'Active'}</p>
                <p className="mt-1">{key.scopes}</p>
              </div>
            </div>
          ))
        ) : (
          <p className="px-5 py-12 text-center text-sm text-white/50">
            No organization API keys found.
          </p>
        )}
      </section>
    </main>
  );
}
