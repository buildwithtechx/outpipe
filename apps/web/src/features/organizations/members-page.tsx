import { useMembers } from './hooks/use-members';
import { useOrganization } from './hooks/use-organization';

export function MembersPage({ orgSlug }: { orgSlug: string }) {
  const organizationQuery = useOrganization(orgSlug);
  const query = useMembers(organizationQuery.organization?.id);
  if (organizationQuery.isLoading || query.isLoading)
    return <p className="p-8 text-sm text-white/55">Loading members…</p>;
  if (
    organizationQuery.isError ||
    query.isError ||
    !organizationQuery.organization
  )
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load workspace members.
      </p>
    );
  const organization = organizationQuery.organization;
  return (
    <main className="mx-auto w-full max-w-6xl px-6 py-12 text-white sm:px-8 lg:py-16">
      <header className="border-b border-white/10 pb-8">
        <p className="mb-3 text-sm text-indigo-200">{organization.name}</p>
        <h1 className="text-3xl font-semibold tracking-tight">Members</h1>
        <p className="mt-3 text-sm text-white/55">
          People with access to this workspace.
        </p>
      </header>
      <section className="mt-8 overflow-hidden rounded-2xl border border-white/10 bg-white/[0.025]">
        {query.data?.length ? (
          query.data.map((member) => (
            <div
              key={member.id}
              className="flex items-center justify-between gap-4 border-b border-white/5 px-5 py-4 last:border-0"
            >
              <div>
                <p className="font-mono text-sm text-white/80">
                  {member.userId}
                </p>
                <p className="mt-1 text-xs text-white/40">
                  Joined {formatDate(member.createdAt)}
                </p>
              </div>
              <span className="rounded-full border border-white/10 px-3 py-1 text-xs text-white/55">
                {member.role}
              </span>
            </div>
          ))
        ) : (
          <p className="px-5 py-12 text-center text-sm text-white/50">
            No members found.
          </p>
        )}
      </section>
    </main>
  );
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(
    new Date(value),
  );
}
