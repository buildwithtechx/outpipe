import { OrganizationDangerZone } from './components/organization-danger-zone';
import { useMembers } from './hooks/use-members';
import { useOrganization } from './hooks/use-organization';

export function OrganizationSettingsPage({ orgSlug }: { orgSlug: string }) {
  const query = useOrganization(orgSlug);
  const members = useMembers(query.organization?.id);

  if (query.isLoading) {
    return (
      <p className="p-8 text-sm text-white/55">Loading workspace settings…</p>
    );
  }

  if (query.isError || !query.organization) {
    return (
      <p className="p-8 text-sm text-rose-200">
        We could not load workspace settings.
      </p>
    );
  }

  const { organization } = query;

  return (
    <div className="w-full max-w-4xl space-y-6 pb-12 text-white">
      <header className="border-b border-white/10 pb-6">
        <p className="mb-2 text-xs font-semibold uppercase tracking-wider text-indigo-400">{organization.name}</p>
        <h1 className="text-3xl font-semibold tracking-[-0.04em]">
          Organization settings
        </h1>
        <p className="mt-1 text-sm text-white/55">
          Workspace identity and membership context.
        </p>
      </header>
      <section className="mt-6 grid gap-5 rounded-2xl border border-white/10 bg-white/[0.025] p-6">
        <div>
          <p className="text-xs uppercase tracking-wider text-white/40">
            Workspace name
          </p>
          <p className="mt-2 text-white/85">{organization.name}</p>
        </div>
        <div>
          <p className="text-xs uppercase tracking-wider text-white/40">
            Workspace slug
          </p>
          <p className="mt-2 font-mono text-white/85">{organization.slug}</p>
        </div>
        <div>
          <p className="text-xs uppercase tracking-wider text-white/40">
            Owner
          </p>
          <p className="mt-2 font-mono text-white/65">{organization.ownerId}</p>
        </div>
      </section>
      <OrganizationDangerZone
        organizationId={organization.id}
        members={members.data ?? []}
      />
    </div>
  );
}
